import pytest
from web3 import Web3

from .network import setup_ethermint
from .utils import send_transaction, ADDRS


@pytest.fixture(scope="module")
def custom_ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("fee-history")
    yield from setup_ethermint(path, 26500, long_timeout_commit=True)


@pytest.fixture(
    scope="module", params=["ethermint", "geth", "ethermint-ws"]
)
def cluster(request, custom_ethermint, geth):
    """
    run on both ethermint and geth
    """
    provider = request.param
    if provider == "ethermint":
        yield custom_ethermint
    elif provider == "geth":
        yield geth
    elif provider == "ethermint-ws":
        ethermint_ws = custom_ethermint.copy()
        ethermint_ws.use_websocket()
        yield ethermint_ws
    else:
        raise NotImplementedError


def test_basic(cluster):
    w3: Web3 = cluster.w3
    eth_rpc = w3.provider
    tx_value = 10
    gas_price = w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": gas_price}
    send_transaction(w3, tx)
    for provider in [eth_rpc]:
        fee_history = provider.make_request("eth_feeHistory", [4, "latest", [100]])
        assert len(fee_history["result"]["baseFeePerGas"]) == 5
