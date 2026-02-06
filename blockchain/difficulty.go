// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"math/big"
	"time"

	"github.com/ltcsuite/ltcd/chaincfg/chainhash"
)

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)
)

// HashToBig converts a chainhash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *chainhash.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *hash
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
// - the most significant 8 bits represent the unsigned base 256 exponent
// - bit 23 (the 24th bit) represents the sign bit
// - the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
//
//	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used in litecoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with litecoind.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

// CalcWork calculates a work value from difficulty bits. Litecoin increases
// the difficulty for generating a block by decreasing the value which the
// generated hash must be less than.  This difficulty target is stored in each
// block header using a compact representation as described in the documentation
// for CompactToBig.  The main chain is selected by choosing the chain that has
// the most proof of work (highest difficulty).  Since a lower target difficulty
// value equates to higher actual difficulty, the work value which will be
// accumulated must be the inverse of the difficulty.  Also, in order to avoid
// potential division by zero and really small floating point numbers, the
// result adds 1 to the denominator and multiplies the numerator by 2^256.
func CalcWork(bits uint32) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(oneLsh256, denominator)
}

// calcEasiestDifficulty calculates the easiest possible difficulty that a block
// can have given starting difficulty bits and a duration.  It is mainly used to
// verify that claimed proof of work by a block is sane as compared to a
// known good checkpoint.
func (b *BlockChain) calcEasiestDifficulty(bits uint32, duration time.Duration) uint32 {
	// Convert types used in the calculations below.
	durationVal := int64(duration / time.Second)
	adjustmentFactor := big.NewInt(b.chainParams.RetargetAdjustmentFactor)

	// The test network rules allow minimum difficulty blocks after more
	// than twice the desired amount of time needed to generate a block has
	// elapsed.
	if b.chainParams.ReduceMinDifficulty {
		reductionTime := int64(b.chainParams.MinDiffReductionTime /
			time.Second)
		if durationVal > reductionTime {
			return b.chainParams.PowLimitBits
		}
	}

	// Since easier difficulty equates to higher numbers, the easiest
	// difficulty for a given duration is the largest value possible given
	// the number of retargets for the duration and starting difficulty
	// multiplied by the max adjustment factor.
	newTarget := CompactToBig(bits)
	for durationVal > 0 && newTarget.Cmp(b.chainParams.PowLimit) < 0 {
		newTarget.Mul(newTarget, adjustmentFactor)
		durationVal -= b.maxRetargetTimespan
	}

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(b.chainParams.PowLimit) > 0 {
		newTarget.Set(b.chainParams.PowLimit)
	}

	return BigToCompact(newTarget)
}

// findPrevTestNetDifficulty returns the difficulty of the previous block which
// did not have the special testnet minimum difficulty rule applied.
func findPrevTestNetDifficulty(startNode HeaderCtx, c ChainCtx) uint32 {
	// Search backwards through the chain for the last block without
	// the special rule applied.
	iterNode := startNode
	for iterNode != nil && iterNode.Height()%c.BlocksPerRetarget() != 0 &&
		iterNode.Bits() == c.ChainParams().PowLimitBits {

		iterNode = iterNode.Parent()
	}

	// Return the found difficulty or the minimum difficulty if no
	// appropriate block was found.
	lastBits := c.ChainParams().PowLimitBits
	if iterNode != nil {
		lastBits = iterNode.Bits()
	}
	return lastBits
}

// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous HeaderCtx based on the difficulty retarget rules.
// This function differs from the exported CalcNextRequiredDifficulty in that
// the exported version uses the current best chain as the previous HeaderCtx
// while this function accepts any block node. This function accepts a ChainCtx
// parameter that gives the necessary difficulty context variables.
func calcNextRequiredDifficulty(lastNode HeaderCtx, newBlockTime time.Time,
	c ChainCtx) (uint32, error) {

	// Emulate the same behavior as Litecoin Core that for regtest there is
	// no difficulty retargeting.
	if c.ChainParams().PoWNoRetargeting {
		return c.ChainParams().PowLimitBits, nil
	}

	// Genesis block.
	if lastNode == nil {
		return c.ChainParams().PowLimitBits, nil
	}

	// Dispatch to the appropriate difficulty algorithm based on block
	// height. Doriancoin transitioned from the original BTC-style
	// algorithm to LWMA, then LWMAv2, then ASERT.
	nHeight := lastNode.Height() + 1
	if c.ChainParams().ASERTHeight > 0 && nHeight > c.ChainParams().ASERTHeight {
		return calcNextRequiredDifficultyASERT(lastNode, c)
	}
	if c.ChainParams().LWMAFixHeight > 0 && nHeight >= c.ChainParams().LWMAFixHeight {
		return calcNextRequiredDifficultyLWMAv2(lastNode, c)
	}
	if c.ChainParams().LWMAHeight > 0 && nHeight >= c.ChainParams().LWMAHeight {
		return calcNextRequiredDifficultyLWMA(lastNode, c)
	}

	// Original BTC-style difficulty retarget algorithm.

	// Return the previous block's difficulty requirements if this block
	// is not at a difficulty retarget interval.
	if (lastNode.Height()+1)%c.BlocksPerRetarget() != 0 {
		// For networks that support it, allow special reduction of the
		// required difficulty once too much time has elapsed without
		// mining a block.
		if c.ChainParams().ReduceMinDifficulty {
			// Return minimum difficulty when more than the desired
			// amount of time has elapsed without mining a block.
			reductionTime := int64(c.ChainParams().MinDiffReductionTime /
				time.Second)
			allowMinTime := lastNode.Timestamp() + reductionTime
			if newBlockTime.Unix() > allowMinTime {
				return c.ChainParams().PowLimitBits, nil
			}

			// The block was mined within the desired timeframe, so
			// return the difficulty for the last block which did
			// not have the special minimum difficulty rule applied.
			return findPrevTestNetDifficulty(lastNode, c), nil
		}

		// For the main network (or any unrecognized networks), simply
		// return the previous block's difficulty requirements.
		return lastNode.Bits(), nil
	}

	// Litecoin fixes an issue where a 51% can change the difficult at
	// will. We only go back the full period unless it's the first retarget
	// after genesis.
	blocksPerRetarget := c.BlocksPerRetarget() - 1
	if (lastNode.Height() + 1) != c.BlocksPerRetarget() {
		blocksPerRetarget = c.BlocksPerRetarget()
	}

	// Get the block node at the previous retarget (targetTimespan days
	// worth of blocks).
	firstNode := lastNode.RelativeAncestorCtx(blocksPerRetarget)
	if firstNode == nil {
		return 0, AssertError("unable to obtain previous retarget block")
	}

	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	actualTimespan := lastNode.Timestamp() - firstNode.Timestamp()
	adjustedTimespan := actualTimespan
	if actualTimespan < c.MinRetargetTimespan() {
		adjustedTimespan = c.MinRetargetTimespan()
	} else if actualTimespan > c.MaxRetargetTimespan() {
		adjustedTimespan = c.MaxRetargetTimespan()
	}

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Litecoind also uses integer division to calculate this
	// result.
	oldTarget := CompactToBig(lastNode.Bits())
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	targetTimeSpan := int64(c.ChainParams().TargetTimespan / time.Second)
	newTarget.Div(newTarget, big.NewInt(targetTimeSpan))

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(c.ChainParams().PowLimit) > 0 {
		newTarget.Set(c.ChainParams().PowLimit)
	}

	// Log new target difficulty and return it.  The new target logging is
	// intentionally converting the bits back to a number instead of using
	// newTarget since conversion to the compact representation loses
	// precision.
	newTargetBits := BigToCompact(newTarget)
	log.Debugf("Difficulty retarget at block height %d", lastNode.Height()+1)
	log.Debugf("Old target %08x (%064x)", lastNode.Bits(), oldTarget)
	log.Debugf("New target %08x (%064x)", newTargetBits, CompactToBig(newTargetBits))
	log.Debugf("Actual timespan %v, adjusted timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second,
		time.Duration(adjustedTimespan)*time.Second,
		c.ChainParams().TargetTimespan)

	return newTargetBits, nil
}

// calcNextRequiredDifficultyLWMA calculates the required difficulty using the
// LWMA (Linear Weighted Moving Average) algorithm. This weights recent blocks
// more heavily, providing faster response to hashrate changes than the
// original BTC-style algorithm.
//
// Reference: https://github.com/zawy12/difficulty-algorithms/issues/3
func calcNextRequiredDifficultyLWMA(lastNode HeaderCtx, c ChainCtx) (uint32, error) {
	params := c.ChainParams()
	T := int64(params.TargetTimePerBlock / time.Second)
	N := params.LWMAWindow

	height := int64(lastNode.Height()) + 1
	blocks := height - int64(params.LWMAHeight)
	if blocks > N {
		blocks = N
	}

	// Need at least 3 blocks for a meaningful LWMA calculation.
	if blocks < 3 {
		return lastNode.Bits(), nil
	}

	prevTarget := CompactToBig(lastNode.Bits())

	var sumWeightedSolvetimes int64
	var sumWeights int64

	block := lastNode
	for i := blocks; i >= 1; i-- {
		prev := block.Parent()
		if prev == nil {
			break
		}

		solvetime := block.Timestamp() - prev.Timestamp()
		if solvetime < 1 {
			solvetime = 1
		}
		if solvetime > 6*T {
			solvetime = 6 * T
		}

		sumWeightedSolvetimes += solvetime * i
		sumWeights += i

		block = prev
	}

	expectedWeightedSolvetimes := sumWeights * T

	// Symmetric caps: limit adjustment to 10x per calculation.
	minWS := expectedWeightedSolvetimes / 10
	maxWS := expectedWeightedSolvetimes * 10

	if sumWeightedSolvetimes < minWS {
		sumWeightedSolvetimes = minWS
	}
	if sumWeightedSolvetimes > maxWS {
		sumWeightedSolvetimes = maxWS
	}

	// nextTarget = prevTarget * sumWeightedSolvetimes / expectedWeightedSolvetimes
	nextTarget := new(big.Int).Mul(prevTarget, big.NewInt(sumWeightedSolvetimes))
	nextTarget.Div(nextTarget, big.NewInt(expectedWeightedSolvetimes))

	if nextTarget.Cmp(params.PowLimit) > 0 {
		nextTarget.Set(params.PowLimit)
	}

	return BigToCompact(nextTarget), nil
}

// calcNextRequiredDifficultyLWMAv2 calculates the required difficulty using
// the stabilized LWMAv2 algorithm. This fixes a feedback loop instability
// in LWMA v1 by using the target from the start of the window as a reference
// instead of the previous block's target, preventing compounding oscillations.
func calcNextRequiredDifficultyLWMAv2(lastNode HeaderCtx, c ChainCtx) (uint32, error) {
	params := c.ChainParams()
	T := int64(params.TargetTimePerBlock / time.Second)
	N := params.LWMAWindow

	// Use distance from original LWMA activation, not LWMAv2 activation,
	// so the window is already full by the time v2 activates.
	height := int64(lastNode.Height()) + 1
	blocks := height - int64(params.LWMAHeight)
	if blocks > N {
		blocks = N
	}

	if blocks < 3 {
		return lastNode.Bits(), nil
	}

	// Find window start block and use its target as reference.
	// This breaks the feedback loop that caused oscillations in v1.
	windowStart := lastNode
	for i := int64(0); i < blocks; i++ {
		prev := windowStart.Parent()
		if prev == nil {
			break
		}
		windowStart = prev
	}
	referenceTarget := CompactToBig(windowStart.Bits())

	var sumWeightedSolvetimes int64
	var sumWeights int64

	block := lastNode
	for i := blocks; i >= 1; i-- {
		prev := block.Parent()
		if prev == nil {
			break
		}

		solvetime := block.Timestamp() - prev.Timestamp()
		if solvetime < 1 {
			solvetime = 1
		}
		if solvetime > 6*T {
			solvetime = 6 * T
		}

		sumWeightedSolvetimes += solvetime * i
		sumWeights += i

		block = prev
	}

	expectedWeightedSolvetimes := sumWeights * T

	// Tighter caps (3x instead of 10x) since window-start reference
	// is more stable.
	minWS := expectedWeightedSolvetimes / 3
	maxWS := expectedWeightedSolvetimes * 3

	if sumWeightedSolvetimes < minWS {
		sumWeightedSolvetimes = minWS
	}
	if sumWeightedSolvetimes > maxWS {
		sumWeightedSolvetimes = maxWS
	}

	nextTarget := new(big.Int).Mul(referenceTarget, big.NewInt(sumWeightedSolvetimes))
	nextTarget.Div(nextTarget, big.NewInt(expectedWeightedSolvetimes))

	if nextTarget.Cmp(params.PowLimit) > 0 {
		nextTarget.Set(params.PowLimit)
	}

	return BigToCompact(nextTarget), nil
}

// calcNextRequiredDifficultyASERT calculates the required difficulty using the
// ASERT (Absolutely Scheduled Exponentially Rising Targets) algorithm.
// Based on BCH's aserti3-2d by Mark Lundeberg. This computes difficulty from
// total time deviation relative to an ideal block schedule using an exponential
// adjustment. It is mathematically proven to never oscillate and has no window
// lag.
//
// Formula: target = anchor_target * 2^((time_delta - T * height_delta) / halflife)
func calcNextRequiredDifficultyASERT(lastNode HeaderCtx, c ChainCtx) (uint32, error) {
	params := c.ChainParams()

	// Find the anchor block at ASERTHeight.
	anchor := lastNode
	for anchor.Height() > params.ASERTHeight {
		anchor = anchor.Parent()
	}

	anchorParent := anchor.Parent()
	if anchorParent == nil {
		return 0, AssertError("ASERT anchor block has no parent")
	}

	anchorTarget := CompactToBig(params.ASERTAnchorBits)

	anchorParentTime := anchorParent.Timestamp()
	currentParentTime := lastNode.Timestamp()
	timeDelta := currentParentTime - anchorParentTime

	nHeight := int64(lastNode.Height()) + 1
	heightDelta := nHeight - int64(params.ASERTHeight)

	T := int64(params.TargetTimePerBlock / time.Second)
	halfLife := params.ASERTHalfLife

	// Compute exponent in fixed-point with 16 fractional bits:
	// exponent = (timeDelta - T * heightDelta) * 65536 / halfLife
	exponent := ((timeDelta - T*heightDelta) * 65536) / halfLife

	// Decompose into integer shifts and fractional part.
	var shifts int32
	var frac uint16

	if exponent >= 0 {
		shifts = int32(exponent >> 16)
		frac = uint16(exponent & 0xFFFF)
	} else {
		// For negative exponents, ensure frac is in [0, 65536).
		absExponent := -exponent
		shifts = -int32(absExponent >> 16)
		remainder := uint32(absExponent & 0xFFFF)
		if remainder != 0 {
			shifts--
			frac = uint16(65536 - remainder)
		} else {
			frac = 0
		}
	}

	// Compute 2^(frac/65536) * 65536 using cubic polynomial approximation.
	// Coefficients from BCH aserti3-2d, designed to stay within uint64 bounds.
	factor := uint32(65536)
	if frac > 0 {
		f := uint64(frac)
		factor = 65536 + uint32(
			(195766423245049*f+
				971821376*f*f+
				5127*f*f*f+
				(1<<47))>>48)
	}

	// Apply fractional part: nextTarget = anchorTarget * factor / 65536
	nextTarget := new(big.Int).Mul(anchorTarget, big.NewInt(int64(factor)))
	nextTarget.Rsh(nextTarget, 16)

	// Apply integer shifts (left = easier, right = harder).
	if shifts > 0 {
		if shifts >= 256 {
			return BigToCompact(params.PowLimit), nil
		}
		nextTarget.Lsh(nextTarget, uint(shifts))
	} else if shifts < 0 {
		absShifts := -shifts
		if absShifts >= 256 {
			return BigToCompact(big.NewInt(1)), nil
		}
		nextTarget.Rsh(nextTarget, uint(absShifts))
	}

	// Ensure target is at least 1 (maximum possible difficulty).
	if nextTarget.Sign() == 0 {
		nextTarget.SetInt64(1)
	}

	if nextTarget.Cmp(params.PowLimit) > 0 {
		nextTarget.Set(params.PowLimit)
	}

	return BigToCompact(nextTarget), nil
}

// CalcNextRequiredDifficulty calculates the required difficulty for the block
// after the end of the current best chain based on the difficulty retarget
// rules.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcNextRequiredDifficulty(timestamp time.Time) (uint32, error) {
	b.chainLock.Lock()
	difficulty, err := calcNextRequiredDifficulty(b.bestChain.Tip(), timestamp, b)
	b.chainLock.Unlock()
	return difficulty, err
}
