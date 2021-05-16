import { Writer, Reader } from "protobufjs/minimal";
import { Any } from "../../../../google/protobuf/any";
export declare const protobufPackage = "ibc.core.client.v1";
/**
 * IdentifiedClientState defines a client state with an additional client
 * identifier field.
 */
export interface IdentifiedClientState {
    /** client identifier */
    clientId: string;
    /** client state */
    clientState: Any | undefined;
}
/** ConsensusStateWithHeight defines a consensus state with an additional height field. */
export interface ConsensusStateWithHeight {
    /** consensus state height */
    height: Height | undefined;
    /** consensus state */
    consensusState: Any | undefined;
}
/**
 * ClientConsensusStates defines all the stored consensus states for a given
 * client.
 */
export interface ClientConsensusStates {
    /** client identifier */
    clientId: string;
    /** consensus states and their heights associated with the client */
    consensusStates: ConsensusStateWithHeight[];
}
/**
 * ClientUpdateProposal is a governance proposal. If it passes, the client is
 * updated with the provided header. The update may fail if the header is not
 * valid given certain conditions specified by the client implementation.
 */
export interface ClientUpdateProposal {
    /** the title of the update proposal */
    title: string;
    /** the description of the proposal */
    description: string;
    /** the client identifier for the client to be updated if the proposal passes */
    clientId: string;
    /** the header used to update the client if the proposal passes */
    header: Any | undefined;
}
/**
 * Height is a monotonically increasing data type
 * that can be compared against another Height for the purposes of updating and
 * freezing clients
 *
 * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
 * the same. However some consensus algorithms may choose to reset the
 * height in certain conditions e.g. hard forks, state-machine breaking changes
 * In these cases, the RevisionNumber is incremented so that height continues to
 * be monitonically increasing even as the RevisionHeight gets reset
 */
export interface Height {
    /** the revision that the client is currently on */
    revisionNumber: number;
    /** the height within the given revision */
    revisionHeight: number;
}
/** Params defines the set of IBC light client parameters. */
export interface Params {
    /** allowed_clients defines the list of allowed client state types. */
    allowedClients: string[];
}
export declare const IdentifiedClientState: {
    encode(message: IdentifiedClientState, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): IdentifiedClientState;
    fromJSON(object: any): IdentifiedClientState;
    toJSON(message: IdentifiedClientState): unknown;
    fromPartial(object: DeepPartial<IdentifiedClientState>): IdentifiedClientState;
};
export declare const ConsensusStateWithHeight: {
    encode(message: ConsensusStateWithHeight, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ConsensusStateWithHeight;
    fromJSON(object: any): ConsensusStateWithHeight;
    toJSON(message: ConsensusStateWithHeight): unknown;
    fromPartial(object: DeepPartial<ConsensusStateWithHeight>): ConsensusStateWithHeight;
};
export declare const ClientConsensusStates: {
    encode(message: ClientConsensusStates, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ClientConsensusStates;
    fromJSON(object: any): ClientConsensusStates;
    toJSON(message: ClientConsensusStates): unknown;
    fromPartial(object: DeepPartial<ClientConsensusStates>): ClientConsensusStates;
};
export declare const ClientUpdateProposal: {
    encode(message: ClientUpdateProposal, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ClientUpdateProposal;
    fromJSON(object: any): ClientUpdateProposal;
    toJSON(message: ClientUpdateProposal): unknown;
    fromPartial(object: DeepPartial<ClientUpdateProposal>): ClientUpdateProposal;
};
export declare const Height: {
    encode(message: Height, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Height;
    fromJSON(object: any): Height;
    toJSON(message: Height): unknown;
    fromPartial(object: DeepPartial<Height>): Height;
};
export declare const Params: {
    encode(message: Params, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Params;
    fromJSON(object: any): Params;
    toJSON(message: Params): unknown;
    fromPartial(object: DeepPartial<Params>): Params;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
