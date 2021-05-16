/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Height } from "../../../../ibc/core/client/v1/client";
export const protobufPackage = "ibc.core.channel.v1";
/**
 * State defines if a channel is in one of the following states:
 * CLOSED, INIT, TRYOPEN, OPEN or UNINITIALIZED.
 */
export var State;
(function (State) {
    /** STATE_UNINITIALIZED_UNSPECIFIED - Default State */
    State[State["STATE_UNINITIALIZED_UNSPECIFIED"] = 0] = "STATE_UNINITIALIZED_UNSPECIFIED";
    /** STATE_INIT - A channel has just started the opening handshake. */
    State[State["STATE_INIT"] = 1] = "STATE_INIT";
    /** STATE_TRYOPEN - A channel has acknowledged the handshake step on the counterparty chain. */
    State[State["STATE_TRYOPEN"] = 2] = "STATE_TRYOPEN";
    /**
     * STATE_OPEN - A channel has completed the handshake. Open channels are
     * ready to send and receive packets.
     */
    State[State["STATE_OPEN"] = 3] = "STATE_OPEN";
    /**
     * STATE_CLOSED - A channel has been closed and can no longer be used to send or receive
     * packets.
     */
    State[State["STATE_CLOSED"] = 4] = "STATE_CLOSED";
    State[State["UNRECOGNIZED"] = -1] = "UNRECOGNIZED";
})(State || (State = {}));
export function stateFromJSON(object) {
    switch (object) {
        case 0:
        case "STATE_UNINITIALIZED_UNSPECIFIED":
            return State.STATE_UNINITIALIZED_UNSPECIFIED;
        case 1:
        case "STATE_INIT":
            return State.STATE_INIT;
        case 2:
        case "STATE_TRYOPEN":
            return State.STATE_TRYOPEN;
        case 3:
        case "STATE_OPEN":
            return State.STATE_OPEN;
        case 4:
        case "STATE_CLOSED":
            return State.STATE_CLOSED;
        case -1:
        case "UNRECOGNIZED":
        default:
            return State.UNRECOGNIZED;
    }
}
export function stateToJSON(object) {
    switch (object) {
        case State.STATE_UNINITIALIZED_UNSPECIFIED:
            return "STATE_UNINITIALIZED_UNSPECIFIED";
        case State.STATE_INIT:
            return "STATE_INIT";
        case State.STATE_TRYOPEN:
            return "STATE_TRYOPEN";
        case State.STATE_OPEN:
            return "STATE_OPEN";
        case State.STATE_CLOSED:
            return "STATE_CLOSED";
        default:
            return "UNKNOWN";
    }
}
/** Order defines if a channel is ORDERED or UNORDERED */
export var Order;
(function (Order) {
    /** ORDER_NONE_UNSPECIFIED - zero-value for channel ordering */
    Order[Order["ORDER_NONE_UNSPECIFIED"] = 0] = "ORDER_NONE_UNSPECIFIED";
    /**
     * ORDER_UNORDERED - packets can be delivered in any order, which may differ from the order in
     * which they were sent.
     */
    Order[Order["ORDER_UNORDERED"] = 1] = "ORDER_UNORDERED";
    /** ORDER_ORDERED - packets are delivered exactly in the order which they were sent */
    Order[Order["ORDER_ORDERED"] = 2] = "ORDER_ORDERED";
    Order[Order["UNRECOGNIZED"] = -1] = "UNRECOGNIZED";
})(Order || (Order = {}));
export function orderFromJSON(object) {
    switch (object) {
        case 0:
        case "ORDER_NONE_UNSPECIFIED":
            return Order.ORDER_NONE_UNSPECIFIED;
        case 1:
        case "ORDER_UNORDERED":
            return Order.ORDER_UNORDERED;
        case 2:
        case "ORDER_ORDERED":
            return Order.ORDER_ORDERED;
        case -1:
        case "UNRECOGNIZED":
        default:
            return Order.UNRECOGNIZED;
    }
}
export function orderToJSON(object) {
    switch (object) {
        case Order.ORDER_NONE_UNSPECIFIED:
            return "ORDER_NONE_UNSPECIFIED";
        case Order.ORDER_UNORDERED:
            return "ORDER_UNORDERED";
        case Order.ORDER_ORDERED:
            return "ORDER_ORDERED";
        default:
            return "UNKNOWN";
    }
}
const baseChannel = {
    state: 0,
    ordering: 0,
    connectionHops: "",
    version: "",
};
export const Channel = {
    encode(message, writer = Writer.create()) {
        if (message.state !== 0) {
            writer.uint32(8).int32(message.state);
        }
        if (message.ordering !== 0) {
            writer.uint32(16).int32(message.ordering);
        }
        if (message.counterparty !== undefined) {
            Counterparty.encode(message.counterparty, writer.uint32(26).fork()).ldelim();
        }
        for (const v of message.connectionHops) {
            writer.uint32(34).string(v);
        }
        if (message.version !== "") {
            writer.uint32(42).string(message.version);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseChannel };
        message.connectionHops = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.state = reader.int32();
                    break;
                case 2:
                    message.ordering = reader.int32();
                    break;
                case 3:
                    message.counterparty = Counterparty.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.connectionHops.push(reader.string());
                    break;
                case 5:
                    message.version = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseChannel };
        message.connectionHops = [];
        if (object.state !== undefined && object.state !== null) {
            message.state = stateFromJSON(object.state);
        }
        else {
            message.state = 0;
        }
        if (object.ordering !== undefined && object.ordering !== null) {
            message.ordering = orderFromJSON(object.ordering);
        }
        else {
            message.ordering = 0;
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromJSON(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.connectionHops !== undefined && object.connectionHops !== null) {
            for (const e of object.connectionHops) {
                message.connectionHops.push(String(e));
            }
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = String(object.version);
        }
        else {
            message.version = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.state !== undefined && (obj.state = stateToJSON(message.state));
        message.ordering !== undefined &&
            (obj.ordering = orderToJSON(message.ordering));
        message.counterparty !== undefined &&
            (obj.counterparty = message.counterparty
                ? Counterparty.toJSON(message.counterparty)
                : undefined);
        if (message.connectionHops) {
            obj.connectionHops = message.connectionHops.map((e) => e);
        }
        else {
            obj.connectionHops = [];
        }
        message.version !== undefined && (obj.version = message.version);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseChannel };
        message.connectionHops = [];
        if (object.state !== undefined && object.state !== null) {
            message.state = object.state;
        }
        else {
            message.state = 0;
        }
        if (object.ordering !== undefined && object.ordering !== null) {
            message.ordering = object.ordering;
        }
        else {
            message.ordering = 0;
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromPartial(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.connectionHops !== undefined && object.connectionHops !== null) {
            for (const e of object.connectionHops) {
                message.connectionHops.push(e);
            }
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = object.version;
        }
        else {
            message.version = "";
        }
        return message;
    },
};
const baseIdentifiedChannel = {
    state: 0,
    ordering: 0,
    connectionHops: "",
    version: "",
    portId: "",
    channelId: "",
};
export const IdentifiedChannel = {
    encode(message, writer = Writer.create()) {
        if (message.state !== 0) {
            writer.uint32(8).int32(message.state);
        }
        if (message.ordering !== 0) {
            writer.uint32(16).int32(message.ordering);
        }
        if (message.counterparty !== undefined) {
            Counterparty.encode(message.counterparty, writer.uint32(26).fork()).ldelim();
        }
        for (const v of message.connectionHops) {
            writer.uint32(34).string(v);
        }
        if (message.version !== "") {
            writer.uint32(42).string(message.version);
        }
        if (message.portId !== "") {
            writer.uint32(50).string(message.portId);
        }
        if (message.channelId !== "") {
            writer.uint32(58).string(message.channelId);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseIdentifiedChannel };
        message.connectionHops = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.state = reader.int32();
                    break;
                case 2:
                    message.ordering = reader.int32();
                    break;
                case 3:
                    message.counterparty = Counterparty.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.connectionHops.push(reader.string());
                    break;
                case 5:
                    message.version = reader.string();
                    break;
                case 6:
                    message.portId = reader.string();
                    break;
                case 7:
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
        const message = { ...baseIdentifiedChannel };
        message.connectionHops = [];
        if (object.state !== undefined && object.state !== null) {
            message.state = stateFromJSON(object.state);
        }
        else {
            message.state = 0;
        }
        if (object.ordering !== undefined && object.ordering !== null) {
            message.ordering = orderFromJSON(object.ordering);
        }
        else {
            message.ordering = 0;
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromJSON(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.connectionHops !== undefined && object.connectionHops !== null) {
            for (const e of object.connectionHops) {
                message.connectionHops.push(String(e));
            }
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = String(object.version);
        }
        else {
            message.version = "";
        }
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
        message.state !== undefined && (obj.state = stateToJSON(message.state));
        message.ordering !== undefined &&
            (obj.ordering = orderToJSON(message.ordering));
        message.counterparty !== undefined &&
            (obj.counterparty = message.counterparty
                ? Counterparty.toJSON(message.counterparty)
                : undefined);
        if (message.connectionHops) {
            obj.connectionHops = message.connectionHops.map((e) => e);
        }
        else {
            obj.connectionHops = [];
        }
        message.version !== undefined && (obj.version = message.version);
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseIdentifiedChannel };
        message.connectionHops = [];
        if (object.state !== undefined && object.state !== null) {
            message.state = object.state;
        }
        else {
            message.state = 0;
        }
        if (object.ordering !== undefined && object.ordering !== null) {
            message.ordering = object.ordering;
        }
        else {
            message.ordering = 0;
        }
        if (object.counterparty !== undefined && object.counterparty !== null) {
            message.counterparty = Counterparty.fromPartial(object.counterparty);
        }
        else {
            message.counterparty = undefined;
        }
        if (object.connectionHops !== undefined && object.connectionHops !== null) {
            for (const e of object.connectionHops) {
                message.connectionHops.push(e);
            }
        }
        if (object.version !== undefined && object.version !== null) {
            message.version = object.version;
        }
        else {
            message.version = "";
        }
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
const baseCounterparty = { portId: "", channelId: "" };
export const Counterparty = {
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
        const message = { ...baseCounterparty };
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
        const message = { ...baseCounterparty };
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
        const message = { ...baseCounterparty };
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
const basePacket = {
    sequence: 0,
    sourcePort: "",
    sourceChannel: "",
    destinationPort: "",
    destinationChannel: "",
    timeoutTimestamp: 0,
};
export const Packet = {
    encode(message, writer = Writer.create()) {
        if (message.sequence !== 0) {
            writer.uint32(8).uint64(message.sequence);
        }
        if (message.sourcePort !== "") {
            writer.uint32(18).string(message.sourcePort);
        }
        if (message.sourceChannel !== "") {
            writer.uint32(26).string(message.sourceChannel);
        }
        if (message.destinationPort !== "") {
            writer.uint32(34).string(message.destinationPort);
        }
        if (message.destinationChannel !== "") {
            writer.uint32(42).string(message.destinationChannel);
        }
        if (message.data.length !== 0) {
            writer.uint32(50).bytes(message.data);
        }
        if (message.timeoutHeight !== undefined) {
            Height.encode(message.timeoutHeight, writer.uint32(58).fork()).ldelim();
        }
        if (message.timeoutTimestamp !== 0) {
            writer.uint32(64).uint64(message.timeoutTimestamp);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...basePacket };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.sequence = longToNumber(reader.uint64());
                    break;
                case 2:
                    message.sourcePort = reader.string();
                    break;
                case 3:
                    message.sourceChannel = reader.string();
                    break;
                case 4:
                    message.destinationPort = reader.string();
                    break;
                case 5:
                    message.destinationChannel = reader.string();
                    break;
                case 6:
                    message.data = reader.bytes();
                    break;
                case 7:
                    message.timeoutHeight = Height.decode(reader, reader.uint32());
                    break;
                case 8:
                    message.timeoutTimestamp = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...basePacket };
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = Number(object.sequence);
        }
        else {
            message.sequence = 0;
        }
        if (object.sourcePort !== undefined && object.sourcePort !== null) {
            message.sourcePort = String(object.sourcePort);
        }
        else {
            message.sourcePort = "";
        }
        if (object.sourceChannel !== undefined && object.sourceChannel !== null) {
            message.sourceChannel = String(object.sourceChannel);
        }
        else {
            message.sourceChannel = "";
        }
        if (object.destinationPort !== undefined &&
            object.destinationPort !== null) {
            message.destinationPort = String(object.destinationPort);
        }
        else {
            message.destinationPort = "";
        }
        if (object.destinationChannel !== undefined &&
            object.destinationChannel !== null) {
            message.destinationChannel = String(object.destinationChannel);
        }
        else {
            message.destinationChannel = "";
        }
        if (object.data !== undefined && object.data !== null) {
            message.data = bytesFromBase64(object.data);
        }
        if (object.timeoutHeight !== undefined && object.timeoutHeight !== null) {
            message.timeoutHeight = Height.fromJSON(object.timeoutHeight);
        }
        else {
            message.timeoutHeight = undefined;
        }
        if (object.timeoutTimestamp !== undefined &&
            object.timeoutTimestamp !== null) {
            message.timeoutTimestamp = Number(object.timeoutTimestamp);
        }
        else {
            message.timeoutTimestamp = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.sequence !== undefined && (obj.sequence = message.sequence);
        message.sourcePort !== undefined && (obj.sourcePort = message.sourcePort);
        message.sourceChannel !== undefined &&
            (obj.sourceChannel = message.sourceChannel);
        message.destinationPort !== undefined &&
            (obj.destinationPort = message.destinationPort);
        message.destinationChannel !== undefined &&
            (obj.destinationChannel = message.destinationChannel);
        message.data !== undefined &&
            (obj.data = base64FromBytes(message.data !== undefined ? message.data : new Uint8Array()));
        message.timeoutHeight !== undefined &&
            (obj.timeoutHeight = message.timeoutHeight
                ? Height.toJSON(message.timeoutHeight)
                : undefined);
        message.timeoutTimestamp !== undefined &&
            (obj.timeoutTimestamp = message.timeoutTimestamp);
        return obj;
    },
    fromPartial(object) {
        const message = { ...basePacket };
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = object.sequence;
        }
        else {
            message.sequence = 0;
        }
        if (object.sourcePort !== undefined && object.sourcePort !== null) {
            message.sourcePort = object.sourcePort;
        }
        else {
            message.sourcePort = "";
        }
        if (object.sourceChannel !== undefined && object.sourceChannel !== null) {
            message.sourceChannel = object.sourceChannel;
        }
        else {
            message.sourceChannel = "";
        }
        if (object.destinationPort !== undefined &&
            object.destinationPort !== null) {
            message.destinationPort = object.destinationPort;
        }
        else {
            message.destinationPort = "";
        }
        if (object.destinationChannel !== undefined &&
            object.destinationChannel !== null) {
            message.destinationChannel = object.destinationChannel;
        }
        else {
            message.destinationChannel = "";
        }
        if (object.data !== undefined && object.data !== null) {
            message.data = object.data;
        }
        else {
            message.data = new Uint8Array();
        }
        if (object.timeoutHeight !== undefined && object.timeoutHeight !== null) {
            message.timeoutHeight = Height.fromPartial(object.timeoutHeight);
        }
        else {
            message.timeoutHeight = undefined;
        }
        if (object.timeoutTimestamp !== undefined &&
            object.timeoutTimestamp !== null) {
            message.timeoutTimestamp = object.timeoutTimestamp;
        }
        else {
            message.timeoutTimestamp = 0;
        }
        return message;
    },
};
const basePacketState = { portId: "", channelId: "", sequence: 0 };
export const PacketState = {
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
        if (message.data.length !== 0) {
            writer.uint32(34).bytes(message.data);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...basePacketState };
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
                case 4:
                    message.data = reader.bytes();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...basePacketState };
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
        if (object.data !== undefined && object.data !== null) {
            message.data = bytesFromBase64(object.data);
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.portId !== undefined && (obj.portId = message.portId);
        message.channelId !== undefined && (obj.channelId = message.channelId);
        message.sequence !== undefined && (obj.sequence = message.sequence);
        message.data !== undefined &&
            (obj.data = base64FromBytes(message.data !== undefined ? message.data : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        const message = { ...basePacketState };
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
        if (object.data !== undefined && object.data !== null) {
            message.data = object.data;
        }
        else {
            message.data = new Uint8Array();
        }
        return message;
    },
};
const baseAcknowledgement = {};
export const Acknowledgement = {
    encode(message, writer = Writer.create()) {
        if (message.result !== undefined) {
            writer.uint32(170).bytes(message.result);
        }
        if (message.error !== undefined) {
            writer.uint32(178).string(message.error);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseAcknowledgement };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 21:
                    message.result = reader.bytes();
                    break;
                case 22:
                    message.error = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseAcknowledgement };
        if (object.result !== undefined && object.result !== null) {
            message.result = bytesFromBase64(object.result);
        }
        if (object.error !== undefined && object.error !== null) {
            message.error = String(object.error);
        }
        else {
            message.error = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.result !== undefined &&
            (obj.result =
                message.result !== undefined
                    ? base64FromBytes(message.result)
                    : undefined);
        message.error !== undefined && (obj.error = message.error);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseAcknowledgement };
        if (object.result !== undefined && object.result !== null) {
            message.result = object.result;
        }
        else {
            message.result = undefined;
        }
        if (object.error !== undefined && object.error !== null) {
            message.error = object.error;
        }
        else {
            message.error = undefined;
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
