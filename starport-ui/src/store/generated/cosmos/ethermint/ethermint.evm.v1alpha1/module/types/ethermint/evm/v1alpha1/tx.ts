/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { TxData, TransactionLogs } from "../../../ethermint/evm/v1alpha1/evm";

export const protobufPackage = "ethermint.evm.v1alpha1";

/** MsgEthereumTx encapsulates an Ethereum transaction as an SDK message. */
export interface MsgEthereumTx {
  /** inner transaction data */
  data: TxData | undefined;
  /** encoded storage size of the transaction */
  size: number;
  /** transaction hash in hex format */
  hash: string;
  /**
   * ethereum signer address in hex format. This address value is checked against
   * the address derived from the signature (V, R, S) using the secp256k1
   * elliptic curve
   */
  from: string;
}

export interface ExtensionOptionsEthereumTx {}

export interface ExtensionOptionsWeb3Tx {}

/** MsgEthereumTxResponse defines the Msg/EthereumTx response type. */
export interface MsgEthereumTxResponse {
  /**
   * contract_address contains the ethereum address of the created contract (if
   * any). If the state transition is an evm.Call, the contract address will be
   * empty.
   */
  contractAddress: string;
  /** bloom represents the bloom filter bytes */
  bloom: Uint8Array;
  /**
   * tx_logs contains the transaction hash and the proto-compatible ethereum
   * logs.
   */
  txLogs: TransactionLogs | undefined;
  /** ret defines the bytes from the execution. */
  ret: Uint8Array;
  /** reverted flag is set to true when the call has been reverted */
  reverted: boolean;
}

const baseMsgEthereumTx: object = { size: 0, hash: "", from: "" };

export const MsgEthereumTx = {
  encode(message: MsgEthereumTx, writer: Writer = Writer.create()): Writer {
    if (message.data !== undefined) {
      TxData.encode(message.data, writer.uint32(10).fork()).ldelim();
    }
    if (message.size !== 0) {
      writer.uint32(17).double(message.size);
    }
    if (message.hash !== "") {
      writer.uint32(26).string(message.hash);
    }
    if (message.from !== "") {
      writer.uint32(34).string(message.from);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgEthereumTx {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgEthereumTx } as MsgEthereumTx;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.data = TxData.decode(reader, reader.uint32());
          break;
        case 2:
          message.size = reader.double();
          break;
        case 3:
          message.hash = reader.string();
          break;
        case 4:
          message.from = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgEthereumTx {
    const message = { ...baseMsgEthereumTx } as MsgEthereumTx;
    if (object.data !== undefined && object.data !== null) {
      message.data = TxData.fromJSON(object.data);
    } else {
      message.data = undefined;
    }
    if (object.size !== undefined && object.size !== null) {
      message.size = Number(object.size);
    } else {
      message.size = 0;
    }
    if (object.hash !== undefined && object.hash !== null) {
      message.hash = String(object.hash);
    } else {
      message.hash = "";
    }
    if (object.from !== undefined && object.from !== null) {
      message.from = String(object.from);
    } else {
      message.from = "";
    }
    return message;
  },

  toJSON(message: MsgEthereumTx): unknown {
    const obj: any = {};
    message.data !== undefined &&
      (obj.data = message.data ? TxData.toJSON(message.data) : undefined);
    message.size !== undefined && (obj.size = message.size);
    message.hash !== undefined && (obj.hash = message.hash);
    message.from !== undefined && (obj.from = message.from);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgEthereumTx>): MsgEthereumTx {
    const message = { ...baseMsgEthereumTx } as MsgEthereumTx;
    if (object.data !== undefined && object.data !== null) {
      message.data = TxData.fromPartial(object.data);
    } else {
      message.data = undefined;
    }
    if (object.size !== undefined && object.size !== null) {
      message.size = object.size;
    } else {
      message.size = 0;
    }
    if (object.hash !== undefined && object.hash !== null) {
      message.hash = object.hash;
    } else {
      message.hash = "";
    }
    if (object.from !== undefined && object.from !== null) {
      message.from = object.from;
    } else {
      message.from = "";
    }
    return message;
  },
};

const baseExtensionOptionsEthereumTx: object = {};

export const ExtensionOptionsEthereumTx = {
  encode(
    _: ExtensionOptionsEthereumTx,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): ExtensionOptionsEthereumTx {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseExtensionOptionsEthereumTx,
    } as ExtensionOptionsEthereumTx;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): ExtensionOptionsEthereumTx {
    const message = {
      ...baseExtensionOptionsEthereumTx,
    } as ExtensionOptionsEthereumTx;
    return message;
  },

  toJSON(_: ExtensionOptionsEthereumTx): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<ExtensionOptionsEthereumTx>
  ): ExtensionOptionsEthereumTx {
    const message = {
      ...baseExtensionOptionsEthereumTx,
    } as ExtensionOptionsEthereumTx;
    return message;
  },
};

const baseExtensionOptionsWeb3Tx: object = {};

export const ExtensionOptionsWeb3Tx = {
  encode(_: ExtensionOptionsWeb3Tx, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): ExtensionOptionsWeb3Tx {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseExtensionOptionsWeb3Tx } as ExtensionOptionsWeb3Tx;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): ExtensionOptionsWeb3Tx {
    const message = { ...baseExtensionOptionsWeb3Tx } as ExtensionOptionsWeb3Tx;
    return message;
  },

  toJSON(_: ExtensionOptionsWeb3Tx): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<ExtensionOptionsWeb3Tx>): ExtensionOptionsWeb3Tx {
    const message = { ...baseExtensionOptionsWeb3Tx } as ExtensionOptionsWeb3Tx;
    return message;
  },
};

const baseMsgEthereumTxResponse: object = {
  contractAddress: "",
  reverted: false,
};

export const MsgEthereumTxResponse = {
  encode(
    message: MsgEthereumTxResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.contractAddress !== "") {
      writer.uint32(10).string(message.contractAddress);
    }
    if (message.bloom.length !== 0) {
      writer.uint32(18).bytes(message.bloom);
    }
    if (message.txLogs !== undefined) {
      TransactionLogs.encode(message.txLogs, writer.uint32(26).fork()).ldelim();
    }
    if (message.ret.length !== 0) {
      writer.uint32(34).bytes(message.ret);
    }
    if (message.reverted === true) {
      writer.uint32(40).bool(message.reverted);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgEthereumTxResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgEthereumTxResponse } as MsgEthereumTxResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.contractAddress = reader.string();
          break;
        case 2:
          message.bloom = reader.bytes();
          break;
        case 3:
          message.txLogs = TransactionLogs.decode(reader, reader.uint32());
          break;
        case 4:
          message.ret = reader.bytes();
          break;
        case 5:
          message.reverted = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgEthereumTxResponse {
    const message = { ...baseMsgEthereumTxResponse } as MsgEthereumTxResponse;
    if (
      object.contractAddress !== undefined &&
      object.contractAddress !== null
    ) {
      message.contractAddress = String(object.contractAddress);
    } else {
      message.contractAddress = "";
    }
    if (object.bloom !== undefined && object.bloom !== null) {
      message.bloom = bytesFromBase64(object.bloom);
    }
    if (object.txLogs !== undefined && object.txLogs !== null) {
      message.txLogs = TransactionLogs.fromJSON(object.txLogs);
    } else {
      message.txLogs = undefined;
    }
    if (object.ret !== undefined && object.ret !== null) {
      message.ret = bytesFromBase64(object.ret);
    }
    if (object.reverted !== undefined && object.reverted !== null) {
      message.reverted = Boolean(object.reverted);
    } else {
      message.reverted = false;
    }
    return message;
  },

  toJSON(message: MsgEthereumTxResponse): unknown {
    const obj: any = {};
    message.contractAddress !== undefined &&
      (obj.contractAddress = message.contractAddress);
    message.bloom !== undefined &&
      (obj.bloom = base64FromBytes(
        message.bloom !== undefined ? message.bloom : new Uint8Array()
      ));
    message.txLogs !== undefined &&
      (obj.txLogs = message.txLogs
        ? TransactionLogs.toJSON(message.txLogs)
        : undefined);
    message.ret !== undefined &&
      (obj.ret = base64FromBytes(
        message.ret !== undefined ? message.ret : new Uint8Array()
      ));
    message.reverted !== undefined && (obj.reverted = message.reverted);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgEthereumTxResponse>
  ): MsgEthereumTxResponse {
    const message = { ...baseMsgEthereumTxResponse } as MsgEthereumTxResponse;
    if (
      object.contractAddress !== undefined &&
      object.contractAddress !== null
    ) {
      message.contractAddress = object.contractAddress;
    } else {
      message.contractAddress = "";
    }
    if (object.bloom !== undefined && object.bloom !== null) {
      message.bloom = object.bloom;
    } else {
      message.bloom = new Uint8Array();
    }
    if (object.txLogs !== undefined && object.txLogs !== null) {
      message.txLogs = TransactionLogs.fromPartial(object.txLogs);
    } else {
      message.txLogs = undefined;
    }
    if (object.ret !== undefined && object.ret !== null) {
      message.ret = object.ret;
    } else {
      message.ret = new Uint8Array();
    }
    if (object.reverted !== undefined && object.reverted !== null) {
      message.reverted = object.reverted;
    } else {
      message.reverted = false;
    }
    return message;
  },
};

/** Msg defines the evm Msg service. */
export interface Msg {
  /** EthereumTx defines a method submitting Ethereum transactions. */
  EthereumTx(request: MsgEthereumTx): Promise<MsgEthereumTxResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  EthereumTx(request: MsgEthereumTx): Promise<MsgEthereumTxResponse> {
    const data = MsgEthereumTx.encode(request).finish();
    const promise = this.rpc.request(
      "ethermint.evm.v1alpha1.Msg",
      "EthereumTx",
      data
    );
    return promise.then((data) =>
      MsgEthereumTxResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (let i = 0; i < arr.byteLength; ++i) {
    bin.push(String.fromCharCode(arr[i]));
  }
  return btoa(bin.join(""));
}

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
