// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgSubmitProposal } from "./types/cosmos/gov/v1beta1/tx";
import { MsgDeposit } from "./types/cosmos/gov/v1beta1/tx";
import { MsgVote } from "./types/cosmos/gov/v1beta1/tx";


const types = [
  ["/cosmos.gov.v1beta1.MsgSubmitProposal", MsgSubmitProposal],
  ["/cosmos.gov.v1beta1.MsgDeposit", MsgDeposit],
  ["/cosmos.gov.v1beta1.MsgVote", MsgVote],
  
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
    msgSubmitProposal: (data: MsgSubmitProposal): EncodeObject => ({ typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal", value: data }),
    msgDeposit: (data: MsgDeposit): EncodeObject => ({ typeUrl: "/cosmos.gov.v1beta1.MsgDeposit", value: data }),
    msgVote: (data: MsgVote): EncodeObject => ({ typeUrl: "/cosmos.gov.v1beta1.MsgVote", value: data }),
    
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
