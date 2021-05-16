/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Any } from "../../../../google/protobuf/any";
import {
  Height,
  IdentifiedClientState,
  ConsensusStateWithHeight,
  Params,
} from "../../../../ibc/core/client/v1/client";
import {
  PageRequest,
  PageResponse,
} from "../../../../cosmos/base/query/v1beta1/pagination";

export const protobufPackage = "ibc.core.client.v1";

/**
 * QueryClientStateRequest is the request type for the Query/ClientState RPC
 * method
 */
export interface QueryClientStateRequest {
  /** client state unique identifier */
  clientId: string;
}

/**
 * QueryClientStateResponse is the response type for the Query/ClientState RPC
 * method. Besides the client state, it includes a proof and the height from
 * which the proof was retrieved.
 */
export interface QueryClientStateResponse {
  /** client state associated with the request identifier */
  clientState: Any | undefined;
  /** merkle proof of existence */
  proof: Uint8Array;
  /** height at which the proof was retrieved */
  proofHeight: Height | undefined;
}

/**
 * QueryClientStatesRequest is the request type for the Query/ClientStates RPC
 * method
 */
export interface QueryClientStatesRequest {
  /** pagination request */
  pagination: PageRequest | undefined;
}

/**
 * QueryClientStatesResponse is the response type for the Query/ClientStates RPC
 * method.
 */
export interface QueryClientStatesResponse {
  /** list of stored ClientStates of the chain. */
  clientStates: IdentifiedClientState[];
  /** pagination response */
  pagination: PageResponse | undefined;
}

/**
 * QueryConsensusStateRequest is the request type for the Query/ConsensusState
 * RPC method. Besides the consensus state, it includes a proof and the height
 * from which the proof was retrieved.
 */
export interface QueryConsensusStateRequest {
  /** client identifier */
  clientId: string;
  /** consensus state revision number */
  revisionNumber: number;
  /** consensus state revision height */
  revisionHeight: number;
  /**
   * latest_height overrrides the height field and queries the latest stored
   * ConsensusState
   */
  latestHeight: boolean;
}

/**
 * QueryConsensusStateResponse is the response type for the Query/ConsensusState
 * RPC method
 */
export interface QueryConsensusStateResponse {
  /** consensus state associated with the client identifier at the given height */
  consensusState: Any | undefined;
  /** merkle proof of existence */
  proof: Uint8Array;
  /** height at which the proof was retrieved */
  proofHeight: Height | undefined;
}

/**
 * QueryConsensusStatesRequest is the request type for the Query/ConsensusStates
 * RPC method.
 */
export interface QueryConsensusStatesRequest {
  /** client identifier */
  clientId: string;
  /** pagination request */
  pagination: PageRequest | undefined;
}

/**
 * QueryConsensusStatesResponse is the response type for the
 * Query/ConsensusStates RPC method
 */
export interface QueryConsensusStatesResponse {
  /** consensus states associated with the identifier */
  consensusStates: ConsensusStateWithHeight[];
  /** pagination response */
  pagination: PageResponse | undefined;
}

/** QueryClientParamsRequest is the request type for the Query/ClientParams RPC method. */
export interface QueryClientParamsRequest {}

/** QueryClientParamsResponse is the response type for the Query/ClientParams RPC method. */
export interface QueryClientParamsResponse {
  /** params defines the parameters of the module. */
  params: Params | undefined;
}

const baseQueryClientStateRequest: object = { clientId: "" };

export const QueryClientStateRequest = {
  encode(
    message: QueryClientStateRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.clientId !== "") {
      writer.uint32(10).string(message.clientId);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryClientStateRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryClientStateRequest,
    } as QueryClientStateRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.clientId = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryClientStateRequest {
    const message = {
      ...baseQueryClientStateRequest,
    } as QueryClientStateRequest;
    if (object.clientId !== undefined && object.clientId !== null) {
      message.clientId = String(object.clientId);
    } else {
      message.clientId = "";
    }
    return message;
  },

  toJSON(message: QueryClientStateRequest): unknown {
    const obj: any = {};
    message.clientId !== undefined && (obj.clientId = message.clientId);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryClientStateRequest>
  ): QueryClientStateRequest {
    const message = {
      ...baseQueryClientStateRequest,
    } as QueryClientStateRequest;
    if (object.clientId !== undefined && object.clientId !== null) {
      message.clientId = object.clientId;
    } else {
      message.clientId = "";
    }
    return message;
  },
};

const baseQueryClientStateResponse: object = {};

export const QueryClientStateResponse = {
  encode(
    message: QueryClientStateResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.clientState !== undefined) {
      Any.encode(message.clientState, writer.uint32(10).fork()).ldelim();
    }
    if (message.proof.length !== 0) {
      writer.uint32(18).bytes(message.proof);
    }
    if (message.proofHeight !== undefined) {
      Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryClientStateResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryClientStateResponse,
    } as QueryClientStateResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.clientState = Any.decode(reader, reader.uint32());
          break;
        case 2:
          message.proof = reader.bytes();
          break;
        case 3:
          message.proofHeight = Height.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryClientStateResponse {
    const message = {
      ...baseQueryClientStateResponse,
    } as QueryClientStateResponse;
    if (object.clientState !== undefined && object.clientState !== null) {
      message.clientState = Any.fromJSON(object.clientState);
    } else {
      message.clientState = undefined;
    }
    if (object.proof !== undefined && object.proof !== null) {
      message.proof = bytesFromBase64(object.proof);
    }
    if (object.proofHeight !== undefined && object.proofHeight !== null) {
      message.proofHeight = Height.fromJSON(object.proofHeight);
    } else {
      message.proofHeight = undefined;
    }
    return message;
  },

  toJSON(message: QueryClientStateResponse): unknown {
    const obj: any = {};
    message.clientState !== undefined &&
      (obj.clientState = message.clientState
        ? Any.toJSON(message.clientState)
        : undefined);
    message.proof !== undefined &&
      (obj.proof = base64FromBytes(
        message.proof !== undefined ? message.proof : new Uint8Array()
      ));
    message.proofHeight !== undefined &&
      (obj.proofHeight = message.proofHeight
        ? Height.toJSON(message.proofHeight)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryClientStateResponse>
  ): QueryClientStateResponse {
    const message = {
      ...baseQueryClientStateResponse,
    } as QueryClientStateResponse;
    if (object.clientState !== undefined && object.clientState !== null) {
      message.clientState = Any.fromPartial(object.clientState);
    } else {
      message.clientState = undefined;
    }
    if (object.proof !== undefined && object.proof !== null) {
      message.proof = object.proof;
    } else {
      message.proof = new Uint8Array();
    }
    if (object.proofHeight !== undefined && object.proofHeight !== null) {
      message.proofHeight = Height.fromPartial(object.proofHeight);
    } else {
      message.proofHeight = undefined;
    }
    return message;
  },
};

const baseQueryClientStatesRequest: object = {};

export const QueryClientStatesRequest = {
  encode(
    message: QueryClientStatesRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryClientStatesRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryClientStatesRequest,
    } as QueryClientStatesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryClientStatesRequest {
    const message = {
      ...baseQueryClientStatesRequest,
    } as QueryClientStatesRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryClientStatesRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryClientStatesRequest>
  ): QueryClientStatesRequest {
    const message = {
      ...baseQueryClientStatesRequest,
    } as QueryClientStatesRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryClientStatesResponse: object = {};

export const QueryClientStatesResponse = {
  encode(
    message: QueryClientStatesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.clientStates) {
      IdentifiedClientState.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryClientStatesResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryClientStatesResponse,
    } as QueryClientStatesResponse;
    message.clientStates = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.clientStates.push(
            IdentifiedClientState.decode(reader, reader.uint32())
          );
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryClientStatesResponse {
    const message = {
      ...baseQueryClientStatesResponse,
    } as QueryClientStatesResponse;
    message.clientStates = [];
    if (object.clientStates !== undefined && object.clientStates !== null) {
      for (const e of object.clientStates) {
        message.clientStates.push(IdentifiedClientState.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryClientStatesResponse): unknown {
    const obj: any = {};
    if (message.clientStates) {
      obj.clientStates = message.clientStates.map((e) =>
        e ? IdentifiedClientState.toJSON(e) : undefined
      );
    } else {
      obj.clientStates = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryClientStatesResponse>
  ): QueryClientStatesResponse {
    const message = {
      ...baseQueryClientStatesResponse,
    } as QueryClientStatesResponse;
    message.clientStates = [];
    if (object.clientStates !== undefined && object.clientStates !== null) {
      for (const e of object.clientStates) {
        message.clientStates.push(IdentifiedClientState.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryConsensusStateRequest: object = {
  clientId: "",
  revisionNumber: 0,
  revisionHeight: 0,
  latestHeight: false,
};

export const QueryConsensusStateRequest = {
  encode(
    message: QueryConsensusStateRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.clientId !== "") {
      writer.uint32(10).string(message.clientId);
    }
    if (message.revisionNumber !== 0) {
      writer.uint32(16).uint64(message.revisionNumber);
    }
    if (message.revisionHeight !== 0) {
      writer.uint32(24).uint64(message.revisionHeight);
    }
    if (message.latestHeight === true) {
      writer.uint32(32).bool(message.latestHeight);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryConsensusStateRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryConsensusStateRequest,
    } as QueryConsensusStateRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.clientId = reader.string();
          break;
        case 2:
          message.revisionNumber = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.revisionHeight = longToNumber(reader.uint64() as Long);
          break;
        case 4:
          message.latestHeight = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryConsensusStateRequest {
    const message = {
      ...baseQueryConsensusStateRequest,
    } as QueryConsensusStateRequest;
    if (object.clientId !== undefined && object.clientId !== null) {
      message.clientId = String(object.clientId);
    } else {
      message.clientId = "";
    }
    if (object.revisionNumber !== undefined && object.revisionNumber !== null) {
      message.revisionNumber = Number(object.revisionNumber);
    } else {
      message.revisionNumber = 0;
    }
    if (object.revisionHeight !== undefined && object.revisionHeight !== null) {
      message.revisionHeight = Number(object.revisionHeight);
    } else {
      message.revisionHeight = 0;
    }
    if (object.latestHeight !== undefined && object.latestHeight !== null) {
      message.latestHeight = Boolean(object.latestHeight);
    } else {
      message.latestHeight = false;
    }
    return message;
  },

  toJSON(message: QueryConsensusStateRequest): unknown {
    const obj: any = {};
    message.clientId !== undefined && (obj.clientId = message.clientId);
    message.revisionNumber !== undefined &&
      (obj.revisionNumber = message.revisionNumber);
    message.revisionHeight !== undefined &&
      (obj.revisionHeight = message.revisionHeight);
    message.latestHeight !== undefined &&
      (obj.latestHeight = message.latestHeight);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryConsensusStateRequest>
  ): QueryConsensusStateRequest {
    const message = {
      ...baseQueryConsensusStateRequest,
    } as QueryConsensusStateRequest;
    if (object.clientId !== undefined && object.clientId !== null) {
      message.clientId = object.clientId;
    } else {
      message.clientId = "";
    }
    if (object.revisionNumber !== undefined && object.revisionNumber !== null) {
      message.revisionNumber = object.revisionNumber;
    } else {
      message.revisionNumber = 0;
    }
    if (object.revisionHeight !== undefined && object.revisionHeight !== null) {
      message.revisionHeight = object.revisionHeight;
    } else {
      message.revisionHeight = 0;
    }
    if (object.latestHeight !== undefined && object.latestHeight !== null) {
      message.latestHeight = object.latestHeight;
    } else {
      message.latestHeight = false;
    }
    return message;
  },
};

const baseQueryConsensusStateResponse: object = {};

export const QueryConsensusStateResponse = {
  encode(
    message: QueryConsensusStateResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.consensusState !== undefined) {
      Any.encode(message.consensusState, writer.uint32(10).fork()).ldelim();
    }
    if (message.proof.length !== 0) {
      writer.uint32(18).bytes(message.proof);
    }
    if (message.proofHeight !== undefined) {
      Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryConsensusStateResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryConsensusStateResponse,
    } as QueryConsensusStateResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.consensusState = Any.decode(reader, reader.uint32());
          break;
        case 2:
          message.proof = reader.bytes();
          break;
        case 3:
          message.proofHeight = Height.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryConsensusStateResponse {
    const message = {
      ...baseQueryConsensusStateResponse,
    } as QueryConsensusStateResponse;
    if (object.consensusState !== undefined && object.consensusState !== null) {
      message.consensusState = Any.fromJSON(object.consensusState);
    } else {
      message.consensusState = undefined;
    }
    if (object.proof !== undefined && object.proof !== null) {
      message.proof = bytesFromBase64(object.proof);
    }
    if (object.proofHeight !== undefined && object.proofHeight !== null) {
      message.proofHeight = Height.fromJSON(object.proofHeight);
    } else {
      message.proofHeight = undefined;
    }
    return message;
  },

  toJSON(message: QueryConsensusStateResponse): unknown {
    const obj: any = {};
    message.consensusState !== undefined &&
      (obj.consensusState = message.consensusState
        ? Any.toJSON(message.consensusState)
        : undefined);
    message.proof !== undefined &&
      (obj.proof = base64FromBytes(
        message.proof !== undefined ? message.proof : new Uint8Array()
      ));
    message.proofHeight !== undefined &&
      (obj.proofHeight = message.proofHeight
        ? Height.toJSON(message.proofHeight)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryConsensusStateResponse>
  ): QueryConsensusStateResponse {
    const message = {
      ...baseQueryConsensusStateResponse,
    } as QueryConsensusStateResponse;
    if (object.consensusState !== undefined && object.consensusState !== null) {
      message.consensusState = Any.fromPartial(object.consensusState);
    } else {
      message.consensusState = undefined;
    }
    if (object.proof !== undefined && object.proof !== null) {
      message.proof = object.proof;
    } else {
      message.proof = new Uint8Array();
    }
    if (object.proofHeight !== undefined && object.proofHeight !== null) {
      message.proofHeight = Height.fromPartial(object.proofHeight);
    } else {
      message.proofHeight = undefined;
    }
    return message;
  },
};

const baseQueryConsensusStatesRequest: object = { clientId: "" };

export const QueryConsensusStatesRequest = {
  encode(
    message: QueryConsensusStatesRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.clientId !== "") {
      writer.uint32(10).string(message.clientId);
    }
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryConsensusStatesRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryConsensusStatesRequest,
    } as QueryConsensusStatesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.clientId = reader.string();
          break;
        case 2:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryConsensusStatesRequest {
    const message = {
      ...baseQueryConsensusStatesRequest,
    } as QueryConsensusStatesRequest;
    if (object.clientId !== undefined && object.clientId !== null) {
      message.clientId = String(object.clientId);
    } else {
      message.clientId = "";
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryConsensusStatesRequest): unknown {
    const obj: any = {};
    message.clientId !== undefined && (obj.clientId = message.clientId);
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryConsensusStatesRequest>
  ): QueryConsensusStatesRequest {
    const message = {
      ...baseQueryConsensusStatesRequest,
    } as QueryConsensusStatesRequest;
    if (object.clientId !== undefined && object.clientId !== null) {
      message.clientId = object.clientId;
    } else {
      message.clientId = "";
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryConsensusStatesResponse: object = {};

export const QueryConsensusStatesResponse = {
  encode(
    message: QueryConsensusStatesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.consensusStates) {
      ConsensusStateWithHeight.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryConsensusStatesResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryConsensusStatesResponse,
    } as QueryConsensusStatesResponse;
    message.consensusStates = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.consensusStates.push(
            ConsensusStateWithHeight.decode(reader, reader.uint32())
          );
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryConsensusStatesResponse {
    const message = {
      ...baseQueryConsensusStatesResponse,
    } as QueryConsensusStatesResponse;
    message.consensusStates = [];
    if (
      object.consensusStates !== undefined &&
      object.consensusStates !== null
    ) {
      for (const e of object.consensusStates) {
        message.consensusStates.push(ConsensusStateWithHeight.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryConsensusStatesResponse): unknown {
    const obj: any = {};
    if (message.consensusStates) {
      obj.consensusStates = message.consensusStates.map((e) =>
        e ? ConsensusStateWithHeight.toJSON(e) : undefined
      );
    } else {
      obj.consensusStates = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryConsensusStatesResponse>
  ): QueryConsensusStatesResponse {
    const message = {
      ...baseQueryConsensusStatesResponse,
    } as QueryConsensusStatesResponse;
    message.consensusStates = [];
    if (
      object.consensusStates !== undefined &&
      object.consensusStates !== null
    ) {
      for (const e of object.consensusStates) {
        message.consensusStates.push(ConsensusStateWithHeight.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryClientParamsRequest: object = {};

export const QueryClientParamsRequest = {
  encode(
    _: QueryClientParamsRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryClientParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryClientParamsRequest,
    } as QueryClientParamsRequest;
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

  fromJSON(_: any): QueryClientParamsRequest {
    const message = {
      ...baseQueryClientParamsRequest,
    } as QueryClientParamsRequest;
    return message;
  },

  toJSON(_: QueryClientParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryClientParamsRequest>
  ): QueryClientParamsRequest {
    const message = {
      ...baseQueryClientParamsRequest,
    } as QueryClientParamsRequest;
    return message;
  },
};

const baseQueryClientParamsResponse: object = {};

export const QueryClientParamsResponse = {
  encode(
    message: QueryClientParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryClientParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryClientParamsResponse,
    } as QueryClientParamsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryClientParamsResponse {
    const message = {
      ...baseQueryClientParamsResponse,
    } as QueryClientParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryClientParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryClientParamsResponse>
  ): QueryClientParamsResponse {
    const message = {
      ...baseQueryClientParamsResponse,
    } as QueryClientParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

/** Query provides defines the gRPC querier service */
export interface Query {
  /** ClientState queries an IBC light client. */
  ClientState(
    request: QueryClientStateRequest
  ): Promise<QueryClientStateResponse>;
  /** ClientStates queries all the IBC light clients of a chain. */
  ClientStates(
    request: QueryClientStatesRequest
  ): Promise<QueryClientStatesResponse>;
  /**
   * ConsensusState queries a consensus state associated with a client state at
   * a given height.
   */
  ConsensusState(
    request: QueryConsensusStateRequest
  ): Promise<QueryConsensusStateResponse>;
  /**
   * ConsensusStates queries all the consensus state associated with a given
   * client.
   */
  ConsensusStates(
    request: QueryConsensusStatesRequest
  ): Promise<QueryConsensusStatesResponse>;
  /** ClientParams queries all parameters of the ibc client. */
  ClientParams(
    request: QueryClientParamsRequest
  ): Promise<QueryClientParamsResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  ClientState(
    request: QueryClientStateRequest
  ): Promise<QueryClientStateResponse> {
    const data = QueryClientStateRequest.encode(request).finish();
    const promise = this.rpc.request(
      "ibc.core.client.v1.Query",
      "ClientState",
      data
    );
    return promise.then((data) =>
      QueryClientStateResponse.decode(new Reader(data))
    );
  }

  ClientStates(
    request: QueryClientStatesRequest
  ): Promise<QueryClientStatesResponse> {
    const data = QueryClientStatesRequest.encode(request).finish();
    const promise = this.rpc.request(
      "ibc.core.client.v1.Query",
      "ClientStates",
      data
    );
    return promise.then((data) =>
      QueryClientStatesResponse.decode(new Reader(data))
    );
  }

  ConsensusState(
    request: QueryConsensusStateRequest
  ): Promise<QueryConsensusStateResponse> {
    const data = QueryConsensusStateRequest.encode(request).finish();
    const promise = this.rpc.request(
      "ibc.core.client.v1.Query",
      "ConsensusState",
      data
    );
    return promise.then((data) =>
      QueryConsensusStateResponse.decode(new Reader(data))
    );
  }

  ConsensusStates(
    request: QueryConsensusStatesRequest
  ): Promise<QueryConsensusStatesResponse> {
    const data = QueryConsensusStatesRequest.encode(request).finish();
    const promise = this.rpc.request(
      "ibc.core.client.v1.Query",
      "ConsensusStates",
      data
    );
    return promise.then((data) =>
      QueryConsensusStatesResponse.decode(new Reader(data))
    );
  }

  ClientParams(
    request: QueryClientParamsRequest
  ): Promise<QueryClientParamsResponse> {
    const data = QueryClientParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "ibc.core.client.v1.Query",
      "ClientParams",
      data
    );
    return promise.then((data) =>
      QueryClientParamsResponse.decode(new Reader(data))
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
