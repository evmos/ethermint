import { ChainConfig, Params, TransactionLogs, State } from "../../../ethermint/evm/v1alpha1/evm";
import { Writer, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "ethermint.evm.v1alpha1";
/** GenesisState defines the evm module's genesis state. */
export interface GenesisState {
    /** accounts is an array containing the ethereum genesis accounts. */
    accounts: GenesisAccount[];
    /** chain_config defines the Ethereum chain configuration. */
    chainConfig: ChainConfig | undefined;
    /** params defines all the paramaters of the module. */
    params: Params | undefined;
    txsLogs: TransactionLogs[];
}
/**
 * GenesisAccount defines an account to be initialized in the genesis state.
 * Its main difference between with Geth's GenesisAccount is that it uses a
 * custom storage type and that it doesn't contain the private key field.
 */
export interface GenesisAccount {
    /** address defines an ethereum hex formated address of an account */
    address: string;
    /** code defines the hex bytes of the account code. */
    code: string;
    /** storage defines the set of state key values for the account. */
    storage: State[];
}
export declare const GenesisState: {
    encode(message: GenesisState, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): GenesisState;
    fromJSON(object: any): GenesisState;
    toJSON(message: GenesisState): unknown;
    fromPartial(object: DeepPartial<GenesisState>): GenesisState;
};
export declare const GenesisAccount: {
    encode(message: GenesisAccount, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): GenesisAccount;
    fromJSON(object: any): GenesisAccount;
    toJSON(message: GenesisAccount): unknown;
    fromPartial(object: DeepPartial<GenesisAccount>): GenesisAccount;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
