/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Counterparty, Version, } from "../../../../ibc/core/connection/v1/connection";
import { Any } from "../../../../google/protobuf/any";
import { Height } from "../../../../ibc/core/client/v1/client";
export const protobufPackage = "ibc.core.connection.v1";
const baseMsgConnectionOpenInit = {
    clientId: "",
    delayPeriod: 0,
    signer: "",
};
export const MsgConnectionOpenInit = {
    encode(message, writer = Writer.create()) {
        if (message.clientId !== "") {
            writer.uint32(10).string(message.clientId);
        }
        if (message.counterparty !== undefined) {
            Counterparty.encode(message.counterparty, writer.uint32(18).fork()).ldelim();
        }
        if (message.version !== undefined) {
            Version.encode(message.version, writer.uint32(26).fork()).ldelim();
        }
        if (message.delayPeriod !== 0) {
            writer.uint32(32).uint64(message.delayPeriod);
        }
        if (message.signer !== "") {
            writer.uint32(42).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgConnectionOpenInit };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.clientId = reader.string();
                    break;
                case 2:
                    message.counterparty = Counterparty.decode(reader, reader.uint32());
                    break;
                case 3:
                    message.version = Version.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.delayPeriod = longToNumber(reader.uint64());
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
        const message = { ...baseMsgConnectionOpenInit };
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = String(object.clientId);
        }
        else {
            message.clientId = "";
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromJSON(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = Version.fromJSON(object.version);
        }
        else {
            message.version = undefined;
        }
        if (object.delayPeriod !== undefined && object.delayPeriod !== null) {
            message.delayPeriod = Number(object.delayPeriod);
        }
        else {
            message.delayPeriod = 0;
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
        message.clientId !== undefined && (obj.clientId = message.clientId);
        message.counterparty !== undefined &&
            (obj.counterparty = message.counterparty
                ? Counterparty.toJSON(message.counterparty)
                : undefined);
        message.version !== undefined &&
            (obj.version = message.version
                ? Version.toJSON(message.version)
                : undefined);
        message.delayPeriod !== undefined &&
            (obj.delayPeriod = message.delayPeriod);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgConnectionOpenInit };
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = object.clientId;
        }
        else {
            message.clientId = "";
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromPartial(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = Version.fromPartial(object.version);
        }
        else {
            message.version = undefined;
        }
        if (object.delayPeriod !== undefined && object.delayPeriod !== null) {
            message.delayPeriod = object.delayPeriod;
        }
        else {
            message.delayPeriod = 0;
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
const baseMsgConnectionOpenInitResponse = {};
export const MsgConnectionOpenInitResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgConnectionOpenInitResponse,
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
            ...baseMsgConnectionOpenInitResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgConnectionOpenInitResponse,
        };
        return message;
    },
};
const baseMsgConnectionOpenTry = {
    clientId: "",
    previousConnectionId: "",
    delayPeriod: 0,
    signer: "",
};
export const MsgConnectionOpenTry = {
    encode(message, writer = Writer.create()) {
        if (message.clientId !== "") {
            writer.uint32(10).string(message.clientId);
        }
        if (message.previousConnectionId !== "") {
            writer.uint32(18).string(message.previousConnectionId);
        }
        if (message.clientState !== undefined) {
            Any.encode(message.clientState, writer.uint32(26).fork()).ldelim();
        }
        if (message.counterparty !== undefined) {
            Counterparty.encode(message.counterparty, writer.uint32(34).fork()).ldelim();
        }
        if (message.delayPeriod !== 0) {
            writer.uint32(40).uint64(message.delayPeriod);
        }
        for (const v of message.counterpartyVersions) {
            Version.encode(v, writer.uint32(50).fork()).ldelim();
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(58).fork()).ldelim();
        }
        if (message.proofInit.length !== 0) {
            writer.uint32(66).bytes(message.proofInit);
        }
        if (message.proofClient.length !== 0) {
            writer.uint32(74).bytes(message.proofClient);
        }
        if (message.proofConsensus.length !== 0) {
            writer.uint32(82).bytes(message.proofConsensus);
        }
        if (message.consensusHeight !== undefined) {
            Height.encode(message.consensusHeight, writer.uint32(90).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(98).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgConnectionOpenTry };
        message.counterpartyVersions = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.clientId = reader.string();
                    break;
                case 2:
                    message.previousConnectionId = reader.string();
                    break;
                case 3:
                    message.clientState = Any.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.counterparty = Counterparty.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.delayPeriod = longToNumber(reader.uint64());
                    break;
                case 6:
                    message.counterpartyVersions.push(Version.decode(reader, reader.uint32()));
                    break;
                case 7:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 8:
                    message.proofInit = reader.bytes();
                    break;
                case 9:
                    message.proofClient = reader.bytes();
                    break;
                case 10:
                    message.proofConsensus = reader.bytes();
                    break;
                case 11:
                    message.consensusHeight = Height.decode(reader, reader.uint32());
                    break;
                case 12:
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
        const message = { ...baseMsgConnectionOpenTry };
        message.counterpartyVersions = [];
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = String(object.clientId);
        }
        else {
            message.clientId = "";
        }
        if (object.previousConnectionId !== undefined &&
            object.previousConnectionId !== null) {
            message.previousConnectionId = String(object.previousConnectionId);
        }
        else {
            message.previousConnectionId = "";
        }
        if (object.clientState !== undefined && object.clientState !== null) {
            message.clientState = Any.fromJSON(object.clientState);
        }
        else {
            message.clientState = undefined;
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromJSON(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.delayPeriod !== undefined && object.delayPeriod !== null) {
            message.delayPeriod = Number(object.delayPeriod);
        }
        else {
            message.delayPeriod = 0;
        }
        if (object.counterpartyVersions !== undefined &&
            object.counterpartyVersions !== null) {
            for (const e of object.counterpartyVersions) {
                message.counterpartyVersions.push(Version.fromJSON(e));
            }
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.proofInit !== undefined && object.proofInit !== null) {
            message.proofInit = bytesFromBase64(object.proofInit);
        }
        if (object.proofClient !== undefined && object.proofClient !== null) {
            message.proofClient = bytesFromBase64(object.proofClient);
        }
        if (object.proofConsensus !== undefined && object.proofConsensus !== null) {
            message.proofConsensus = bytesFromBase64(object.proofConsensus);
        }
        if (object.consensusHeight !== undefined &&
            object.consensusHeight !== null) {
            message.consensusHeight = Height.fromJSON(object.consensusHeight);
        }
        else {
            message.consensusHeight = undefined;
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
        message.clientId !== undefined && (obj.clientId = message.clientId);
        message.previousConnectionId !== undefined &&
            (obj.previousConnectionId = message.previousConnectionId);
        message.clientState !== undefined &&
            (obj.clientState = message.clientState
                ? Any.toJSON(message.clientState)
                : undefined);
        message.counterparty !== undefined &&
            (obj.counterparty = message.counterparty
                ? Counterparty.toJSON(message.counterparty)
                : undefined);
        message.delayPeriod !== undefined &&
            (obj.delayPeriod = message.delayPeriod);
        if (message.counterpartyVersions) {
            obj.counterpartyVersions = message.counterpartyVersions.map((e) => e ? Version.toJSON(e) : undefined);
        }
        else {
            obj.counterpartyVersions = [];
        }
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.proofInit !== undefined &&
            (obj.proofInit = base64FromBytes(message.proofInit !== undefined ? message.proofInit : new Uint8Array()));
        message.proofClient !== undefined &&
            (obj.proofClient = base64FromBytes(message.proofClient !== undefined
                ? message.proofClient
                : new Uint8Array()));
        message.proofConsensus !== undefined &&
            (obj.proofConsensus = base64FromBytes(message.proofConsensus !== undefined
                ? message.proofConsensus
                : new Uint8Array()));
        message.consensusHeight !== undefined &&
            (obj.consensusHeight = message.consensusHeight
                ? Height.toJSON(message.consensusHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgConnectionOpenTry };
        message.counterpartyVersions = [];
        if (object.clientId !== undefined && object.clientId !== null) {
            message.clientId = object.clientId;
        }
        else {
            message.clientId = "";
        }
        if (object.previousConnectionId !== undefined &&
            object.previousConnectionId !== null) {
            message.previousConnectionId = object.previousConnectionId;
        }
        else {
            message.previousConnectionId = "";
        }
        if (object.clientState !== undefined && object.clientState !== null) {
            message.clientState = Any.fromPartial(object.clientState);
        }
        else {
            message.clientState = undefined;
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromPartial(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.delayPeriod !== undefined && object.delayPeriod !== null) {
            message.delayPeriod = object.delayPeriod;
        }
        else {
            message.delayPeriod = 0;
        }
        if (object.counterpartyVersions !== undefined &&
            object.counterpartyVersions !== null) {
            for (const e of object.counterpartyVersions) {
                message.counterpartyVersions.push(Version.fromPartial(e));
            }
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.proofInit !== undefined && object.proofInit !== null) {
            message.proofInit = object.proofInit;
        }
        else {
            message.proofInit = new Uint8Array();
        }
        if (object.proofClient !== undefined && object.proofClient !== null) {
            message.proofClient = object.proofClient;
        }
        else {
            message.proofClient = new Uint8Array();
        }
        if (object.proofConsensus !== undefined && object.proofConsensus !== null) {
            message.proofConsensus = object.proofConsensus;
        }
        else {
            message.proofConsensus = new Uint8Array();
        }
        if (object.consensusHeight !== undefined &&
            object.consensusHeight !== null) {
            message.consensusHeight = Height.fromPartial(object.consensusHeight);
        }
        else {
            message.consensusHeight = undefined;
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
const baseMsgConnectionOpenTryResponse = {};
export const MsgConnectionOpenTryResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgConnectionOpenTryResponse,
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
            ...baseMsgConnectionOpenTryResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgConnectionOpenTryResponse,
        };
        return message;
    },
};
const baseMsgConnectionOpenAck = {
    connectionId: "",
    counterpartyConnectionId: "",
    signer: "",
};
export const MsgConnectionOpenAck = {
    encode(message, writer = Writer.create()) {
        if (message.connectionId !== "") {
            writer.uint32(10).string(message.connectionId);
        }
        if (message.counterpartyConnectionId !== "") {
            writer.uint32(18).string(message.counterpartyConnectionId);
        }
        if (message.version !== undefined) {
            Version.encode(message.version, writer.uint32(26).fork()).ldelim();
        }
        if (message.clientState !== undefined) {
            Any.encode(message.clientState, writer.uint32(34).fork()).ldelim();
        }
        if (message.proofHeight !== undefined) {
            Height.encode(message.proofHeight, writer.uint32(42).fork()).ldelim();
        }
        if (message.proofTry.length !== 0) {
            writer.uint32(50).bytes(message.proofTry);
        }
        if (message.proofClient.length !== 0) {
            writer.uint32(58).bytes(message.proofClient);
        }
        if (message.proofConsensus.length !== 0) {
            writer.uint32(66).bytes(message.proofConsensus);
        }
        if (message.consensusHeight !== undefined) {
            Height.encode(message.consensusHeight, writer.uint32(74).fork()).ldelim();
        }
        if (message.signer !== "") {
            writer.uint32(82).string(message.signer);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgConnectionOpenAck };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connectionId = reader.string();
                    break;
                case 2:
                    message.counterpartyConnectionId = reader.string();
                    break;
                case 3:
                    message.version = Version.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.clientState = Any.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.proofHeight = Height.decode(reader, reader.uint32());
                    break;
                case 6:
                    message.proofTry = reader.bytes();
                    break;
                case 7:
                    message.proofClient = reader.bytes();
                    break;
                case 8:
                    message.proofConsensus = reader.bytes();
                    break;
                case 9:
                    message.consensusHeight = Height.decode(reader, reader.uint32());
                    break;
                case 10:
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
        const message = { ...baseMsgConnectionOpenAck };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = String(object.connectionId);
        }
        else {
            message.connectionId = "";
        }
        if (object.counterpartyConnectionId !== undefined &&
            object.counterpartyConnectionId !== null) {
            message.counterpartyConnectionId = String(object.counterpartyConnectionId);
        }
        else {
            message.counterpartyConnectionId = "";
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = Version.fromJSON(object.version);
        }
        else {
            message.version = undefined;
        }
        if (object.clientState !== undefined && object.clientState !== null) {
            message.clientState = Any.fromJSON(object.clientState);
        }
        else {
            message.clientState = undefined;
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromJSON(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.proofTry !== undefined && object.proofTry !== null) {
            message.proofTry = bytesFromBase64(object.proofTry);
        }
        if (object.proofClient !== undefined && object.proofClient !== null) {
            message.proofClient = bytesFromBase64(object.proofClient);
        }
        if (object.proofConsensus !== undefined && object.proofConsensus !== null) {
            message.proofConsensus = bytesFromBase64(object.proofConsensus);
        }
        if (object.consensusHeight !== undefined &&
            object.consensusHeight !== null) {
            message.consensusHeight = Height.fromJSON(object.consensusHeight);
        }
        else {
            message.consensusHeight = undefined;
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
        message.connectionId !== undefined &&
            (obj.connectionId = message.connectionId);
        message.counterpartyConnectionId !== undefined &&
            (obj.counterpartyConnectionId = message.counterpartyConnectionId);
        message.version !== undefined &&
            (obj.version = message.version
                ? Version.toJSON(message.version)
                : undefined);
        message.clientState !== undefined &&
            (obj.clientState = message.clientState
                ? Any.toJSON(message.clientState)
                : undefined);
        message.proofHeight !== undefined &&
            (obj.proofHeight = message.proofHeight
                ? Height.toJSON(message.proofHeight)
                : undefined);
        message.proofTry !== undefined &&
            (obj.proofTry = base64FromBytes(message.proofTry !== undefined ? message.proofTry : new Uint8Array()));
        message.proofClient !== undefined &&
            (obj.proofClient = base64FromBytes(message.proofClient !== undefined
                ? message.proofClient
                : new Uint8Array()));
        message.proofConsensus !== undefined &&
            (obj.proofConsensus = base64FromBytes(message.proofConsensus !== undefined
                ? message.proofConsensus
                : new Uint8Array()));
        message.consensusHeight !== undefined &&
            (obj.consensusHeight = message.consensusHeight
                ? Height.toJSON(message.consensusHeight)
                : undefined);
        message.signer !== undefined && (obj.signer = message.signer);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgConnectionOpenAck };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = object.connectionId;
        }
        else {
            message.connectionId = "";
        }
        if (object.counterpartyConnectionId !== undefined &&
            object.counterpartyConnectionId !== null) {
            message.counterpartyConnectionId = object.counterpartyConnectionId;
        }
        else {
            message.counterpartyConnectionId = "";
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = Version.fromPartial(object.version);
        }
        else {
            message.version = undefined;
        }
        if (object.clientState !== undefined && object.clientState !== null) {
            message.clientState = Any.fromPartial(object.clientState);
        }
        else {
            message.clientState = undefined;
        }
        if (object.proofHeight !== undefined && object.proofHeight !== null) {
            message.proofHeight = Height.fromPartial(object.proofHeight);
        }
        else {
            message.proofHeight = undefined;
        }
        if (object.proofTry !== undefined && object.proofTry !== null) {
            message.proofTry = object.proofTry;
        }
        else {
            message.proofTry = new Uint8Array();
        }
        if (object.proofClient !== undefined && object.proofClient !== null) {
            message.proofClient = object.proofClient;
        }
        else {
            message.proofClient = new Uint8Array();
        }
        if (object.proofConsensus !== undefined && object.proofConsensus !== null) {
            message.proofConsensus = object.proofConsensus;
        }
        else {
            message.proofConsensus = new Uint8Array();
        }
        if (object.consensusHeight !== undefined &&
            object.consensusHeight !== null) {
            message.consensusHeight = Height.fromPartial(object.consensusHeight);
        }
        else {
            message.consensusHeight = undefined;
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
const baseMsgConnectionOpenAckResponse = {};
export const MsgConnectionOpenAckResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgConnectionOpenAckResponse,
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
            ...baseMsgConnectionOpenAckResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgConnectionOpenAckResponse,
        };
        return message;
    },
};
const baseMsgConnectionOpenConfirm = { connectionId: "", signer: "" };
export const MsgConnectionOpenConfirm = {
    encode(message, writer = Writer.create()) {
        if (message.connectionId !== "") {
            writer.uint32(10).string(message.connectionId);
        }
        if (message.proofAck.length !== 0) {
            writer.uint32(18).bytes(message.proofAck);
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
        const message = {
            ...baseMsgConnectionOpenConfirm,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.connectionId = reader.string();
                    break;
                case 2:
                    message.proofAck = reader.bytes();
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
        const message = {
            ...baseMsgConnectionOpenConfirm,
        };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = String(object.connectionId);
        }
        else {
            message.connectionId = "";
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
        message.connectionId !== undefined &&
            (obj.connectionId = message.connectionId);
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
        const message = {
            ...baseMsgConnectionOpenConfirm,
        };
        if (object.connectionId !== undefined && object.connectionId !== null) {
            message.connectionId = object.connectionId;
        }
        else {
            message.connectionId = "";
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
const baseMsgConnectionOpenConfirmResponse = {};
export const MsgConnectionOpenConfirmResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgConnectionOpenConfirmResponse,
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
            ...baseMsgConnectionOpenConfirmResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgConnectionOpenConfirmResponse,
        };
        return message;
    },
};
export class MsgClientImpl {
    constructor(rpc) {
        this.rpc = rpc;
    }
    ConnectionOpenInit(request) {
        const data = MsgConnectionOpenInit.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Msg", "ConnectionOpenInit", data);
        return promise.then((data) => MsgConnectionOpenInitResponse.decode(new Reader(data)));
    }
    ConnectionOpenTry(request) {
        const data = MsgConnectionOpenTry.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Msg", "ConnectionOpenTry", data);
        return promise.then((data) => MsgConnectionOpenTryResponse.decode(new Reader(data)));
    }
    ConnectionOpenAck(request) {
        const data = MsgConnectionOpenAck.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Msg", "ConnectionOpenAck", data);
        return promise.then((data) => MsgConnectionOpenAckResponse.decode(new Reader(data)));
    }
    ConnectionOpenConfirm(request) {
        const data = MsgConnectionOpenConfirm.encode(request).finish();
        const promise = this.rpc.request("ibc.core.connection.v1.Msg", "ConnectionOpenConfirm", data);
        return promise.then((data) => MsgConnectionOpenConfirmResponse.decode(new Reader(data)));
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
