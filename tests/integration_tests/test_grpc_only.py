import base64
import json
import subprocess
import time
from pathlib import Path

import pytest
import requests
from pystarport import ports

from .network import setup_custom_ethermint
from .utils import (
    CONTRACTS,
    decode_bech32,
    deploy_contract,
    supervisorctl,
    wait_for_block,
    wait_for_port,
)


@pytest.fixture(scope="module")
def custom_ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("grpc-only")

    # reuse rollback-test config because it has an extra fullnode
    yield from setup_custom_ethermint(
        path,
        26400,
        Path(__file__).parent / "configs/rollback-test.jsonnet",
    )


def grpc_eth_call(port: int, args: dict, chain_id=None, proposer_address=None):
    """
    do a eth_call through grpc gateway directly
    """
    params = {
        "args": base64.b64encode(json.dumps(args).encode()).decode(),
    }
    if chain_id is not None:
        params["chain_id"] = str(chain_id)
    if proposer_address is not None:
        params["proposer_address"] = str(proposer_address)
    return requests.get(
        f"http://localhost:{port}/ethermint/evm/v1/eth_call", params
    ).json()


def test_grpc_mode(custom_ethermint):
    """
    - restart a fullnode in grpc-only mode
    - test the grpc queries all works
    """
    w3 = custom_ethermint.w3
    contract, _ = deploy_contract(w3, CONTRACTS["TestChainID"])
    assert 9000 == contract.caller.currentChainID()

    msg = {
        "to": contract.address,
        "data": contract.encodeABI(fn_name="currentChainID"),
    }
    api_port = ports.api_port(custom_ethermint.base_port(1))
    # in normal mode, grpc query works even if we don't pass chain_id explicitly
    success = False
    max_retry = 3
    sleep = 1
    for i in range(max_retry):
        rsp = grpc_eth_call(api_port, msg)
        ret = rsp["ret"]
        valid = ret is not None
        if valid and 9000 == int.from_bytes(base64.b64decode(ret.encode()), "big"):
            success = True
            break
        time.sleep(sleep)
    assert success
    # wait 1 more block for both nodes to avoid node stopped before tnx get included
    for i in range(2):
        wait_for_block(custom_ethermint.cosmos_cli(i), 1)
    supervisorctl(
        custom_ethermint.base_dir / "../tasks.ini", "stop", "ethermint_9000-1-node1"
    )

    # run grpc-only mode directly with existing chain state
    with (custom_ethermint.base_dir / "node1.log").open("a") as logfile:
        proc = subprocess.Popen(
            [
                "ethermintd",
                "start",
                "--grpc-only",
                "--home",
                custom_ethermint.base_dir / "node1",
            ],
            stdout=logfile,
            stderr=subprocess.STDOUT,
        )
        try:
            # wait for grpc and rest api ports
            grpc_port = ports.grpc_port(custom_ethermint.base_port(1))
            wait_for_port(grpc_port)
            wait_for_port(api_port)

            # in grpc-only mode, grpc query don't work if we don't pass chain_id
            rsp = grpc_eth_call(api_port, msg)
            assert rsp["code"] != 0, str(rsp)
            assert "invalid chain ID" in rsp["message"]

            # it don't works without proposer address neither
            rsp = grpc_eth_call(api_port, msg, chain_id=9000)
            assert rsp["code"] != 0, str(rsp)
            assert "validator does not exist" in rsp["message"]

            # pass the first validator's consensus address to grpc query
            addr = custom_ethermint.cosmos_cli(0).consensus_address()
            cons_addr = decode_bech32(addr)

            # should work with both chain_id and proposer_address set
            rsp = grpc_eth_call(
                api_port,
                msg,
                chain_id=100,
                proposer_address=base64.b64encode(cons_addr).decode(),
            )
            assert "code" not in rsp, str(rsp)
            assert 100 == int.from_bytes(base64.b64decode(rsp["ret"].encode()), "big")
        finally:
            proc.terminate()
            proc.wait()
