package types

import (
	// embed compiled smart contract
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// HexString is a byte array that serializes to hex
type HexString []byte

// MarshalJSON serializes ByteArray to hex
func (s HexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", string(s)))
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *HexString) UnmarshalJSON(data []byte) error {
	var x string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	str, err := hex.DecodeString(x)
	if err != nil {
		return err
	}
	*s = str
	return nil
}

// CompiledContract contains compiled bytecode and abi
type CompiledContract struct {
	ABI abi.ABI
	Bin HexString
}

type jsonCompiledContract struct {
	ABI string
	Bin HexString
}

// MarshalJSON serializes ByteArray to hex
func (s CompiledContract) MarshalJSON() ([]byte, error) {
	abi1, err := json.Marshal(s.ABI)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jsonCompiledContract{ABI: string(abi1), Bin: s.Bin})
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *CompiledContract) UnmarshalJSON(data []byte) error {
	var x jsonCompiledContract
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}

	s.Bin = x.Bin
	if err := json.Unmarshal([]byte(x.ABI), &s.ABI); err != nil {
		return fmt.Errorf("failed to unmarshal ABI: %w", err)
	}

	return nil
}

var (
	//go:embed ERC20Contract.json
	erc20JSON []byte

	// ERC20Contract is the compiled test erc20 contract
	ERC20Contract CompiledContract

	//go:embed SimpleStorageContract.json
	simpleStorageJSON []byte

	// SimpleStorageContract is the compiled test simple storage contract
	SimpleStorageContract CompiledContract

	//go:embed TestMessageCall.json
	testMessageCallJSON []byte

	// TestMessageCall is the compiled message call benchmark contract
	TestMessageCall CompiledContract
)

func init() {
	err := json.Unmarshal(erc20JSON, &ERC20Contract)
	if err != nil {
		panic(err)
	}

	if len(ERC20Contract.Bin) == 0 {
		panic("load contract failed")
	}

	err = json.Unmarshal(testMessageCallJSON, &TestMessageCall)
	if err != nil {
		panic(err)
	}

	if len(TestMessageCall.Bin) == 0 {
		panic("load contract failed")
	}

	err = json.Unmarshal(simpleStorageJSON, &SimpleStorageContract)
	if err != nil {
		panic(err)
	}

	if len(TestMessageCall.Bin) == 0 {
		panic("load contract failed")
	}
}
