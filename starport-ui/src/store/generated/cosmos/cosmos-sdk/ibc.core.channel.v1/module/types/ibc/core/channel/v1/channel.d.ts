import { Writer, Reader } from "protobufjs/minimal";
import { Height } from "../../../../ibc/core/client/v1/client";
export declare const protobufPackage = "ibc.core.channel.v1";
/**
 * State defines if a channel is in one of the following states:
 * CLOSED, INIT, TRYOPEN, OPEN or UNINITIALIZED.
 */
export declare enum State {
    /** STATE_UNINITIALIZED_UNSPECIFIED - Default State */
    STATE_UNINITIALIZED_UNSPECIFIED = 0,
    /** STATE_INIT - A channel has just started the opening handshake. */
    STATE_INIT = 1,
    /** STATE_TRYOPEN - A channel has acknowledged the handshake step on the counterparty chain. */
    STATE_TRYOPEN = 2,
    /**
     * STATE_OPEN - A channel has completed the handshake. Open channels are
     * ready to send and receive packets.
     */
    STATE_OPEN = 3,
    /**
     * STATE_CLOSED - A channel has been closed and can no longer be used to send or receive
     * packets.
     */
    STATE_CLOSED = 4,
    UNRECOGNIZED = -1
}
export declare function stateFromJSON(object: any): State;
export declare function stateToJSON(object: State): string;
/** Order defines if a channel is ORDERED or UNORDERED */
export declare enum Order {
    /** ORDER_NONE_UNSPECIFIED - zero-value for channel ordering */
    ORDER_NONE_UNSPECIFIED = 0,
    /**
     * ORDER_UNORDERED - packets can be delivered in any order, which may differ from the order in
     * which they were sent.
     */
    ORDER_UNORDERED = 1,
    /** ORDER_ORDERED - packets are delivered exactly in the order which they were sent */
    ORDER_ORDERED = 2,
    UNRECOGNIZED = -1
}
export declare function orderFromJSON(object: any): Order;
export declare function orderToJSON(object: Order): string;
/**
 * Channel defines pipeline for exactly-once packet delivery between specific
 * modules on separate blockchains, which has at least one end capable of
 * sending packets and one end capable of receiving packets.
 */
export interface Channel {
    /** current state of the channel end */
    state: State;
    /** whether the channel is ordered or unordered */
    ordering: Order;
    /** counterparty channel end */
    counterparty: Counterparty | undefined;
    /**
     * list of connection identifiers, in order, along which packets sent on
     * this channel will travel
     */
    connectionHops: string[];
    /** opaque channel version, which is agreed upon during the handshake */
    version: string;
}
/**
 * IdentifiedChannel defines a channel with additional port and channel
 * identifier fields.
 */
export interface IdentifiedChannel {
    /** current state of the channel end */
    state: State;
    /** whether the channel is ordered or unordered */
    ordering: Order;
    /** counterparty channel end */
    counterparty: Counterparty | undefined;
    /**
     * list of connection identifiers, in order, along which packets sent on
     * this channel will travel
     */
    connectionHops: string[];
    /** opaque channel version, which is agreed upon during the handshake */
    version: string;
    /** port identifier */
    portId: string;
    /** channel identifier */
    channelId: string;
}
/** Counterparty defines a channel end counterparty */
export interface Counterparty {
    /** port on the counterparty chain which owns the other end of the channel. */
    portId: string;
    /** channel end on the counterparty chain */
    channelId: string;
}
/** Packet defines a type that carries data across different chains through IBC */
export interface Packet {
    /**
     * number corresponds to the order of sends and receives, where a Packet
     * with an earlier sequence number must be sent and received before a Packet
     * with a later sequence number.
     */
    sequence: number;
    /** identifies the port on the sending chain. */
    sourcePort: string;
    /** identifies the channel end on the sending chain. */
    sourceChannel: string;
    /** identifies the port on the receiving chain. */
    destinationPort: string;
    /** identifies the channel end on the receiving chain. */
    destinationChannel: string;
    /** actual opaque bytes transferred directly to the application module */
    data: Uint8Array;
    /** block height after which the packet times out */
    timeoutHeight: Height | undefined;
    /** block timestamp (in nanoseconds) after which the packet times out */
    timeoutTimestamp: number;
}
/**
 * PacketState defines the generic type necessary to retrieve and store
 * packet commitments, acknowledgements, and receipts.
 * Caller is responsible for knowing the context necessary to interpret this
 * state as a commitment, acknowledgement, or a receipt.
 */
export interface PacketState {
    /** channel port identifier. */
    portId: string;
    /** channel unique identifier. */
    channelId: string;
    /** packet sequence. */
    sequence: number;
    /** embedded data that represents packet state. */
    data: Uint8Array;
}
/**
 * Acknowledgement is the recommended acknowledgement format to be used by
 * app-specific protocols.
 * NOTE: The field numbers 21 and 22 were explicitly chosen to avoid accidental
 * conflicts with other protobuf message formats used for acknowledgements.
 * The first byte of any message with this format will be the non-ASCII values
 * `0xaa` (result) or `0xb2` (error). Implemented as defined by ICS:
 * https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#acknowledgement-envelope
 */
export interface Acknowledgement {
    result: Uint8Array | undefined;
    error: string | undefined;
}
export declare const Channel: {
    encode(message: Channel, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Channel;
    fromJSON(object: any): Channel;
    toJSON(message: Channel): unknown;
    fromPartial(object: DeepPartial<Channel>): Channel;
};
export declare const IdentifiedChannel: {
    encode(message: IdentifiedChannel, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): IdentifiedChannel;
    fromJSON(object: any): IdentifiedChannel;
    toJSON(message: IdentifiedChannel): unknown;
    fromPartial(object: DeepPartial<IdentifiedChannel>): IdentifiedChannel;
};
export declare const Counterparty: {
    encode(message: Counterparty, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Counterparty;
    fromJSON(object: any): Counterparty;
    toJSON(message: Counterparty): unknown;
    fromPartial(object: DeepPartial<Counterparty>): Counterparty;
};
export declare const Packet: {
    encode(message: Packet, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Packet;
    fromJSON(object: any): Packet;
    toJSON(message: Packet): unknown;
    fromPartial(object: DeepPartial<Packet>): Packet;
};
export declare const PacketState: {
    encode(message: PacketState, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): PacketState;
    fromJSON(object: any): PacketState;
    toJSON(message: PacketState): unknown;
    fromPartial(object: DeepPartial<PacketState>): PacketState;
};
export declare const Acknowledgement: {
    encode(message: Acknowledgement, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Acknowledgement;
    fromJSON(object: any): Acknowledgement;
    toJSON(message: Acknowledgement): unknown;
    fromPartial(object: DeepPartial<Acknowledgement>): Acknowledgement;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
