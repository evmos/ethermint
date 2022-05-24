<!--
order: 1
-->

# Concepts

## EVM

The Ethereum Virtual Machine (EVM) is a computation engine which can be thought of as one single entity maintained by thousands of connected computers running an Ethereum client. It is considered to be the part of the Ethereum protocol that handles the deployment and execution of [smart contracts](https://ethereum.org/en/developers/docs/smart-contracts/). To make a clear distinction:

* The Ethereum protocol describes a blockchain, in which all Ethereum accounts and smart contracts live. It has only one canonical state (a data structure, which keeps all accounts) at any given block in the chain.
* The EVM, however, is the [state machine](https://en.wikipedia.org/wiki/Finite-state_machine) that defines the rules for computing a new valid state from block to block. It is an isolated runtime, which means that code running inside the EVM has no access to network, filesystem, or other processes.

The `x/evm` module implements the EVM as a Cosmos SDK module. It allows users to interact with the EVM by submitting Ethereum transactions and executing their containing messages on the given state to evoke a state transition.

### EVM

EVM is a virtual machine ([VM](https://en.wikipedia.org/wiki/Virtual_machine)) that is responisble for making changes to the state. Every node has to get the exact same result give an identical starting state and transaction (deterministic).

regardless of the environment (hardware and OS)

Ethereum node also runs the EVM

EVM is just some code running on Ethereum nodes. In the [go-etheruem source code](https://github.com/ethereum/go-ethereum/blob/master/core/vm/instructions.go) you can see the EVM opcodes.

### Types of Accounts

There are two types of accounts in the EVM:

* **Externally Owned Account (EOA)**: Has nonce and balance
* **Contract**: Has nonce, balance, storage root, (immutable) code hash

### State

The Ethereum state is a data structure, implemented as a Merkle Patricia Trie, that is a collection of all accounts on the chain.

Each contract account's storage root is implemented using the same data structure, another Merkle Patricia Trie.

Any change in the data structure results in a different State Root.

Block hearder (parent hash, block number, time stamp, nonce, receipts,...)



### Architecture

The EVM is a simple stack-based architecture:

- Virtual ROM: Contract code is pulled into this read only memory during processing Txs
- Machine state (volatile): changes as the EMV runs and is wiped clean after processing each tx
  - Program counter (PC)
  - Gas: keeps track of how much gas is used
  - Stack: compute state changes
  - Memory:
- Access to account storage (persistent)

### State Transition with Smart Contracts

A state transition on the EVM can be initiated through a transaction that either deploys or calls a smart contract.

Smart contracts are just like regular accounts on the blockchain, which additionally store executable code in an Ethereum-specific binary format (EVM bytecode). They are typically written in an Ethereum high level language, compiled into byte code using an EVM compiler, and finally deployed on the blockchain, by submitting a transaction using an Ethereum client. Whenever another account makes a message call to that deployed contract, it executes its EVM bytecode to  perform calculations and further transactions.

For each transaction:
* code of contract being called is loaded,
* program counter set to zero,
* contract starage is loaded
* memory is set to all zeros
* all bkock and environment variables are set

If gas limit is his (out of gas execption): no changes to Ethereum state are applied, expect sender's nonce incrementin and their balance going down to pay for wasting the EVM's time.


### State transition (in geth)

### Executing EVM bytecode

Solidity is compiled down to EVM bytecode, which consists of basic operations (add, multiply, store, etc...), called Opcodes. Each opcode has a [gas cost](https://www.evm.codes/), that reflects the complexity of its execution (e.g. `ADD = 3gas` and `SSTORE = 100gas`) and the actial cost of running these operations on actual computer hardware. The **gas price** can change depending on the demand of the network at the time. If the network is under heavy load, you might have to pay a highter gas price to get your transaction executed.
The EVM bytecode is stored on chain, not the solidity code.

### Opcodes

The EVM operates as a stack-based machine, where transactions carry a payload of Opcodes, that are used to specify the interaction with a smart contract.

Typically contracts expose a public ABI, which is a list of supported ways a user can interact with a contract. To interact with a contract, a user will submit a transaction carrying any amount of wei (including 0) and a data payload formatted according to the ABI, specifying the type of interaction and any additional parameters. Each Opcode execution requires gas that needs to be payed with the transaction. The EVM is therefore considered quasi-turing complete, as it allows any arbitrary computation, but the amount of computations during a contract execution is limited to the amount gas provided in the transaction.

For further reading, please refer to:

- [EVM](https://eth.wiki/concepts/evm/evm)
- [EVM Architecture](https://cypherpunks-core.github.io/ethereumbook/13evm.html#evm_architecture)
- [What is Ethereum](https://ethdocs.org/en/latest/introduction/what-is-ethereum.html#what-is-ethereum)
- [Opcodes](https://www.ethervm.io/)

## StateDB

The `StateDB` interface from [go-ethereum](https://github.com/ethereum/go-ethereum/blob/master/core/vm/interface.go) represents an EVM database for full state querying. EVM state transitions are enabled by this interface, which in the `x/evm` module is implemented by the `Keeper`. The implementation of this interface is what makes Ethermint EVM compatible.

## Consensus Engine

The application using the `x/evm` module interacts with the Tendermint Core Consensus Engine over an Application Blockchain Interface (ABCI). Together, the application and Tendermint Core form the programs that run a complete blockchain and combine business logic with decentralized data storage.

Ethereum transactions that are submitted to the `x/evm` module take part in a this consensus process before being executed and changing the application state. We encourage to understand the basics of the [Tendermint consensus engine](https://docs.tendermint.com/master/introduction/what-is-tendermint.html#intro-to-abci) in order to understand state transitions in detail.

## JSON-RPC

JSON-RPC is a stateless, lightweight remote procedure call (RPC) protocol. Primarily this specification defines several data structures and the rules around their processing. It is transport agnostic in that the concepts can be used within the same process, over sockets, over HTTP, or in many various message passing environments. It uses JSON (RFC 4627) as a data format.

Ethermint supports all standard web3 [JSON-RPC](https://evmos.dev/api/json-rpc/server.html) APIs. For more info check the client section.

## Transaction Logs

On every `x/evm` transaction, the result contains the Ethereum `Log`s from the state machine execution that are used by the JSON-RPC Web3 server for filter querying and for processing the EVM Hooks.

The tx logs are stored in the transient store during tx execution and then emitted through cosmos events after the transaction has been processed. They can be queried via gRPC and JSON-RPC.

## Block Bloom

Bloom is the bloom filter value in bytes for each block that can be used for filter queries. The block bloom value is stored in the transient store and then emitted through a cosmos event during `EndBlock` processing. They can be queried via gRPC and JSON-RPC.

::: tip
👉 **Note**: Since they are not stored on state, Transaction Logs and Block Blooms are not persisted after upgrades. A user must use an archival node after upgrades in order to obtain legacy chain events.
:::