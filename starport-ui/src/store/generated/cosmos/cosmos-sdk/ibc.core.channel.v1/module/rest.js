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
/**
* - ORDER_NONE_UNSPECIFIED: zero-value for channel ordering
 - ORDER_UNORDERED: packets can be delivered in any order, which may differ from the order in
which they were sent.
 - ORDER_ORDERED: packets are delivered exactly in the order which they were sent
*/
export var V1Order;
(function (V1Order) {
    V1Order["ORDER_NONE_UNSPECIFIED"] = "ORDER_NONE_UNSPECIFIED";
    V1Order["ORDER_UNORDERED"] = "ORDER_UNORDERED";
    V1Order["ORDER_ORDERED"] = "ORDER_ORDERED";
})(V1Order || (V1Order = {}));
/**
* State defines if a channel is in one of the following states:
CLOSED, INIT, TRYOPEN, OPEN or UNINITIALIZED.

 - STATE_UNINITIALIZED_UNSPECIFIED: Default State
 - STATE_INIT: A channel has just started the opening handshake.
 - STATE_TRYOPEN: A channel has acknowledged the handshake step on the counterparty chain.
 - STATE_OPEN: A channel has completed the handshake. Open channels are
ready to send and receive packets.
 - STATE_CLOSED: A channel has been closed and can no longer be used to send or receive
packets.
*/
export var V1State;
(function (V1State) {
    V1State["STATE_UNINITIALIZED_UNSPECIFIED"] = "STATE_UNINITIALIZED_UNSPECIFIED";
    V1State["STATE_INIT"] = "STATE_INIT";
    V1State["STATE_TRYOPEN"] = "STATE_TRYOPEN";
    V1State["STATE_OPEN"] = "STATE_OPEN";
    V1State["STATE_CLOSED"] = "STATE_CLOSED";
})(V1State || (V1State = {}));
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
 * @title ibc/core/channel/v1/tx.proto
 * @version version not set
 */
export class Api extends HttpClient {
    constructor() {
        super(...arguments);
        /**
         * No description
         *
         * @tags Query
         * @name QueryChannels
         * @summary Channels queries all the IBC channels of a chain.
         * @request GET:/ibc/core/channel/v1beta1/channels
         */
        this.queryChannels = (query, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels`,
            method: "GET",
            query: query,
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryChannel
         * @summary Channel queries an IBC Channel.
         * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}
         */
        this.queryChannel = (channelId, portId, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryChannelClientState
       * @summary ChannelClientState queries for the client state for the channel associated
      with the provided channel identifiers.
       * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/client_state
       */
        this.queryChannelClientState = (channelId, portId, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/client_state`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryChannelConsensusState
       * @summary ChannelConsensusState queries for the consensus state for the channel
      associated with the provided channel identifiers.
       * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/consensus_state/revision/{revisionNumber}/height/{revisionHeight}
       */
        this.queryChannelConsensusState = (channelId, portId, revisionNumber, revisionHeight, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/consensus_state/revision/${revisionNumber}/height/${revisionHeight}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryNextSequenceReceive
         * @summary NextSequenceReceive returns the next receive sequence for a given channel.
         * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/next_sequence
         */
        this.queryNextSequenceReceive = (channelId, portId, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/next_sequence`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryPacketAcknowledgements
       * @summary PacketAcknowledgements returns all the packet acknowledgements associated
      with a channel.
       * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_acknowledgements
       */
        this.queryPacketAcknowledgements = (channelId, portId, query, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_acknowledgements`,
            method: "GET",
            query: query,
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryPacketAcknowledgement
         * @summary PacketAcknowledgement queries a stored packet acknowledgement hash.
         * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_acks/{sequence}
         */
        this.queryPacketAcknowledgement = (channelId, portId, sequence, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_acks/${sequence}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryPacketCommitments
       * @summary PacketCommitments returns all the packet commitments hashes associated
      with a channel.
       * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_commitments
       */
        this.queryPacketCommitments = (channelId, portId, query, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_commitments`,
            method: "GET",
            query: query,
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryUnreceivedAcks
       * @summary UnreceivedAcks returns all the unreceived IBC acknowledgements associated with a
      channel and sequences.
       * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_commitments/{packetAckSequences}/unreceived_acks
       */
        this.queryUnreceivedAcks = (channelId, portId, packetAckSequences, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_commitments/${packetAckSequences}/unreceived_acks`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryUnreceivedPackets
       * @summary UnreceivedPackets returns all the unreceived IBC packets associated with a
      channel and sequences.
       * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_commitments/{packetCommitmentSequences}/unreceived_packets
       */
        this.queryUnreceivedPackets = (channelId, portId, packetCommitmentSequences, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_commitments/${packetCommitmentSequences}/unreceived_packets`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryPacketCommitment
         * @summary PacketCommitment queries a stored packet commitment hash.
         * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_commitments/{sequence}
         */
        this.queryPacketCommitment = (channelId, portId, sequence, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_commitments/${sequence}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
         * No description
         *
         * @tags Query
         * @name QueryPacketReceipt
         * @summary PacketReceipt queries if a given packet sequence has been received on the queried chain
         * @request GET:/ibc/core/channel/v1beta1/channels/{channelId}/ports/{portId}/packet_receipts/{sequence}
         */
        this.queryPacketReceipt = (channelId, portId, sequence, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/channels/${channelId}/ports/${portId}/packet_receipts/${sequence}`,
            method: "GET",
            format: "json",
            ...params,
        });
        /**
       * No description
       *
       * @tags Query
       * @name QueryConnectionChannels
       * @summary ConnectionChannels queries all the channels associated with a connection
      end.
       * @request GET:/ibc/core/channel/v1beta1/connections/{connection}/channels
       */
        this.queryConnectionChannels = (connection, query, params = {}) => this.request({
            path: `/ibc/core/channel/v1beta1/connections/${connection}/channels`,
            method: "GET",
            query: query,
            format: "json",
            ...params,
        });
    }
}
