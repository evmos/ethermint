# x/evm

The `x/evm` module is responsible for executing Ethereum Virtual Machine (EVM) state transitions.

## Usage

1. Import the module and the dependency packages.

   ```go
   import (
      "github.com/cosmos/cosmos-sdk/x/auth"
      "github.com/cosmos/cosmos-sdk/x/bank"
       
      "github.com/tharsis/ethermint/app/ante"
      ethermint "github.com/tharsis/ethermint/types"
      "github.com/tharsis/ethermint/x/evm"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        evm.AppModuleBasic{},
      )
    )
    ```

3. Create the module's parameter subspace in your application constructor.

   ```go
   func NewApp(...) *App {
     // ...
     app.subspaces[evm.ModuleName] = app.ParamsKeeper.Subspace(evm.DefaultParamspace)
   }
   ```

4. Define the Ethermint `ProtoAccount` for the `AccountKeeper`

   ```go
   func NewApp(...) *App {
      // ...
        app.AccountKeeper = auth.NewAccountKeeper(
            cdc, keys[auth.StoreKey], app.subspaces[auth.ModuleName], ethermint.ProtoAccount,
        )
   }
   ```

5. Create the keeper.

   ```go
   func NewApp(...) *App {
      // ...
        app.EvmKeeper = evm.NewKeeper(
         app.cdc, keys[evm.StoreKey], app.subspaces[evm.ModuleName], app.AccountKeeper,
        )
   }
   ```

6. Add the `x/evm` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       evm.NewAppModule(app.EvmKeeper, app.AccountKeeper),
       // ...
     )
   }
   ```

7. Set the `x/evm` module `BeginBlock` and `EndBlock` ordering:

    ```go
      app.mm.SetOrderBeginBlockers(
        evm.ModuleName, ...
      )
      app.mm.SetOrderEndBlockers(
        evm.ModuleName, ...
      )
    ```

8. Set the `x/evm` module genesis order. The module must go after the `auth` and `bank` modules.

    ```go
    func NewApp(...) *App {
      // ...
      app.mm.SetOrderInitGenesis(auth.ModuleName, bank.ModuleName, ... , evm.ModuleName, ...)
    }
    ```

9. Set the Ethermint `AnteHandler` to support EVM transactions. Note,
the default `AnteHandler` provided by the `x/evm` module depends on the `x/auth` and `x/supply`
modules.

   ```go
   func NewApp(...) *App {
     // ...
     app.SetAnteHandler(ante.NewAnteHandler(
      app.AccountKeeper, app.EvmKeeper, app.SupplyKeeper
     ))
   }
   ```

## Genesis

The `x/evm` module defines its genesis state as follows:

```go
type GenesisState struct {
  Accounts    []GenesisAccount  `json:"accounts"`
  TxsLogs     []TransactionLogs `json:"txs_logs"`
  ChainConfig ChainConfig       `json:"chain_config"`
  Params      Params            `json:"params"`
}
```

Which relies on the following types:

```go
type GenesisAccount struct {
  Address string        `json:"address"`
  Balance sdk.Int       `json:"balance"`
  Code    hexutil.Bytes `json:"code,omitempty"`
  Storage Storage       `json:"storage,omitempty"`
}


type TransactionLogs struct {
  Hash ethcmn.Hash     `json:"hash"`
  Logs []*ethtypes.Log `json:"logs"`
}

type ChainConfig struct {
  HomesteadBlock sdk.Int `json:"homestead_block" yaml:"homestead_block"` // Homestead switch block (< 0 no fork, 0 = already homestead)

  DAOForkBlock   sdk.Int `json:"dao_fork_block" yaml:"dao_fork_block"`     // TheDAO hard-fork switch block (< 0 no fork)
  DAOForkSupport bool    `json:"dao_fork_support" yaml:"dao_fork_support"` // Whether the nodes supports or opposes the DAO hard-fork

  // EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
  EIP150Block sdk.Int `json:"eip150_block" yaml:"eip150_block"` // EIP150 HF block (< 0 no fork)
  EIP150Hash  string  `json:"eip150_hash" yaml:"eip150_hash"`   // EIP150 HF hash (needed for header only clients as only gas pricing changed)

  EIP155Block sdk.Int `json:"eip155_block" yaml:"eip155_block"` // EIP155 HF block
  EIP158Block sdk.Int `json:"eip158_block" yaml:"eip158_block"` // EIP158 HF block

  ByzantiumBlock      sdk.Int `json:"byzantium_block" yaml:"byzantium_block"`           // Byzantium switch block (< 0 no fork, 0 = already on byzantium)
  ConstantinopleBlock sdk.Int `json:"constantinople_block" yaml:"constantinople_block"` // Constantinople switch block (< 0 no fork, 0 = already activated)
  PetersburgBlock     sdk.Int `json:"petersburg_block" yaml:"petersburg_block"`         // Petersburg switch block (< 0 same as Constantinople)
  IstanbulBlock       sdk.Int `json:"istanbul_block" yaml:"istanbul_block"`             // Istanbul switch block (< 0 no fork, 0 = already on istanbul)
  MuirGlacierBlock    sdk.Int `json:"muir_glacier_block" yaml:"muir_glacier_block"`     // Eip-2384 (bomb delay) switch block (< 0 no fork, 0 = already activated)

  YoloV2Block sdk.Int `json:"yoloV2_block" yaml:"yoloV2_block"` // YOLO v1: https://github.com/ethereum/EIPs/pull/2657 (Ephemeral testnet)
  EWASMBlock  sdk.Int `json:"ewasm_block" yaml:"ewasm_block"`   // EWASM switch block (< 0 no fork, 0 = already activated)
}

type Params struct {
  // EVMDenom defines the token denomination used for state transitions on the
  // EVM module.
  EvmDenom string `json:"evm_denom" yaml:"evm_denom"`
  // EnableCreate toggles state transitions that use the vm.Create function
  EnableCreate bool `json:"enable_create" yaml:"enable_create"`
  // EnableCall toggles state transitions that use the vm.Call function
  EnableCall bool `json:"enable_call" yaml:"enable_call"`
  // ExtraEIPs defines the additional EIPs for the vm.Config
  ExtraEIPs []int `json:"extra_eips" yaml:"extra_eips"`
}
```

## Client

### JSON-RPC

See the Ethermint [JSON-RPC docs](https://docs.ethermint.zone/basics/json_rpc.html) for reference.

## Documentation and Specification

* Ethermint documentation: [https://docs.ethermint.zone](https://docs.ethermint.zone)
