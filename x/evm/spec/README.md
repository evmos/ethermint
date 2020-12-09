<!--
order: 0
title: EVM Overview
parent:
  title: "evm"
-->

# `evm`

## Abstract

This document defines the specification of the Ethereum Virtual Machine (EVM) as a Cosmos SDK module.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Messages](04_messages.md)**
5. **[ABCI](05_abci.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_params.md)**

## Module Architecture

> **NOTE:**: If you're not familiar with the overall module structure from
the SDK modules, please check this [document](https://docs.cosmos.network/master/building-modules/structure.html) as
prerequisite reading.

```shell
evm/
├── client
│   └── cli
│       ├── query.go      # CLI query commands for the module
│       └── tx.go         # CLI transaction commands for the module
├── keeper
│   ├── keeper.go         # ABCI BeginBlock and EndBlock logic
│   ├── keeper.go         # Store keeper that handles the business logic of the module and has access to a specific subtree of the state tree.
│   ├── params.go         # Parameter getter and setter
│   ├── querier.go        # State query functions
│   └── statedb.go        # Functions from types/statedb with a passed in sdk.Context
├── types
│   ├── chain_config.go
│   ├── codec.go          # Type registration for encoding
│   ├── errors.go         # Module-specific errors
│   ├── events.go         # Events exposed to the Tendermint PubSub/Websocket
│   ├── genesis.go        # Genesis state for the module
│   ├── journal.go        # Ethereum Journal of state transitions
│   ├── keys.go           # Store keys and utility functions
│   ├── logs.go           # Types for persisting Ethereum tx logs on state after chain upgrades
│   ├── msg.go            # EVM module transaction messages
│   ├── params.go         # Module parameters that can be customized with governance parameter change proposals
│   ├── state_object.go   # EVM state object
│   ├── statedb.go        # Implementation of the StateDb interface
│   ├── storage.go        # Implementation of the Ethereum state storage map using arrays to prevent non-determinism
│   └── tx_data.go        # Ethereum transaction data types
├── genesis.go            # ABCI InitGenesis and ExportGenesis functionality
├── handler.go            # Message routing
└── module.go             # Module setup for the module manager
```
