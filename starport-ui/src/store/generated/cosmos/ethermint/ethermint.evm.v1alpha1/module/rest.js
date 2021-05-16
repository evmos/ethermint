/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */
export var ContentType;
(function (ContentType) {
    ContentType["Json"] = "application/json";
    ContentType["FormData"] = "multipart/form-data";
    ContentType["UrlEncoded"] = "application/x-www-form-urlencoded";
})(ContentType || (ContentType = {}));
export class HttpClient {
    constructor(apiConfig = {}) {
        this.baseUrl = "";
        this.securityData = null;
        this.securityWorker = null;
        this.abortControllers = new Map();
        this.baseApiParams = {
            credentials: "same-origin",
            headers: {},
            redirect: "follow",
            referrerPolicy: "no-referrer",
        };
        this.setSecurityData = (data) => {
            this.securityData = data;
        };
        this.contentFormatters = {
            [ContentType.Json]: (input) => input !== null && (typeof input === "object" || typeof input === "string") ? JSON.stringify(input) : input,
            [ContentType.FormData]: (input) => Object.keys(input || {}).reduce((data, key) => {
                data.append(key, input[key]);
                return data;
            }, new FormData()),
            [ContentType.UrlEncoded]: (input) => this.toQueryString(input),
        };
        this.createAbortSignal = (cancelToken) => {
            if (this.abortControllers.has(cancelToken)) {
                const abortController = this.abortControllers.get(cancelToken);
                if (abortController) {
                    return abortController.signal;
                }
                return void 0;
            }
            const abortController = new AbortController();
            this.abortControllers.set(cancelToken, abortController);
            return abortController.signal;
        };
        this.abortRequest = (cancelToken) => {
            const abortController = this.abortControllers.get(cancelToken);
            if (abortController) {
                abortController.abort();
                this.abortControllers.delete(cancelToken);
            }
        };
        this.request = ({ body, secure, path, type, query, format = "json", baseUrl, cancelToken, ...params }) => {
            const secureParams = (secure && this.securityWorker && this.securityWorker(this.securityData)) || {};
            const requestParams = this.mergeRequestParams(params, secureParams);
            const queryString = query && this.toQueryString(query);
            const payloadFormatter = this.contentFormatters[type || ContentType.Json];
            return fetch(`${baseUrl || this.baseUrl || ""}${path}${queryString ? `?${queryString}` : ""}`, {
                ...requestParams,
                headers: {
                    ...(type && type !== ContentType.FormData ? { "Content-Type": type } : {}),
                    ...(requestParams.headers || {}),
                },
                signal: cancelToken ? this.createAbortSignal(cancelToken) : void 0,
                body: typeof body === "undefined" || body === null ? null : payloadFormatter(body),
            }).then(async (response) => {
                const r = response;
                r.data = null;
                r.error = null;
                const data = await response[format]()
                    .then((data) => {
                    if (r.ok) {
                        r.data = data;
                    }
                    else {
                        r.error = data;
                    }
                    return r;
                })
                    .catch((e) => {
                    r.error = e;
                    return r;
                });
                if (cancelToken) {
                    this.abortControllers.delete(cancelToken);
                }
                if (!response.ok)
                    throw data;
                return data;
            });
        };
        Object.assign(this, apiConfig);
    }
    addQueryParam(query, key) {
        const value = query[key];
        return (encodeURIComponent(key) +
            "=" +
            encodeURIComponent(Array.isArray(value) ? value.join(",") : typeof value === "number" ? value : `${value}`));
    }
    toQueryString(rawQuery) {
        const query = rawQuery || {};
        const keys = Object.keys(query).filter((key) => "undefined" !== typeof query[key]);
        return keys
            .map((key) => typeof query[key] === "object" && !Array.isArray(query[key])
            ? this.toQueryString(query[key])
            : this.addQueryParam(query, key))
            .join("&");
    }
    addQueryParams(rawQuery) {
        const queryString = this.toQueryString(rawQuery);
        return queryString ? `?${queryString}` : "";
    }
    mergeRequestParams(params1, params2) {
        return {
            ...this.baseApiParams,
            ...params1,
            ...(params2 || {}),
            headers: {
                ...(this.baseApiParams.headers || {}),
                ...(params1.headers || {}),
                ...((params2 && params2.headers) || {}),
            },
        };
    }
}
/**
 * @title ethermint/evm/v1alpha1/tx.proto
 * @version version not set
 */
export class Api extends HttpClient {
    constructor() {
        super(...arguments);
        /**
         * No description
         *
         * @tags Query
         * @name QueryAccount
         * @summary Account queries an Ethereum account.
         * @request GET:/ethermint/evm/v1alpha1/account/{address}
         */
        this.queryAccount = (address, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/account/${address}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryBalance
       * @summary Balance queries the balance of a the EVM denomination for a single
      EthAccount.
       * @request GET:/ethermint/evm/v1alpha1/balances/{address}
       */
        this.queryBalance = (address, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/balances/${address}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryBlockBloom
         * @summary BlockBloom queries the block bloom filter bytes at a given height.
         * @request GET:/ethermint/evm/v1alpha1/block_bloom
         */
        this.queryBlockBloom = (query, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/block_bloom`,
            method: "GET",
            query: query,
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryBlockLogs
         * @summary BlockLogs queries all the ethereum logs for a given block hash.
         * @request GET:/ethermint/evm/v1alpha1/block_logs/{hash}
         */
        this.queryBlockLogs = (hash, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/block_logs/${hash}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryCode
         * @summary Code queries the balance of all coins for a single account.
         * @request GET:/ethermint/evm/v1alpha1/codes/{address}
         */
        this.queryCode = (address, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/codes/${address}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryCosmosAccount
         * @summary Account queries an Ethereum account's Cosmos Address.
         * @request GET:/ethermint/evm/v1alpha1/cosmos_account/{address}
         */
        this.queryCosmosAccount = (address, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/cosmos_account/${address}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryParams
         * @summary Params queries the parameters of x/evm module.
         * @request GET:/ethermint/evm/v1alpha1/params
         */
        this.queryParams = (params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/params`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryStaticCall
         * @summary StaticCall queries the static call value of x/evm module.
         * @request GET:/ethermint/evm/v1alpha1/static_call
         */
        this.queryStaticCall = (query, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/static_call`,
            method: "GET",
            query: query,
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryStorage
         * @summary Storage queries the balance of all coins for a single account.
         * @request GET:/ethermint/evm/v1alpha1/storage/{address}/{key}
         */
        this.queryStorage = (address, key, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/storage/${address}/${key}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryTxLogs
         * @summary TxLogs queries ethereum logs from a transaction.
         * @request GET:/ethermint/evm/v1alpha1/tx_logs/{hash}
         */
        this.queryTxLogs = (hash, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/tx_logs/${hash}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryTxReceipt
         * @summary TxReceipt queries a receipt by a transaction hash.
         * @request GET:/ethermint/evm/v1alpha1/tx_receipt/{hash}
         */
        this.queryTxReceipt = (hash, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/tx_receipt/${hash}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryTxReceiptsByBlockHeight
         * @summary TxReceiptsByBlockHeight queries tx receipts by a block height.
         * @request GET:/ethermint/evm/v1alpha1/tx_receipts_block/{height}
         */
        this.queryTxReceiptsByBlockHeight = (height, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/tx_receipts_block/${height}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryTxReceiptsByBlockHash
         * @summary TxReceiptsByBlockHash queries tx receipts by a block hash.
         * @request GET:/ethermint/evm/v1alpha1/tx_receipts_block_hash/{hash}
         */
        this.queryTxReceiptsByBlockHash = (hash, params = {}) => this.request({
            path: `/ethermint/evm/v1alpha1/tx_receipts_block_hash/${hash}`,
            method: "GET",
            format: "json",
            ...params,
        });
    }
}
