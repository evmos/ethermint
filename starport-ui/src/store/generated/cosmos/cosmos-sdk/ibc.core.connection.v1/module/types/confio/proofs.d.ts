import { Writer, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "ics23";
export declare enum HashOp {
    /** NO_HASH - NO_HASH is the default if no data passed. Note this is an illegal argument some places. */
    NO_HASH = 0,
    SHA256 = 1,
    SHA512 = 2,
    KECCAK = 3,
    RIPEMD160 = 4,
    /** BITCOIN - ripemd160(sha256(x)) */
    BITCOIN = 5,
    UNRECOGNIZED = -1
}
export declare function hashOpFromJSON(object: any): HashOp;
export declare function hashOpToJSON(object: HashOp): string;
/**
 * LengthOp defines how to process the key and value of the LeafOp
 * to include length information. After encoding the length with the given
 * algorithm, the length will be prepended to the key and value bytes.
 * (Each one with it's own encoded length)
 */
export declare enum LengthOp {
    /** NO_PREFIX - NO_PREFIX don't include any length info */
    NO_PREFIX = 0,
    /** VAR_PROTO - VAR_PROTO uses protobuf (and go-amino) varint encoding of the length */
    VAR_PROTO = 1,
    /** VAR_RLP - VAR_RLP uses rlp int encoding of the length */
    VAR_RLP = 2,
    /** FIXED32_BIG - FIXED32_BIG uses big-endian encoding of the length as a 32 bit integer */
    FIXED32_BIG = 3,
    /** FIXED32_LITTLE - FIXED32_LITTLE uses little-endian encoding of the length as a 32 bit integer */
    FIXED32_LITTLE = 4,
    /** FIXED64_BIG - FIXED64_BIG uses big-endian encoding of the length as a 64 bit integer */
    FIXED64_BIG = 5,
    /** FIXED64_LITTLE - FIXED64_LITTLE uses little-endian encoding of the length as a 64 bit integer */
    FIXED64_LITTLE = 6,
    /** REQUIRE_32_BYTES - REQUIRE_32_BYTES is like NONE, but will fail if the input is not exactly 32 bytes (sha256 output) */
    REQUIRE_32_BYTES = 7,
    /** REQUIRE_64_BYTES - REQUIRE_64_BYTES is like NONE, but will fail if the input is not exactly 64 bytes (sha512 output) */
    REQUIRE_64_BYTES = 8,
    UNRECOGNIZED = -1
}
export declare function lengthOpFromJSON(object: any): LengthOp;
export declare function lengthOpToJSON(object: LengthOp): string;
/**
 * ExistenceProof takes a key and a value and a set of steps to perform on it.
 * The result of peforming all these steps will provide a "root hash", which can
 * be compared to the value in a header.
 *
 * Since it is computationally infeasible to produce a hash collission for any of the used
 * cryptographic hash functions, if someone can provide a series of operations to transform
 * a given key and value into a root hash that matches some trusted root, these key and values
 * must be in the referenced merkle tree.
 *
 * The only possible issue is maliablity in LeafOp, such as providing extra prefix data,
 * which should be controlled by a spec. Eg. with lengthOp as NONE,
 * prefix = FOO, key = BAR, value = CHOICE
 * and
 * prefix = F, key = OOBAR, value = CHOICE
 * would produce the same value.
 *
 * With LengthOp this is tricker but not impossible. Which is why the "leafPrefixEqual" field
 * in the ProofSpec is valuable to prevent this mutability. And why all trees should
 * length-prefix the data before hashing it.
 */
export interface ExistenceProof {
    key: Uint8Array;
    value: Uint8Array;
    leaf: LeafOp | undefined;
    path: InnerOp[];
}
/**
 * NonExistenceProof takes a proof of two neighbors, one left of the desired key,
 * one right of the desired key. If both proofs are valid AND they are neighbors,
 * then there is no valid proof for the given key.
 */
export interface NonExistenceProof {
    /** TODO: remove this as unnecessary??? we prove a range */
    key: Uint8Array;
    left: ExistenceProof | undefined;
    right: ExistenceProof | undefined;
}
/** CommitmentProof is either an ExistenceProof or a NonExistenceProof, or a Batch of such messages */
export interface CommitmentProof {
    exist: ExistenceProof | undefined;
    nonexist: NonExistenceProof | undefined;
    batch: BatchProof | undefined;
    compressed: CompressedBatchProof | undefined;
}
/**
 * LeafOp represents the raw key-value data we wish to prove, and
 * must be flexible to represent the internal transformation from
 * the original key-value pairs into the basis hash, for many existing
 * merkle trees.
 *
 * key and value are passed in. So that the signature of this operation is:
 * leafOp(key, value) -> output
 *
 * To process this, first prehash the keys and values if needed (ANY means no hash in this case):
 * hkey = prehashKey(key)
 * hvalue = prehashValue(value)
 *
 * Then combine the bytes, and hash it
 * output = hash(prefix || length(hkey) || hkey || length(hvalue) || hvalue)
 */
export interface LeafOp {
    hash: HashOp;
    prehashKey: HashOp;
    prehashValue: HashOp;
    length: LengthOp;
    /**
     * prefix is a fixed bytes that may optionally be included at the beginning to differentiate
     * a leaf node from an inner node.
     */
    prefix: Uint8Array;
}
/**
 * InnerOp represents a merkle-proof step that is not a leaf.
 * It represents concatenating two children and hashing them to provide the next result.
 *
 * The result of the previous step is passed in, so the signature of this op is:
 * innerOp(child) -> output
 *
 * The result of applying InnerOp should be:
 * output = op.hash(op.prefix || child || op.suffix)
 *
 * where the || operator is concatenation of binary data,
 * and child is the result of hashing all the tree below this step.
 *
 * Any special data, like prepending child with the length, or prepending the entire operation with
 * some value to differentiate from leaf nodes, should be included in prefix and suffix.
 * If either of prefix or suffix is empty, we just treat it as an empty string
 */
export interface InnerOp {
    hash: HashOp;
    prefix: Uint8Array;
    suffix: Uint8Array;
}
/**
 * ProofSpec defines what the expected parameters are for a given proof type.
 * This can be stored in the client and used to validate any incoming proofs.
 *
 * verify(ProofSpec, Proof) -> Proof | Error
 *
 * As demonstrated in tests, if we don't fix the algorithm used to calculate the
 * LeafHash for a given tree, there are many possible key-value pairs that can
 * generate a given hash (by interpretting the preimage differently).
 * We need this for proper security, requires client knows a priori what
 * tree format server uses. But not in code, rather a configuration object.
 */
export interface ProofSpec {
    /**
     * any field in the ExistenceProof must be the same as in this spec.
     * except Prefix, which is just the first bytes of prefix (spec can be longer)
     */
    leafSpec: LeafOp | undefined;
    innerSpec: InnerSpec | undefined;
    /** max_depth (if > 0) is the maximum number of InnerOps allowed (mainly for fixed-depth tries) */
    maxDepth: number;
    /** min_depth (if > 0) is the minimum number of InnerOps allowed (mainly for fixed-depth tries) */
    minDepth: number;
}
/**
 * InnerSpec contains all store-specific structure info to determine if two proofs from a
 * given store are neighbors.
 *
 * This enables:
 *
 * isLeftMost(spec: InnerSpec, op: InnerOp)
 * isRightMost(spec: InnerSpec, op: InnerOp)
 * isLeftNeighbor(spec: InnerSpec, left: InnerOp, right: InnerOp)
 */
export interface InnerSpec {
    /**
     * Child order is the ordering of the children node, must count from 0
     * iavl tree is [0, 1] (left then right)
     * merk is [0, 2, 1] (left, right, here)
     */
    childOrder: number[];
    childSize: number;
    minPrefixLength: number;
    maxPrefixLength: number;
    /** empty child is the prehash image that is used when one child is nil (eg. 20 bytes of 0) */
    emptyChild: Uint8Array;
    /** hash is the algorithm that must be used for each InnerOp */
    hash: HashOp;
}
/** BatchProof is a group of multiple proof types than can be compressed */
export interface BatchProof {
    entries: BatchEntry[];
}
/** Use BatchEntry not CommitmentProof, to avoid recursion */
export interface BatchEntry {
    exist: ExistenceProof | undefined;
    nonexist: NonExistenceProof | undefined;
}
export interface CompressedBatchProof {
    entries: CompressedBatchEntry[];
    lookupInners: InnerOp[];
}
/** Use BatchEntry not CommitmentProof, to avoid recursion */
export interface CompressedBatchEntry {
    exist: CompressedExistenceProof | undefined;
    nonexist: CompressedNonExistenceProof | undefined;
}
export interface CompressedExistenceProof {
    key: Uint8Array;
    value: Uint8Array;
    leaf: LeafOp | undefined;
    /** these are indexes into the lookup_inners table in CompressedBatchProof */
    path: number[];
}
export interface CompressedNonExistenceProof {
    /** TODO: remove this as unnecessary??? we prove a range */
    key: Uint8Array;
    left: CompressedExistenceProof | undefined;
    right: CompressedExistenceProof | undefined;
}
export declare const ExistenceProof: {
    encode(message: ExistenceProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ExistenceProof;
    fromJSON(object: any): ExistenceProof;
    toJSON(message: ExistenceProof): unknown;
    fromPartial(object: DeepPartial<ExistenceProof>): ExistenceProof;
};
export declare const NonExistenceProof: {
    encode(message: NonExistenceProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): NonExistenceProof;
    fromJSON(object: any): NonExistenceProof;
    toJSON(message: NonExistenceProof): unknown;
    fromPartial(object: DeepPartial<NonExistenceProof>): NonExistenceProof;
};
export declare const CommitmentProof: {
    encode(message: CommitmentProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): CommitmentProof;
    fromJSON(object: any): CommitmentProof;
    toJSON(message: CommitmentProof): unknown;
    fromPartial(object: DeepPartial<CommitmentProof>): CommitmentProof;
};
export declare const LeafOp: {
    encode(message: LeafOp, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): LeafOp;
    fromJSON(object: any): LeafOp;
    toJSON(message: LeafOp): unknown;
    fromPartial(object: DeepPartial<LeafOp>): LeafOp;
};
export declare const InnerOp: {
    encode(message: InnerOp, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): InnerOp;
    fromJSON(object: any): InnerOp;
    toJSON(message: InnerOp): unknown;
    fromPartial(object: DeepPartial<InnerOp>): InnerOp;
};
export declare const ProofSpec: {
    encode(message: ProofSpec, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): ProofSpec;
    fromJSON(object: any): ProofSpec;
    toJSON(message: ProofSpec): unknown;
    fromPartial(object: DeepPartial<ProofSpec>): ProofSpec;
};
export declare const InnerSpec: {
    encode(message: InnerSpec, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): InnerSpec;
    fromJSON(object: any): InnerSpec;
    toJSON(message: InnerSpec): unknown;
    fromPartial(object: DeepPartial<InnerSpec>): InnerSpec;
};
export declare const BatchProof: {
    encode(message: BatchProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): BatchProof;
    fromJSON(object: any): BatchProof;
    toJSON(message: BatchProof): unknown;
    fromPartial(object: DeepPartial<BatchProof>): BatchProof;
};
export declare const BatchEntry: {
    encode(message: BatchEntry, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): BatchEntry;
    fromJSON(object: any): BatchEntry;
    toJSON(message: BatchEntry): unknown;
    fromPartial(object: DeepPartial<BatchEntry>): BatchEntry;
};
export declare const CompressedBatchProof: {
    encode(message: CompressedBatchProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): CompressedBatchProof;
    fromJSON(object: any): CompressedBatchProof;
    toJSON(message: CompressedBatchProof): unknown;
    fromPartial(object: DeepPartial<CompressedBatchProof>): CompressedBatchProof;
};
export declare const CompressedBatchEntry: {
    encode(message: CompressedBatchEntry, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): CompressedBatchEntry;
    fromJSON(object: any): CompressedBatchEntry;
    toJSON(message: CompressedBatchEntry): unknown;
    fromPartial(object: DeepPartial<CompressedBatchEntry>): CompressedBatchEntry;
};
export declare const CompressedExistenceProof: {
    encode(message: CompressedExistenceProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): CompressedExistenceProof;
    fromJSON(object: any): CompressedExistenceProof;
    toJSON(message: CompressedExistenceProof): unknown;
    fromPartial(object: DeepPartial<CompressedExistenceProof>): CompressedExistenceProof;
};
export declare const CompressedNonExistenceProof: {
    encode(message: CompressedNonExistenceProof, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): CompressedNonExistenceProof;
    fromJSON(object: any): CompressedNonExistenceProof;
    toJSON(message: CompressedNonExistenceProof): unknown;
    fromPartial(object: DeepPartial<CompressedNonExistenceProof>): CompressedNonExistenceProof;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
