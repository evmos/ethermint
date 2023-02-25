import asyncio
import json
from collections import defaultdict

import websockets
from pystarport import ports


def test_single_request_netversion(ethermint):
    ethermint.use_websocket()
    eth_ws = ethermint.w3.provider

    response = eth_ws.make_request("net_version", [])

    # net_version should be 9000
    assert response["result"] == "9000", "got " + response["result"] + ", expected 9000"


# note:
# batch requests still not implemented in web3.py
# todo: follow https://github.com/ethereum/web3.py/issues/832, add tests when complete

# eth_subscribe and eth_unsubscribe support still not implemented in web3.py
# todo: follow https://github.com/ethereum/web3.py/issues/1402, add tests when complete


class Client:
    def __init__(self, ws):
        self._ws = ws
        self._subs = defaultdict(asyncio.Queue)
        self._rsps = defaultdict(asyncio.Queue)

    async def receive_loop(self):
        while True:
            msg = json.loads(await self._ws.recv())
            if "id" in msg:
                # responses
                await self._rsps[msg["id"]].put(msg)

    async def recv_response(self, rpcid):
        rsp = await self._rsps[rpcid].get()
        del self._rsps[rpcid]
        return rsp

    async def send(self, id):
        await self._ws.send(
            json.dumps({"id": id, "method": "web3_clientVersion", "params": []})
        )
        rsp = await self.recv_response(id)
        assert "error" not in rsp


def test_web3_client_version(ethermint):
    ethermint_ws = ethermint.copy()
    ethermint_ws.use_websocket()
    port = ethermint_ws.base_port(0)
    url = f"ws://127.0.0.1:{ports.evmrpc_ws_port(port)}"
    loop = asyncio.get_event_loop()

    async def async_test():
        async with websockets.connect(url) as ws:
            c = Client(ws)
            t = asyncio.create_task(c.receive_loop())
            # run send concurrently
            await asyncio.gather(*[c.send(id) for id in ["0", 1, 2.0]])
            t.cancel()
            try:
                await t
            except asyncio.CancelledError:
                # allow retry
                pass

    loop.run_until_complete(async_test())


def test_batch_request_netversion(ethermint):
    return


def test_ws_subscribe_log(ethermint):
    return


def test_ws_subscribe_newheads(ethermint):
    return
