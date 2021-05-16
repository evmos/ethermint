import { Reader, Writer } from "protobufjs/minimal";
import { Log, TxReceipt, TransactionLogs, Params } from "../../../ethermint/evm/v1alpha1/evm";
export declare const protobufPackage = "ethermint.evm.v1alpha1";
/** QueryAccountRequest is the request type for the Query/Account RPC method. */
export interface QueryAccountRequest {
    /** address is the ethereum hex address to query the account for. */
    address: string;
}
/** QueryAccountResponse is the response type for the Query/Account RPC method. */
export interface QueryAccountResponse {
    /** balance is the balance of the EVM denomination. */
    balance: string;
    /** code_hash is the code bytes from the EOA. */
    codeHash: Uint8Array;
    /** nonce is the account's sequence number. */
    nonce: number;
}
/** QueryCosmosAccountRequest is the request type for the Query/CosmosAccount RPC method. */
export interface QueryCosmosAccountRequest {
    /** address is the ethereum hex address to query the account for. */
    address: string;
}
/** QueryCosmosAccountResponse is the response type for the Query/CosmosAccount RPC method. */
export interface QueryCosmosAccountResponse {
    /** cosmos_address is the cosmos address of the account. */
    cosmosAddress: string;
    /** sequence is the account's sequence number. */
    sequence: number;
    /** account_number is the account numbert */
    accountNumber: number;
}
/** QueryBalanceRequest is the request type for the Query/Balance RPC method. */
export interface QueryBalanceRequest {
    /** address is the ethereum hex address to query the balance for. */
    address: string;
}
/** QueryBalanceResponse is the response type for the Query/Balance RPC method. */
export interface QueryBalanceResponse {
    /** balance is the balance of the EVM denomination. */
    balance: string;
}
/** QueryStorageRequest is the request type for the Query/Storage RPC method. */
export interface QueryStorageRequest {
    /** / address is the ethereum hex address to query the storage state for. */
    address: string;
    /** key defines the key of the storage state */
    key: string;
}
/**
 * QueryStorageResponse is the response type for the Query/Storage RPC
 * method.
 */
export interface QueryStorageResponse {
    /** key defines the storage state value hash associated with the given key. */
    value: string;
}
/** QueryCodeRequest is the request type for the Query/Code RPC method. */
export interface QueryCodeRequest {
    /** address is the ethereum hex address to query the code for. */
    address: string;
}
/**
 * QueryCodeResponse is the response type for the Query/Code RPC
 * method.
 */
export interface QueryCodeResponse {
    /** code represents the code bytes from an ethereum address. */
    code: Uint8Array;
}
/** QueryTxLogsRequest is the request type for the Query/TxLogs RPC method. */
export interface QueryTxLogsRequest {
    /** hash is the ethereum transaction hex hash to query the logs for. */
    hash: string;
}
/** QueryTxLogs is the response type for the Query/TxLogs RPC method. */
export interface QueryTxLogsResponse {
    /** logs represents the ethereum logs generated from the given transaction. */
    logs: Log[];
}
/** QueryTxReceiptRequest is the request type for the Query/TxReceipt RPC method. */
export interface QueryTxReceiptRequest {
    /** hash is the ethereum transaction hex hash to query the receipt for. */
    hash: string;
}
/** QueryTxReceiptResponse is the response type for the Query/TxReceipt RPC method. */
export interface QueryTxReceiptResponse {
    /** receipt represents the ethereum receipt for the given transaction. */
    receipt: TxReceipt | undefined;
}
/** QueryTxReceiptsByBlockHeightRequest is the request type for the Query/TxReceiptsByBlockHeight RPC method. */
export interface QueryTxReceiptsByBlockHeightRequest {
    /** height is the block height to query tx receipts for */
    height: number;
}
/** QueryTxReceiptsByBlockHeightResponse is the response type for the Query/TxReceiptsByBlockHeight RPC method. */
export interface QueryTxReceiptsByBlockHeightResponse {
    /** tx receipts list for the block */
    receipts: TxReceipt[];
}
/** QueryTxReceiptsByBlockHashRequest is the request type for the Query/TxReceiptsByBlockHash RPC method. */
export interface QueryTxReceiptsByBlockHashRequest {
    /** hash is the ethereum transaction hex hash to query the receipt for. */
    hash: string;
}
/** QueryTxReceiptsByBlockHashResponse is the response type for the Query/TxReceiptsByBlockHash RPC method. */
export interface QueryTxReceiptsByBlockHashResponse {
    /** tx receipts list for the block */
    receipts: TxReceipt[];
}
/** QueryBlockLogsRequest is the request type for the Query/BlockLogs RPC method. */
export interface QueryBlockLogsRequest {
    /** hash is the block hash to query the logs for. */
    hash: string;
}
/** QueryTxLogs is the response type for the Query/BlockLogs RPC method. */
export interface QueryBlockLogsResponse {
    /** logs represents the ethereum logs generated at the given block hash. */
    txLogs: TransactionLogs[];
}
/**
 * QueryBlockBloomRequest is the request type for the Query/BlockBloom RPC
 * method.
 */
export interface QueryBlockBloomRequest {
    height: number;
}
/**
 * QueryBlockBloomResponse is the response type for the Query/BlockBloom RPC
 * method.
 */
export interface QueryBlockBloomResponse {
    /** bloom represents bloom filter for the given block hash. */
    bloom: Uint8Array;
}
/** QueryParamsRequest defines the request type for querying x/evm parameters. */
export interface QueryParamsRequest {
}
/** QueryParamsResponse defines the response type for querying x/evm parameters. */
export interface QueryParamsResponse {
    /** params define the evm module parameters. */
    params: Params | undefined;
}
/** QueryStaticCallRequest defines static call request */
export interface QueryStaticCallRequest {
    /** address is the ethereum contract hex address to for static call. */
    address: string;
    /** static call input generated from abi */
    input: Uint8Array;
}
/** // QueryStaticCallRequest defines static call response */
export interface QueryStaticCallResponse {
    data: Uint8Array;
}
export declare const QueryAccountRequest: {
    encode(message: QueryAccountRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryAccountRequest;
    fromJSON(object: any): QueryAccountRequest;
    toJSON(message: QueryAccountRequest): unknown;
    fromPartial(object: DeepPartial<QueryAccountRequest>): QueryAccountRequest;
};
export declare const QueryAccountResponse: {
    encode(message: QueryAccountResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryAccountResponse;
    fromJSON(object: any): QueryAccountResponse;
    toJSON(message: QueryAccountResponse): unknown;
    fromPartial(object: DeepPartial<QueryAccountResponse>): QueryAccountResponse;
};
export declare const QueryCosmosAccountRequest: {
    encode(message: QueryCosmosAccountRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryCosmosAccountRequest;
    fromJSON(object: any): QueryCosmosAccountRequest;
    toJSON(message: QueryCosmosAccountRequest): unknown;
    fromPartial(object: DeepPartial<QueryCosmosAccountRequest>): QueryCosmosAccountRequest;
};
export declare const QueryCosmosAccountResponse: {
    encode(message: QueryCosmosAccountResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryCosmosAccountResponse;
    fromJSON(object: any): QueryCosmosAccountResponse;
    toJSON(message: QueryCosmosAccountResponse): unknown;
    fromPartial(object: DeepPartial<QueryCosmosAccountResponse>): QueryCosmosAccountResponse;
};
export declare const QueryBalanceRequest: {
    encode(message: QueryBalanceRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryBalanceRequest;
    fromJSON(object: any): QueryBalanceRequest;
    toJSON(message: QueryBalanceRequest): unknown;
    fromPartial(object: DeepPartial<QueryBalanceRequest>): QueryBalanceRequest;
};
export declare const QueryBalanceResponse: {
    encode(message: QueryBalanceResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryBalanceResponse;
    fromJSON(object: any): QueryBalanceResponse;
    toJSON(message: QueryBalanceResponse): unknown;
    fromPartial(object: DeepPartial<QueryBalanceResponse>): QueryBalanceResponse;
};
export declare const QueryStorageRequest: {
    encode(message: QueryStorageRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryStorageRequest;
    fromJSON(object: any): QueryStorageRequest;
    toJSON(message: QueryStorageRequest): unknown;
    fromPartial(object: DeepPartial<QueryStorageRequest>): QueryStorageRequest;
};
export declare const QueryStorageResponse: {
    encode(message: QueryStorageResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryStorageResponse;
    fromJSON(object: any): QueryStorageResponse;
    toJSON(message: QueryStorageResponse): unknown;
    fromPartial(object: DeepPartial<QueryStorageResponse>): QueryStorageResponse;
};
export declare const QueryCodeRequest: {
    encode(message: QueryCodeRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryCodeRequest;
    fromJSON(object: any): QueryCodeRequest;
    toJSON(message: QueryCodeRequest): unknown;
    fromPartial(object: DeepPartial<QueryCodeRequest>): QueryCodeRequest;
};
export declare const QueryCodeResponse: {
    encode(message: QueryCodeResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryCodeResponse;
    fromJSON(object: any): QueryCodeResponse;
    toJSON(message: QueryCodeResponse): unknown;
    fromPartial(object: DeepPartial<QueryCodeResponse>): QueryCodeResponse;
};
export declare const QueryTxLogsRequest: {
    encode(message: QueryTxLogsRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxLogsRequest;
    fromJSON(object: any): QueryTxLogsRequest;
    toJSON(message: QueryTxLogsRequest): unknown;
    fromPartial(object: DeepPartial<QueryTxLogsRequest>): QueryTxLogsRequest;
};
export declare const QueryTxLogsResponse: {
    encode(message: QueryTxLogsResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxLogsResponse;
    fromJSON(object: any): QueryTxLogsResponse;
    toJSON(message: QueryTxLogsResponse): unknown;
    fromPartial(object: DeepPartial<QueryTxLogsResponse>): QueryTxLogsResponse;
};
export declare const QueryTxReceiptRequest: {
    encode(message: QueryTxReceiptRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxReceiptRequest;
    fromJSON(object: any): QueryTxReceiptRequest;
    toJSON(message: QueryTxReceiptRequest): unknown;
    fromPartial(object: DeepPartial<QueryTxReceiptRequest>): QueryTxReceiptRequest;
};
export declare const QueryTxReceiptResponse: {
    encode(message: QueryTxReceiptResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxReceiptResponse;
    fromJSON(object: any): QueryTxReceiptResponse;
    toJSON(message: QueryTxReceiptResponse): unknown;
    fromPartial(object: DeepPartial<QueryTxReceiptResponse>): QueryTxReceiptResponse;
};
export declare const QueryTxReceiptsByBlockHeightRequest: {
    encode(message: QueryTxReceiptsByBlockHeightRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxReceiptsByBlockHeightRequest;
    fromJSON(object: any): QueryTxReceiptsByBlockHeightRequest;
    toJSON(message: QueryTxReceiptsByBlockHeightRequest): unknown;
    fromPartial(object: DeepPartial<QueryTxReceiptsByBlockHeightRequest>): QueryTxReceiptsByBlockHeightRequest;
};
export declare const QueryTxReceiptsByBlockHeightResponse: {
    encode(message: QueryTxReceiptsByBlockHeightResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxReceiptsByBlockHeightResponse;
    fromJSON(object: any): QueryTxReceiptsByBlockHeightResponse;
    toJSON(message: QueryTxReceiptsByBlockHeightResponse): unknown;
    fromPartial(object: DeepPartial<QueryTxReceiptsByBlockHeightResponse>): QueryTxReceiptsByBlockHeightResponse;
};
export declare const QueryTxReceiptsByBlockHashRequest: {
    encode(message: QueryTxReceiptsByBlockHashRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxReceiptsByBlockHashRequest;
    fromJSON(object: any): QueryTxReceiptsByBlockHashRequest;
    toJSON(message: QueryTxReceiptsByBlockHashRequest): unknown;
    fromPartial(object: DeepPartial<QueryTxReceiptsByBlockHashRequest>): QueryTxReceiptsByBlockHashRequest;
};
export declare const QueryTxReceiptsByBlockHashResponse: {
    encode(message: QueryTxReceiptsByBlockHashResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryTxReceiptsByBlockHashResponse;
    fromJSON(object: any): QueryTxReceiptsByBlockHashResponse;
    toJSON(message: QueryTxReceiptsByBlockHashResponse): unknown;
    fromPartial(object: DeepPartial<QueryTxReceiptsByBlockHashResponse>): QueryTxReceiptsByBlockHashResponse;
};
export declare const QueryBlockLogsRequest: {
    encode(message: QueryBlockLogsRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryBlockLogsRequest;
    fromJSON(object: any): QueryBlockLogsRequest;
    toJSON(message: QueryBlockLogsRequest): unknown;
    fromPartial(object: DeepPartial<QueryBlockLogsRequest>): QueryBlockLogsRequest;
};
export declare const QueryBlockLogsResponse: {
    encode(message: QueryBlockLogsResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryBlockLogsResponse;
    fromJSON(object: any): QueryBlockLogsResponse;
    toJSON(message: QueryBlockLogsResponse): unknown;
    fromPartial(object: DeepPartial<QueryBlockLogsResponse>): QueryBlockLogsResponse;
};
export declare const QueryBlockBloomRequest: {
    encode(message: QueryBlockBloomRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryBlockBloomRequest;
    fromJSON(object: any): QueryBlockBloomRequest;
    toJSON(message: QueryBlockBloomRequest): unknown;
    fromPartial(object: DeepPartial<QueryBlockBloomRequest>): QueryBlockBloomRequest;
};
export declare const QueryBlockBloomResponse: {
    encode(message: QueryBlockBloomResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryBlockBloomResponse;
    fromJSON(object: any): QueryBlockBloomResponse;
    toJSON(message: QueryBlockBloomResponse): unknown;
    fromPartial(object: DeepPartial<QueryBlockBloomResponse>): QueryBlockBloomResponse;
};
export declare const QueryParamsRequest: {
    encode(_: QueryParamsRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest;
    fromJSON(_: any): QueryParamsRequest;
    toJSON(_: QueryParamsRequest): unknown;
    fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest;
};
export declare const QueryParamsResponse: {
    encode(message: QueryParamsResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse;
    fromJSON(object: any): QueryParamsResponse;
    toJSON(message: QueryParamsResponse): unknown;
    fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse;
};
export declare const QueryStaticCallRequest: {
    encode(message: QueryStaticCallRequest, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryStaticCallRequest;
    fromJSON(object: any): QueryStaticCallRequest;
    toJSON(message: QueryStaticCallRequest): unknown;
    fromPartial(object: DeepPartial<QueryStaticCallRequest>): QueryStaticCallRequest;
};
export declare const QueryStaticCallResponse: {
    encode(message: QueryStaticCallResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): QueryStaticCallResponse;
    fromJSON(object: any): QueryStaticCallResponse;
    toJSON(message: QueryStaticCallResponse): unknown;
    fromPartial(object: DeepPartial<QueryStaticCallResponse>): QueryStaticCallResponse;
};
/** Query defines the gRPC querier service. */
export interface Query {
    /** Account queries an Ethereum account. */
    Account(request: QueryAccountRequest): Promise<QueryAccountResponse>;
    /** Account queries an Ethereum account's Cosmos Address. */
    CosmosAccount(request: QueryCosmosAccountRequest): Promise<QueryCosmosAccountResponse>;
    /**
     * Balance queries the balance of a the EVM denomination for a single
     * EthAccount.
     */
    Balance(request: QueryBalanceRequest): Promise<QueryBalanceResponse>;
    /** Storage queries the balance of all coins for a single account. */
    Storage(request: QueryStorageRequest): Promise<QueryStorageResponse>;
    /** Code queries the balance of all coins for a single account. */
    Code(request: QueryCodeRequest): Promise<QueryCodeResponse>;
    /** TxLogs queries ethereum logs from a transaction. */
    TxLogs(request: QueryTxLogsRequest): Promise<QueryTxLogsResponse>;
    /** TxReceipt queries a receipt by a transaction hash. */
    TxReceipt(request: QueryTxReceiptRequest): Promise<QueryTxReceiptResponse>;
    /** TxReceiptsByBlockHeight queries tx receipts by a block height. */
    TxReceiptsByBlockHeight(request: QueryTxReceiptsByBlockHeightRequest): Promise<QueryTxReceiptsByBlockHeightResponse>;
    /** TxReceiptsByBlockHash queries tx receipts by a block hash. */
    TxReceiptsByBlockHash(request: QueryTxReceiptsByBlockHashRequest): Promise<QueryTxReceiptsByBlockHashResponse>;
    /** BlockLogs queries all the ethereum logs for a given block hash. */
    BlockLogs(request: QueryBlockLogsRequest): Promise<QueryBlockLogsResponse>;
    /** BlockBloom queries the block bloom filter bytes at a given height. */
    BlockBloom(request: QueryBlockBloomRequest): Promise<QueryBlockBloomResponse>;
    /** Params queries the parameters of x/evm module. */
    Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
    /** StaticCall queries the static call value of x/evm module. */
    StaticCall(request: QueryStaticCallRequest): Promise<QueryStaticCallResponse>;
}
export declare class QueryClientImpl implements Query {
    private readonly rpc;
    constructor(rpc: Rpc);
    Account(request: QueryAccountRequest): Promise<QueryAccountResponse>;
    CosmosAccount(request: QueryCosmosAccountRequest): Promise<QueryCosmosAccountResponse>;
    Balance(request: QueryBalanceRequest): Promise<QueryBalanceResponse>;
    Storage(request: QueryStorageRequest): Promise<QueryStorageResponse>;
    Code(request: QueryCodeRequest): Promise<QueryCodeResponse>;
    TxLogs(request: QueryTxLogsRequest): Promise<QueryTxLogsResponse>;
    TxReceipt(request: QueryTxReceiptRequest): Promise<QueryTxReceiptResponse>;
    TxReceiptsByBlockHeight(request: QueryTxReceiptsByBlockHeightRequest): Promise<QueryTxReceiptsByBlockHeightResponse>;
    TxReceiptsByBlockHash(request: QueryTxReceiptsByBlockHashRequest): Promise<QueryTxReceiptsByBlockHashResponse>;
    BlockLogs(request: QueryBlockLogsRequest): Promise<QueryBlockLogsResponse>;
    BlockBloom(request: QueryBlockBloomRequest): Promise<QueryBlockBloomResponse>;
    Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
    StaticCall(request: QueryStaticCallRequest): Promise<QueryStaticCallResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
