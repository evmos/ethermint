/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { IdentifiedChannel, PacketState, } from "../../../../ibc/core/channel/v1/channel";
export const protobufPackage = "ibc.core.channel.v1";
const baseGenesisState = { nextChannelSequence: 0 };
export const GenesisState = {
    encode(message, writer = Writer.create()) {
        for (const v of message.channels) {
            IdentifiedChannel.encode(v, writer.uint32(10).fork()).ldelim();
        }
        for (const v of message.acknowledgements) {
            PacketState.encode(v, writer.uint32(18).fork()).ldelim();
        }
        for (const v of message.commitments) {
            PacketState.encode(v, writer.uint32(26).fork()).ldelim();
        }
        for (const v of message.receipts) {
            PacketState.encode(v, writer.uint32(34).fork()).ldelim();
        }
        for (const v of message.sendSequences) {
            PacketSequence.encode(v, writer.uint32(42).fork()).ldelim();
        }
        for (const v of message.recvSequences) {
            PacketSequence.encode(v, writer.uint32(50).fork()).ldelim();
        }
        for (const v of message.ackSequences) {
            PacketSequence.encode(v, writer.uint32(58).fork()).ldelim();
        }
        if (message.nextChannelSequence !== 0) {
            writer.uint32(64).uint64(message.nextChannelSequence);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseGenesisState };
        message.channels = [];
        message.acknowledgements = [];
        message.commitments = [];
        message.receipts = [];
        message.sendSequences = [];
        message.recvSequences = [];
        message.ackSequences = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.channels.push(IdentifiedChannel.decode(reader, reader.uint32()));
                    break;
                case 2:
                    message.acknowledgements.push(PacketState.decode(reader, reader.uint32()));
                    break;
                case 3:
                    message.commitments.push(PacketState.decode(reader, reader.uint32()));
                    break;
                case 4:
                    message.receipts.push(PacketState.decode(reader, reader.uint32()));
                    break;
                case 5:
                    message.sendSequences.push(PacketSequence.decode(reader, reader.uint32()));
                    break;
                case 6:
                    message.recvSequences.push(PacketSequence.decode(reader, reader.uint32()));
                    break;
                case 7:
                    message.ackSequences.push(PacketSequence.decode(reader, reader.uint32()));
                    break;
                case 8:
                    message.nextChannelSequence = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseGenesisState };
        message.channels = [];
        message.acknowledgements = [];
        message.commitments = [];
        message.receipts = [];
        message.sendSequences = [];
        message.recvSequences = [];
        message.ackSequences = [];
        if (object.channels !== undefined && object.channels !== null) {
            for (const e of object.channels) {
                message.channels.push(IdentifiedChannel.fromJSON(e));
            }
        }
        if (object.acknowledgements !== undefined &&
            object.acknowledgements !== null) {
            for (const e of object.acknowledgements) {
                message.acknowledgements.push(PacketState.fromJSON(e));
            }
        }
        if (object.commitments !== undefined && object.commitments !== null) {
            for (const e of object.commitments) {
                message.commitments.push(PacketState.fromJSON(e));
            }
        }
        if (object.receipts !== undefined && object.receipts !== null) {
            for (const e of object.receipts) {
                message.receipts.push(PacketState.fromJSON(e));
            }
        }
        if (object.sendSequences !== undefined && object.sendSequences !== null) {
            for (const e of object.sendSequences) {
                message.sendSequences.push(PacketSequence.fromJSON(e));
            }
        }
        if (object.recvSequences !== undefined && object.recvSequences !== null) {
            for (const e of object.recvSequences) {
                message.recvSequences.push(PacketSequence.fromJSON(e));
            }
        }
        if (object.ackSequences !== undefined && object.ackSequences !== null) {
            for (const e of object.ackSequences) {
                message.ackSequences.push(PacketSequence.fromJSON(e));
            }
        }
        if (object.nextChannelSequence !== undefined &&
            object.nextChannelSequence !== null) {
            message.nextChannelSequence = Number(object.nextChannelSequence);
        }
        else {
            message.nextChannelSequence = 0;
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
        if (message.acknowledgements) {
            obj.acknowledgements = message.acknowledgements.map((e) => e ? PacketState.toJSON(e) : undefined);
        }
        else {
            obj.acknowledgements = [];
        }
        if (message.commitments) {
            obj.commitments = message.commitments.map((e) => e ? PacketState.toJSON(e) : undefined);
        }
        else {
            obj.commitments = [];
        }
        if (message.receipts) {
            obj.receipts = message.receipts.map((e) => e ? PacketState.toJSON(e) : undefined);
        }
        else {
            obj.receipts = [];
        }
        if (message.sendSequences) {
            obj.sendSequences = message.sendSequences.map((e) => e ? PacketSequence.toJSON(e) : undefined);
        }
        else {
            obj.sendSequences = [];
        }
        if (message.recvSequences) {
            obj.recvSequences = message.recvSequences.map((e) => e ? PacketSequence.toJSON(e) : undefined);
        }
        else {
            obj.recvSequences = [];
        }
        if (message.ackSequences) {
            obj.ackSequences = message.ackSequences.map((e) => e ? PacketSequence.toJSON(e) : undefined);
        }
        else {
            obj.ackSequences = [];
        }
        message.nextChannelSequence !== undefined &&
            (obj.nextChannelSequence = message.nextChannelSequence);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseGenesisState };
        message.channels = [];
        message.acknowledgements = [];
        message.commitments = [];
        message.receipts = [];
        message.sendSequences = [];
        message.recvSequences = [];
        message.ackSequences = [];
        if (object.channels !== undefined && object.channels !== null) {
            for (const e of object.channels) {
                message.channels.push(IdentifiedChannel.fromPartial(e));
            }
        }
        if (object.acknowledgements !== undefined &&
            object.acknowledgements !== null) {
            for (const e of object.acknowledgements) {
                message.acknowledgements.push(PacketState.fromPartial(e));
            }
        }
        if (object.commitments !== undefined && object.commitments !== null) {
            for (const e of object.commitments) {
                message.commitments.push(PacketState.fromPartial(e));
            }
        }
        if (object.receipts !== undefined && object.receipts !== null) {
            for (const e of object.receipts) {
                message.receipts.push(PacketState.fromPartial(e));
            }
        }
        if (object.sendSequences !== undefined && object.sendSequences !== null) {
            for (const e of object.sendSequences) {
                message.sendSequences.push(PacketSequence.fromPartial(e));
            }
        }
        if (object.recvSequences !== undefined && object.recvSequences !== null) {
            for (const e of object.recvSequences) {
                message.recvSequences.push(PacketSequence.fromPartial(e));
            }
        }
        if (object.ackSequences !== undefined && object.ackSequences !== null) {
            for (const e of object.ackSequences) {
                message.ackSequences.push(PacketSequence.fromPartial(e));
            }
        }
        if (object.nextChannelSequence !== undefined &&
            object.nextChannelSequence !== null) {
            message.nextChannelSequence = object.nextChannelSequence;
        }
        else {
            message.nextChannelSequence = 0;
        }
        return message;
    },
};
const basePacketSequence = { portId: "", channelId: "", sequence: 0 };
export const PacketSequence = {
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
        const message = { ...basePacketSequence };
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
        const message = { ...basePacketSequence };
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
        const message = { ...basePacketSequence };
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
