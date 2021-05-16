// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
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

const registry = new Registry(<any>types);

const defaultFee = {
  amount: [],
  gas: "200000",
};

interface TxClientOptions {
  addr: string
}

interface SignAndBroadcastOptions {
  fee: StdFee,
  memo?: string
}

const txClient = async (wallet: OfflineSigner, { addr: addr }: TxClientOptions = { addr: "http://localhost:26657" }) => {
  if (!wallet) throw new Error("wallet is required");

  const client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
  const { address } = (await wallet.getAccounts())[0];

  return {
    signAndBroadcast: (msgs: EncodeObject[], { fee=defaultFee, memo=null }: SignAndBroadcastOptions) => memo?client.signAndBroadcast(address, msgs, fee,memo):client.signAndBroadcast(address, msgs, fee),
    msgChannelOpenInit: (data: MsgChannelOpenInit): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenInit", value: data }),
    msgTimeout: (data: MsgTimeout): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgTimeout", value: data }),
    msgChannelOpenAck: (data: MsgChannelOpenAck): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenAck", value: data }),
    msgChannelCloseInit: (data: MsgChannelCloseInit): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelCloseInit", value: data }),
    msgChannelCloseConfirm: (data: MsgChannelCloseConfirm): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelCloseConfirm", value: data }),
    msgChannelOpenConfirm: (data: MsgChannelOpenConfirm): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenConfirm", value: data }),
    msgChannelOpenTry: (data: MsgChannelOpenTry): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgChannelOpenTry", value: data }),
    msgRecvPacket: (data: MsgRecvPacket): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgRecvPacket", value: data }),
    msgTimeoutOnClose: (data: MsgTimeoutOnClose): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgTimeoutOnClose", value: data }),
    msgAcknowledgement: (data: MsgAcknowledgement): EncodeObject => ({ typeUrl: "/ibc.core.channel.v1.MsgAcknowledgement", value: data }),
    
  };
};

interface QueryClientOptions {
  addr: string
}

const queryClient = async ({ addr: addr }: QueryClientOptions = { addr: "http://localhost:1317" }) => {
  return new Api({ baseUrl: addr });
};

export {
  txClient,
  queryClient,
};
