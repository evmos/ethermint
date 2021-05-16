/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import {
  IdentifiedConnection,
  ConnectionPaths,
} from "../../../../ibc/core/connection/v1/connection";

export const protobufPackage = "ibc.core.connection.v1";

/** GenesisState defines the ibc connection submodule's genesis state. */
export interface GenesisState {
  connections: IdentifiedConnection[];
  clientConnectionPaths: ConnectionPaths[];
  /** the sequence for the next generated connection identifier */
  nextConnectionSequence: number;
}

const baseGenesisState: object = { nextConnectionSequence: 0 };

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    for (const v of message.connections) {
      IdentifiedConnection.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.clientConnectionPaths) {
      ConnectionPaths.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.nextConnectionSequence !== 0) {
      writer.uint32(24).uint64(message.nextConnectionSequence);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.connections = [];
    message.clientConnectionPaths = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.connections.push(
            IdentifiedConnection.decode(reader, reader.uint32())
          );
          break;
        case 2:
          message.clientConnectionPaths.push(
            ConnectionPaths.decode(reader, reader.uint32())
          );
          break;
        case 3:
          message.nextConnectionSequence = longToNumber(
            reader.uint64() as Long
          );
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
    message.connections = [];
    message.clientConnectionPaths = [];
    if (object.connections !== undefined && object.connections !== null) {
      for (const e of object.connections) {
        message.connections.push(IdentifiedConnection.fromJSON(e));
      }
    }
    if (
      object.clientConnectionPaths !== undefined &&
      object.clientConnectionPaths !== null
    ) {
      for (const e of object.clientConnectionPaths) {
        message.clientConnectionPaths.push(ConnectionPaths.fromJSON(e));
      }
    }
    if (
      object.nextConnectionSequence !== undefined &&
      object.nextConnectionSequence !== null
    ) {
      message.nextConnectionSequence = Number(object.nextConnectionSequence);
    } else {
      message.nextConnectionSequence = 0;
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    if (message.connections) {
      obj.connections = message.connections.map((e) =>
        e ? IdentifiedConnection.toJSON(e) : undefined
      );
    } else {
      obj.connections = [];
    }
    if (message.clientConnectionPaths) {
      obj.clientConnectionPaths = message.clientConnectionPaths.map((e) =>
        e ? ConnectionPaths.toJSON(e) : undefined
      );
    } else {
      obj.clientConnectionPaths = [];
    }
    message.nextConnectionSequence !== undefined &&
      (obj.nextConnectionSequence = message.nextConnectionSequence);
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.connections = [];
    message.clientConnectionPaths = [];
    if (object.connections !== undefined && object.connections !== null) {
      for (const e of object.connections) {
        message.connections.push(IdentifiedConnection.fromPartial(e));
      }
    }
    if (
      object.clientConnectionPaths !== undefined &&
      object.clientConnectionPaths !== null
    ) {
      for (const e of object.clientConnectionPaths) {
        message.clientConnectionPaths.push(ConnectionPaths.fromPartial(e));
      }
    }
    if (
      object.nextConnectionSequence !== undefined &&
      object.nextConnectionSequence !== null
    ) {
      message.nextConnectionSequence = object.nextConnectionSequence;
    } else {
      message.nextConnectionSequence = 0;
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
