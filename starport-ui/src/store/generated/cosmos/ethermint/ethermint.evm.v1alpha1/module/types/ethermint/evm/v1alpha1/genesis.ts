/* eslint-disable */
import {
  ChainConfig,
  Params,
  TransactionLogs,
  State,
} from "../../../ethermint/evm/v1alpha1/evm";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "ethermint.evm.v1alpha1";

/** GenesisState defines the evm module's genesis state. */
export interface GenesisState {
  /** accounts is an array containing the ethereum genesis accounts. */
  accounts: GenesisAccount[];
  /** chain_config defines the Ethereum chain configuration. */
  chainConfig: ChainConfig | undefined;
  /** params defines all the paramaters of the module. */
  params: Params | undefined;
  txsLogs: TransactionLogs[];
}

/**
 * GenesisAccount defines an account to be initialized in the genesis state.
 * Its main difference between with Geth's GenesisAccount is that it uses a
 * custom storage type and that it doesn't contain the private key field.
 */
export interface GenesisAccount {
  /** address defines an ethereum hex formated address of an account */
  address: string;
  /** code defines the hex bytes of the account code. */
  code: string;
  /** storage defines the set of state key values for the account. */
  storage: State[];
}

const baseGenesisState: object = {};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    for (const v of message.accounts) {
      GenesisAccount.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.chainConfig !== undefined) {
      ChainConfig.encode(
        message.chainConfig,
        writer.uint32(18).fork()
      ).ldelim();
    }
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.txsLogs) {
      TransactionLogs.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.accounts = [];
    message.txsLogs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.accounts.push(GenesisAccount.decode(reader, reader.uint32()));
          break;
        case 2:
          message.chainConfig = ChainConfig.decode(reader, reader.uint32());
          break;
        case 3:
          message.params = Params.decode(reader, reader.uint32());
          break;
        case 4:
          message.txsLogs.push(TransactionLogs.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.accounts = [];
    message.txsLogs = [];
    if (object.accounts !== undefined && object.accounts !== null) {
      for (const e of object.accounts) {
        message.accounts.push(GenesisAccount.fromJSON(e));
      }
    }
    if (object.chainConfig !== undefined && object.chainConfig !== null) {
      message.chainConfig = ChainConfig.fromJSON(object.chainConfig);
    } else {
      message.chainConfig = undefined;
    }
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    if (object.txsLogs !== undefined && object.txsLogs !== null) {
      for (const e of object.txsLogs) {
        message.txsLogs.push(TransactionLogs.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    if (message.accounts) {
      obj.accounts = message.accounts.map((e) =>
        e ? GenesisAccount.toJSON(e) : undefined
      );
    } else {
      obj.accounts = [];
    }
    message.chainConfig !== undefined &&
      (obj.chainConfig = message.chainConfig
        ? ChainConfig.toJSON(message.chainConfig)
        : undefined);
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    if (message.txsLogs) {
      obj.txsLogs = message.txsLogs.map((e) =>
        e ? TransactionLogs.toJSON(e) : undefined
      );
    } else {
      obj.txsLogs = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.accounts = [];
    message.txsLogs = [];
    if (object.accounts !== undefined && object.accounts !== null) {
      for (const e of object.accounts) {
        message.accounts.push(GenesisAccount.fromPartial(e));
      }
    }
    if (object.chainConfig !== undefined && object.chainConfig !== null) {
      message.chainConfig = ChainConfig.fromPartial(object.chainConfig);
    } else {
      message.chainConfig = undefined;
    }
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    if (object.txsLogs !== undefined && object.txsLogs !== null) {
      for (const e of object.txsLogs) {
        message.txsLogs.push(TransactionLogs.fromPartial(e));
      }
    }
    return message;
  },
};

const baseGenesisAccount: object = { address: "", code: "" };

export const GenesisAccount = {
  encode(message: GenesisAccount, writer: Writer = Writer.create()): Writer {
    if (message.address !== "") {
      writer.uint32(10).string(message.address);
    }
    if (message.code !== "") {
      writer.uint32(18).string(message.code);
    }
    for (const v of message.storage) {
      State.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisAccount {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisAccount } as GenesisAccount;
    message.storage = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.address = reader.string();
          break;
        case 2:
          message.code = reader.string();
          break;
        case 3:
          message.storage.push(State.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisAccount {
    const message = { ...baseGenesisAccount } as GenesisAccount;
    message.storage = [];
    if (object.address !== undefined && object.address !== null) {
      message.address = String(object.address);
    } else {
      message.address = "";
    }
    if (object.code !== undefined && object.code !== null) {
      message.code = String(object.code);
    } else {
      message.code = "";
    }
    if (object.storage !== undefined && object.storage !== null) {
      for (const e of object.storage) {
        message.storage.push(State.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GenesisAccount): unknown {
    const obj: any = {};
    message.address !== undefined && (obj.address = message.address);
    message.code !== undefined && (obj.code = message.code);
    if (message.storage) {
      obj.storage = message.storage.map((e) =>
        e ? State.toJSON(e) : undefined
      );
    } else {
      obj.storage = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisAccount>): GenesisAccount {
    const message = { ...baseGenesisAccount } as GenesisAccount;
    message.storage = [];
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = "";
    }
    if (object.code !== undefined && object.code !== null) {
      message.code = object.code;
    } else {
      message.code = "";
    }
    if (object.storage !== undefined && object.storage !== null) {
      for (const e of object.storage) {
        message.storage.push(State.fromPartial(e));
      }
    }
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;
