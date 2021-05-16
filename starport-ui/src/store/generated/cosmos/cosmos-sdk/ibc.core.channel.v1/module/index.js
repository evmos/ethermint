// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgChannelOpenInit } from "./types/ibc/core/channel/v1/tx";
import { MsgTimeout } from "./types/ibc/core/channel/v1/tx";
import { MsgChannelOpenAck } from "./types/ibc/core/channel/v1/tx";
import { MsgChannelCloseInit } from "./types/ibc/core/channel/v1/tx";
import { MsgChannelCloseConfirm } from "./types/ibc/core/channel/v1/tx";
import { MsgChannelOpenConfirm } from "./types/ibc/core/channel/v1/tx";
import { MsgChannelOpenTry } from "./types/ibc/core/channel/v1/tx";
import { MsgRecvPacket } from "./types/ibc/core/channel/v1/tx";
import { MsgTimeoutOnClose } from "./types/ibc/core/channel/v1/tx";
import { MsgAcknowledgement } from "./types/ibc/core/channel/v1/tx";
const types = [
    ["/ibc.core.channel.v1.MsgChannelOpenInit", MsgChannelOpenInit],
    ["/ibc.core.channel.v1.MsgTimeout", MsgTimeout],
    ["/ibc.core.channel.v1.MsgChannelOpenAck", MsgChannelOpenAck],
    ["/ibc.core.channel.v1.MsgChannelCloseInit", MsgChannelCloseInit],
    ["/ibc.core.channel.v1.MsgChannelCloseConfirm", MsgChannelCloseConfirm],
    ["/ibc.core.channel.v1.MsgChannelOpenConfirm", MsgChannelOpenConfirm],
    ["/ibc.core.channel.v1.MsgChannelOpenTry", MsgChannelOpenTry],
    ["/ibc.core.channel.v1.MsgRecvPacket", MsgRecvPacket],
    ["/ibc.core.channel.v1.MsgTimeoutOnClose", MsgTimeoutOnClose],
    ["/ibc.core.channel.v1.MsgAcknowledgement", MsgAcknowledgement],
];
const registry = new Registry(types);
const defaultFee = {
    amount: [],
    gas: "200000",
};
const txClient = async (wallet, { addr: addr } = { addr: "http://localhost:26657" }) => {
    if (!wallet)
        throw new Error("wallet is required");
    const client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
    const { address } = (await wallet.getAccounts())[0];
    return {
        signAndBroadcast: (msgs, { fee = defaultFee, memo = null }) => memo ? client.signAndBroadcast(address, msgs, fee, memo) : client.signAndBroadcast(address, msgs, fee),
        msgChannelOpenInit: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenInit", value: data }),
        msgTimeout: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgTimeout", value: data }),
        msgChannelOpenAck: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenAck", value: data }),
        msgChannelCloseInit: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelCloseInit", value: data }),
        msgChannelCloseConfirm: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelCloseConfirm", value: data }),
        msgChannelOpenConfirm: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenConfirm", value: data }),
        msgChannelOpenTry: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenTry", value: data }),
        msgRecvPacket: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgRecvPacket", value: data }),
        msgTimeoutOnClose: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgTimeoutOnClose", value: data }),
        msgAcknowledgement: (data) => ({ typeUrl: "/ibc.core.channel.v1.MsgAcknowledgement", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
