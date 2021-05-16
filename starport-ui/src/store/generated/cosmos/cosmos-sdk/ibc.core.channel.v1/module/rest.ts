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
* Channel defines pipeline for exactly-once packet delivery between specific
modules on separate blockchains, which has at least one end capable of
sending packets and one end capable of receiving packets.
*/
export interface V1Channel {
  /**
   * State defines if a channel is in one of the following states:
   * CLOSED, INIT, TRYOPEN, OPEN or UNINITIALIZED.
   *
   *  - STATE_UNINITIALIZED_UNSPECIFIED: Default State
   *  - STATE_INIT: A channel has just started the opening handshake.
   *  - STATE_TRYOPEN: A channel has acknowledged the handshake step on the counterparty chain.
   *  - STATE_OPEN: A channel has completed the handshake. Open channels are
   * ready to send and receive packets.
   *  - STATE_CLOSED: A channel has been closed and can no longer be used to send or receive
   * packets.
   */
  state?: V1State;

  /**
   * - ORDER_NONE_UNSPECIFIED: zero-value for channel ordering
   *  - ORDER_UNORDERED: packets can be delivered in any order, which may differ from the order in
   * which they were sent.
   *  - ORDER_ORDERED: packets are delivered exactly in the order which they were sent
   */
  ordering?: V1Order;
  counterparty?: V1Counterparty;
  connectionHops?: string[];
  version?: string;
}

export interface V1Counterparty {
  /** port on the counterparty chain which owns the other end of the channel. */
  portId?: string;
  channelId?: string;
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
* IdentifiedChannel defines a channel with additional port and channel
identifier fields.
*/
export interface V1IdentifiedChannel {
  /**
   * State defines if a channel is in one of the following states:
   * CLOSED, INIT, TRYOPEN, OPEN or UNINITIALIZED.
   *
   *  - STATE_UNINITIALIZED_UNSPECIFIED: Default State
   *  - STATE_INIT: A channel has just started the opening handshake.
   *  - STATE_TRYOPEN: A channel has acknowledged the handshake step on the counterparty chain.
   *  - STATE_OPEN: A channel has completed the handshake. Open channels are
   * ready to send and receive packets.
   *  - STATE_CLOSED: A channel has been closed and can no longer be used to send or receive
   * packets.
   */
  state?: V1State;

  /**
   * - ORDER_NONE_UNSPECIFIED: zero-value for channel ordering
   *  - ORDER_UNORDERED: packets can be delivered in any order, which may differ from the order in
   * which they were sent.
   *  - ORDER_ORDERED: packets are delivered exactly in the order which they were sent
   */
  ordering?: V1Order;
  counterparty?: V1Counterparty;
  connectionHops?: string[];
  version?: string;
  portId?: string;
  channelId?: string;
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
 * MsgAcknowledgementResponse defines the Msg/Acknowledgement response type.
 */
export type V1MsgAcknowledgementResponse = object;

/**
 * MsgChannelCloseConfirmResponse defines the Msg/ChannelCloseConfirm response type.
 */
export type V1MsgChannelCloseConfirmResponse = object;

/**
 * MsgChannelCloseInitResponse defines the Msg/ChannelCloseInit response type.
 */
export type V1MsgChannelCloseInitResponse = object;

/**
 * MsgChannelOpenAckResponse defines the Msg/ChannelOpenAck response type.
 */
export type V1MsgChannelOpenAckResponse = object;

/**
 * MsgChannelOpenConfirmResponse defines the Msg/ChannelOpenConfirm response type.
 */
export type V1MsgChannelOpenConfirmResponse = object;

/**
 * MsgChannelOpenInitResponse defines the Msg/ChannelOpenInit response type.
 */
export type V1MsgChannelOpenInitResponse = object;

/**
 * MsgChannelOpenTryResponse defines the Msg/ChannelOpenTry response type.
 */
export type V1MsgChannelOpenTryResponse = object;

/**
 * MsgRecvPacketResponse defines the Msg/RecvPacket response type.
 */
export type V1MsgRecvPacketResponse = object;

/**
 * MsgTimeoutOnCloseResponse defines the Msg/TimeoutOnClose response type.
 */
export type V1MsgTimeoutOnCloseResponse = object;

/**
 * MsgTimeoutResponse defines the Msg/Timeout response type.
 */
export type V1MsgTimeoutResponse = object;

/**
* - ORDER_NONE_UNSPECIFIED: zero-value for channel ordering
 - ORDER_UNORDERED: packets can be delivered in any order, which may differ from the order in
which they were sent.
 - ORDER_ORDERED: packets are delivered exactly in the order which they were sent
*/
export enum V1Order {
  ORDER_NONE_UNSPECIFIED = "ORDER_NONE_UNSPECIFIED",
  ORDER_UNORDERED = "ORDER_UNORDERED",
  ORDER_ORDERED = "ORDER_ORDERED",
}

export interface V1Packet {
  /**
   * number corresponds to the order of sends and receives, where a Packet
   * with an earlier sequence number must be sent and received before a Packet
   * with a later sequence number.
   * @format uint64
   */
  sequence?: string;

  /** identifies the port on the sending chain. */
  sourcePort?: string;

  /** identifies the channel end on the sending chain. */
  sourceChannel?: string;

  /** identifies the port on the receiving chain. */
  destinationPort?: string;

  /** identifies the channel end on the receiving chain. */
  destinationChannel?: string;

  /** @format byte */
  data?: string;

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  timeoutHeight?: V1Height;

  /** @format uint64 */
  timeoutTimestamp?: string;
}

/**
* PacketState defines the generic type necessary to retrieve and store
packet commitments, acknowledgements, and receipts.
Caller is responsible for knowing the context necessary to interpret this
state as a commitment, acknowledgement, or a receipt.
*/
export interface V1PacketState {
  /** channel port identifier. */
  portId?: string;

  /** channel unique identifier. */
  channelId?: string;

  /**
   * packet sequence.
   * @format uint64
   */
  sequence?: string;

  /**
   * embedded data that represents packet state.
   * @format byte
   */
  data?: string;
}

export interface V1QueryChannelClientStateResponse {
  /**
   * IdentifiedClientState defines a client state with an additional client
   * identifier field.
   */
  identifiedClientState?: V1IdentifiedClientState;

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

export interface V1QueryChannelConsensusStateResponse {
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
  clientId?: string;

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
* QueryChannelResponse is the response type for the Query/Channel RPC method.
Besides the Channel end, it includes a proof and the height from which the
proof was retrieved.
*/
export interface V1QueryChannelResponse {
  /**
   * Channel defines pipeline for exactly-once packet delivery between specific
   * modules on separate blockchains, which has at least one end capable of
   * sending packets and one end capable of receiving packets.
   */
  channel?: V1Channel;

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
 * QueryChannelsResponse is the response type for the Query/Channels RPC method.
 */
export interface V1QueryChannelsResponse {
  /** list of stored channels of the chain. */
  channels?: V1IdentifiedChannel[];

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

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  height?: V1Height;
}

export interface V1QueryConnectionChannelsResponse {
  /** list of channels associated with a connection. */
  channels?: V1IdentifiedChannel[];

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

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  height?: V1Height;
}

export interface V1QueryNextSequenceReceiveResponse {
  /** @format uint64 */
  nextSequenceReceive?: string;

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

export interface V1QueryPacketAcknowledgementResponse {
  /** @format byte */
  acknowledgement?: string;

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

export interface V1QueryPacketAcknowledgementsResponse {
  acknowledgements?: V1PacketState[];

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

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  height?: V1Height;
}

export interface V1QueryPacketCommitmentResponse {
  /** @format byte */
  commitment?: string;

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

export interface V1QueryPacketCommitmentsResponse {
  commitments?: V1PacketState[];

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

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  height?: V1Height;
}

export interface V1QueryPacketReceiptResponse {
  received?: boolean;

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

export interface V1QueryUnreceivedAcksResponse {
  sequences?: string[];

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  height?: V1Height;
}

export interface V1QueryUnreceivedPacketsResponse {
  sequences?: string[];

  /**
   * Normally the RevisionHeight is incremented at each height while keeping RevisionNumber
   * the same. However some consensus algorithms may choose to reset the
   * height in certain conditions e.g. hard forks, state-machine breaking changes
   * In these cases, the RevisionNumber is incremented so that height continues to
   * be monitonically increasing even as the RevisionHeight gets reset
   */
  height?: V1Height;
}

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
export enum V1State {
  STATE_UNINITIALIZED_UNSPECIFIED = "STATE_UNINITIALIZED_UNSPECIFIED",
  STATE_INIT = "STATE_INIT",
  STATE_TRYOPEN = "STATE_TRYOPEN",
  STATE_OPEN = "STATE_OPEN",
  STATE_CLOSED = "STATE_CLOSED",
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

export type QueryParamsType = Record<string | number, any>;
export type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;

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

export type RequestParams = Omit<FullRequestParams, "body" | "method" | "query" | "path">;

export interface ApiConfig<SecurityDataType = unknown> {
  baseUrl?: string;
  baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
  securityWorker?: (securityData: SecurityDataType) => RequestParams | void;
}

export interface HttpResponse<D extends unknown, E extends unknown = unknown> extends Response {
  data: D;
  error: E;
}

type CancelToken = Symbol | string | number;

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
}

export class HttpClient<SecurityDataType = unknown> {
  public baseUrl: string = "";
  private securityData: SecurityDataType = null as any;
  private securityWorker: null | ApiConfig<SecurityDataType>["securityWorker"] = null;
  private abortControllers = new Map<CancelToken, AbortController>();

  private baseApiParams: RequestParams = {
    credentials: "same-origin",
    headers: {},
    redirect: "follow",
    referrerPolicy: "no-referrer",
  };

  constructor(apiConfig: ApiConfig<SecurityDataType> = {}) {
    Object.assign(this, apiConfig);
  }

  public setSecurityData = (data: SecurityDataType) => {
    this.securityData = data;
  };

  private addQueryParam(query: QueryParamsType, key: string) {
    const value = query[key];

    return (
      encodeURIComponent(key) +
      "=" +
      encodeURIComponent(Array.isArray(value) ? value.join(",") : typeof value === "number" ? value : `${value}`)
    );
  }

  protected toQueryString(rawQuery?: QueryParamsType): string {
    const query = rawQuery || {};
    const keys = Object.keys(query).filter((key) => "undefined" !== typeof query[key]);
    return keys
      .map((key) =>
        typeof query[key] === "object" && !Array.isArray(query[key])
          ? this.toQueryString(query[key] as QueryParamsType)
          : this.addQueryParam(query, key),
      )
      .join("&");
  }

  protected addQueryParams(rawQuery?: QueryParamsType): string {
    const queryString = this.toQueryString(rawQuery);
    return queryString ? `?${queryString}` : "";
  }

  private contentFormatters: Record<ContentType, (input: any) => any> = {
    [ContentType.Json]: (input: any) =>
      input !== null && (typeof input === "object" || typeof input === "string") ? JSON.stringify(input) : input,
    [ContentType.FormData]: (input: any) =>
      Object.keys(input || {}).reduce((data, key) => {
        data.append(key, input[key]);
        return data;
      }, new FormData()),
    [ContentType.UrlEncoded]: (input: any) => this.toQueryString(input),
  };

  private mergeRequestParams(params1: RequestParams, params2?: RequestParams): RequestParams {
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

  private createAbortSignal = (cancelToken: CancelToken): AbortSignal | undefined => {
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

  public abortRequest = (cancelToken: CancelToken) => {
    const abortController = this.abortControllers.get(cancelToken);

    if (abortController) {
      abortController.abort();
      this.abortControllers.delete(cancelToken);
    }
  };

  public request = <T = any, E = any>({
    body,
    secure,
    path,
    type,
    query,
    format = "json",
    baseUrl,
    cancelToken,
    ...params
  }: FullRequestParams): Promise<HttpResponse<T, E>> => {
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
      const r = response as HttpResponse<T, E>;
      r.data = (null as unknown) as T;
      r.error = (null as unknown) as E;

      const data = await response[format]()
        .then((data) => {
          if (r.ok) {
            r.data = data;
          } else {
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

      if (!response.ok) throw data;
      return data;
    });
  };
}

/**
 * @title ibc/core/channel/v1/tx.proto
 * @version version not set
 */
export class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
  /**
   * No description
   *
   * @tags Query
   * @name QueryChannels
   * @summary Channels queries all the IBC channels of a chain.
   * @request GET:/ibc/core/channel/v1beta1/channels
   */
  queryChannels = (
    query?: {
      "pagination.key"?: string;
      "pagination.offset"?: string;
      "pagination.limit"?: string;
      "pagination.countTotal"?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.request<V1QueryChannelsResponse, RpcStatus>({
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
  queryChannel = (channelId: string, portId: string, params: RequestParams = {}) =>
    this.request<V1QueryChannelResponse, RpcStatus>({
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
  queryChannelClientState = (channelId: string, portId: string, params: RequestParams = {}) =>
    this.request<V1QueryChannelClientStateResponse, RpcStatus>({
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
  queryChannelConsensusState = (
    channelId: string,
    portId: string,
    revisionNumber: string,
    revisionHeight: string,
    params: RequestParams = {},
  ) =>
    this.request<V1QueryChannelConsensusStateResponse, RpcStatus>({
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
  queryNextSequenceReceive = (channelId: string, portId: string, params: RequestParams = {}) =>
    this.request<V1QueryNextSequenceReceiveResponse, RpcStatus>({
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
  queryPacketAcknowledgements = (
    channelId: string,
    portId: string,
    query?: {
      "pagination.key"?: string;
      "pagination.offset"?: string;
      "pagination.limit"?: string;
      "pagination.countTotal"?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.request<V1QueryPacketAcknowledgementsResponse, RpcStatus>({
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
  queryPacketAcknowledgement = (channelId: string, portId: string, sequence: string, params: RequestParams = {}) =>
    this.request<V1QueryPacketAcknowledgementResponse, RpcStatus>({
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
  queryPacketCommitments = (
    channelId: string,
    portId: string,
    query?: {
      "pagination.key"?: string;
      "pagination.offset"?: string;
      "pagination.limit"?: string;
      "pagination.countTotal"?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.request<V1QueryPacketCommitmentsResponse, RpcStatus>({
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
  queryUnreceivedAcks = (channelId: string, portId: string, packetAckSequences: string[], params: RequestParams = {}) =>
    this.request<V1QueryUnreceivedAcksResponse, RpcStatus>({
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
  queryUnreceivedPackets = (
    channelId: string,
    portId: string,
    packetCommitmentSequences: string[],
    params: RequestParams = {},
  ) =>
    this.request<V1QueryUnreceivedPacketsResponse, RpcStatus>({
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
  queryPacketCommitment = (channelId: string, portId: string, sequence: string, params: RequestParams = {}) =>
    this.request<V1QueryPacketCommitmentResponse, RpcStatus>({
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
  queryPacketReceipt = (channelId: string, portId: string, sequence: string, params: RequestParams = {}) =>
    this.request<V1QueryPacketReceiptResponse, RpcStatus>({
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
  queryConnectionChannels = (
    connection: string,
    query?: {
      "pagination.key"?: string;
      "pagination.offset"?: string;
      "pagination.limit"?: string;
      "pagination.countTotal"?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.request<V1QueryConnectionChannelsResponse, RpcStatus>({
      path: `/ibc/core/channel/v1beta1/connections/${connection}/channels`,
      method: "GET",
      query: query,
      format: "json",
      ...params,
    });
}
