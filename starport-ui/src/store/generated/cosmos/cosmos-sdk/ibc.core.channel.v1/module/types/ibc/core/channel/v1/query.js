/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Channel, IdentifiedChannel, PacketState, } from "../../../../ibc/core/channel/v1/channel";
import { Height, IdentifiedClientState, } from "../../../../ibc/core/client/v1/client";
import { PageRequest, PageResponse, } from "../../../../cosmos/base/query/v1beta1/pagination";
import { Any } from "../../../../google/protobuf/any";
export const protobufPackage = "ibc.core.channel.v1";
const baseQueryChannelRequest = { portId: "", channelId: "" };
export const QueryChannelRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryChannelRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryChannelRequest };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryChannelRequest };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        return message;
    },
};
const baseQueryChannelResponse = {};
export const QueryChannelResponse = {
    encode(message, writer = Writer.create()) {
        if (message.channel !== undefined) {
            Channel.encode(message.channel, writer.uint32(10).fork()).ldelim();
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
        const message = { ...baseQueryChannelResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.channel = Channel.decode(reader, reader.uint32());
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
        const message = { ...baseQueryChannelResponse };
        if (object.channel !== undefined && object.channel !== null) {
            message.channel = Channel.fromJSON(object.channel);
        }
        else {
            message.channel = undefined;
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
        message.channel !== undefined &&
            (obj.channel = message.channel
                ? Channel.toJSON(message.channel)
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
        const message = { ...baseQueryChannelResponse };
        if (object.channel !== undefined && object.channel !== null) {
            message.channel = Channel.fromPartial(object.channel);
        }
        else {
            message.channel = undefined;
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
const baseQueryChannelsRequest = {};
export const QueryChannelsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryChannelsRequest };
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
        const message = { ...baseQueryChannelsRequest };
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
        const message = { ...baseQueryChannelsRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryChannelsResponse = {};
export const QueryChannelsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.channels) {
            IdentifiedChannel.encode(v, writer.uint32(10).fork()).ldelim();
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
        const message = { ...baseQueryChannelsResponse };
        message.channels = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.channels.push(IdentifiedChannel.decode(reader, reader.uint32()));
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
        const message = { ...baseQueryChannelsResponse };
        message.channels = [];
        if (object.channels !== undefined && object.channels !== null) {
            for (const e of object.channels) {
                message.channels.push(IdentifiedChannel.fromJSON(e));
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
        if (message.channels) {
            obj.channels = message.channels.map((e) => e ? IdentifiedChannel.toJSON(e) : undefined);
        }
        else {
            obj.channels = [];
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
        const message = { ...baseQueryChannelsResponse };
        message.channels = [];
        if (object.channels !== undefined && object.channels !== null) {
            for (const e of object.channels) {
                message.channels.push(IdentifiedChannel.fromPartial(e));
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
const baseQueryConnectionChannelsRequest = { connection: "" };
export const QueryConnectionChannelsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.connection !== "") {
            writer.uint32(10).string(message.connection);
        }
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryConnectionChannelsRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connection = reader.string();
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
    fromJSON(object) {
        const message = {
            ...baseQueryConnectionChannelsRequest,
        };
        if (object.connection !== undefined && object.connection !== null) {
            message.connection = String(object.connection);
        }
        else {
            message.connection = "";
        }
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
        message.connection !== undefined && (obj.connection = message.connection);
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryConnectionChannelsRequest,
        };
        if (object.connection !== undefined && object.connection !== null) {
            message.connection = object.connection;
        }
        else {
            message.connection = "";
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryConnectionChannelsResponse = {};
export const QueryConnectionChannelsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.channels) {
            IdentifiedChannel.encode(v, writer.uint32(10).fork()).ldelim();
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
            ...baseQueryConnectionChannelsResponse,
        };
        message.channels = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.channels.push(IdentifiedChannel.decode(reader, reader.uint32()));
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
            ...baseQueryConnectionChannelsResponse,
        };
        message.channels = [];
        if (object.channels !== undefined && object.channels !== null) {
            for (const e of object.channels) {
                message.channels.push(IdentifiedChannel.fromJSON(e));
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
        if (message.channels) {
            obj.channels = message.channels.map((e) => e ? IdentifiedChannel.toJSON(e) : undefined);
        }
        else {
            obj.channels = [];
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
            ...baseQueryConnectionChannelsResponse,
        };
        message.channels = [];
        if (object.channels !== undefined && object.channels !== null) {
            for (const e of object.channels) {
                message.channels.push(IdentifiedChannel.fromPartial(e));
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
const baseQueryChannelClientStateRequest = {
    portId: "",
    channelId: "",
};
export const QueryChannelClientStateRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryChannelClientStateRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
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
            ...baseQueryChannelClientStateRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryChannelClientStateRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        return message;
    },
};
const baseQueryChannelClientStateResponse = {};
export const QueryChannelClientStateResponse = {
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
            ...baseQueryChannelClientStateResponse,
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
            ...baseQueryChannelClientStateResponse,
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
            ...baseQueryChannelClientStateResponse,
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
const baseQueryChannelConsensusStateRequest = {
    portId: "",
    channelId: "",
    revisionNumber: 0,
    revisionHeight: 0,
};
export const QueryChannelConsensusStateRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.revisionNumber !== 0) {
            writer.uint32(24).uint64(message.revisionNumber);
        }
        if (message.revisionHeight !== 0) {
            writer.uint32(32).uint64(message.revisionHeight);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryChannelConsensusStateRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
                    message.revisionNumber = longToNumber(reader.uint64());
                    break;
                case 4:
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
            ...baseQueryChannelConsensusStateRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
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
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.revisionNumber !== undefined &&
            (obj.revisionNumber = message.revisionNumber);
        message.revisionHeight !== undefined &&
            (obj.revisionHeight = message.revisionHeight);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryChannelConsensusStateRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
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
const baseQueryChannelConsensusStateResponse = { clientId: "" };
export const QueryChannelConsensusStateResponse = {
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
            ...baseQueryChannelConsensusStateResponse,
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
            ...baseQueryChannelConsensusStateResponse,
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
            ...baseQueryChannelConsensusStateResponse,
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
const baseQueryPacketCommitmentRequest = {
    portId: "",
    channelId: "",
    sequence: 0,
};
export const QueryPacketCommitmentRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.sequence !== 0) {
            writer.uint32(24).uint64(message.sequence);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryPacketCommitmentRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
                    message.sequence = longToNumber(reader.uint64());
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
            ...baseQueryPacketCommitmentRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = Number(object.sequence);
        }
        else {
            message.sequence = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.sequence !== undefined && (obj.sequence = message.sequence);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryPacketCommitmentRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = object.sequence;
        }
        else {
            message.sequence = 0;
        }
        return message;
    },
};
const baseQueryPacketCommitmentResponse = {};
export const QueryPacketCommitmentResponse = {
    encode(message, writer = Writer.create()) {
        if (message.commitment.length !== 0) {
            writer.uint32(10).bytes(message.commitment);
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
            ...baseQueryPacketCommitmentResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.commitment = reader.bytes();
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
            ...baseQueryPacketCommitmentResponse,
        };
        if (object.commitment !== undefined && object.commitment !== null) {
            message.commitment = bytesFromBase64(object.commitment);
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
        message.commitment !== undefined &&
            (obj.commitment = base64FromBytes(message.commitment !== undefined ? message.commitment : new Uint8Array()));
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
            ...baseQueryPacketCommitmentResponse,
        };
        if (object.commitment !== undefined && object.commitment !== null) {
            message.commitment = object.commitment;
        }
        else {
            message.commitment = new Uint8Array();
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
const baseQueryPacketCommitmentsRequest = { portId: "", channelId: "" };
export const QueryPacketCommitmentsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(26).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryPacketCommitmentsRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
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
            ...baseQueryPacketCommitmentsRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
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
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryPacketCommitmentsRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryPacketCommitmentsResponse = {};
export const QueryPacketCommitmentsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.commitments) {
            PacketState.encode(v, writer.uint32(10).fork()).ldelim();
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
            ...baseQueryPacketCommitmentsResponse,
        };
        message.commitments = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.commitments.push(PacketState.decode(reader, reader.uint32()));
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
            ...baseQueryPacketCommitmentsResponse,
        };
        message.commitments = [];
        if (object.commitments !== undefined && object.commitments !== null) {
            for (const e of object.commitments) {
                message.commitments.push(PacketState.fromJSON(e));
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
        if (message.commitments) {
            obj.commitments = message.commitments.map((e) => e ? PacketState.toJSON(e) : undefined);
        }
        else {
            obj.commitments = [];
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
            ...baseQueryPacketCommitmentsResponse,
        };
        message.commitments = [];
        if (object.commitments !== undefined && object.commitments !== null) {
            for (const e of object.commitments) {
                message.commitments.push(PacketState.fromPartial(e));
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
const baseQueryPacketReceiptRequest = {
    portId: "",
    channelId: "",
    sequence: 0,
};
export const QueryPacketReceiptRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.sequence !== 0) {
            writer.uint32(24).uint64(message.sequence);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryPacketReceiptRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
                    message.sequence = longToNumber(reader.uint64());
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
            ...baseQueryPacketReceiptRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = Number(object.sequence);
        }
        else {
            message.sequence = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.sequence !== undefined && (obj.sequence = message.sequence);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryPacketReceiptRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = object.sequence;
        }
        else {
            message.sequence = 0;
        }
        return message;
    },
};
const baseQueryPacketReceiptResponse = { received: false };
export const QueryPacketReceiptResponse = {
    encode(message, writer = Writer.create()) {
        if (message.received === true) {
            writer.uint32(16).bool(message.received);
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
            ...baseQueryPacketReceiptResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 2:
                    message.received = reader.bool();
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
            ...baseQueryPacketReceiptResponse,
        };
        if (object.received !== undefined && object.received !== null) {
            message.received = Boolean(object.received);
        }
        else {
            message.received = false;
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
        message.received !== undefined && (obj.received = message.received);
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
            ...baseQueryPacketReceiptResponse,
        };
        if (object.received !== undefined && object.received !== null) {
            message.received = object.received;
        }
        else {
            message.received = false;
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
const baseQueryPacketAcknowledgementRequest = {
    portId: "",
    channelId: "",
    sequence: 0,
};
export const QueryPacketAcknowledgementRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.sequence !== 0) {
            writer.uint32(24).uint64(message.sequence);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryPacketAcknowledgementRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
                    message.sequence = longToNumber(reader.uint64());
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
            ...baseQueryPacketAcknowledgementRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = Number(object.sequence);
        }
        else {
            message.sequence = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.sequence !== undefined && (obj.sequence = message.sequence);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryPacketAcknowledgementRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = object.sequence;
        }
        else {
            message.sequence = 0;
        }
        return message;
    },
};
const baseQueryPacketAcknowledgementResponse = {};
export const QueryPacketAcknowledgementResponse = {
    encode(message, writer = Writer.create()) {
        if (message.acknowledgement.length !== 0) {
            writer.uint32(10).bytes(message.acknowledgement);
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
            ...baseQueryPacketAcknowledgementResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.acknowledgement = reader.bytes();
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
            ...baseQueryPacketAcknowledgementResponse,
        };
        if (object.acknowledgement !== undefined &&
            object.acknowledgement !== null) {
            message.acknowledgement = bytesFromBase64(object.acknowledgement);
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
        message.acknowledgement !== undefined &&
            (obj.acknowledgement = base64FromBytes(message.acknowledgement !== undefined
                ? message.acknowledgement
                : new Uint8Array()));
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
            ...baseQueryPacketAcknowledgementResponse,
        };
        if (object.acknowledgement !== undefined &&
            object.acknowledgement !== null) {
            message.acknowledgement = object.acknowledgement;
        }
        else {
            message.acknowledgement = new Uint8Array();
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
const baseQueryPacketAcknowledgementsRequest = {
    portId: "",
    channelId: "",
};
export const QueryPacketAcknowledgementsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(26).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryPacketAcknowledgementsRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
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
            ...baseQueryPacketAcknowledgementsRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
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
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryPacketAcknowledgementsRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryPacketAcknowledgementsResponse = {};
export const QueryPacketAcknowledgementsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.acknowledgements) {
            PacketState.encode(v, writer.uint32(10).fork()).ldelim();
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
            ...baseQueryPacketAcknowledgementsResponse,
        };
        message.acknowledgements = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.acknowledgements.push(PacketState.decode(reader, reader.uint32()));
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
            ...baseQueryPacketAcknowledgementsResponse,
        };
        message.acknowledgements = [];
        if (object.acknowledgements !== undefined &&
            object.acknowledgements !== null) {
            for (const e of object.acknowledgements) {
                message.acknowledgements.push(PacketState.fromJSON(e));
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
        if (message.acknowledgements) {
            obj.acknowledgements = message.acknowledgements.map((e) => e ? PacketState.toJSON(e) : undefined);
        }
        else {
            obj.acknowledgements = [];
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
            ...baseQueryPacketAcknowledgementsResponse,
        };
        message.acknowledgements = [];
        if (object.acknowledgements !== undefined &&
            object.acknowledgements !== null) {
            for (const e of object.acknowledgements) {
                message.acknowledgements.push(PacketState.fromPartial(e));
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
const baseQueryUnreceivedPacketsRequest = {
    portId: "",
    channelId: "",
    packetCommitmentSequences: 0,
};
export const QueryUnreceivedPacketsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        writer.uint32(26).fork();
        for (const v of message.packetCommitmentSequences) {
            writer.uint64(v);
        }
        writer.ldelim();
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryUnreceivedPacketsRequest,
        };
        message.packetCommitmentSequences = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
                    if ((tag & 7) === 2) {
                        const end2 = reader.uint32() + reader.pos;
                        while (reader.pos < end2) {
                            message.packetCommitmentSequences.push(longToNumber(reader.uint64()));
                        }
                    }
                    else {
                        message.packetCommitmentSequences.push(longToNumber(reader.uint64()));
                    }
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
            ...baseQueryUnreceivedPacketsRequest,
        };
        message.packetCommitmentSequences = [];
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        if (object.packetCommitmentSequences !== undefined &&
            object.packetCommitmentSequences !== null) {
            for (const e of object.packetCommitmentSequences) {
                message.packetCommitmentSequences.push(Number(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        if (message.packetCommitmentSequences) {
            obj.packetCommitmentSequences = message.packetCommitmentSequences.map((e) => e);
        }
        else {
            obj.packetCommitmentSequences = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryUnreceivedPacketsRequest,
        };
        message.packetCommitmentSequences = [];
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.packetCommitmentSequences !== undefined &&
            object.packetCommitmentSequences !== null) {
            for (const e of object.packetCommitmentSequences) {
                message.packetCommitmentSequences.push(e);
            }
        }
        return message;
    },
};
const baseQueryUnreceivedPacketsResponse = { sequences: 0 };
export const QueryUnreceivedPacketsResponse = {
    encode(message, writer = Writer.create()) {
        writer.uint32(10).fork();
        for (const v of message.sequences) {
            writer.uint64(v);
        }
        writer.ldelim();
        if (message.height !== undefined) {
            Height.encode(message.height, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryUnreceivedPacketsResponse,
        };
        message.sequences = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    if ((tag & 7) === 2) {
                        const end2 = reader.uint32() + reader.pos;
                        while (reader.pos < end2) {
                            message.sequences.push(longToNumber(reader.uint64()));
                        }
                    }
                    else {
                        message.sequences.push(longToNumber(reader.uint64()));
                    }
                    break;
                case 2:
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
            ...baseQueryUnreceivedPacketsResponse,
        };
        message.sequences = [];
        if (object.sequences !== undefined && object.sequences !== null) {
            for (const e of object.sequences) {
                message.sequences.push(Number(e));
            }
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
        if (message.sequences) {
            obj.sequences = message.sequences.map((e) => e);
        }
        else {
            obj.sequences = [];
        }
        message.height !== undefined &&
            (obj.height = message.height ? Height.toJSON(message.height) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryUnreceivedPacketsResponse,
        };
        message.sequences = [];
        if (object.sequences !== undefined && object.sequences !== null) {
            for (const e of object.sequences) {
                message.sequences.push(e);
            }
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
const baseQueryUnreceivedAcksRequest = {
    portId: "",
    channelId: "",
    packetAckSequences: 0,
};
export const QueryUnreceivedAcksRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        writer.uint32(26).fork();
        for (const v of message.packetAckSequences) {
            writer.uint64(v);
        }
        writer.ldelim();
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryUnreceivedAcksRequest,
        };
        message.packetAckSequences = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
                    break;
                case 3:
                    if ((tag & 7) === 2) {
                        const end2 = reader.uint32() + reader.pos;
                        while (reader.pos < end2) {
                            message.packetAckSequences.push(longToNumber(reader.uint64()));
                        }
                    }
                    else {
                        message.packetAckSequences.push(longToNumber(reader.uint64()));
                    }
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
            ...baseQueryUnreceivedAcksRequest,
        };
        message.packetAckSequences = [];
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        if (object.packetAckSequences !== undefined &&
            object.packetAckSequences !== null) {
            for (const e of object.packetAckSequences) {
                message.packetAckSequences.push(Number(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        if (message.packetAckSequences) {
            obj.packetAckSequences = message.packetAckSequences.map((e) => e);
        }
        else {
            obj.packetAckSequences = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryUnreceivedAcksRequest,
        };
        message.packetAckSequences = [];
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        if (object.packetAckSequences !== undefined &&
            object.packetAckSequences !== null) {
            for (const e of object.packetAckSequences) {
                message.packetAckSequences.push(e);
            }
        }
        return message;
    },
};
const baseQueryUnreceivedAcksResponse = { sequences: 0 };
export const QueryUnreceivedAcksResponse = {
    encode(message, writer = Writer.create()) {
        writer.uint32(10).fork();
        for (const v of message.sequences) {
            writer.uint64(v);
        }
        writer.ldelim();
        if (message.height !== undefined) {
            Height.encode(message.height, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryUnreceivedAcksResponse,
        };
        message.sequences = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    if ((tag & 7) === 2) {
                        const end2 = reader.uint32() + reader.pos;
                        while (reader.pos < end2) {
                            message.sequences.push(longToNumber(reader.uint64()));
                        }
                    }
                    else {
                        message.sequences.push(longToNumber(reader.uint64()));
                    }
                    break;
                case 2:
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
            ...baseQueryUnreceivedAcksResponse,
        };
        message.sequences = [];
        if (object.sequences !== undefined && object.sequences !== null) {
            for (const e of object.sequences) {
                message.sequences.push(Number(e));
            }
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
        if (message.sequences) {
            obj.sequences = message.sequences.map((e) => e);
        }
        else {
            obj.sequences = [];
        }
        message.height !== undefined &&
            (obj.height = message.height ? Height.toJSON(message.height) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryUnreceivedAcksResponse,
        };
        message.sequences = [];
        if (object.sequences !== undefined && object.sequences !== null) {
            for (const e of object.sequences) {
                message.sequences.push(e);
            }
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
const baseQueryNextSequenceReceiveRequest = {
    portId: "",
    channelId: "",
};
export const QueryNextSequenceReceiveRequest = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryNextSequenceReceiveRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channelId = reader.string();
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
            ...baseQueryNextSequenceReceiveRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = String(object.channelId);
        }
        else {
            message.channelId = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryNextSequenceReceiveRequest,
        };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channelId !== undefined && object.channelId !== null) {
            message.channelId = object.channelId;
        }
        else {
            message.channelId = "";
        }
        return message;
    },
};
const baseQueryNextSequenceReceiveResponse = { nextSequenceReceive: 0 };
export const QueryNextSequenceReceiveResponse = {
    encode(message, writer = Writer.create()) {
        if (message.nextSequenceReceive !== 0) {
            writer.uint32(8).uint64(message.nextSequenceReceive);
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
            ...baseQueryNextSequenceReceiveResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.nextSequenceReceive = longToNumber(reader.uint64());
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
            ...baseQueryNextSequenceReceiveResponse,
        };
        if (object.nextSequenceReceive !== undefined &&
            object.nextSequenceReceive !== null) {
            message.nextSequenceReceive = Number(object.nextSequenceReceive);
        }
        else {
            message.nextSequenceReceive = 0;
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
        message.nextSequenceReceive !== undefined &&
            (obj.nextSequenceReceive = message.nextSequenceReceive);
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
            ...baseQueryNextSequenceReceiveResponse,
        };
        if (object.nextSequenceReceive !== undefined &&
            object.nextSequenceReceive !== null) {
            message.nextSequenceReceive = object.nextSequenceReceive;
        }
        else {
            message.nextSequenceReceive = 0;
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
    Channel(request) {
        const data = QueryChannelRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "Channel", data);
        return promise.then((data) => QueryChannelResponse.decode(new Reader(data)));
    }
    Channels(request) {
        const data = QueryChannelsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "Channels", data);
        return promise.then((data) => QueryChannelsResponse.decode(new Reader(data)));
    }
    ConnectionChannels(request) {
        const data = QueryConnectionChannelsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "ConnectionChannels", data);
        return promise.then((data) => QueryConnectionChannelsResponse.decode(new Reader(data)));
    }
    ChannelClientState(request) {
        const data = QueryChannelClientStateRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "ChannelClientState", data);
        return promise.then((data) => QueryChannelClientStateResponse.decode(new Reader(data)));
    }
    ChannelConsensusState(request) {
        const data = QueryChannelConsensusStateRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "ChannelConsensusState", data);
        return promise.then((data) => QueryChannelConsensusStateResponse.decode(new Reader(data)));
    }
    PacketCommitment(request) {
        const data = QueryPacketCommitmentRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "PacketCommitment", data);
        return promise.then((data) => QueryPacketCommitmentResponse.decode(new Reader(data)));
    }
    PacketCommitments(request) {
        const data = QueryPacketCommitmentsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "PacketCommitments", data);
        return promise.then((data) => QueryPacketCommitmentsResponse.decode(new Reader(data)));
    }
    PacketReceipt(request) {
        const data = QueryPacketReceiptRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "PacketReceipt", data);
        return promise.then((data) => QueryPacketReceiptResponse.decode(new Reader(data)));
    }
    PacketAcknowledgement(request) {
        const data = QueryPacketAcknowledgementRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "PacketAcknowledgement", data);
        return promise.then((data) => QueryPacketAcknowledgementResponse.decode(new Reader(data)));
    }
    PacketAcknowledgements(request) {
        const data = QueryPacketAcknowledgementsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "PacketAcknowledgements", data);
        return promise.then((data) => QueryPacketAcknowledgementsResponse.decode(new Reader(data)));
    }
    UnreceivedPackets(request) {
        const data = QueryUnreceivedPacketsRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "UnreceivedPackets", data);
        return promise.then((data) => QueryUnreceivedPacketsResponse.decode(new Reader(data)));
    }
    UnreceivedAcks(request) {
        const data = QueryUnreceivedAcksRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "UnreceivedAcks", data);
        return promise.then((data) => QueryUnreceivedAcksResponse.decode(new Reader(data)));
    }
    NextSequenceReceive(request) {
        const data = QueryNextSequenceReceiveRequest.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Query", "NextSequenceReceive", data);
        return promise.then((data) => QueryNextSequenceReceiveResponse.decode(new Reader(data)));
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
