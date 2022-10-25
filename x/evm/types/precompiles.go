package types

// RunStateful defines the function used by StatefulPrecompiledContracts
//type RunStateful func(evm evm.EVM, addr common.Address, input []byte, value *big.Int) (ret []byte, err error)

type BasePrecompile struct{}

// func NewBasePrecompile(precompile evm.StatefulPrecompiledContract) *BasePrecompile {
// 	return &BasePrecompile{
// 		runStateful: precompile.RunStateful,
// 	}
// }

func (bpc *BasePrecompile) Run(input []byte) ([]byte, error) {
	return []byte{}, nil
}
