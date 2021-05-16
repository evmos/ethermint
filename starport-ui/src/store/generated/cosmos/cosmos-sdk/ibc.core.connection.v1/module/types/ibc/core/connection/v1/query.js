/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { ConnectionEnd, IdentifiedConnection, } from "../../../../ibc/core/connection/v1/connection";
import { Height, IdentifiedClientState, } from "../../../../ibc/core/client/v1/client";
import { PageRequest, PageResponse, } from "../../../../cosmos/base/query/v1beta1/pagination";
import { Any } from "../../../../google/protobuf/any";
export const protobufPackage = "ibc.core.connection.v1";
const baseQueryConnectionRequest = { connectionId: "" };
export const QueryConnectionRequest = {
    encode(message, writer = Writer.create()) {
        if (message.connectionId !== "") {
            writer.uint32(10).string(message.connectionId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryConnectionRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connectionId = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryConnectionRequest };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = String(object.connectionId);
        }
        else {
            message.connectionId = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.connectionId !== undefined &&
            (obj.connectionId = message.connectionId);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryConnectionRequest };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = object.connectionId;
        }
        else {
            message.connectionId = "";
        }
        return message;
    },
};
const baseQueryConnectionResponse = {};
export const QueryConnectionResponse = {
    encode(message, writer = Writer.create()) {
        if (message.connection !== undefined) {
            ConnectionEnd.encode(message.connection, writer.uint32(10).fork()).ldelim();
        }
        if (message.proof.length !== 0) {
            writer.uint32(18).bytes(message.proof);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connection = ConnectionEnd.decode(reader, reader.uint32());
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
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionResponse,
        };
        if (object.connection !== undefined && object.connection !== null) {
            message.connection = ConnectionEnd.fromJSON(object.connection);
        }
        else {
            message.connection = undefined;
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = bytesFromBase64(object.proof);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.connection !== undefined &&
            (obj.connection = message.connection
                ? ConnectionEnd.toJSON(message.connection)
                : undefined);
        message.proof !== undefined &&
            (obj.proof = base64FromBytes(message.proof !== undefined ? message.proof : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionResponse,
        };
        if (object.connection !== undefined && object.connection !== null) {
            message.connection = ConnectionEnd.fromPartial(object.connection);
        }
        else {
            message.connection = undefined;
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = object.proof;
        }
        else {
            message.proof = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
};
const baseQueryConnectionsRequest = {};
export const QueryConnectionsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionsRequest,
        };
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
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionsRequest,
        };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionsRequest,
        };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryConnectionsResponse = {};
export const QueryConnectionsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.connections) {
            IdentifiedConnection.encode(v, writer.uint32(10).fork()).ldelim();
        }
        if (message.pagination !== undefined) {
            PageResponse.encode(message.pagination, writer.uint32(18).fork()).ldelim();
        }
        if (message.height !== undefined) {
            Height.encode(message.height, writer.uint32(26).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionsResponse,
        };
        message.connections = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connections.push(IdentifiedConnection.decode(reader, reader.uint32()));
                    break;
                case 2:
                    message.pagination = PageResponse.decode(reader, reader.uint32());
                    break;
                case 3:
                    message.height = Height.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionsResponse,
        };
        message.connections = [];
        if (object.connections !== undefined && object.connections !== null) {
            for (const e of object.connections) {
                message.connections.push(IdentifiedConnection.fromJSON(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        if (object.height !== undefined && object.height !== null) {
            message.height = Height.fromJSON(object.height);
        }
        else {
            message.height = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.connections) {
            obj.connections = message.connections.map((e) => e ? IdentifiedConnection.toJSON(e) : undefined);
        }
        else {
            obj.connections = [];
        }
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageResponse.toJSON(message.pagination)
                : undefined);
        message.height !== undefined &&
            (obj.height = message.height ? Height.toJSON(message.height) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionsResponse,
        };
        message.connections = [];
        if (object.connections !== undefined && object.connections !== null) {
            for (const e of object.connections) {
                message.connections.push(IdentifiedConnection.fromPartial(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        if (object.height !== undefined && object.height !== null) {
            message.height = Height.fromPartial(object.height);
        }
        else {
            message.height = undefined;
        }
        return message;
    },
};
const baseQueryClientConnectionsRequest = { clientId: "" };
export const QueryClientConnectionsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.clientId !== "") {
            writer.uint32(10).string(message.clientId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryClientConnectionsRequest,
        };
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
    fromJSON(object) {
        const message = {
            ...baseQueryClientConnectionsRequest,
        };
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = String(object.clientId);
        }
        else {
            message.clientId = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.clientId !== undefined && (obj.clientId = message.clientId);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryClientConnectionsRequest,
        };
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = object.clientId;
        }
        else {
            message.clientId = "";
        }
        return message;
    },
};
const baseQueryClientConnectionsResponse = { connectionPaths: "" };
export const QueryClientConnectionsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.connectionPaths) {
            writer.uint32(10).string(v);
        }
        if (message.proof.length !== 0) {
            writer.uint32(18).bytes(message.proof);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryClientConnectionsResponse,
        };
        message.connectionPaths = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connectionPaths.push(reader.string());
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
    fromJSON(object) {
        const message = {
            ...baseQueryClientConnectionsResponse,
        };
        message.connectionPaths = [];
        if (object.connectionPaths !== undefined &&
            object.connectionPaths !== null) {
            for (const e of object.connectionPaths) {
                message.connectionPaths.push(String(e));
            }
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = bytesFromBase64(object.proof);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.connectionPaths) {
            obj.connectionPaths = message.connectionPaths.map((e) => e);
        }
        else {
            obj.connectionPaths = [];
        }
        message.proof !== undefined &&
            (obj.proof = base64FromBytes(message.proof !== undefined ? message.proof : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryClientConnectionsResponse,
        };
        message.connectionPaths = [];
        if (object.connectionPaths !== undefined &&
            object.connectionPaths !== null) {
            for (const e of object.connectionPaths) {
                message.connectionPaths.push(e);
            }
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = object.proof;
        }
        else {
            message.proof = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
};
const baseQueryConnectionClientStateRequest = { connectionId: "" };
export const QueryConnectionClientStateRequest = {
    encode(message, writer = Writer.create()) {
        if (message.connectionId !== "") {
            writer.uint32(10).string(message.connectionId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionClientStateRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connectionId = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionClientStateRequest,
        };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = String(object.connectionId);
        }
        else {
            message.connectionId = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.connectionId !== undefined &&
            (obj.connectionId = message.connectionId);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionClientStateRequest,
        };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = object.connectionId;
        }
        else {
            message.connectionId = "";
        }
        return message;
    },
};
const baseQueryConnectionClientStateResponse = {};
export const QueryConnectionClientStateResponse = {
    encode(message, writer = Writer.create()) {
        if (message.identifiedClientState !== undefined) {
            IdentifiedClientState.encode(message.identifiedClientState, writer.uint32(10).fork()).ldelim();
        }
        if (message.proof.length !== 0) {
            writer.uint32(18).bytes(message.proof);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionClientStateResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.identifiedClientState = IdentifiedClientState.decode(reader, reader.uint32());
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
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionClientStateResponse,
        };
        if (object.identifiedClientState !== undefined &&
            object.identifiedClientState !== null) {
            message.identifiedClientState = IdentifiedClientState.fromJSON(object.identifiedClientState);
        }
        else {
            message.identifiedClientState = undefined;
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = bytesFromBase64(object.proof);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.identifiedClientState !== undefined &&
            (obj.identifiedClientState = message.identifiedClientState
                ? IdentifiedClientState.toJSON(message.identifiedClientState)
                : undefined);
        message.proof !== undefined &&
            (obj.proof = base64FromBytes(message.proof !== undefined ? message.proof : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionClientStateResponse,
        };
        if (object.identifiedClientState !== undefined &&
            object.identifiedClientState !== null) {
            message.identifiedClientState = IdentifiedClientState.fromPartial(object.identifiedClientState);
        }
        else {
            message.identifiedClientState = undefined;
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = object.proof;
        }
        else {
            message.proof = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
};
const baseQueryConnectionConsensusStateRequest = {
    connectionId: "",
    revisionNumber: 0,
    revisionHeight: 0,
};
export const QueryConnectionConsensusStateRequest = {
    encode(message, writer = Writer.create()) {
        if (message.connectionId !== "") {
            writer.uint32(10).string(message.connectionId);
        }
        if (message.revisionNumber !== 0) {
            writer.uint32(16).uint64(message.revisionNumber);
        }
        if (message.revisionHeight !== 0) {
            writer.uint32(24).uint64(message.revisionHeight);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionConsensusStateRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connectionId = reader.string();
                    break;
                case 2:
                    message.revisionNumber = longToNumber(reader.uint64());
                    break;
                case 3:
                    message.revisionHeight = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionConsensusStateRequest,
        };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = String(object.connectionId);
        }
        else {
            message.connectionId = "";
        }
        if (object.revisionNumber !== undefined && object.revisionNumber !== null) {
            message.revisionNumber = Number(object.revisionNumber);
        }
        else {
            message.revisionNumber = 0;
        }
        if (object.revisionHeight !== undefined && object.revisionHeight !== null) {
            message.revisionHeight = Number(object.revisionHeight);
        }
        else {
            message.revisionHeight = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.connectionId !== undefined &&
            (obj.connectionId = message.connectionId);
        message.revisionNumber !== undefined &&
            (obj.revisionNumber = message.revisionNumber);
        message.revisionHeight !== undefined &&
            (obj.revisionHeight = message.revisionHeight);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionConsensusStateRequest,
        };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = object.connectionId;
        }
        else {
            message.connectionId = "";
        }
        if (object.revisionNumber !== undefined && object.revisionNumber !== null) {
            message.revisionNumber = object.revisionNumber;
        }
        else {
            message.revisionNumber = 0;
        }
        if (object.revisionHeight !== undefined && object.revisionHeight !== null) {
            message.revisionHeight = object.revisionHeight;
        }
        else {
            message.revisionHeight = 0;
        }
        return message;
    },
};
const baseQueryConnectionConsensusStateResponse = { clientId: "" };
export const QueryConnectionConsensusStateResponse = {
    encode(message, writer = Writer.create()) {
        if (message.consensusState !== undefined) {
            Any.encode(message.consensusState, writer.uint32(10).fork()).ldelim();
        }
        if (message.clientId !== "") {
            writer.uint32(18).string(message.clientId);
        }
        if (message.proof.length !== 0) {
            writer.uint32(26).bytes(message.proof);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(34).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionConsensusStateResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.consensusState = Any.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.clientId = reader.string();
                    break;
                case 3:
                    message.proof = reader.bytes();
                    break;
                case 4:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionConsensusStateResponse,
        };
        if (object.consensusState !== undefined && object.consensusState !== null) {
            message.consensusState = Any.fromJSON(object.consensusState);
        }
        else {
            message.consensusState = undefined;
        }
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = String(object.clientId);
        }
        else {
            message.clientId = "";
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = bytesFromBase64(object.proof);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.consensusState !== undefined &&
            (obj.consensusState = message.consensusState
                ? Any.toJSON(message.consensusState)
                : undefined);
        message.clientId !== undefined && (obj.clientId = message.clientId);
        message.proof !== undefined &&
            (obj.proof = base64FromBytes(message.proof !== undefined ? message.proof : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionConsensusStateResponse,
        };
        if (object.consensusState !== undefined && object.consensusState !== null) {
            message.consensusState = Any.fromPartial(object.consensusState);
        }
        else {
            message.consensusState = undefined;
        }
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = object.clientId;
        }
        else {
            message.clientId = "";
        }
        if (object.proof !== undefined && object.proof !== null) {
            message.proof = object.proof;
        }
        else {
            message.proof = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        return message;
    },
};
export class QueryClientImpl {
    constructor(rpc) {
        this.rpc = rpc;
    }
    Connection(request) {
        const data = QueryConnectionRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Query", "Connection", data);
        return promise.then((data) => QueryConnectionResponse.decode(new Reader(data)));
    }
    Connections(request) {
        const data = QueryConnectionsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Query", "Connections", data);
        return promise.then((data) => QueryConnectionsResponse.decode(new Reader(data)));
    }
    ClientConnections(request) {
        const data = QueryClientConnectionsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Query", "ClientConnections", data);
        return promise.then((data) => QueryClientConnectionsResponse.decode(new Reader(data)));
    }
    ConnectionClientState(request) {
        const data = QueryConnectionClientStateRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Query", "ConnectionClientState", data);
        return promise.then((data) => QueryConnectionClientStateResponse.decode(new Reader(data)));
    }
    ConnectionConsensusState(request) {
        const data = QueryConnectionConsensusStateRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Query", "ConnectionConsensusState", data);
        return promise.then((data) => QueryConnectionConsensusStateResponse.decode(new Reader(data)));
    }
}
var globalThis = (() => {
    if (typeof globalThis !== "undefined")
        return globalThis;
    if (typeof self !== "undefined")
        return self;
    if (typeof window !== "undefined")
        return window;
    if (typeof global !== "undefined")
        return global;
    throw "Unable to locate global object";
})();
const atob = globalThis.atob ||
    ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64) {
    const bin = atob(b64);
    const arr = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; ++i) {
        arr[i] = bin.charCodeAt(i);
    }
    return arr;
}
const btoa = globalThis.btoa ||
    ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr) {
    const bin = [];
    for (let i = 0; i < arr.byteLength; ++i) {
        bin.push(String.fromCharCode(arr[i]));
    }
    return btoa(bin.join(""));
}
function longToNumber(long) {
    if (long.gt(Number.MAX_SAFE_INTEGER)) {
        throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
    }
    return long.toNumber();
}
if (util.Long !== Long) {
    util.Long = Long;
    configure();
}
