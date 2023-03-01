import asyncio
import json
import time
from collections import defaultdict

import websockets
from web3 import Web3

from .network import Ethermint
from .utils import (
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_raw_transactions,
    sign_transaction,
    wait_for_new_blocks,
)

# note:
# batch requests still not implemented in web3.py
# todo: follow https://github.com/ethereum/web3.py/issues/832, add tests when complete

# eth_subscribe and eth_unsubscribe support still not implemented in web3.py
# todo: follow https://github.com/ethereum/web3.py/issues/1402, add tests when complete


def test_batch_request_netversion(ethermint):
    return


class Client:
    def __init__(self, ws):
        self._ws = ws
        self._gen_id = 0
        self._subs = defaultdict(asyncio.Queue)
        self._rsps = defaultdict(asyncio.Queue)

    def gen_id(self):
        self._gen_id += 1
        return self._gen_id

    async def receive_loop(self):
        while True:
            msg = json.loads(await self._ws.recv())
            if "id" in msg:
                # responses
                await self._rsps[msg["id"]].put(msg)
            else:
                # subscriptions
                assert msg["method"] == "eth_subscription"
                sub_id = msg["params"]["subscription"]
                await self._subs[sub_id].put(msg["params"]["result"])

    async def recv_response(self, rpcid):
        rsp = await self._rsps[rpcid].get()
        del self._rsps[rpcid]
        return rsp

    async def recv_subscription(self, sub_id):
        return await self._subs[sub_id].get()

    async def subscribe(self, *args):
        rpcid = self.gen_id()
        await self._ws.send(
            json.dumps({"id": rpcid, "method": "eth_subscribe", "params": args})
        )
        rsp = await self.recv_response(rpcid)
        assert "error" not in rsp
        return rsp["result"]

    async def send(self, method, *args):
        id = self.gen_id()
        await self._ws.send(json.dumps({"id": id, "method": method, "params": args}))
        rsp = await self.recv_response(id)
        assert "error" not in rsp
        return rsp["result"]

    def sub_qsize(self, sub_id):
        return self._subs[sub_id].qsize()

    async def unsubscribe(self, sub_id):
        rpcid = self.gen_id()
        await self._ws.send(
            json.dumps({"id": rpcid, "method": "eth_unsubscribe", "params": [sub_id]})
        )
        rsp = await self.recv_response(rpcid)
        assert "error" not in rsp
        return rsp["result"]


# TestEvent topic from TestMessageCall contract calculated from event signature
TEST_EVENT_TOPIC = Web3.keccak(text="TestEvent(uint256)")


def test_subscribe_basic(ethermint: Ethermint):
    """
    test basic subscribe and unsubscribe
    """
    cli = ethermint.cosmos_cli()
    loop = asyncio.get_event_loop()

    async def assert_unsubscribe(c: Client, sub_id):
        assert await c.unsubscribe(sub_id)
        # check no more messages
        await loop.run_in_executor(None, wait_for_new_blocks, cli, 2)
        assert c.sub_qsize(sub_id) == 0
        # unsubscribe again return False
        assert not await c.unsubscribe(sub_id)

    async def subscriber_test(c: Client):
        sub_id = await c.subscribe("newHeads")
        # wait for three new blocks
        msgs = [await c.recv_subscription(sub_id) for i in range(3)]
        # check blocks are consecutive
        assert int(msgs[1]["number"], 0) == int(msgs[0]["number"], 0) + 1
        assert int(msgs[2]["number"], 0) == int(msgs[1]["number"], 0) + 1
        await assert_unsubscribe(c, sub_id)

    async def logs_test(c: Client, w3, contract, address):
        sub_id = await c.subscribe("logs", {"address": address})
        iterations = 10000
        tx = contract.functions.test(iterations).build_transaction()
        raw_transactions = []
        for key_from in KEYS.values():
            signed = sign_transaction(w3, tx, key_from)
            raw_transactions.append(signed.rawTransaction)
        send_raw_transactions(w3, raw_transactions)
        total = len(KEYS) * iterations
        msgs = [await c.recv_subscription(sub_id) for i in range(total)]
        assert len(msgs) == total
        assert all(msg["topics"] == [TEST_EVENT_TOPIC.hex()] for msg in msgs)
        await assert_unsubscribe(c, sub_id)

    async def net_version_test(c: Client):
        version = await c.send("net_version")
        # net_version should be 9000
        assert version == "9000", "got " + version + ", expected 9000"

    async def async_test():
        async with websockets.connect(ethermint.w3_ws_endpoint) as ws:
            c = Client(ws)
            t = asyncio.create_task(c.receive_loop())
            # run three subscribers concurrently
            await asyncio.gather(*[subscriber_test(c) for i in range(3)])
            await asyncio.gather(*[net_version_test(c)])
            contract, _ = deploy_contract(ethermint.w3, CONTRACTS["TestMessageCall"])
            inner = contract.caller.inner()
            begin = time.time()
            await asyncio.gather(*[logs_test(c, ethermint.w3, contract, inner)])
            print("msg call time", time.time() - begin)
            t.cancel()
            try:
                await t
            except asyncio.CancelledError:
                # allow retry
                pass

    timeout = 100
    loop.run_until_complete(asyncio.wait_for(async_test(), timeout))
