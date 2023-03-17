import configparser
import subprocess
from pathlib import Path

import pytest
import requests
from pystarport.cluster import SUPERVISOR_CONFIG_FILE

from .network import setup_custom_ethermint
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    sign_transaction,
    wait_for_new_blocks,
)


def update_node_cmd(path, cmd, i):
    ini_path = path / SUPERVISOR_CONFIG_FILE
    ini = configparser.RawConfigParser()
    ini.read(ini_path)
    for section in ini.sections():
        if section == f"program:ethermint_9000-1-node{i}":
            ini[section].update(
                {
                    "command": f"{cmd} start --home %(here)s/node{i}",
                    "autorestart": "false",  # don't restart when stopped
                }
            )
    with ini_path.open("w") as fp:
        ini.write(fp)


def post_init(patch_binary):
    def inner(path, base_port, config):
        chain_id = "ethermint_9000-1"
        update_node_cmd(path / chain_id, patch_binary, 1)

    return inner


@pytest.fixture(scope="module")
def custom_ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("rollback")

    cmd = [
        "nix-build",
        "--no-out-link",
        Path(__file__).parent / "configs/cache-access-list-ethermintd.nix",
    ]
    print(*cmd)
    patch_binary = (
        Path(subprocess.check_output(cmd).strip().decode()) / "bin/ethermintd"
    )
    print(patch_binary)

    # init with genesis binary
    yield from setup_custom_ethermint(
        path,
        27000,
        Path(__file__).parent / "configs/default.jsonnet",
        post_init=post_init(patch_binary),
        wait_port=True,
    )


def multi_transfer(w3, contract, sender, key, receiver):
    amt = 100
    txhashes = []
    nonce = w3.eth.get_transaction_count(sender)
    for i in range(2):
        tx = contract.functions.transfer(sender, amt).build_transaction({
            "from": receiver,
            "nonce": nonce + i,
        })
        signed = sign_transaction(w3, tx, key)
        txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
        txhashes.append(txhash)
    return txhashes


def test_patch(custom_ethermint):
    cli = custom_ethermint.cosmos_cli()
    w3 = custom_ethermint.w3
    validator = ADDRS["validator"]
    community = ADDRS["community"]
    contract, _ = deploy_contract(w3, CONTRACTS["TestERC20A"])
    amt = 3000
    # fund community
    params = {"from": validator}
    tx = contract.functions.transfer(community, amt).build_transaction(params)
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1

    wait_for_new_blocks(cli, 1, sleep=0.1)
    txhashes = multi_transfer(w3, contract, validator, KEYS["validator"], community)
    txhashes += multi_transfer(w3, contract, community, KEYS["community"], validator)

    for txhash in txhashes:
        receipt = w3.eth.wait_for_transaction_receipt(txhash)
        assert receipt.status == 1

    params = {
        "method": "debug_traceTransaction",
        "id": 1,
        "jsonrpc": "2.0",
    }
    gas = 29506
    diff = 2000
    for i, txhash in enumerate(txhashes):
        params["params"] = [txhash.hex()]
        rsp0 = requests.post(custom_ethermint.w3_http_endpoint(0), json=params)
        rsp1 = requests.post(custom_ethermint.w3_http_endpoint(1), json=params)
        assert rsp0.status_code == 200
        assert rsp1.status_code == 200
        gas0 = rsp0.json()["result"]["gas"]
        gas1 = rsp1.json()["result"]["gas"]
        assert gas0 == gas
        assert gas1 == gas if i % 2 == 0 else gas1 == gas - diff
