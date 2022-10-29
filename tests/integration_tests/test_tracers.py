from .expected_constants import (
    EXPECTED_CALLTRACERS,
    EXPECTED_STRUCT_TRACER
    )
from web3 import Web3

from .utils import (
    ADDRS,
    KEYS,
    send_transaction,
)

def send_and_get_hash(w3, tx_value=100):
    # Do an ethereum transfer
    gas_price = w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": gas_price*10}
    return send_transaction(w3, tx, KEYS["validator"])["transactionHash"].hex()



def test_pending_transaction_filter(ethermint_rpc_ws):
    w3: Web3 = ethermint_rpc_ws.w3
    eth_rpc = w3.provider
    tx_hash = send_and_get_hash(w3)

    tx_res = eth_rpc.make_request("debug_traceTransaction", [tx_hash])
    assert tx_res['result'] == EXPECTED_STRUCT_TRACER, ""

    tx_res = eth_rpc.make_request("debug_traceTransaction", [tx_hash, {"tracer":"callTracer"}])
    assert tx_res['result'] == EXPECTED_CALLTRACERS, ""

    tx_res = eth_rpc.make_request("debug_traceTransaction", [tx_hash, {"tracer":"callTracer", "tracerConfig": "{'onlyTopCall':True}"}])
    assert tx_res['result'] == EXPECTED_CALLTRACERS, ""