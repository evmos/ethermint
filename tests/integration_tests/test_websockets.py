from .expected_constants import (
    EXPECTED_FEE_HISTORY,
    EXPECTED_GET_PROOF,
    EXPECTED_GET_STORAGE_AT,
    EXPECTED_GET_TRANSACTION,
    EXPECTED_GET_TRANSACTION_RECEIPT,
)
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    w3_wait_for_block,
    w3_wait_for_new_blocks,
)

def test_single_request_netversion(ethermint):
    ethermint.use_websocket()
    eth_ws = ethermint.w3.provider

    response = eth_ws.make_request("net_version", [])

    # net_version should be 9000
    assert response["result"] == "9000", "got " + response["result"] + ", expected 9000"

# note:

# batch requests still not implemented in web3.py (no equivalent of createBatch() from web3.js)
# todo: follow https://github.com/ethereum/web3.py/issues/832, add tests when complete

# eth_subscribe and eth_unsubscribe support still not implemented in web3.py
# todo: follow https://github.com/ethereum/web3.py/issues/1402, add tests when complete

def test_batch_request_netversion(ethermint):
    return

def test_ws_subscribe_log(ethermint):
    return

def test_ws_subscribe_newHeads(ethermint):
    return