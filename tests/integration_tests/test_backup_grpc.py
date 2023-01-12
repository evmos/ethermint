import os
import signal
import subprocess
from pathlib import Path
from typing import NamedTuple

import pytest
import requests
from pystarport import ports

from .network import Ethermint
from .utils import ADDRS, wait_for_fn, wait_for_port


class Network(NamedTuple):
    main: Ethermint
    backup: Ethermint


def exec(config, path, base_port):
    cmd = [
        "pystarport",
        "init",
        "--config",
        config,
        "--data",
        path,
        "--base_port",
        str(base_port),
        "--no_remove",
    ]
    print(*cmd)
    subprocess.run(cmd, check=True)
    return subprocess.Popen(
        ["pystarport", "start", "--data", path, "--quiet"],
        preexec_fn=os.setsid,
    )


@pytest.fixture(scope="module")
def network(tmp_path_factory):
    chain_id = "ethermint_9000-1"
    base = Path(__file__).parent / "configs"

    # backup
    path0 = tmp_path_factory.mktemp("ethermint-backup")
    base_port0 = 26770
    procs = [exec(base / "backup.jsonnet", path0, base_port0)]

    wait_for_port(base_port0)

    # main
    path1 = tmp_path_factory.mktemp("ethermint-main")
    base_port1 = 26750
    procs.append(exec(base / "main.jsonnet", path1, base_port1))

    try:
        wait_for_port(ports.evmrpc_port(base_port0))
        wait_for_port(ports.evmrpc_ws_port(base_port0))
        wait_for_port(ports.grpc_port(base_port1))
        yield Network(Ethermint(path0 / chain_id), Ethermint(path1 / chain_id))
    finally:
        for proc in procs:
            os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
            proc.wait()
            print("killed:", proc.pid)


def grpc_call(p, address):
    url = f"http://127.0.0.1:{p}/cosmos/bank/v1beta1/balances/{address}"
    response = requests.get(url)
    if not response.ok:
        # retry until file get synced
        return -1
        # raise Exception(
        #     f"response code: {response.status_code}, "
        #     f"{response.reason}, {response.json()}"
        # )
    result = response.json()
    if result.get("code"):
        raise Exception(result["raw_log"])
    return result["balances"]


def test_basic(network):
    pw3 = network.main.w3
    pcli = network.main.cosmos_cli()
    validator = pcli.address("validator")
    community = pcli.address("community")
    print("address: ", validator, community)
    backup_grpc_port = ports.api_port(network.backup.base_port(0))

    def check_balances():
        mbalances = [pcli.balances(community), pcli.balances(validator)]
        bbalances = [
            grpc_call(backup_grpc_port, community),
            grpc_call(backup_grpc_port, validator),
        ]
        print("main", mbalances)
        print("backup", bbalances)
        return mbalances == bbalances

    txhash = pw3.eth.send_transaction(
        {
            "from": ADDRS["validator"],
            "to": ADDRS["community"],
            "value": 1000,
        }
    )
    receipt = pw3.eth.wait_for_transaction_receipt(txhash)
    assert receipt.status == 1
    wait_for_fn("cross-check-balances", check_balances)
