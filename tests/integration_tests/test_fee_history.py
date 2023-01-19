from concurrent.futures import ThreadPoolExecutor, as_completed

import pytest
from web3 import Web3

from .network import setup_ethermint
from .utils import ADDRS, send_transaction


@pytest.fixture(scope="module")
def custom_ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("fee-history")
    yield from setup_ethermint(path, 26500, long_timeout_commit=True)


@pytest.fixture(scope="module", params=["ethermint", "geth"])
def cluster(request, custom_ethermint, geth):
    """
    run on both ethermint and geth
    """
    provider = request.param
    if provider == "ethermint":
        yield custom_ethermint
    elif provider == "geth":
        yield geth
    else:
        raise NotImplementedError


def test_basic_fee(cluster):
    w3: Web3 = cluster.w3
    call = w3.provider.make_request
    tx_value = 10
    gas_price = w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": gas_price}
    send_transaction(w3, tx)
    range = 4
    method = "eth_feeHistory"
    field = "baseFeePerGas"
    percentiles = [100]
    height = w3.eth.block_number
    latest = dict(
        blocks=["latest", hex(height)],
        # range of base fee + next fee
        expect=range + 1,
    )
    earliest = dict(
        blocks=["earliest", "0x0"],
        # only 1 base fee + next fee
        expect=2,
    )
    for tc in [latest, earliest]:
        res = []
        with ThreadPoolExecutor(len(tc["blocks"])) as exec:
            tasks = [
                exec.submit(call, method, [range, b, percentiles]) for b in tc["blocks"]
            ]
            res = [future.result()["result"][field] for future in as_completed(tasks)]
        assert len(res) == len(tc["blocks"])
        assert res[0] == res[1]
        assert len(res[0]) == tc["expect"]
