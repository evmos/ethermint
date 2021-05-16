/**
* `Any` contains an arbitrary serialized protocol buffer message along with a
URL that describes the type of the serialized message.

Protobuf library provides support to pack/unpack Any values in the form
of utility functions or additional generated methods of the Any type.

Example 1: Pack and unpack a message in C++.

    Foo foo = ...;
    Any any;
    any.PackFrom(foo);
    ...
    if (any.UnpackTo(&foo)) {
      ...
    }

Example 2: Pack and unpack a message in Java.

    Foo foo = ...;
    Any any = Any.pack(foo);
    ...
    if (any.is(Foo.class)) {
      foo = any.unpack(Foo.class);
    }

 Example 3: Pack and unpack a message in Python.

    foo = Foo(...)
    any = Any()
    any.Pack(foo)
    ...
    if any.Is(Foo.DESCRIPTOR):
      any.Unpack(foo)
      ...

 Example 4: Pack and unpack a message in Go

     foo := &pb.Foo{...}
     any, err := ptypes.MarshalAny(foo)
     ...
     foo := &pb.Foo{}
     if err := ptypes.UnmarshalAny(any, foo); err != nil {
       ...
     }

The pack methods provided by protobuf library will by default use
'type.googleapis.com/full.type.name' as the type URL and the unpack
methods only use the fully qualified type name after the last '/'
in the type URL, for example "foo.bar.com/x/y.z" will yield type
name "y.z".


JSON
====
The JSON representation of an `Any` value uses the regular
representation of the deserialized, embedded message, with an
additional field `@type` which contains the type URL. Example:

    package google.profile;
    message Person {
      string first_name = 1;
      string last_name = 2;
    }

    {
      "@type": "type.googleapis.com/google.profile.Person",
      "firstName": <string>,
      "lastName": <string>
    }

If the embedded message type is well-known and has a custom JSON
representation, that representation will be embedded adding a field
`value` which holds the custom JSON in addition to the `@type`
field. Example (for message [google.protobuf.Duration][]):

    {
      "@type": "type.googleapis.com/google.protobuf.Duration",
      "value": "1.212s"
    }
*/
export interface ProtobufAny {
    /**
     * A URL/resource name that uniquely identifies the type of the serialized
     * protocol buffer message. This string must contain at least
     * one "/" character. The last segment of the URL's path must represent
     * the fully qualified name of the type (as in
     * `path/google.protobuf.Duration`). The name should be in a canonical form
     * (e.g., leading "." is not accepted).
     *
     * In practice, teams usually precompile into the binary all types that they
     * expect it to use in the context of Any. However, for URLs which use the
     * scheme `http`, `https`, or no scheme, one can optionally set up a type
     * server that maps type URLs to message definitions as follows:
     *
     * * If no scheme is provided, `https` is assumed.
     * * An HTTP GET on the URL must yield a [google.protobuf.Type][]
     *   value in binary format, or produce an error.
     * * Applications are allowed to cache lookup results based on the
     *   URL, or have them precompiled into a binary to avoid any
     *   lookup. Therefore, binary compatibility needs to be preserved
     *   on changes to types. (Use versioned type names to manage
     *   breaking changes.)
     *
     * Note: this functionality is not currently available in the official
     * protobuf release, and it is not used for type URLs beginning with
     * type.googleapis.com.
     *
     * Schemes other than `http`, `https` (or the empty scheme) might be
     * used with implementation specific semantics.
     */
    typeUrl?: string;
    /**
     * Must be a valid serialized protocol buffer of the above specified type.
     * @format byte
     */
    value?: string;
}
export interface RpcStatus {
    /** @format int32 */
    code?: number;
    message?: string;
    details?: ProtobufAny[];
}
/**
 * ConsensusStateWithHeight defines a consensus state with an additional height field.
 */
export interface V1ConsensusStateWithHeight {
    /**
     * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
     * the same. However some consensus algorithms may choose to reset the
     * height in certain conditions e.g. hard forks, state-machine breaking changes
     * In these cases, the RevisionNumber is incremented so that height continues to
     * be monitonically increasing even as the RevisionHeight gets reset
     */
    height?: V1Height;
    /**
     * `Any` contains an arbitrary serialized protocol buffer message along with a
     * URL that describes the type of the serialized message.
     *
     * Protobuf library provides support to pack/unpack Any values in the form
     * of utility functions or additional generated methods of the Any type.
     *
     * Example 1: Pack and unpack a message in C++.
     *
     *     Foo foo = ...;
     *     Any any;
     *     any.PackFrom(foo);
     *     ...
     *     if (any.UnpackTo(&foo)) {
     *       ...
     *     }
     *
     * Example 2: Pack and unpack a message in Java.
     *
     *     Foo foo = ...;
     *     Any any = Any.pack(foo);
     *     ...
     *     if (any.is(Foo.class)) {
     *       foo = any.unpack(Foo.class);
     *     }
     *
     *  Example 3: Pack and unpack a message in Python.
     *
     *     foo = Foo(...)
     *     any = Any()
     *     any.Pack(foo)
     *     ...
     *     if any.Is(Foo.DESCRIPTOR):
     *       any.Unpack(foo)
     *       ...
     *
     *  Example 4: Pack and unpack a message in Go
     *
     *      foo := &pb.Foo{...}
     *      any, err := ptypes.MarshalAny(foo)
     *      ...
     *      foo := &pb.Foo{}
     *      if err := ptypes.UnmarshalAny(any, foo); err != nil {
     *        ...
     *      }
     *
     * The pack methods provided by protobuf library will by default use
     * 'type.googleapis.com/full.type.name' as the type URL and the unpack
     * methods only use the fully qualified type name after the last '/'
     * in the type URL, for example "foo.bar.com/x/y.z" will yield type
     * name "y.z".
     *
     *
     * JSON
     * ====
     * The JSON representation of an `Any` value uses the regular
     * representation of the deserialized, embedded message, with an
     * additional field `@type` which contains the type URL. Example:
     *
     *     package google.profile;
     *     message Person {
     *       string first_name = 1;
     *       string last_name = 2;
     *     }
     *
     *     {
     *       "@type": "type.googleapis.com/google.profile.Person",
     *       "firstName": <string>,
     *       "lastName": <string>
     *     }
     *
     * If the embedded message type is well-known and has a custom JSON
     * representation, that representation will be embedded adding a field
     * `value` which holds the custom JSON in addition to the `@type`
     * field. Example (for message [google.protobuf.Duration][]):
     *
     *     {
     *       "@type": "type.googleapis.com/google.protobuf.Duration",
     *       "value": "1.212s"
     *     }
     */
    consensusState?: ProtobufAny;
}
/**
* Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
the same. However some consensus algorithms may choose to reset the
height in certain conditions e.g. hard forks, state-machine breaking changes
In these cases, the RevisionNumber is incremented so that height continues to
be monitonically increasing even as the RevisionHeight gets reset
*/
export interface V1Height {
    /** @format uint64 */
    revisionNumber?: string;
    /** @format uint64 */
    revisionHeight?: string;
}
/**
* IdentifiedClientState defines a client state with an additional client
identifier field.
*/
export interface V1IdentifiedClientState {
    clientId?: string;
    /**
     * `Any` contains an arbitrary serialized protocol buffer message along with a
     * URL that describes the type of the serialized message.
     *
     * Protobuf library provides support to pack/unpack Any values in the form
     * of utility functions or additional generated methods of the Any type.
     *
     * Example 1: Pack and unpack a message in C++.
     *
     *     Foo foo = ...;
     *     Any any;
     *     any.PackFrom(foo);
     *     ...
     *     if (any.UnpackTo(&foo)) {
     *       ...
     *     }
     *
     * Example 2: Pack and unpack a message in Java.
     *
     *     Foo foo = ...;
     *     Any any = Any.pack(foo);
     *     ...
     *     if (any.is(Foo.class)) {
     *       foo = any.unpack(Foo.class);
     *     }
     *
     *  Example 3: Pack and unpack a message in Python.
     *
     *     foo = Foo(...)
     *     any = Any()
     *     any.Pack(foo)
     *     ...
     *     if any.Is(Foo.DESCRIPTOR):
     *       any.Unpack(foo)
     *       ...
     *
     *  Example 4: Pack and unpack a message in Go
     *
     *      foo := &pb.Foo{...}
     *      any, err := ptypes.MarshalAny(foo)
     *      ...
     *      foo := &pb.Foo{}
     *      if err := ptypes.UnmarshalAny(any, foo); err != nil {
     *        ...
     *      }
     *
     * The pack methods provided by protobuf library will by default use
     * 'type.googleapis.com/full.type.name' as the type URL and the unpack
     * methods only use the fully qualified type name after the last '/'
     * in the type URL, for example "foo.bar.com/x/y.z" will yield type
     * name "y.z".
     *
     *
     * JSON
     * ====
     * The JSON representation of an `Any` value uses the regular
     * representation of the deserialized, embedded message, with an
     * additional field `@type` which contains the type URL. Example:
     *
     *     package google.profile;
     *     message Person {
     *       string first_name = 1;
     *       string last_name = 2;
     *     }
     *
     *     {
     *       "@type": "type.googleapis.com/google.profile.Person",
     *       "firstName": <string>,
     *       "lastName": <string>
     *     }
     *
     * If the embedded message type is well-known and has a custom JSON
     * representation, that representation will be embedded adding a field
     * `value` which holds the custom JSON in addition to the `@type`
     * field. Example (for message [google.protobuf.Duration][]):
     *
     *     {
     *       "@type": "type.googleapis.com/google.protobuf.Duration",
     *       "value": "1.212s"
     *     }
     */
    clientState?: ProtobufAny;
}
/**
 * MsgCreateClientResponse defines the Msg/CreateClient response type.
 */
export declare type V1MsgCreateClientResponse = object;
/**
 * MsgSubmitMisbehaviourResponse defines the Msg/SubmitMisbehaviour response type.
 */
export declare type V1MsgSubmitMisbehaviourResponse = object;
/**
 * MsgUpdateClientResponse defines the Msg/UpdateClient response type.
 */
export declare type V1MsgUpdateClientResponse = object;
/**
 * MsgUpgradeClientResponse defines the Msg/UpgradeClient response type.
 */
export declare type V1MsgUpgradeClientResponse = object;
/**
 * Params defines the set of IBC light client parameters.
 */
export interface V1Params {
    /** allowed_clients defines the list of allowed client state types. */
    allowedClients?: string[];
}
/**
 * QueryClientParamsResponse is the response type for the Query/ClientParams RPC method.
 */
export interface V1QueryClientParamsResponse {
    /** params defines the parameters of the module. */
    params?: V1Params;
}
/**
* QueryClientStateResponse is the response type for the Query/ClientState RPC
method. Besides the client state, it includes a proof and the height from
which the proof was retrieved.
*/
export interface V1QueryClientStateResponse {
    /**
     * `Any` contains an arbitrary serialized protocol buffer message along with a
     * URL that describes the type of the serialized message.
     *
     * Protobuf library provides support to pack/unpack Any values in the form
     * of utility functions or additional generated methods of the Any type.
     *
     * Example 1: Pack and unpack a message in C++.
     *
     *     Foo foo = ...;
     *     Any any;
     *     any.PackFrom(foo);
     *     ...
     *     if (any.UnpackTo(&foo)) {
     *       ...
     *     }
     *
     * Example 2: Pack and unpack a message in Java.
     *
     *     Foo foo = ...;
     *     Any any = Any.pack(foo);
     *     ...
     *     if (any.is(Foo.class)) {
     *       foo = any.unpack(Foo.class);
     *     }
     *
     *  Example 3: Pack and unpack a message in Python.
     *
     *     foo = Foo(...)
     *     any = Any()
     *     any.Pack(foo)
     *     ...
     *     if any.Is(Foo.DESCRIPTOR):
     *       any.Unpack(foo)
     *       ...
     *
     *  Example 4: Pack and unpack a message in Go
     *
     *      foo := &pb.Foo{...}
     *      any, err := ptypes.MarshalAny(foo)
     *      ...
     *      foo := &pb.Foo{}
     *      if err := ptypes.UnmarshalAny(any, foo); err != nil {
     *        ...
     *      }
     *
     * The pack methods provided by protobuf library will by default use
     * 'type.googleapis.com/full.type.name' as the type URL and the unpack
     * methods only use the fully qualified type name after the last '/'
     * in the type URL, for example "foo.bar.com/x/y.z" will yield type
     * name "y.z".
     *
     *
     * JSON
     * ====
     * The JSON representation of an `Any` value uses the regular
     * representation of the deserialized, embedded message, with an
     * additional field `@type` which contains the type URL. Example:
     *
     *     package google.profile;
     *     message Person {
     *       string first_name = 1;
     *       string last_name = 2;
     *     }
     *
     *     {
     *       "@type": "type.googleapis.com/google.profile.Person",
     *       "firstName": <string>,
     *       "lastName": <string>
     *     }
     *
     * If the embedded message type is well-known and has a custom JSON
     * representation, that representation will be embedded adding a field
     * `value` which holds the custom JSON in addition to the `@type`
     * field. Example (for message [google.protobuf.Duration][]):
     *
     *     {
     *       "@type": "type.googleapis.com/google.protobuf.Duration",
     *       "value": "1.212s"
     *     }
     */
    clientState?: ProtobufAny;
    /** @format byte */
    proof?: string;
    /**
     * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
     * the same. However some consensus algorithms may choose to reset the
     * height in certain conditions e.g. hard forks, state-machine breaking changes
     * In these cases, the RevisionNumber is incremented so that height continues to
     * be monitonically increasing even as the RevisionHeight gets reset
     */
    proofHeight?: V1Height;
}
/**
* QueryClientStatesResponse is the response type for the Query/ClientStates RPC
method.
*/
export interface V1QueryClientStatesResponse {
    /** list of stored ClientStates of the chain. */
    clientStates?: V1IdentifiedClientState[];
    /**
     * PageResponse is to be embedded in gRPC response messages where the
     * corresponding request message has used PageRequest.
     *
     *  message SomeResponse {
     *          repeated Bar results = 1;
     *          PageResponse page = 2;
     *  }
     */
    pagination?: V1Beta1PageResponse;
}
export interface V1QueryConsensusStateResponse {
    /**
     * `Any` contains an arbitrary serialized protocol buffer message along with a
     * URL that describes the type of the serialized message.
     *
     * Protobuf library provides support to pack/unpack Any values in the form
     * of utility functions or additional generated methods of the Any type.
     *
     * Example 1: Pack and unpack a message in C++.
     *
     *     Foo foo = ...;
     *     Any any;
     *     any.PackFrom(foo);
     *     ...
     *     if (any.UnpackTo(&foo)) {
     *       ...
     *     }
     *
     * Example 2: Pack and unpack a message in Java.
     *
     *     Foo foo = ...;
     *     Any any = Any.pack(foo);
     *     ...
     *     if (any.is(Foo.class)) {
     *       foo = any.unpack(Foo.class);
     *     }
     *
     *  Example 3: Pack and unpack a message in Python.
     *
     *     foo = Foo(...)
     *     any = Any()
     *     any.Pack(foo)
     *     ...
     *     if any.Is(Foo.DESCRIPTOR):
     *       any.Unpack(foo)
     *       ...
     *
     *  Example 4: Pack and unpack a message in Go
     *
     *      foo := &pb.Foo{...}
     *      any, err := ptypes.MarshalAny(foo)
     *      ...
     *      foo := &pb.Foo{}
     *      if err := ptypes.UnmarshalAny(any, foo); err != nil {
     *        ...
     *      }
     *
     * The pack methods provided by protobuf library will by default use
     * 'type.googleapis.com/full.type.name' as the type URL and the unpack
     * methods only use the fully qualified type name after the last '/'
     * in the type URL, for example "foo.bar.com/x/y.z" will yield type
     * name "y.z".
     *
     *
     * JSON
     * ====
     * The JSON representation of an `Any` value uses the regular
     * representation of the deserialized, embedded message, with an
     * additional field `@type` which contains the type URL. Example:
     *
     *     package google.profile;
     *     message Person {
     *       string first_name = 1;
     *       string last_name = 2;
     *     }
     *
     *     {
     *       "@type": "type.googleapis.com/google.profile.Person",
     *       "firstName": <string>,
     *       "lastName": <string>
     *     }
     *
     * If the embedded message type is well-known and has a custom JSON
     * representation, that representation will be embedded adding a field
     * `value` which holds the custom JSON in addition to the `@type`
     * field. Example (for message [google.protobuf.Duration][]):
     *
     *     {
     *       "@type": "type.googleapis.com/google.protobuf.Duration",
     *       "value": "1.212s"
     *     }
     */
    consensusState?: ProtobufAny;
    /** @format byte */
    proof?: string;
    /**
     * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
     * the same. However some consensus algorithms may choose to reset the
     * height in certain conditions e.g. hard forks, state-machine breaking changes
     * In these cases, the RevisionNumber is incremented so that height continues to
     * be monitonically increasing even as the RevisionHeight gets reset
     */
    proofHeight?: V1Height;
}
export interface V1QueryConsensusStatesResponse {
    consensusStates?: V1ConsensusStateWithHeight[];
    /**
     * PageResponse is to be embedded in gRPC response messages where the
     * corresponding request message has used PageRequest.
     *
     *  message SomeResponse {
     *          repeated Bar results = 1;
     *          PageResponse page = 2;
     *  }
     */
    pagination?: V1Beta1PageResponse;
}
/**
* message SomeRequest {
         Foo some_parameter = 1;
         PageRequest pagination = 2;
 }
*/
export interface V1Beta1PageRequest {
    /**
     * key is a value returned in PageResponse.next_key to begin
     * querying the next page most efficiently. Only one of offset or key
     * should be set.
     * @format byte
     */
    key?: string;
    /**
     * offset is a numeric offset that can be used when key is unavailable.
     * It is less efficient than using key. Only one of offset or key should
     * be set.
     * @format uint64
     */
    offset?: string;
    /**
     * limit is the total number of results to be returned in the result page.
     * If left empty it will default to a value to be set by each app.
     * @format uint64
     */
    limit?: string;
    /**
     * count_total is set to true  to indicate that the result set should include
     * a count of the total number of items available for pagination in UIs.
     * count_total is only respected when offset is used. It is ignored when key
     * is set.
     */
    countTotal?: boolean;
}
/**
* PageResponse is to be embedded in gRPC response messages where the
corresponding request message has used PageRequest.

 message SomeResponse {
         repeated Bar results = 1;
         PageResponse page = 2;
 }
*/
export interface V1Beta1PageResponse {
    /** @format byte */
    nextKey?: string;
    /** @format uint64 */
    total?: string;
}
export declare type QueryParamsType = Record<string | number, any>;
export declare type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;
export interface FullRequestParams extends Omit<RequestInit, "body"> {
    /** set parameter to `true` for call `securityWorker` for this request */
    secure?: boolean;
    /** request path */
    path: string;
    /** content type of request body */
    type?: ContentType;
    /** query params */
    query?: QueryParamsType;
    /** format of response (i.e. response.json() -> format: "json") */
    format?: keyof Omit<Body, "body" | "bodyUsed">;
    /** request body */
    body?: unknown;
    /** base url */
    baseUrl?: string;
    /** request cancellation token */
    cancelToken?: CancelToken;
}
export declare type RequestParams = Omit<FullRequestParams, "body" | "method" | "query" | "path">;
export interface ApiConfig<SecurityDataType = unknown> {
    baseUrl?: string;
    baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
    securityWorker?: (securityData: SecurityDataType) => RequestParams | void;
}
export interface HttpResponse<D extends unknown, E extends unknown = unknown> extends Response {
    data: D;
    error: E;
}
declare type CancelToken = Symbol | string | number;
export declare enum ContentType {
    Json = "application/json",
    FormData = "multipart/form-data",
    UrlEncoded = "application/x-www-form-urlencoded"
}
export declare class HttpClient<SecurityDataType = unknown> {
    baseUrl: string;
    private securityData;
    private securityWorker;
    private abortControllers;
    private baseApiParams;
    constructor(apiConfig?: ApiConfig<SecurityDataType>);
    setSecurityData: (data: SecurityDataType) => void;
    private addQueryParam;
    protected toQueryString(rawQuery?: QueryParamsType): string;
    protected addQueryParams(rawQuery?: QueryParamsType): string;
    private contentFormatters;
    private mergeRequestParams;
    private createAbortSignal;
    abortRequest: (cancelToken: CancelToken) => void;
    request: <T = any, E = any>({ body, secure, path, type, query, format, baseUrl, cancelToken, ...params }: FullRequestParams) => Promise<HttpResponse<T, E>>;
}
/**
 * @title ibc/core/client/v1/client.proto
 * @version version not set
 */
export declare class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
    /**
     * No description
     *
     * @tags Query
     * @name QueryClientParams
     * @summary ClientParams queries all parameters of the ibc client.
     * @request GET:/ibc/client/v1beta1/params
     */
    queryClientParams: (params?: RequestParams) => Promise<HttpResponse<V1QueryClientParamsResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryClientStates
     * @summary ClientStates queries all the IBC light clients of a chain.
     * @request GET:/ibc/core/client/v1beta1/client_states
     */
    queryClientStates: (query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<V1QueryClientStatesResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryClientState
     * @summary ClientState queries an IBC light client.
     * @request GET:/ibc/core/client/v1beta1/client_states/{clientId}
     */
    queryClientState: (clientId: string, params?: RequestParams) => Promise<HttpResponse<V1QueryClientStateResponse, RpcStatus>>;
    /**
   * No description
   *
   * @tags Query
   * @name QueryConsensusStates
   * @summary ConsensusStates queries all the consensus state associated with a given
  client.
   * @request GET:/ibc/core/client/v1beta1/consensus_states/{clientId}
   */
    queryConsensusStates: (clientId: string, query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<V1QueryConsensusStatesResponse, RpcStatus>>;
    /**
   * No description
   *
   * @tags Query
   * @name QueryConsensusState
   * @summary ConsensusState queries a consensus state associated with a client state at
  a given height.
   * @request GET:/ibc/core/client/v1beta1/consensus_states/{clientId}/revision/{revisionNumber}/height/{revisionHeight}
   */
    queryConsensusState: (clientId: string, revisionNumber: string, revisionHeight: string, query?: {
        latestHeight?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<V1QueryConsensusStateResponse, RpcStatus>>;
}
export {};
