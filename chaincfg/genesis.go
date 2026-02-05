// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"time"

	"github.com/ltcsuite/ltcd/chaincfg/chainhash"
	"github.com/ltcsuite/ltcd/wire"
)

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var genesisCoinbaseTx = wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash:  chainhash.Hash{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, 0x52, 0x4c, 0x41, 0x20, 0x54, 0x69, 0x6d, 0x65, 0x73, // |.......RLA Times|
				0x20, 0x30, 0x38, 0x2f, 0x4d, 0x61, 0x72, 0x2f, 0x32, 0x30, 0x31, 0x34, 0x20, 0x46, 0x6f, 0x72, // | 08/Mar/2014 For|
				0x20, 0x44, 0x6f, 0x72, 0x69, 0x61, 0x6e, 0x20, 0x4e, 0x61, 0x6b, 0x61, 0x6d, 0x6f, 0x74, 0x6f, // | Dorian Nakamoto|
				0x2c, 0x20, 0x62, 0x69, 0x74, 0x63, 0x6f, 0x69, 0x6e, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, // |, bitcoin articl|
				0x65, 0x20, 0x62, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x20, 0x64, 0x65, 0x6e, 0x69, 0x61, 0x6c, 0x73, // |e brings denials|
				0x2c, 0x20, 0x69, 0x6e, 0x74, 0x72, 0x69, 0x67, 0x75, 0x65, // |, intrigue|

			},
			Sequence: 0xffffffff,
		},
	},
	TxOut: []*wire.TxOut{
		{
			Value: 0x12a05f200,
			PkScript: []byte{
				0x41, 0x4, 0x1, 0x84, 0x71, 0xf, 0xa6, 0x89,
				0xad, 0x50, 0x23, 0x69, 0xc, 0x80, 0xf3, 0xa4,
				0x9c, 0x8f, 0x13, 0xf8, 0xd4, 0x5b, 0x8c, 0x85,
				0x7f, 0xbc, 0xbc, 0x8b, 0xc4, 0xa8, 0xe4, 0xd3,
				0xeb, 0x4b, 0x10, 0xf4, 0xd4, 0x60, 0x4f, 0xa0,
				0x8d, 0xce, 0x60, 0x1a, 0xaf, 0xf, 0x47, 0x2,
				0x16, 0xfe, 0x1b, 0x51, 0x85, 0xb, 0x4a, 0xcf,
				0x21, 0xb1, 0x79, 0xc4, 0x50, 0x70, 0xac, 0x7b,
				0x3, 0xa9, 0xac,
			},
		},
	},
	LockTime: 0,
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0xb9, 0xd1, 0xe7, 0xd1, 0xc7, 0x23, 0x06, 0xcb,
	0xb5, 0x71, 0x02, 0x41, 0x1b, 0x09, 0xc6, 0xeb,
	0xe2, 0xfe, 0xc5, 0x69, 0x7d, 0x08, 0x56, 0x74,
	0x0b, 0xd2, 0x7b, 0x27, 0x5e, 0xa2, 0x1d, 0xd2,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x5d, 0x6d, 0x74, 0x19, 0xeb, 0xe7, 0x91, 0x72,
	0x14, 0x6a, 0x2f, 0xee, 0xaf, 0x15, 0x1b, 0xfc,
	0x01, 0x60, 0xae, 0x1e, 0x5b, 0x7e, 0xcb, 0xe3,
	0xfa, 0x46, 0x6e, 0x28, 0x0a, 0x7d, 0x7b, 0xa2,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},  // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: genesisMerkleRoot, // a27b7d0a286e46fae3cb7e5b1eae6001fc1b15afee2f6a147291e7eb19746d5d
		Timestamp:  time.Unix(1394325760, 0),
		Bits:       0x1e0ffff0,
		Nonce:      385834689,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x51, 0xbf, 0x2f, 0x59, 0xfd, 0xe1, 0x8e, 0x5b,
	0x96, 0x5c, 0x32, 0x52, 0x18, 0x28, 0x45, 0x63,
	0x22, 0x72, 0x0e, 0x5f, 0xbc, 0xcd, 0x75, 0x7b,
	0xdd, 0x9f, 0xb5, 0x4e, 0x46, 0x69, 0x77, 0x70,
})

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: regTestGenesisMerkleRoot, // a27b7d0a286e46fae3cb7e5b1eae6001fc1b15afee2f6a147291e7eb19746d5d
		Timestamp:  time.Unix(1394325759, 0),
		Bits:       0x1e0ffff0,
		Nonce:      149343,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet4GenesisHash is the hash of the first block in the block chain for the
// test network (version 4).
var testNet4GenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x51, 0xbf, 0x2f, 0x59, 0xfd, 0xe1, 0x8e, 0x5b,
	0x96, 0x5c, 0x32, 0x52, 0x18, 0x28, 0x45, 0x63,
	0x22, 0x72, 0x0e, 0x5f, 0xbc, 0xcd, 0x75, 0x7b,
	0xdd, 0x9f, 0xb5, 0x4e, 0x46, 0x69, 0x77, 0x70,
})

// testNet4GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 4).  It is the same as the merkle root
// for the main network.
var testNet4GenesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x5d, 0x6d, 0x74, 0x19, 0xeb, 0xe7, 0x91, 0x72,
	0x14, 0x6a, 0x2f, 0xee, 0xaf, 0x15, 0x1b, 0xfc,
	0x01, 0x60, 0xae, 0x1e, 0x5b, 0x7e, 0xcb, 0xe3,
	0xfa, 0x46, 0x6e, 0x28, 0x0a, 0x7d, 0x7b, 0xa2,
})

// testNet4GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 4).
var testNet4GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},          // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: testNet4GenesisMerkleRoot, // a27b7d0a286e46fae3cb7e5b1eae6001fc1b15afee2f6a147291e7eb19746d5d
		Timestamp:  time.Unix(1394325759, 0),
		Bits:       0x1e0ffff0,
		Nonce:      149343,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x70, 0x8f, 0x2f, 0xd3, 0x31, 0x35, 0x4e, 0xc2,
	0x42, 0xb2, 0x19, 0xfb, 0x19, 0xb5, 0x60, 0xe5,
	0x98, 0x88, 0xca, 0x96, 0x79, 0x39, 0x41, 0xfd,
	0xae, 0x68, 0x06, 0x14, 0x98, 0x87, 0x84, 0xe4,
})

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: simNetGenesisMerkleRoot,  // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:  time.Unix(1401292357, 0), // 2014-05-28 15:52:37 +0000 UTC
		Bits:       0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:      2,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// sigNetGenesisHash is the hash of the first block in the block chain for the
// signet test network. Doriancoin doesn't have signet, so we use the regtest
// genesis hash to ensure signet doesn't accidentally match mainnet.
var sigNetGenesisHash = chainhash.Hash{
	0x51, 0xbf, 0x2f, 0x59, 0xfd, 0xe1, 0x8e, 0x5b,
	0x96, 0x5c, 0x32, 0x52, 0x18, 0x28, 0x45, 0x63,
	0x22, 0x72, 0x0e, 0x5f, 0xbc, 0xcd, 0x75, 0x7b,
	0xdd, 0x9f, 0xb5, 0x4e, 0x46, 0x69, 0x77, 0x70,
}

// sigNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the signet test network. It is the same as the merkle root for
// the main network.
var sigNetGenesisMerkleRoot = genesisMerkleRoot

// sigNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the signet test network.
var sigNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},        // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: sigNetGenesisMerkleRoot, // a27b7d0a286e46fae3cb7e5b1eae6001fc1b15afee2f6a147291e7eb19746d5d
		Timestamp:  time.Unix(1394325760, 0),
		Bits:       0x1e0ffff0,
		Nonce:      385834689,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
