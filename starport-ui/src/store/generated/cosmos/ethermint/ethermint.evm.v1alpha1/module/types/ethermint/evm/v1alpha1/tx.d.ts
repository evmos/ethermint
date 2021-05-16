import { Reader, Writer } from "protobufjs/minimal";
import { TxData, TransactionLogs } from "../../../ethermint/evm/v1alpha1/evm";
export declare const protobufPackage = "ethermint.evm.v1alpha1";
/** MsgEthereumTx encapsulates an Ethereum transaction as an SDK message. */
export interface MsgEthereumTx {
    /** inner transaction data */
    data: TxData | undefined;
    /** encoded storage size of the transaction */
    size: number;
    /** transaction hash in hex format */
    hash: string;
    /**
     * ethereum signer address in hex format. This address value is checked against
     * the address derived from the signature (V, R, S) using the secp256k1
     * elliptic curve
     */
    from: string;
}
export interface ExtensionOptionsEthereumTx {
}
export interface ExtensionOptionsWeb3Tx {
}
/** MsgEthereumTxResponse defines the Msg/EthereumTx response type. */
export interface MsgEthereumTxResponse {
    /**
     * contract_address contains the ethereum address of the created contract (if
     * any). If the state transition is an evm.Call, the contract address will be
     * empty.
     */
    contractAddress: string;
    /** bloom represents the bloom filter bytes */
    bloom: Uint8Array;
    /**
     * tx_logs contains the transaction hash and the proto-compatible ethereum
     * logs.
     */
    txLogs: TransactionLogs | undefined;
    /** ret defines the bytes from the execution. */
    ret: Uint8Array;
    /** reverted flag is set to true when the call has been reverted */
    reverted: boolean;
}
export declare const MsgEthereumTx: {
    encode(message: MsgEthereumTx, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgEthereumTx;
    fromJSON(object: any): MsgEthereumTx;
    toJSON(message: MsgEthereumTx): unknown;
    fromPartial(object: DeepPartial<MsgEthereumTx>): MsgEthereumTx;
};
export declare const ExtensionOptionsEthereumTx: {
    encode(_: ExtensionOptionsEthereumTx, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ExtensionOptionsEthereumTx;
    fromJSON(_: any): ExtensionOptionsEthereumTx;
    toJSON(_: ExtensionOptionsEthereumTx): unknown;
    fromPartial(_: DeepPartial<ExtensionOptionsEthereumTx>): ExtensionOptionsEthereumTx;
};
export declare const ExtensionOptionsWeb3Tx: {
    encode(_: ExtensionOptionsWeb3Tx, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ExtensionOptionsWeb3Tx;
    fromJSON(_: any): ExtensionOptionsWeb3Tx;
    toJSON(_: ExtensionOptionsWeb3Tx): unknown;
    fromPartial(_: DeepPartial<ExtensionOptionsWeb3Tx>): ExtensionOptionsWeb3Tx;
};
export declare const MsgEthereumTxResponse: {
    encode(message: MsgEthereumTxResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgEthereumTxResponse;
    fromJSON(object: any): MsgEthereumTxResponse;
    toJSON(message: MsgEthereumTxResponse): unknown;
    fromPartial(object: DeepPartial<MsgEthereumTxResponse>): MsgEthereumTxResponse;
};
/** Msg defines the evm Msg service. */
export interface Msg {
    /** EthereumTx defines a method submitting Ethereum transactions. */
    EthereumTx(request: MsgEthereumTx): Promise<MsgEthereumTxResponse>;
}
export declare class MsgClientImpl implements Msg {
    private readonly rpc;
    constructor(rpc: Rpc);
    EthereumTx(request: MsgEthereumTx): Promise<MsgEthereumTxResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
