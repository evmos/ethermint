import { Writer, Reader } from "protobufjs/minimal";
import { Duration } from "../../../google/protobuf/duration";
export declare const protobufPackage = "cosmos.slashing.v1beta1";
/**
 * ValidatorSigningInfo defines a validator's signing info for monitoring their
 * liveness activity.
 */
export interface ValidatorSigningInfo {
    address: string;
    /** height at which validator was first a candidate OR was unjailed */
    startHeight: number;
    /** index offset into signed block bit array */
    indexOffset: number;
    /** timestamp validator cannot be unjailed until */
    jailedUntil: Date | undefined;
    /**
     * whether or not a validator has been tombstoned (killed out of validator
     * set)
     */
    tombstoned: boolean;
    /** missed blocks counter (to avoid scanning the array every time) */
    missedBlocksCounter: number;
}
/** Params represents the parameters used for by the slashing module. */
export interface Params {
    signedBlocksWindow: number;
    minSignedPerWindow: Uint8Array;
    downtimeJailDuration: Duration | undefined;
    slashFractionDoubleSign: Uint8Array;
    slashFractionDowntime: Uint8Array;
}
export declare const ValidatorSigningInfo: {
    encode(message: ValidatorSigningInfo, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ValidatorSigningInfo;
    fromJSON(object: any): ValidatorSigningInfo;
    toJSON(message: ValidatorSigningInfo): unknown;
    fromPartial(object: DeepPartial<ValidatorSigningInfo>): ValidatorSigningInfo;
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
