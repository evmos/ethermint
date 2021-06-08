package types

import (
	log "github.com/xlab/suplog"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/crypto"
)

var EmptyCodeHash = crypto.Keccak256(nil)

// EncodeTxResponse takes all of the necessary data from the EVM execution
// and returns the data as a byte slice encoded with protobuf.
func EncodeTxResponse(res *MsgEthereumTxResponse) ([]byte, error) {
	return proto.Marshal(res)
}

// DecodeTxResponse decodes an protobuf-encoded byte slice into TxResponse
func DecodeTxResponse(in []byte) (*MsgEthereumTxResponse, error) {
	var txMsgData sdk.TxMsgData
	if err := proto.Unmarshal(in, &txMsgData); err != nil {
		log.WithError(err).Errorln("failed to unmarshal TxMsgData")
		return nil, err
	}

	dataList := txMsgData.GetData()
	if len(dataList) == 0 {
		return &MsgEthereumTxResponse{}, nil
	}

	var res MsgEthereumTxResponse

	err := proto.Unmarshal(dataList[0].GetData(), &res)
	if err != nil {
		err = errors.Wrap(err, "proto.Unmarshal failed")
		return nil, err
	}

	return &res, nil
}

// EncodeTransactionLogs encodes TransactionLogs slice into a protobuf-encoded byte slice.
func EncodeTransactionLogs(res *TransactionLogs) ([]byte, error) {
	return proto.Marshal(res)
}

// DecodeTxResponse decodes an protobuf-encoded byte slice into TransactionLogs
func DecodeTransactionLogs(data []byte) (TransactionLogs, error) {
	var logs TransactionLogs
	err := proto.Unmarshal(data, &logs)
	if err != nil {
		return TransactionLogs{}, err
	}
	return logs, nil
}
