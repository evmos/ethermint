/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Log, TxReceipt, TransactionLogs, Params, } from "../../../ethermint/evm/v1alpha1/evm";
export const protobufPackage = "ethermint.evm.v1alpha1";
const baseQueryAccountRequest = { address: "" };
export const QueryAccountRequest = {
    encode(message, writer = Writer.create()) {
        if (message.address !== "") {
            writer.uint32(10).string(message.address);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAccountRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.address = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryAccountRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = String(object.address);
        }
        else {
            message.address = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.address !== undefined && (obj.address = message.address);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAccountRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = object.address;
        }
        else {
            message.address = "";
        }
        return message;
    },
};
const baseQueryAccountResponse = { balance: "", nonce: 0 };
export const QueryAccountResponse = {
    encode(message, writer = Writer.create()) {
        if (message.balance !== "") {
            writer.uint32(10).string(message.balance);
        }
        if (message.codeHash.length !== 0) {
            writer.uint32(18).bytes(message.codeHash);
        }
        if (message.nonce !== 0) {
            writer.uint32(24).uint64(message.nonce);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAccountResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.balance = reader.string();
                    break;
                case 2:
                    message.codeHash = reader.bytes();
                    break;
                case 3:
                    message.nonce = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryAccountResponse };
        if (object.balance !== undefined && object.balance !== null) {
            message.balance = String(object.balance);
        }
        else {
            message.balance = "";
        }
        if (object.codeHash !== undefined && object.codeHash !== null) {
            message.codeHash = bytesFromBase64(object.codeHash);
        }
        if (object.nonce !== undefined && object.nonce !== null) {
            message.nonce = Number(object.nonce);
        }
        else {
            message.nonce = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.balance !== undefined && (obj.balance = message.balance);
        message.codeHash !== undefined &&
            (obj.codeHash = base64FromBytes(message.codeHash !== undefined ? message.codeHash : new Uint8Array()));
        message.nonce !== undefined && (obj.nonce = message.nonce);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAccountResponse };
        if (object.balance !== undefined && object.balance !== null) {
            message.balance = object.balance;
        }
        else {
            message.balance = "";
        }
        if (object.codeHash !== undefined && object.codeHash !== null) {
            message.codeHash = object.codeHash;
        }
        else {
            message.codeHash = new Uint8Array();
        }
        if (object.nonce !== undefined && object.nonce !== null) {
            message.nonce = object.nonce;
        }
        else {
            message.nonce = 0;
        }
        return message;
    },
};
const baseQueryCosmosAccountRequest = { address: "" };
export const QueryCosmosAccountRequest = {
    encode(message, writer = Writer.create()) {
        if (message.address !== "") {
            writer.uint32(10).string(message.address);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryCosmosAccountRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.address = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryCosmosAccountRequest,
        };
        if (object.address !== undefined && object.address !== null) {
            message.address = String(object.address);
        }
        else {
            message.address = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.address !== undefined && (obj.address = message.address);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryCosmosAccountRequest,
        };
        if (object.address !== undefined && object.address !== null) {
            message.address = object.address;
        }
        else {
            message.address = "";
        }
        return message;
    },
};
const baseQueryCosmosAccountResponse = {
    cosmosAddress: "",
    sequence: 0,
    accountNumber: 0,
};
export const QueryCosmosAccountResponse = {
    encode(message, writer = Writer.create()) {
        if (message.cosmosAddress !== "") {
            writer.uint32(10).string(message.cosmosAddress);
        }
        if (message.sequence !== 0) {
            writer.uint32(16).uint64(message.sequence);
        }
        if (message.accountNumber !== 0) {
            writer.uint32(24).uint64(message.accountNumber);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryCosmosAccountResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.cosmosAddress = reader.string();
                    break;
                case 2:
                    message.sequence = longToNumber(reader.uint64());
                    break;
                case 3:
                    message.accountNumber = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryCosmosAccountResponse,
        };
        if (object.cosmosAddress !== undefined && object.cosmosAddress !== null) {
            message.cosmosAddress = String(object.cosmosAddress);
        }
        else {
            message.cosmosAddress = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = Number(object.sequence);
        }
        else {
            message.sequence = 0;
        }
        if (object.accountNumber !== undefined && object.accountNumber !== null) {
            message.accountNumber = Number(object.accountNumber);
        }
        else {
            message.accountNumber = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.cosmosAddress !== undefined &&
            (obj.cosmosAddress = message.cosmosAddress);
        message.sequence !== undefined && (obj.sequence = message.sequence);
        message.accountNumber !== undefined &&
            (obj.accountNumber = message.accountNumber);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryCosmosAccountResponse,
        };
        if (object.cosmosAddress !== undefined && object.cosmosAddress !== null) {
            message.cosmosAddress = object.cosmosAddress;
        }
        else {
            message.cosmosAddress = "";
        }
        if (object.sequence !== undefined && object.sequence !== null) {
            message.sequence = object.sequence;
        }
        else {
            message.sequence = 0;
        }
        if (object.accountNumber !== undefined && object.accountNumber !== null) {
            message.accountNumber = object.accountNumber;
        }
        else {
            message.accountNumber = 0;
        }
        return message;
    },
};
const baseQueryBalanceRequest = { address: "" };
export const QueryBalanceRequest = {
    encode(message, writer = Writer.create()) {
        if (message.address !== "") {
            writer.uint32(10).string(message.address);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryBalanceRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.address = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryBalanceRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = String(object.address);
        }
        else {
            message.address = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.address !== undefined && (obj.address = message.address);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryBalanceRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = object.address;
        }
        else {
            message.address = "";
        }
        return message;
    },
};
const baseQueryBalanceResponse = { balance: "" };
export const QueryBalanceResponse = {
    encode(message, writer = Writer.create()) {
        if (message.balance !== "") {
            writer.uint32(10).string(message.balance);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryBalanceResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.balance = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryBalanceResponse };
        if (object.balance !== undefined && object.balance !== null) {
            message.balance = String(object.balance);
        }
        else {
            message.balance = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.balance !== undefined && (obj.balance = message.balance);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryBalanceResponse };
        if (object.balance !== undefined && object.balance !== null) {
            message.balance = object.balance;
        }
        else {
            message.balance = "";
        }
        return message;
    },
};
const baseQueryStorageRequest = { address: "", key: "" };
export const QueryStorageRequest = {
    encode(message, writer = Writer.create()) {
        if (message.address !== "") {
            writer.uint32(10).string(message.address);
        }
        if (message.key !== "") {
            writer.uint32(18).string(message.key);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryStorageRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.address = reader.string();
                    break;
                case 2:
                    message.key = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryStorageRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = String(object.address);
        }
        else {
            message.address = "";
        }
        if (object.key !== undefined && object.key !== null) {
            message.key = String(object.key);
        }
        else {
            message.key = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.address !== undefined && (obj.address = message.address);
        message.key !== undefined && (obj.key = message.key);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryStorageRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = object.address;
        }
        else {
            message.address = "";
        }
        if (object.key !== undefined && object.key !== null) {
            message.key = object.key;
        }
        else {
            message.key = "";
        }
        return message;
    },
};
const baseQueryStorageResponse = { value: "" };
export const QueryStorageResponse = {
    encode(message, writer = Writer.create()) {
        if (message.value !== "") {
            writer.uint32(10).string(message.value);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryStorageResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.value = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryStorageResponse };
        if (object.value !== undefined && object.value !== null) {
            message.value = String(object.value);
        }
        else {
            message.value = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.value !== undefined && (obj.value = message.value);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryStorageResponse };
        if (object.value !== undefined && object.value !== null) {
            message.value = object.value;
        }
        else {
            message.value = "";
        }
        return message;
    },
};
const baseQueryCodeRequest = { address: "" };
export const QueryCodeRequest = {
    encode(message, writer = Writer.create()) {
        if (message.address !== "") {
            writer.uint32(10).string(message.address);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryCodeRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.address = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryCodeRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = String(object.address);
        }
        else {
            message.address = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.address !== undefined && (obj.address = message.address);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryCodeRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = object.address;
        }
        else {
            message.address = "";
        }
        return message;
    },
};
const baseQueryCodeResponse = {};
export const QueryCodeResponse = {
    encode(message, writer = Writer.create()) {
        if (message.code.length !== 0) {
            writer.uint32(10).bytes(message.code);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryCodeResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.code = reader.bytes();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryCodeResponse };
        if (object.code !== undefined && object.code !== null) {
            message.code = bytesFromBase64(object.code);
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.code !== undefined &&
            (obj.code = base64FromBytes(message.code !== undefined ? message.code : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryCodeResponse };
        if (object.code !== undefined && object.code !== null) {
            message.code = object.code;
        }
        else {
            message.code = new Uint8Array();
        }
        return message;
    },
};
const baseQueryTxLogsRequest = { hash: "" };
export const QueryTxLogsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.hash !== "") {
            writer.uint32(10).string(message.hash);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryTxLogsRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.hash = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryTxLogsRequest };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = String(object.hash);
        }
        else {
            message.hash = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.hash !== undefined && (obj.hash = message.hash);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryTxLogsRequest };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = object.hash;
        }
        else {
            message.hash = "";
        }
        return message;
    },
};
const baseQueryTxLogsResponse = {};
export const QueryTxLogsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.logs) {
            Log.encode(v, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryTxLogsResponse };
        message.logs = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.logs.push(Log.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryTxLogsResponse };
        message.logs = [];
        if (object.logs !== undefined && object.logs !== null) {
            for (const e of object.logs) {
                message.logs.push(Log.fromJSON(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.logs) {
            obj.logs = message.logs.map((e) => (e ? Log.toJSON(e) : undefined));
        }
        else {
            obj.logs = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryTxLogsResponse };
        message.logs = [];
        if (object.logs !== undefined && object.logs !== null) {
            for (const e of object.logs) {
                message.logs.push(Log.fromPartial(e));
            }
        }
        return message;
    },
};
const baseQueryTxReceiptRequest = { hash: "" };
export const QueryTxReceiptRequest = {
    encode(message, writer = Writer.create()) {
        if (message.hash !== "") {
            writer.uint32(10).string(message.hash);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryTxReceiptRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.hash = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryTxReceiptRequest };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = String(object.hash);
        }
        else {
            message.hash = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.hash !== undefined && (obj.hash = message.hash);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryTxReceiptRequest };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = object.hash;
        }
        else {
            message.hash = "";
        }
        return message;
    },
};
const baseQueryTxReceiptResponse = {};
export const QueryTxReceiptResponse = {
    encode(message, writer = Writer.create()) {
        if (message.receipt !== undefined) {
            TxReceipt.encode(message.receipt, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryTxReceiptResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.receipt = TxReceipt.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryTxReceiptResponse };
        if (object.receipt !== undefined && object.receipt !== null) {
            message.receipt = TxReceipt.fromJSON(object.receipt);
        }
        else {
            message.receipt = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.receipt !== undefined &&
            (obj.receipt = message.receipt
                ? TxReceipt.toJSON(message.receipt)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryTxReceiptResponse };
        if (object.receipt !== undefined && object.receipt !== null) {
            message.receipt = TxReceipt.fromPartial(object.receipt);
        }
        else {
            message.receipt = undefined;
        }
        return message;
    },
};
const baseQueryTxReceiptsByBlockHeightRequest = { height: 0 };
export const QueryTxReceiptsByBlockHeightRequest = {
    encode(message, writer = Writer.create()) {
        if (message.height !== 0) {
            writer.uint32(8).int64(message.height);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryTxReceiptsByBlockHeightRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.height = longToNumber(reader.int64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHeightRequest,
        };
        if (object.height !== undefined && object.height !== null) {
            message.height = Number(object.height);
        }
        else {
            message.height = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.height !== undefined && (obj.height = message.height);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHeightRequest,
        };
        if (object.height !== undefined && object.height !== null) {
            message.height = object.height;
        }
        else {
            message.height = 0;
        }
        return message;
    },
};
const baseQueryTxReceiptsByBlockHeightResponse = {};
export const QueryTxReceiptsByBlockHeightResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.receipts) {
            TxReceipt.encode(v, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryTxReceiptsByBlockHeightResponse,
        };
        message.receipts = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.receipts.push(TxReceipt.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHeightResponse,
        };
        message.receipts = [];
        if (object.receipts !== undefined && object.receipts !== null) {
            for (const e of object.receipts) {
                message.receipts.push(TxReceipt.fromJSON(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.receipts) {
            obj.receipts = message.receipts.map((e) => e ? TxReceipt.toJSON(e) : undefined);
        }
        else {
            obj.receipts = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHeightResponse,
        };
        message.receipts = [];
        if (object.receipts !== undefined && object.receipts !== null) {
            for (const e of object.receipts) {
                message.receipts.push(TxReceipt.fromPartial(e));
            }
        }
        return message;
    },
};
const baseQueryTxReceiptsByBlockHashRequest = { hash: "" };
export const QueryTxReceiptsByBlockHashRequest = {
    encode(message, writer = Writer.create()) {
        if (message.hash !== "") {
            writer.uint32(10).string(message.hash);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryTxReceiptsByBlockHashRequest,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.hash = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHashRequest,
        };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = String(object.hash);
        }
        else {
            message.hash = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.hash !== undefined && (obj.hash = message.hash);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHashRequest,
        };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = object.hash;
        }
        else {
            message.hash = "";
        }
        return message;
    },
};
const baseQueryTxReceiptsByBlockHashResponse = {};
export const QueryTxReceiptsByBlockHashResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.receipts) {
            TxReceipt.encode(v, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryTxReceiptsByBlockHashResponse,
        };
        message.receipts = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.receipts.push(TxReceipt.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHashResponse,
        };
        message.receipts = [];
        if (object.receipts !== undefined && object.receipts !== null) {
            for (const e of object.receipts) {
                message.receipts.push(TxReceipt.fromJSON(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.receipts) {
            obj.receipts = message.receipts.map((e) => e ? TxReceipt.toJSON(e) : undefined);
        }
        else {
            obj.receipts = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryTxReceiptsByBlockHashResponse,
        };
        message.receipts = [];
        if (object.receipts !== undefined && object.receipts !== null) {
            for (const e of object.receipts) {
                message.receipts.push(TxReceipt.fromPartial(e));
            }
        }
        return message;
    },
};
const baseQueryBlockLogsRequest = { hash: "" };
export const QueryBlockLogsRequest = {
    encode(message, writer = Writer.create()) {
        if (message.hash !== "") {
            writer.uint32(10).string(message.hash);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryBlockLogsRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.hash = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryBlockLogsRequest };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = String(object.hash);
        }
        else {
            message.hash = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.hash !== undefined && (obj.hash = message.hash);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryBlockLogsRequest };
        if (object.hash !== undefined && object.hash !== null) {
            message.hash = object.hash;
        }
        else {
            message.hash = "";
        }
        return message;
    },
};
const baseQueryBlockLogsResponse = {};
export const QueryBlockLogsResponse = {
    encode(message, writer = Writer.create()) {
        for (const v of message.txLogs) {
            TransactionLogs.encode(v, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryBlockLogsResponse };
        message.txLogs = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.txLogs.push(TransactionLogs.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryBlockLogsResponse };
        message.txLogs = [];
        if (object.txLogs !== undefined && object.txLogs !== null) {
            for (const e of object.txLogs) {
                message.txLogs.push(TransactionLogs.fromJSON(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.txLogs) {
            obj.txLogs = message.txLogs.map((e) => e ? TransactionLogs.toJSON(e) : undefined);
        }
        else {
            obj.txLogs = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryBlockLogsResponse };
        message.txLogs = [];
        if (object.txLogs !== undefined && object.txLogs !== null) {
            for (const e of object.txLogs) {
                message.txLogs.push(TransactionLogs.fromPartial(e));
            }
        }
        return message;
    },
};
const baseQueryBlockBloomRequest = { height: 0 };
export const QueryBlockBloomRequest = {
    encode(message, writer = Writer.create()) {
        if (message.height !== 0) {
            writer.uint32(8).int64(message.height);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryBlockBloomRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.height = longToNumber(reader.int64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryBlockBloomRequest };
        if (object.height !== undefined && object.height !== null) {
            message.height = Number(object.height);
        }
        else {
            message.height = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.height !== undefined && (obj.height = message.height);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryBlockBloomRequest };
        if (object.height !== undefined && object.height !== null) {
            message.height = object.height;
        }
        else {
            message.height = 0;
        }
        return message;
    },
};
const baseQueryBlockBloomResponse = {};
export const QueryBlockBloomResponse = {
    encode(message, writer = Writer.create()) {
        if (message.bloom.length !== 0) {
            writer.uint32(10).bytes(message.bloom);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryBlockBloomResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.bloom = reader.bytes();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryBlockBloomResponse,
        };
        if (object.bloom !== undefined && object.bloom !== null) {
            message.bloom = bytesFromBase64(object.bloom);
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.bloom !== undefined &&
            (obj.bloom = base64FromBytes(message.bloom !== undefined ? message.bloom : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryBlockBloomResponse,
        };
        if (object.bloom !== undefined && object.bloom !== null) {
            message.bloom = object.bloom;
        }
        else {
            message.bloom = new Uint8Array();
        }
        return message;
    },
};
const baseQueryParamsRequest = {};
export const QueryParamsRequest = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryParamsRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(_) {
        const message = { ...baseQueryParamsRequest };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseQueryParamsRequest };
        return message;
    },
};
const baseQueryParamsResponse = {};
export const QueryParamsResponse = {
    encode(message, writer = Writer.create()) {
        if (message.params !== undefined) {
            Params.encode(message.params, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryParamsResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.params = Params.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryParamsResponse };
        if (object.params !== undefined && object.params !== null) {
            message.params = Params.fromJSON(object.params);
        }
        else {
            message.params = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.params !== undefined &&
            (obj.params = message.params ? Params.toJSON(message.params) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryParamsResponse };
        if (object.params !== undefined && object.params !== null) {
            message.params = Params.fromPartial(object.params);
        }
        else {
            message.params = undefined;
        }
        return message;
    },
};
const baseQueryStaticCallRequest = { address: "" };
export const QueryStaticCallRequest = {
    encode(message, writer = Writer.create()) {
        if (message.address !== "") {
            writer.uint32(10).string(message.address);
        }
        if (message.input.length !== 0) {
            writer.uint32(18).bytes(message.input);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryStaticCallRequest };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.address = reader.string();
                    break;
                case 2:
                    message.input = reader.bytes();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseQueryStaticCallRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = String(object.address);
        }
        else {
            message.address = "";
        }
        if (object.input !== undefined && object.input !== null) {
            message.input = bytesFromBase64(object.input);
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.address !== undefined && (obj.address = message.address);
        message.input !== undefined &&
            (obj.input = base64FromBytes(message.input !== undefined ? message.input : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryStaticCallRequest };
        if (object.address !== undefined && object.address !== null) {
            message.address = object.address;
        }
        else {
            message.address = "";
        }
        if (object.input !== undefined && object.input !== null) {
            message.input = object.input;
        }
        else {
            message.input = new Uint8Array();
        }
        return message;
    },
};
const baseQueryStaticCallResponse = {};
export const QueryStaticCallResponse = {
    encode(message, writer = Writer.create()) {
        if (message.data.length !== 0) {
            writer.uint32(10).bytes(message.data);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseQueryStaticCallResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.data = reader.bytes();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseQueryStaticCallResponse,
        };
        if (object.data !== undefined && object.data !== null) {
            message.data = bytesFromBase64(object.data);
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.data !== undefined &&
            (obj.data = base64FromBytes(message.data !== undefined ? message.data : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseQueryStaticCallResponse,
        };
        if (object.data !== undefined && object.data !== null) {
            message.data = object.data;
        }
        else {
            message.data = new Uint8Array();
        }
        return message;
    },
};
export class QueryClientImpl {
    constructor(rpc) {
        this.rpc = rpc;
    }
    Account(request) {
        const data = QueryAccountRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "Account", data);
        return promise.then((data) => QueryAccountResponse.decode(new Reader(data)));
    }
    CosmosAccount(request) {
        const data = QueryCosmosAccountRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "CosmosAccount", data);
        return promise.then((data) => QueryCosmosAccountResponse.decode(new Reader(data)));
    }
    Balance(request) {
        const data = QueryBalanceRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "Balance", data);
        return promise.then((data) => QueryBalanceResponse.decode(new Reader(data)));
    }
    Storage(request) {
        const data = QueryStorageRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "Storage", data);
        return promise.then((data) => QueryStorageResponse.decode(new Reader(data)));
    }
    Code(request) {
        const data = QueryCodeRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "Code", data);
        return promise.then((data) => QueryCodeResponse.decode(new Reader(data)));
    }
    TxLogs(request) {
        const data = QueryTxLogsRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "TxLogs", data);
        return promise.then((data) => QueryTxLogsResponse.decode(new Reader(data)));
    }
    TxReceipt(request) {
        const data = QueryTxReceiptRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "TxReceipt", data);
        return promise.then((data) => QueryTxReceiptResponse.decode(new Reader(data)));
    }
    TxReceiptsByBlockHeight(request) {
        const data = QueryTxReceiptsByBlockHeightRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "TxReceiptsByBlockHeight", data);
        return promise.then((data) => QueryTxReceiptsByBlockHeightResponse.decode(new Reader(data)));
    }
    TxReceiptsByBlockHash(request) {
        const data = QueryTxReceiptsByBlockHashRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "TxReceiptsByBlockHash", data);
        return promise.then((data) => QueryTxReceiptsByBlockHashResponse.decode(new Reader(data)));
    }
    BlockLogs(request) {
        const data = QueryBlockLogsRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "BlockLogs", data);
        return promise.then((data) => QueryBlockLogsResponse.decode(new Reader(data)));
    }
    BlockBloom(request) {
        const data = QueryBlockBloomRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "BlockBloom", data);
        return promise.then((data) => QueryBlockBloomResponse.decode(new Reader(data)));
    }
    Params(request) {
        const data = QueryParamsRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "Params", data);
        return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
    }
    StaticCall(request) {
        const data = QueryStaticCallRequest.encode(request).finish();
        const promise = this.rpc.request("ethermint.evm.v1alpha1.Query", "StaticCall", data);
        return promise.then((data) => QueryStaticCallResponse.decode(new Reader(data)));
    }
}
var globalThis = (() => {
    if (typeof globalThis !== "undefined")
        return globalThis;
    if (typeof self !== "undefined")
        return self;
    if (typeof window !== "undefined")
        return window;
    if (typeof global !== "undefined")
        return global;
    throw "Unable to locate global object";
})();
const atob = globalThis.atob ||
    ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64) {
    const bin = atob(b64);
    const arr = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; ++i) {
        arr[i] = bin.charCodeAt(i);
    }
    return arr;
}
const btoa = globalThis.btoa ||
    ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr) {
    const bin = [];
    for (let i = 0; i < arr.byteLength; ++i) {
        bin.push(String.fromCharCode(arr[i]));
    }
    return btoa(bin.join(""));
}
function longToNumber(long) {
    if (long.gt(Number.MAX_SAFE_INTEGER)) {
        throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
    }
    return long.toNumber();
}
if (util.Long !== Long) {
    util.Long = Long;
    configure();
}
