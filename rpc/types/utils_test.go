package types

// TODO TestFormatBlock
// expBlock = map[string]interface{}{
// 	"number":           hexutil.Uint64(header.Height),
// 	"hash":             hexutil.Bytes(header.Hash()),
// 	"parentHash":       common.BytesToHash(header.LastBlockID.Hash.Bytes()),
// 	"nonce":            ethtypes.BlockNonce{},   // PoW specific
// 	"sha3Uncles":       ethtypes.EmptyUncleHash, // No uncles in Tendermint
// 	"logsBloom":        bloom,
// 	"stateRoot":        hexutil.Bytes(header.AppHash),
// 	"miner":            common.BytesToAddress(tc.validator.Bytes()),
// 	"mixHash":          common.Hash{},
// 	"difficulty":       (*hexutil.Big)(big.NewInt(0)),
// 	"extraData":        "0x",
// 	"size":             hexutil.Uint64(tc.resBlock.Block.Size()),
// 	"gasLimit":         hexutil.Uint64(gasLimit), // Static gas limit
// 	"gasUsed":          (*hexutil.Big)(gasUsed),
// 	"timestamp":        hexutil.Uint64(header.Time.Unix()),
// 	"transactionsRoot": transactionsRoot,
// 	"receiptsRoot":     ethtypes.EmptyRootHash,

// 	"uncles":          []common.Hash{},
// 	"transactions":    ethRPCTxs,
// 	"totalDifficulty": (*hexutil.Big)(big.NewInt(0)),
// }

// if tc.baseFee != nil {
// 	expBlock["baseFeePerGas"] = (*hexutil.Big)(tc.baseFee)
// }
