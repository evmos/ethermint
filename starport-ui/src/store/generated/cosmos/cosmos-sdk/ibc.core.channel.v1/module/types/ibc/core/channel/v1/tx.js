/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Channel, Packet } from "../../../../ibc/core/channel/v1/channel";
import { Height } from "../../../../ibc/core/client/v1/client";
export const protobufPackage = "ibc.core.channel.v1";
const baseMsgChannelOpenInit = { portId: "", signer: "" };
export const MsgChannelOpenInit = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channel !== undefined) {
            Channel.encode(message.channel, writer.uint32(18).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(26).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgChannelOpenInit };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.channel = Channel.decode(reader, reader.uint32());
                    break;
                case 3:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgChannelOpenInit };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.channel !== undefined && object.channel !== null) {
            message.channel = Channel.fromJSON(object.channel);
        }
        else {
            message.channel = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channel !== undefined &&
            (obj.channel = message.channel
                ? Channel.toJSON(message.channel)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgChannelOpenInit };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.channel !== undefined && object.channel !== null) {
            message.channel = Channel.fromPartial(object.channel);
        }
        else {
            message.channel = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgChannelOpenInitResponse = {};
export const MsgChannelOpenInitResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgChannelOpenInitResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgChannelOpenInitResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgChannelOpenInitResponse,
        };
        return message;
    },
};
const baseMsgChannelOpenTry = {
    portId: "",
    previousChannelId: "",
    counterpartyVersion: "",
    signer: "",
};
export const MsgChannelOpenTry = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.previousChannelId !== "") {
            writer.uint32(18).string(message.previousChannelId);
        }
        if (message.channel !== undefined) {
            Channel.encode(message.channel, writer.uint32(26).fork()).ldelim();
        }
        if (message.counterpartyVersion !== "") {
            writer.uint32(34).string(message.counterpartyVersion);
        }
        if (message.proofInit.length !== 0) {
            writer.uint32(42).bytes(message.proofInit);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(50).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(58).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgChannelOpenTry };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.portId = reader.string();
                    break;
                case 2:
                    message.previousChannelId = reader.string();
                    break;
                case 3:
                    message.channel = Channel.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.counterpartyVersion = reader.string();
                    break;
                case 5:
                    message.proofInit = reader.bytes();
                    break;
                case 6:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 7:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgChannelOpenTry };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = String(object.portId);
        }
        else {
            message.portId = "";
        }
        if (object.previousChannelId !== undefined &&
            object.previousChannelId !== null) {
            message.previousChannelId = String(object.previousChannelId);
        }
        else {
            message.previousChannelId = "";
        }
        if (object.channel !== undefined && object.channel !== null) {
            message.channel = Channel.fromJSON(object.channel);
        }
        else {
            message.channel = undefined;
        }
        if (object.counterpartyVersion !== undefined &&
            object.counterpartyVersion !== null) {
            message.counterpartyVersion = String(object.counterpartyVersion);
        }
        else {
            message.counterpartyVersion = "";
        }
        if (object.proofInit !== undefined && object.proofInit !== null) {
            message.proofInit = bytesFromBase64(object.proofInit);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.previousChannelId !== undefined &&
            (obj.previousChannelId = message.previousChannelId);
        message.channel !== undefined &&
            (obj.channel = message.channel
                ? Channel.toJSON(message.channel)
                : undefined);
        message.counterpartyVersion !== undefined &&
            (obj.counterpartyVersion = message.counterpartyVersion);
        message.proofInit !== undefined &&
            (obj.proofInit = base64FromBytes(message.proofInit !== undefined ? message.proofInit : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgChannelOpenTry };
        if (object.portId !== undefined && object.portId !== null) {
            message.portId = object.portId;
        }
        else {
            message.portId = "";
        }
        if (object.previousChannelId !== undefined &&
            object.previousChannelId !== null) {
            message.previousChannelId = object.previousChannelId;
        }
        else {
            message.previousChannelId = "";
        }
        if (object.channel !== undefined && object.channel !== null) {
            message.channel = Channel.fromPartial(object.channel);
        }
        else {
            message.channel = undefined;
        }
        if (object.counterpartyVersion !== undefined &&
            object.counterpartyVersion !== null) {
            message.counterpartyVersion = object.counterpartyVersion;
        }
        else {
            message.counterpartyVersion = "";
        }
        if (object.proofInit !== undefined && object.proofInit !== null) {
            message.proofInit = object.proofInit;
        }
        else {
            message.proofInit = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgChannelOpenTryResponse = {};
export const MsgChannelOpenTryResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgChannelOpenTryResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgChannelOpenTryResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgChannelOpenTryResponse,
        };
        return message;
    },
};
const baseMsgChannelOpenAck = {
    portId: "",
    channelId: "",
    counterpartyChannelId: "",
    counterpartyVersion: "",
    signer: "",
};
export const MsgChannelOpenAck = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.counterpartyChannelId !== "") {
            writer.uint32(26).string(message.counterpartyChannelId);
        }
        if (message.counterpartyVersion !== "") {
            writer.uint32(34).string(message.counterpartyVersion);
        }
        if (message.proofTry.length !== 0) {
            writer.uint32(42).bytes(message.proofTry);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(50).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(58).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgChannelOpenAck };
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
                    message.counterpartyChannelId = reader.string();
                    break;
                case 4:
                    message.counterpartyVersion = reader.string();
                    break;
                case 5:
                    message.proofTry = reader.bytes();
                    break;
                case 6:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 7:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgChannelOpenAck };
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
        if (object.counterpartyChannelId !== undefined &&
            object.counterpartyChannelId !== null) {
            message.counterpartyChannelId = String(object.counterpartyChannelId);
        }
        else {
            message.counterpartyChannelId = "";
        }
        if (object.counterpartyVersion !== undefined &&
            object.counterpartyVersion !== null) {
            message.counterpartyVersion = String(object.counterpartyVersion);
        }
        else {
            message.counterpartyVersion = "";
        }
        if (object.proofTry !== undefined && object.proofTry !== null) {
            message.proofTry = bytesFromBase64(object.proofTry);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.counterpartyChannelId !== undefined &&
            (obj.counterpartyChannelId = message.counterpartyChannelId);
        message.counterpartyVersion !== undefined &&
            (obj.counterpartyVersion = message.counterpartyVersion);
        message.proofTry !== undefined &&
            (obj.proofTry = base64FromBytes(message.proofTry !== undefined ? message.proofTry : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgChannelOpenAck };
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
        if (object.counterpartyChannelId !== undefined &&
            object.counterpartyChannelId !== null) {
            message.counterpartyChannelId = object.counterpartyChannelId;
        }
        else {
            message.counterpartyChannelId = "";
        }
        if (object.counterpartyVersion !== undefined &&
            object.counterpartyVersion !== null) {
            message.counterpartyVersion = object.counterpartyVersion;
        }
        else {
            message.counterpartyVersion = "";
        }
        if (object.proofTry !== undefined && object.proofTry !== null) {
            message.proofTry = object.proofTry;
        }
        else {
            message.proofTry = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgChannelOpenAckResponse = {};
export const MsgChannelOpenAckResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgChannelOpenAckResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgChannelOpenAckResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgChannelOpenAckResponse,
        };
        return message;
    },
};
const baseMsgChannelOpenConfirm = {
    portId: "",
    channelId: "",
    signer: "",
};
export const MsgChannelOpenConfirm = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.proofAck.length !== 0) {
            writer.uint32(26).bytes(message.proofAck);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(34).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(42).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgChannelOpenConfirm };
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
                    message.proofAck = reader.bytes();
                    break;
                case 4:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgChannelOpenConfirm };
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
        if (object.proofAck !== undefined && object.proofAck !== null) {
            message.proofAck = bytesFromBase64(object.proofAck);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.proofAck !== undefined &&
            (obj.proofAck = base64FromBytes(message.proofAck !== undefined ? message.proofAck : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgChannelOpenConfirm };
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
        if (object.proofAck !== undefined && object.proofAck !== null) {
            message.proofAck = object.proofAck;
        }
        else {
            message.proofAck = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgChannelOpenConfirmResponse = {};
export const MsgChannelOpenConfirmResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgChannelOpenConfirmResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgChannelOpenConfirmResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgChannelOpenConfirmResponse,
        };
        return message;
    },
};
const baseMsgChannelCloseInit = {
    portId: "",
    channelId: "",
    signer: "",
};
export const MsgChannelCloseInit = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.signer !== "") {
            writer.uint32(26).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgChannelCloseInit };
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
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgChannelCloseInit };
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
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgChannelCloseInit };
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
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgChannelCloseInitResponse = {};
export const MsgChannelCloseInitResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgChannelCloseInitResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgChannelCloseInitResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgChannelCloseInitResponse,
        };
        return message;
    },
};
const baseMsgChannelCloseConfirm = {
    portId: "",
    channelId: "",
    signer: "",
};
export const MsgChannelCloseConfirm = {
    encode(message, writer = Writer.create()) {
        if (message.portId !== "") {
            writer.uint32(10).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(18).string(message.channelId);
        }
        if (message.proofInit.length !== 0) {
            writer.uint32(26).bytes(message.proofInit);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(34).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(42).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgChannelCloseConfirm };
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
                    message.proofInit = reader.bytes();
                    break;
                case 4:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgChannelCloseConfirm };
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
        if (object.proofInit !== undefined && object.proofInit !== null) {
            message.proofInit = bytesFromBase64(object.proofInit);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.proofInit !== undefined &&
            (obj.proofInit = base64FromBytes(message.proofInit !== undefined ? message.proofInit : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgChannelCloseConfirm };
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
        if (object.proofInit !== undefined && object.proofInit !== null) {
            message.proofInit = object.proofInit;
        }
        else {
            message.proofInit = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgChannelCloseConfirmResponse = {};
export const MsgChannelCloseConfirmResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgChannelCloseConfirmResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgChannelCloseConfirmResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgChannelCloseConfirmResponse,
        };
        return message;
    },
};
const baseMsgRecvPacket = { signer: "" };
export const MsgRecvPacket = {
    encode(message, writer = Writer.create()) {
        if (message.packet !== undefined) {
            Packet.encode(message.packet, writer.uint32(10).fork()).ldelim();
        }
        if (message.proofCommitment.length !== 0) {
            writer.uint32(18).bytes(message.proofCommitment);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(34).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgRecvPacket };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.packet = Packet.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.proofCommitment = reader.bytes();
                    break;
                case 3:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgRecvPacket };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromJSON(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.proofCommitment !== undefined &&
            object.proofCommitment !== null) {
            message.proofCommitment = bytesFromBase64(object.proofCommitment);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.packet !== undefined &&
            (obj.packet = message.packet ? Packet.toJSON(message.packet) : undefined);
        message.proofCommitment !== undefined &&
            (obj.proofCommitment = base64FromBytes(message.proofCommitment !== undefined
                ? message.proofCommitment
                : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgRecvPacket };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromPartial(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.proofCommitment !== undefined &&
            object.proofCommitment !== null) {
            message.proofCommitment = object.proofCommitment;
        }
        else {
            message.proofCommitment = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgRecvPacketResponse = {};
export const MsgRecvPacketResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgRecvPacketResponse };
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
    fromJSON(_) {
        const message = { ...baseMsgRecvPacketResponse };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseMsgRecvPacketResponse };
        return message;
    },
};
const baseMsgTimeout = { nextSequenceRecv: 0, signer: "" };
export const MsgTimeout = {
    encode(message, writer = Writer.create()) {
        if (message.packet !== undefined) {
            Packet.encode(message.packet, writer.uint32(10).fork()).ldelim();
        }
        if (message.proofUnreceived.length !== 0) {
            writer.uint32(18).bytes(message.proofUnreceived);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(26).fork()).ldelim();
        }
        if (message.nextSequenceRecv !== 0) {
            writer.uint32(32).uint64(message.nextSequenceRecv);
        }
        if (message.signer !== "") {
            writer.uint32(42).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgTimeout };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.packet = Packet.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.proofUnreceived = reader.bytes();
                    break;
                case 3:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.nextSequenceRecv = longToNumber(reader.uint64());
                    break;
                case 5:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgTimeout };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromJSON(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.proofUnreceived !== undefined &&
            object.proofUnreceived !== null) {
            message.proofUnreceived = bytesFromBase64(object.proofUnreceived);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.nextSequenceRecv !== undefined &&
            object.nextSequenceRecv !== null) {
            message.nextSequenceRecv = Number(object.nextSequenceRecv);
        }
        else {
            message.nextSequenceRecv = 0;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.packet !== undefined &&
            (obj.packet = message.packet ? Packet.toJSON(message.packet) : undefined);
        message.proofUnreceived !== undefined &&
            (obj.proofUnreceived = base64FromBytes(message.proofUnreceived !== undefined
                ? message.proofUnreceived
                : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.nextSequenceRecv !== undefined &&
            (obj.nextSequenceRecv = message.nextSequenceRecv);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgTimeout };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromPartial(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.proofUnreceived !== undefined &&
            object.proofUnreceived !== null) {
            message.proofUnreceived = object.proofUnreceived;
        }
        else {
            message.proofUnreceived = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.nextSequenceRecv !== undefined &&
            object.nextSequenceRecv !== null) {
            message.nextSequenceRecv = object.nextSequenceRecv;
        }
        else {
            message.nextSequenceRecv = 0;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgTimeoutResponse = {};
export const MsgTimeoutResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgTimeoutResponse };
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
    fromJSON(_) {
        const message = { ...baseMsgTimeoutResponse };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseMsgTimeoutResponse };
        return message;
    },
};
const baseMsgTimeoutOnClose = { nextSequenceRecv: 0, signer: "" };
export const MsgTimeoutOnClose = {
    encode(message, writer = Writer.create()) {
        if (message.packet !== undefined) {
            Packet.encode(message.packet, writer.uint32(10).fork()).ldelim();
        }
        if (message.proofUnreceived.length !== 0) {
            writer.uint32(18).bytes(message.proofUnreceived);
        }
        if (message.proofClose.length !== 0) {
            writer.uint32(26).bytes(message.proofClose);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(34).fork()).ldelim();
        }
        if (message.nextSequenceRecv !== 0) {
            writer.uint32(40).uint64(message.nextSequenceRecv);
        }
        if (message.signer !== "") {
            writer.uint32(50).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgTimeoutOnClose };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.packet = Packet.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.proofUnreceived = reader.bytes();
                    break;
                case 3:
                    message.proofClose = reader.bytes();
                    break;
                case 4:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.nextSequenceRecv = longToNumber(reader.uint64());
                    break;
                case 6:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgTimeoutOnClose };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromJSON(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.proofUnreceived !== undefined &&
            object.proofUnreceived !== null) {
            message.proofUnreceived = bytesFromBase64(object.proofUnreceived);
        }
        if (object.proofClose !== undefined && object.proofClose !== null) {
            message.proofClose = bytesFromBase64(object.proofClose);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.nextSequenceRecv !== undefined &&
            object.nextSequenceRecv !== null) {
            message.nextSequenceRecv = Number(object.nextSequenceRecv);
        }
        else {
            message.nextSequenceRecv = 0;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.packet !== undefined &&
            (obj.packet = message.packet ? Packet.toJSON(message.packet) : undefined);
        message.proofUnreceived !== undefined &&
            (obj.proofUnreceived = base64FromBytes(message.proofUnreceived !== undefined
                ? message.proofUnreceived
                : new Uint8Array()));
        message.proofClose !== undefined &&
            (obj.proofClose = base64FromBytes(message.proofClose !== undefined ? message.proofClose : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.nextSequenceRecv !== undefined &&
            (obj.nextSequenceRecv = message.nextSequenceRecv);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgTimeoutOnClose };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromPartial(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.proofUnreceived !== undefined &&
            object.proofUnreceived !== null) {
            message.proofUnreceived = object.proofUnreceived;
        }
        else {
            message.proofUnreceived = new Uint8Array();
        }
        if (object.proofClose !== undefined && object.proofClose !== null) {
            message.proofClose = object.proofClose;
        }
        else {
            message.proofClose = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.nextSequenceRecv !== undefined &&
            object.nextSequenceRecv !== null) {
            message.nextSequenceRecv = object.nextSequenceRecv;
        }
        else {
            message.nextSequenceRecv = 0;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgTimeoutOnCloseResponse = {};
export const MsgTimeoutOnCloseResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgTimeoutOnCloseResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgTimeoutOnCloseResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgTimeoutOnCloseResponse,
        };
        return message;
    },
};
const baseMsgAcknowledgement = { signer: "" };
export const MsgAcknowledgement = {
    encode(message, writer = Writer.create()) {
        if (message.packet !== undefined) {
            Packet.encode(message.packet, writer.uint32(10).fork()).ldelim();
        }
        if (message.acknowledgement.length !== 0) {
            writer.uint32(18).bytes(message.acknowledgement);
        }
        if (message.proofAcked.length !== 0) {
            writer.uint32(26).bytes(message.proofAcked);
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(34).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(42).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgAcknowledgement };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.packet = Packet.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.acknowledgement = reader.bytes();
                    break;
                case 3:
                    message.proofAcked = reader.bytes();
                    break;
                case 4:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.signer = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgAcknowledgement };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromJSON(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.acknowledgement !== undefined &&
            object.acknowledgement !== null) {
            message.acknowledgement = bytesFromBase64(object.acknowledgement);
        }
        if (object.proofAcked !== undefined && object.proofAcked !== null) {
            message.proofAcked = bytesFromBase64(object.proofAcked);
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = String(object.signer);
        }
        else {
            message.signer = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.packet !== undefined &&
            (obj.packet = message.packet ? Packet.toJSON(message.packet) : undefined);
        message.acknowledgement !== undefined &&
            (obj.acknowledgement = base64FromBytes(message.acknowledgement !== undefined
                ? message.acknowledgement
                : new Uint8Array()));
        message.proofAcked !== undefined &&
            (obj.proofAcked = base64FromBytes(message.proofAcked !== undefined ? message.proofAcked : new Uint8Array()));
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgAcknowledgement };
        if (object.packet !== undefined && object.packet !== null) {
            message.packet = Packet.fromPartial(object.packet);
        }
        else {
            message.packet = undefined;
        }
        if (object.acknowledgement !== undefined &&
            object.acknowledgement !== null) {
            message.acknowledgement = object.acknowledgement;
        }
        else {
            message.acknowledgement = new Uint8Array();
        }
        if (object.proofAcked !== undefined && object.proofAcked !== null) {
            message.proofAcked = object.proofAcked;
        }
        else {
            message.proofAcked = new Uint8Array();
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.signer !== undefined && object.signer !== null) {
            message.signer = object.signer;
        }
        else {
            message.signer = "";
        }
        return message;
    },
};
const baseMsgAcknowledgementResponse = {};
export const MsgAcknowledgementResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgAcknowledgementResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgAcknowledgementResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgAcknowledgementResponse,
        };
        return message;
    },
};
export class MsgClientImpl {
    constructor(rpc) {
        this.rpc = rpc;
    }
    ChannelOpenInit(request) {
        const data = MsgChannelOpenInit.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "ChannelOpenInit", data);
        return promise.then((data) => MsgChannelOpenInitResponse.decode(new Reader(data)));
    }
    ChannelOpenTry(request) {
        const data = MsgChannelOpenTry.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "ChannelOpenTry", data);
        return promise.then((data) => MsgChannelOpenTryResponse.decode(new Reader(data)));
    }
    ChannelOpenAck(request) {
        const data = MsgChannelOpenAck.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "ChannelOpenAck", data);
        return promise.then((data) => MsgChannelOpenAckResponse.decode(new Reader(data)));
    }
    ChannelOpenConfirm(request) {
        const data = MsgChannelOpenConfirm.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "ChannelOpenConfirm", data);
        return promise.then((data) => MsgChannelOpenConfirmResponse.decode(new Reader(data)));
    }
    ChannelCloseInit(request) {
        const data = MsgChannelCloseInit.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "ChannelCloseInit", data);
        return promise.then((data) => MsgChannelCloseInitResponse.decode(new Reader(data)));
    }
    ChannelCloseConfirm(request) {
        const data = MsgChannelCloseConfirm.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "ChannelCloseConfirm", data);
        return promise.then((data) => MsgChannelCloseConfirmResponse.decode(new Reader(data)));
    }
    RecvPacket(request) {
        const data = MsgRecvPacket.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "RecvPacket", data);
        return promise.then((data) => MsgRecvPacketResponse.decode(new Reader(data)));
    }
    Timeout(request) {
        const data = MsgTimeout.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "Timeout", data);
        return promise.then((data) => MsgTimeoutResponse.decode(new Reader(data)));
    }
    TimeoutOnClose(request) {
        const data = MsgTimeoutOnClose.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "TimeoutOnClose", data);
        return promise.then((data) => MsgTimeoutOnCloseResponse.decode(new Reader(data)));
    }
    Acknowledgement(request) {
        const data = MsgAcknowledgement.encode(request).finish();
        const promise = this.rpc.request("ibc.core.channel.v1.Msg", "Acknowledgement", data);
        return promise.then((data) => MsgAcknowledgementResponse.decode(new Reader(data)));
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
