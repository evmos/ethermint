/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Any } from "../../../../google/protobuf/any";
import { Event } from "../../../../tendermint/abci/types";

export const protobufPackage = "cosmos.base.abci.v1beta1";

/**
 * TxResponse defines a structure containing relevant tx data and metadata. The
 * tags are stringified and the log is JSON decoded.
 */
export interface TxResponse {
  /** The block height */
  height: number;
  /** The transaction hash. */
  txhash: string;
  /** Namespace for the Code */
  codespace: string;
  /** Response code. */
  code: number;
  /** Result bytes, if any. */
  data: string;
  /**
   * The output of the application's logger (raw string). May be
   * non-deterministic.
   */
  rawLog: string;
  /** The output of the application's logger (typed). May be non-deterministic. */
  logs: ABCIMessageLog[];
  /** Additional information. May be non-deterministic. */
  info: string;
  /** Amount of gas requested for transaction. */
  gasWanted: number;
  /** Amount of gas consumed by transaction. */
  gasUsed: number;
  /** The request transaction bytes. */
  tx: Any | undefined;
  /**
   * Time of the previous block. For heights > 1, it's the weighted median of
   * the timestamps of the valid votes in the block.LastCommit. For height == 1,
   * it's genesis time.
   */
  timestamp: string;
}

/** ABCIMessageLog defines a structure containing an indexed tx ABCI message log. */
export interface ABCIMessageLog {
  msgIndex: number;
  log: string;
  /**
   * Events contains a slice of Event objects that were emitted during some
   * execution.
   */
  events: StringEvent[];
}

/**
 * StringEvent defines en Event object wrapper where all the attributes
 * contain key/value pairs that are strings instead of raw bytes.
 */
export interface StringEvent {
  type: string;
  attributes: Attribute[];
}

/**
 * Attribute defines an attribute wrapper where the key and value are
 * strings instead of raw bytes.
 */
export interface Attribute {
  key: string;
  value: string;
}

/** GasInfo defines tx execution gas context. */
export interface GasInfo {
  /** GasWanted is the maximum units of work we allow this tx to perform. */
  gasWanted: number;
  /** GasUsed is the amount of gas actually consumed. */
  gasUsed: number;
}

/** Result is the union of ResponseFormat and ResponseCheckTx. */
export interface Result {
  /**
   * Data is any data returned from message or handler execution. It MUST be
   * length prefixed in order to separate data from multiple message executions.
   */
  data: Uint8Array;
  /** Log contains the log information from message or handler execution. */
  log: string;
  /**
   * Events contains a slice of Event objects that were emitted during message
   * or handler execution.
   */
  events: Event[];
}

/**
 * SimulationResponse defines the response generated when a transaction is
 * successfully simulated.
 */
export interface SimulationResponse {
  gasInfo: GasInfo | undefined;
  result: Result | undefined;
}

/**
 * MsgData defines the data returned in a Result object during message
 * execution.
 */
export interface MsgData {
  msgType: string;
  data: Uint8Array;
}

/**
 * TxMsgData defines a list of MsgData. A transaction will have a MsgData object
 * for each message.
 */
export interface TxMsgData {
  data: MsgData[];
}

/** SearchTxsResult defines a structure for querying txs pageable */
export interface SearchTxsResult {
  /** Count of all txs */
  totalCount: number;
  /** Count of txs in current page */
  count: number;
  /** Index of current page, start from 1 */
  pageNumber: number;
  /** Count of total pages */
  pageTotal: number;
  /** Max count txs per page */
  limit: number;
  /** List of txs in current page */
  txs: TxResponse[];
}

const baseTxResponse: object = {
  height: 0,
  txhash: "",
  codespace: "",
  code: 0,
  data: "",
  rawLog: "",
  info: "",
  gasWanted: 0,
  gasUsed: 0,
  timestamp: "",
};

export const TxResponse = {
  encode(message: TxResponse, writer: Writer = Writer.create()): Writer {
    if (message.height !== 0) {
      writer.uint32(8).int64(message.height);
    }
    if (message.txhash !== "") {
      writer.uint32(18).string(message.txhash);
    }
    if (message.codespace !== "") {
      writer.uint32(26).string(message.codespace);
    }
    if (message.code !== 0) {
      writer.uint32(32).uint32(message.code);
    }
    if (message.data !== "") {
      writer.uint32(42).string(message.data);
    }
    if (message.rawLog !== "") {
      writer.uint32(50).string(message.rawLog);
    }
    for (const v of message.logs) {
      ABCIMessageLog.encode(v!, writer.uint32(58).fork()).ldelim();
    }
    if (message.info !== "") {
      writer.uint32(66).string(message.info);
    }
    if (message.gasWanted !== 0) {
      writer.uint32(72).int64(message.gasWanted);
    }
    if (message.gasUsed !== 0) {
      writer.uint32(80).int64(message.gasUsed);
    }
    if (message.tx !== undefined) {
      Any.encode(message.tx, writer.uint32(90).fork()).ldelim();
    }
    if (message.timestamp !== "") {
      writer.uint32(98).string(message.timestamp);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): TxResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTxResponse } as TxResponse;
    message.logs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.height = longToNumber(reader.int64() as Long);
          break;
        case 2:
          message.txhash = reader.string();
          break;
        case 3:
          message.codespace = reader.string();
          break;
        case 4:
          message.code = reader.uint32();
          break;
        case 5:
          message.data = reader.string();
          break;
        case 6:
          message.rawLog = reader.string();
          break;
        case 7:
          message.logs.push(ABCIMessageLog.decode(reader, reader.uint32()));
          break;
        case 8:
          message.info = reader.string();
          break;
        case 9:
          message.gasWanted = longToNumber(reader.int64() as Long);
          break;
        case 10:
          message.gasUsed = longToNumber(reader.int64() as Long);
          break;
        case 11:
          message.tx = Any.decode(reader, reader.uint32());
          break;
        case 12:
          message.timestamp = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): TxResponse {
    const message = { ...baseTxResponse } as TxResponse;
    message.logs = [];
    if (object.height !== undefined && object.height !== null) {
      message.height = Number(object.height);
    } else {
      message.height = 0;
    }
    if (object.txhash !== undefined && object.txhash !== null) {
      message.txhash = String(object.txhash);
    } else {
      message.txhash = "";
    }
    if (object.codespace !== undefined && object.codespace !== null) {
      message.codespace = String(object.codespace);
    } else {
      message.codespace = "";
    }
    if (object.code !== undefined && object.code !== null) {
      message.code = Number(object.code);
    } else {
      message.code = 0;
    }
    if (object.data !== undefined && object.data !== null) {
      message.data = String(object.data);
    } else {
      message.data = "";
    }
    if (object.rawLog !== undefined && object.rawLog !== null) {
      message.rawLog = String(object.rawLog);
    } else {
      message.rawLog = "";
    }
    if (object.logs !== undefined && object.logs !== null) {
      for (const e of object.logs) {
        message.logs.push(ABCIMessageLog.fromJSON(e));
      }
    }
    if (object.info !== undefined && object.info !== null) {
      message.info = String(object.info);
    } else {
      message.info = "";
    }
    if (object.gasWanted !== undefined && object.gasWanted !== null) {
      message.gasWanted = Number(object.gasWanted);
    } else {
      message.gasWanted = 0;
    }
    if (object.gasUsed !== undefined && object.gasUsed !== null) {
      message.gasUsed = Number(object.gasUsed);
    } else {
      message.gasUsed = 0;
    }
    if (object.tx !== undefined && object.tx !== null) {
      message.tx = Any.fromJSON(object.tx);
    } else {
      message.tx = undefined;
    }
    if (object.timestamp !== undefined && object.timestamp !== null) {
      message.timestamp = String(object.timestamp);
    } else {
      message.timestamp = "";
    }
    return message;
  },

  toJSON(message: TxResponse): unknown {
    const obj: any = {};
    message.height !== undefined && (obj.height = message.height);
    message.txhash !== undefined && (obj.txhash = message.txhash);
    message.codespace !== undefined && (obj.codespace = message.codespace);
    message.code !== undefined && (obj.code = message.code);
    message.data !== undefined && (obj.data = message.data);
    message.rawLog !== undefined && (obj.rawLog = message.rawLog);
    if (message.logs) {
      obj.logs = message.logs.map((e) =>
        e ? ABCIMessageLog.toJSON(e) : undefined
      );
    } else {
      obj.logs = [];
    }
    message.info !== undefined && (obj.info = message.info);
    message.gasWanted !== undefined && (obj.gasWanted = message.gasWanted);
    message.gasUsed !== undefined && (obj.gasUsed = message.gasUsed);
    message.tx !== undefined &&
      (obj.tx = message.tx ? Any.toJSON(message.tx) : undefined);
    message.timestamp !== undefined && (obj.timestamp = message.timestamp);
    return obj;
  },

  fromPartial(object: DeepPartial<TxResponse>): TxResponse {
    const message = { ...baseTxResponse } as TxResponse;
    message.logs = [];
    if (object.height !== undefined && object.height !== null) {
      message.height = object.height;
    } else {
      message.height = 0;
    }
    if (object.txhash !== undefined && object.txhash !== null) {
      message.txhash = object.txhash;
    } else {
      message.txhash = "";
    }
    if (object.codespace !== undefined && object.codespace !== null) {
      message.codespace = object.codespace;
    } else {
      message.codespace = "";
    }
    if (object.code !== undefined && object.code !== null) {
      message.code = object.code;
    } else {
      message.code = 0;
    }
    if (object.data !== undefined && object.data !== null) {
      message.data = object.data;
    } else {
      message.data = "";
    }
    if (object.rawLog !== undefined && object.rawLog !== null) {
      message.rawLog = object.rawLog;
    } else {
      message.rawLog = "";
    }
    if (object.logs !== undefined && object.logs !== null) {
      for (const e of object.logs) {
        message.logs.push(ABCIMessageLog.fromPartial(e));
      }
    }
    if (object.info !== undefined && object.info !== null) {
      message.info = object.info;
    } else {
      message.info = "";
    }
    if (object.gasWanted !== undefined && object.gasWanted !== null) {
      message.gasWanted = object.gasWanted;
    } else {
      message.gasWanted = 0;
    }
    if (object.gasUsed !== undefined && object.gasUsed !== null) {
      message.gasUsed = object.gasUsed;
    } else {
      message.gasUsed = 0;
    }
    if (object.tx !== undefined && object.tx !== null) {
      message.tx = Any.fromPartial(object.tx);
    } else {
      message.tx = undefined;
    }
    if (object.timestamp !== undefined && object.timestamp !== null) {
      message.timestamp = object.timestamp;
    } else {
      message.timestamp = "";
    }
    return message;
  },
};

const baseABCIMessageLog: object = { msgIndex: 0, log: "" };

export const ABCIMessageLog = {
  encode(message: ABCIMessageLog, writer: Writer = Writer.create()): Writer {
    if (message.msgIndex !== 0) {
      writer.uint32(8).uint32(message.msgIndex);
    }
    if (message.log !== "") {
      writer.uint32(18).string(message.log);
    }
    for (const v of message.events) {
      StringEvent.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): ABCIMessageLog {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseABCIMessageLog } as ABCIMessageLog;
    message.events = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.msgIndex = reader.uint32();
          break;
        case 2:
          message.log = reader.string();
          break;
        case 3:
          message.events.push(StringEvent.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ABCIMessageLog {
    const message = { ...baseABCIMessageLog } as ABCIMessageLog;
    message.events = [];
    if (object.msgIndex !== undefined && object.msgIndex !== null) {
      message.msgIndex = Number(object.msgIndex);
    } else {
      message.msgIndex = 0;
    }
    if (object.log !== undefined && object.log !== null) {
      message.log = String(object.log);
    } else {
      message.log = "";
    }
    if (object.events !== undefined && object.events !== null) {
      for (const e of object.events) {
        message.events.push(StringEvent.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: ABCIMessageLog): unknown {
    const obj: any = {};
    message.msgIndex !== undefined && (obj.msgIndex = message.msgIndex);
    message.log !== undefined && (obj.log = message.log);
    if (message.events) {
      obj.events = message.events.map((e) =>
        e ? StringEvent.toJSON(e) : undefined
      );
    } else {
      obj.events = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<ABCIMessageLog>): ABCIMessageLog {
    const message = { ...baseABCIMessageLog } as ABCIMessageLog;
    message.events = [];
    if (object.msgIndex !== undefined && object.msgIndex !== null) {
      message.msgIndex = object.msgIndex;
    } else {
      message.msgIndex = 0;
    }
    if (object.log !== undefined && object.log !== null) {
      message.log = object.log;
    } else {
      message.log = "";
    }
    if (object.events !== undefined && object.events !== null) {
      for (const e of object.events) {
        message.events.push(StringEvent.fromPartial(e));
      }
    }
    return message;
  },
};

const baseStringEvent: object = { type: "" };

export const StringEvent = {
  encode(message: StringEvent, writer: Writer = Writer.create()): Writer {
    if (message.type !== "") {
      writer.uint32(10).string(message.type);
    }
    for (const v of message.attributes) {
      Attribute.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): StringEvent {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseStringEvent } as StringEvent;
    message.attributes = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.type = reader.string();
          break;
        case 2:
          message.attributes.push(Attribute.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): StringEvent {
    const message = { ...baseStringEvent } as StringEvent;
    message.attributes = [];
    if (object.type !== undefined && object.type !== null) {
      message.type = String(object.type);
    } else {
      message.type = "";
    }
    if (object.attributes !== undefined && object.attributes !== null) {
      for (const e of object.attributes) {
        message.attributes.push(Attribute.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: StringEvent): unknown {
    const obj: any = {};
    message.type !== undefined && (obj.type = message.type);
    if (message.attributes) {
      obj.attributes = message.attributes.map((e) =>
        e ? Attribute.toJSON(e) : undefined
      );
    } else {
      obj.attributes = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<StringEvent>): StringEvent {
    const message = { ...baseStringEvent } as StringEvent;
    message.attributes = [];
    if (object.type !== undefined && object.type !== null) {
      message.type = object.type;
    } else {
      message.type = "";
    }
    if (object.attributes !== undefined && object.attributes !== null) {
      for (const e of object.attributes) {
        message.attributes.push(Attribute.fromPartial(e));
      }
    }
    return message;
  },
};

const baseAttribute: object = { key: "", value: "" };

export const Attribute = {
  encode(message: Attribute, writer: Writer = Writer.create()): Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(18).string(message.value);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Attribute {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseAttribute } as Attribute;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.value = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Attribute {
    const message = { ...baseAttribute } as Attribute;
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = String(object.value);
    } else {
      message.value = "";
    }
    return message;
  },

  toJSON(message: Attribute): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial(object: DeepPartial<Attribute>): Attribute {
    const message = { ...baseAttribute } as Attribute;
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = "";
    }
    return message;
  },
};

const baseGasInfo: object = { gasWanted: 0, gasUsed: 0 };

export const GasInfo = {
  encode(message: GasInfo, writer: Writer = Writer.create()): Writer {
    if (message.gasWanted !== 0) {
      writer.uint32(8).uint64(message.gasWanted);
    }
    if (message.gasUsed !== 0) {
      writer.uint32(16).uint64(message.gasUsed);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GasInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGasInfo } as GasInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.gasWanted = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.gasUsed = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GasInfo {
    const message = { ...baseGasInfo } as GasInfo;
    if (object.gasWanted !== undefined && object.gasWanted !== null) {
      message.gasWanted = Number(object.gasWanted);
    } else {
      message.gasWanted = 0;
    }
    if (object.gasUsed !== undefined && object.gasUsed !== null) {
      message.gasUsed = Number(object.gasUsed);
    } else {
      message.gasUsed = 0;
    }
    return message;
  },

  toJSON(message: GasInfo): unknown {
    const obj: any = {};
    message.gasWanted !== undefined && (obj.gasWanted = message.gasWanted);
    message.gasUsed !== undefined && (obj.gasUsed = message.gasUsed);
    return obj;
  },

  fromPartial(object: DeepPartial<GasInfo>): GasInfo {
    const message = { ...baseGasInfo } as GasInfo;
    if (object.gasWanted !== undefined && object.gasWanted !== null) {
      message.gasWanted = object.gasWanted;
    } else {
      message.gasWanted = 0;
    }
    if (object.gasUsed !== undefined && object.gasUsed !== null) {
      message.gasUsed = object.gasUsed;
    } else {
      message.gasUsed = 0;
    }
    return message;
  },
};

const baseResult: object = { log: "" };

export const Result = {
  encode(message: Result, writer: Writer = Writer.create()): Writer {
    if (message.data.length !== 0) {
      writer.uint32(10).bytes(message.data);
    }
    if (message.log !== "") {
      writer.uint32(18).string(message.log);
    }
    for (const v of message.events) {
      Event.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Result {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseResult } as Result;
    message.events = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.data = reader.bytes();
          break;
        case 2:
          message.log = reader.string();
          break;
        case 3:
          message.events.push(Event.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Result {
    const message = { ...baseResult } as Result;
    message.events = [];
    if (object.data !== undefined && object.data !== null) {
      message.data = bytesFromBase64(object.data);
    }
    if (object.log !== undefined && object.log !== null) {
      message.log = String(object.log);
    } else {
      message.log = "";
    }
    if (object.events !== undefined && object.events !== null) {
      for (const e of object.events) {
        message.events.push(Event.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: Result): unknown {
    const obj: any = {};
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array()
      ));
    message.log !== undefined && (obj.log = message.log);
    if (message.events) {
      obj.events = message.events.map((e) => (e ? Event.toJSON(e) : undefined));
    } else {
      obj.events = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<Result>): Result {
    const message = { ...baseResult } as Result;
    message.events = [];
    if (object.data !== undefined && object.data !== null) {
      message.data = object.data;
    } else {
      message.data = new Uint8Array();
    }
    if (object.log !== undefined && object.log !== null) {
      message.log = object.log;
    } else {
      message.log = "";
    }
    if (object.events !== undefined && object.events !== null) {
      for (const e of object.events) {
        message.events.push(Event.fromPartial(e));
      }
    }
    return message;
  },
};

const baseSimulationResponse: object = {};

export const SimulationResponse = {
  encode(
    message: SimulationResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.gasInfo !== undefined) {
      GasInfo.encode(message.gasInfo, writer.uint32(10).fork()).ldelim();
    }
    if (message.result !== undefined) {
      Result.encode(message.result, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): SimulationResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSimulationResponse } as SimulationResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.gasInfo = GasInfo.decode(reader, reader.uint32());
          break;
        case 2:
          message.result = Result.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SimulationResponse {
    const message = { ...baseSimulationResponse } as SimulationResponse;
    if (object.gasInfo !== undefined && object.gasInfo !== null) {
      message.gasInfo = GasInfo.fromJSON(object.gasInfo);
    } else {
      message.gasInfo = undefined;
    }
    if (object.result !== undefined && object.result !== null) {
      message.result = Result.fromJSON(object.result);
    } else {
      message.result = undefined;
    }
    return message;
  },

  toJSON(message: SimulationResponse): unknown {
    const obj: any = {};
    message.gasInfo !== undefined &&
      (obj.gasInfo = message.gasInfo
        ? GasInfo.toJSON(message.gasInfo)
        : undefined);
    message.result !== undefined &&
      (obj.result = message.result ? Result.toJSON(message.result) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<SimulationResponse>): SimulationResponse {
    const message = { ...baseSimulationResponse } as SimulationResponse;
    if (object.gasInfo !== undefined && object.gasInfo !== null) {
      message.gasInfo = GasInfo.fromPartial(object.gasInfo);
    } else {
      message.gasInfo = undefined;
    }
    if (object.result !== undefined && object.result !== null) {
      message.result = Result.fromPartial(object.result);
    } else {
      message.result = undefined;
    }
    return message;
  },
};

const baseMsgData: object = { msgType: "" };

export const MsgData = {
  encode(message: MsgData, writer: Writer = Writer.create()): Writer {
    if (message.msgType !== "") {
      writer.uint32(10).string(message.msgType);
    }
    if (message.data.length !== 0) {
      writer.uint32(18).bytes(message.data);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgData } as MsgData;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.msgType = reader.string();
          break;
        case 2:
          message.data = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgData {
    const message = { ...baseMsgData } as MsgData;
    if (object.msgType !== undefined && object.msgType !== null) {
      message.msgType = String(object.msgType);
    } else {
      message.msgType = "";
    }
    if (object.data !== undefined && object.data !== null) {
      message.data = bytesFromBase64(object.data);
    }
    return message;
  },

  toJSON(message: MsgData): unknown {
    const obj: any = {};
    message.msgType !== undefined && (obj.msgType = message.msgType);
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<MsgData>): MsgData {
    const message = { ...baseMsgData } as MsgData;
    if (object.msgType !== undefined && object.msgType !== null) {
      message.msgType = object.msgType;
    } else {
      message.msgType = "";
    }
    if (object.data !== undefined && object.data !== null) {
      message.data = object.data;
    } else {
      message.data = new Uint8Array();
    }
    return message;
  },
};

const baseTxMsgData: object = {};

export const TxMsgData = {
  encode(message: TxMsgData, writer: Writer = Writer.create()): Writer {
    for (const v of message.data) {
      MsgData.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): TxMsgData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTxMsgData } as TxMsgData;
    message.data = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.data.push(MsgData.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): TxMsgData {
    const message = { ...baseTxMsgData } as TxMsgData;
    message.data = [];
    if (object.data !== undefined && object.data !== null) {
      for (const e of object.data) {
        message.data.push(MsgData.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: TxMsgData): unknown {
    const obj: any = {};
    if (message.data) {
      obj.data = message.data.map((e) => (e ? MsgData.toJSON(e) : undefined));
    } else {
      obj.data = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<TxMsgData>): TxMsgData {
    const message = { ...baseTxMsgData } as TxMsgData;
    message.data = [];
    if (object.data !== undefined && object.data !== null) {
      for (const e of object.data) {
        message.data.push(MsgData.fromPartial(e));
      }
    }
    return message;
  },
};

const baseSearchTxsResult: object = {
  totalCount: 0,
  count: 0,
  pageNumber: 0,
  pageTotal: 0,
  limit: 0,
};

export const SearchTxsResult = {
  encode(message: SearchTxsResult, writer: Writer = Writer.create()): Writer {
    if (message.totalCount !== 0) {
      writer.uint32(8).uint64(message.totalCount);
    }
    if (message.count !== 0) {
      writer.uint32(16).uint64(message.count);
    }
    if (message.pageNumber !== 0) {
      writer.uint32(24).uint64(message.pageNumber);
    }
    if (message.pageTotal !== 0) {
      writer.uint32(32).uint64(message.pageTotal);
    }
    if (message.limit !== 0) {
      writer.uint32(40).uint64(message.limit);
    }
    for (const v of message.txs) {
      TxResponse.encode(v!, writer.uint32(50).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): SearchTxsResult {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSearchTxsResult } as SearchTxsResult;
    message.txs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.totalCount = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.count = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.pageNumber = longToNumber(reader.uint64() as Long);
          break;
        case 4:
          message.pageTotal = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          message.limit = longToNumber(reader.uint64() as Long);
          break;
        case 6:
          message.txs.push(TxResponse.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SearchTxsResult {
    const message = { ...baseSearchTxsResult } as SearchTxsResult;
    message.txs = [];
    if (object.totalCount !== undefined && object.totalCount !== null) {
      message.totalCount = Number(object.totalCount);
    } else {
      message.totalCount = 0;
    }
    if (object.count !== undefined && object.count !== null) {
      message.count = Number(object.count);
    } else {
      message.count = 0;
    }
    if (object.pageNumber !== undefined && object.pageNumber !== null) {
      message.pageNumber = Number(object.pageNumber);
    } else {
      message.pageNumber = 0;
    }
    if (object.pageTotal !== undefined && object.pageTotal !== null) {
      message.pageTotal = Number(object.pageTotal);
    } else {
      message.pageTotal = 0;
    }
    if (object.limit !== undefined && object.limit !== null) {
      message.limit = Number(object.limit);
    } else {
      message.limit = 0;
    }
    if (object.txs !== undefined && object.txs !== null) {
      for (const e of object.txs) {
        message.txs.push(TxResponse.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: SearchTxsResult): unknown {
    const obj: any = {};
    message.totalCount !== undefined && (obj.totalCount = message.totalCount);
    message.count !== undefined && (obj.count = message.count);
    message.pageNumber !== undefined && (obj.pageNumber = message.pageNumber);
    message.pageTotal !== undefined && (obj.pageTotal = message.pageTotal);
    message.limit !== undefined && (obj.limit = message.limit);
    if (message.txs) {
      obj.txs = message.txs.map((e) => (e ? TxResponse.toJSON(e) : undefined));
    } else {
      obj.txs = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<SearchTxsResult>): SearchTxsResult {
    const message = { ...baseSearchTxsResult } as SearchTxsResult;
    message.txs = [];
    if (object.totalCount !== undefined && object.totalCount !== null) {
      message.totalCount = object.totalCount;
    } else {
      message.totalCount = 0;
    }
    if (object.count !== undefined && object.count !== null) {
      message.count = object.count;
    } else {
      message.count = 0;
    }
    if (object.pageNumber !== undefined && object.pageNumber !== null) {
      message.pageNumber = object.pageNumber;
    } else {
      message.pageNumber = 0;
    }
    if (object.pageTotal !== undefined && object.pageTotal !== null) {
      message.pageTotal = object.pageTotal;
    } else {
      message.pageTotal = 0;
    }
    if (object.limit !== undefined && object.limit !== null) {
      message.limit = object.limit;
    } else {
      message.limit = 0;
    }
    if (object.txs !== undefined && object.txs !== null) {
      for (const e of object.txs) {
        message.txs.push(TxResponse.fromPartial(e));
      }
    }
    return message;
  },
};

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

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
