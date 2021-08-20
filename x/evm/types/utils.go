package types

import (
	"encoding/binary"
	"fmt"

	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/crypto"
)

var EmptyCodeHash = crypto.Keccak256(nil)

// DecodeTxResponse decodes an protobuf-encoded byte slice into TxResponse
func DecodeTxResponse(in []byte) (*MsgEthereumTxResponse, error) {
	var txMsgData sdk.TxMsgData
	if err := proto.Unmarshal(in, &txMsgData); err != nil {
		return nil, err
	}

	data := txMsgData.GetData()
	if len(data) == 0 {
		return &MsgEthereumTxResponse{}, nil
	}

	var res MsgEthereumTxResponse

	err := proto.Unmarshal(data[0].GetData(), &res)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal tx response message data")
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

// UnwrapEthereumMsg extract MsgEthereumTx from wrapping sdk.Tx
func UnwrapEthereumMsg(tx *sdk.Tx) (*MsgEthereumTx, error) {
	if tx == nil {
		return nil, fmt.Errorf("invalid tx: nil")
	}

	if len((*tx).GetMsgs()) != 1 {
		return nil, fmt.Errorf("invalid tx type: %T", tx)
	}
	msg, ok := (*tx).GetMsgs()[0].(*MsgEthereumTx)
	if !ok {
		return nil, fmt.Errorf("invalid tx type: %T", tx)
	}

	return msg, nil
}

// BinSearch execute the binary search and hone in on an executable gas limit
func BinSearch(lo uint64, hi uint64, executable func(uint64) (bool, *MsgEthereumTxResponse, error)) (uint64, error) {
	for lo+1 < hi {
		mid := (hi + lo) / 2
		failed, _, err := executable(mid)

		// If the error is not nil(consensus error), it means the provided message
		// call or transaction will never be accepted no matter how much gas it is
		// assigned. Return the error directly, don't struggle any more.
		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi, nil
}

// Uint16ToBytes transfer value to variable size []byte
func Uint16ToBytes(v uint16) []byte {
	bz := []byte{0x00, 0x00}
	binary.BigEndian.PutUint16(bz, v)

	if v < 256 {
		return bz[1:]
	}

	return bz
}

// BytesToUint16 transfer []byte to uint16.
// Note: If the bytes length greater than two, the value lower than 256 will be returned.
func BytesToUint16(bz []byte) uint16 {
	l := len(bz)
	if l == 0 {
		return 0
	} else if l == 1 {
		return uint16(bz[0])
	}

	return binary.BigEndian.Uint16(bz[l-2:])
}
