import configparser
import json
import re
import subprocess
from pathlib import Path

import pytest
import requests
from pystarport import ports
from pystarport.cluster import SUPERVISOR_CONFIG_FILE

from .network import setup_custom_ethermint
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    sign_transaction,
    supervisorctl,
    wait_for_new_blocks,
    wait_for_port,
)


def init_cosmovisor(home):
    cosmovisor = home / "cosmovisor"
    cosmovisor.mkdir()
    (cosmovisor / "patched").symlink_to("../../../patched")
    (cosmovisor / "genesis").symlink_to("./patched/genesis")


def post_init(path, base_port, config):
    """
    prepare cosmovisor for each node
    """
    chain_id = "ethermint_9000-1"
    cfg = json.loads((path / chain_id / "config.json").read_text())
    for i, _ in enumerate(cfg["validators"]):
        home = path / chain_id / f"node{i}"
        init_cosmovisor(home)

    # patch supervisord ini config
    ini_path = path / chain_id / SUPERVISOR_CONFIG_FILE
    ini = configparser.RawConfigParser()
    ini.read(ini_path)
    reg = re.compile(rf"^program:{chain_id}-node(\d+)")
    for section in ini.sections():
        m = reg.match(section)
        if m:
            i = m.group(1)
            ini[section].update(
                {
                    "command": f"cosmovisor start --home %(here)s/node{i}",
                    "environment": (
                        f"DAEMON_NAME=ethermintd,DAEMON_HOME=%(here)s/node{i}"
                    ),
                }
            )
    with ini_path.open("w") as fp:
        ini.write(fp)


@pytest.fixture(scope="module")
def custom_ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("patch")
    cmd = [
        "nix-build",
        Path(__file__).parent / "configs/cache-access-list-ethermintd.nix",
        "-o",
        path / "patched",
    ]
    print(*cmd)
    subprocess.run(cmd, check=True)
    # init with patch binary
    yield from setup_custom_ethermint(
        path,
        27000,
        Path(__file__).parent / "configs/cosmovisor.jsonnet",
        post_init=post_init,
        chain_binary=str(path / "patched/genesis/bin/ethermintd"),
        wait_port=True,
    )


def multi_transfer(w3, contract, sender, key, receiver):
    amt = 100
    txhashes = []
    nonce = w3.eth.get_transaction_count(sender)
    for i in range(2):
        tx = contract.functions.transfer(sender, amt).build_transaction(
            {
                "from": receiver,
                "nonce": nonce + i,
            }
        )
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

    sleep = 0.1
    target_height = wait_for_new_blocks(cli, 1, sleep)
    txhashes = multi_transfer(w3, contract, validator, KEYS["validator"], community)
    txhashes += multi_transfer(w3, contract, community, KEYS["community"], validator)

    for txhash in txhashes:
        receipt = w3.eth.wait_for_transaction_receipt(txhash)
        assert receipt.status == 1

    nodes = [0, 1]
    (wait_for_new_blocks(custom_ethermint.cosmos_cli(n), 1, sleep) for n in nodes)

    params = {
        "method": "debug_traceTransaction",
        "id": 1,
        "jsonrpc": "2.0",
    }
    gas = 29506
    diff = 2000

    def assert_debug_tx(i, all_eq=False):
        results = []
        for txhash in txhashes:
            params["params"] = [txhash.hex()]
            rsp = requests.post(custom_ethermint.w3_http_endpoint(i), json=params)
            assert rsp.status_code == 200
            result = rsp.json()["result"]["gas"]
            results.append(result)
        for i, result in enumerate(results):
            if all_eq:
                assert result == gas
            else:
                # costs less gas when cache access list
                assert result == gas if i % 2 == 0 else result == gas - diff

    (assert_debug_tx(n) for n in nodes)

    base_dir = custom_ethermint.base_dir
    for n in nodes:
        supervisorctl(base_dir / "../tasks.ini", "stop", f"ethermint_9000-1-node{n}")

    procs = []

    def append_proc(log, cmd):
        with (base_dir / log).open("a") as logfile:
            procs.append(
                subprocess.Popen(
                    cmd,
                    stdout=logfile,
                    stderr=subprocess.STDOUT,
                )
            )

    path = Path(custom_ethermint.chain_binary).parent.parent.parent
    grpc_port1 = ports.grpc_port(custom_ethermint.base_port(1))

    for blk_end in [target_height - 1, target_height]:
        try:
            append_proc(
                "node1.log",
                [
                    f"{str(path)}/genesis/bin/ethermintd",
                    "start",
                    "--home",
                    base_dir / "node1",
                ],
            )
            append_proc(
                "node0.log",
                [
                    f"{str(path)}/integration-test-patch/bin/ethermintd",
                    "start",
                    "--json-rpc.backup-grpc-address-block-range",
                    f'{{"0.0.0.0:{grpc_port1}": [0, {blk_end}]}}',
                    "--home",
                    base_dir / "node0",
                ],
            )
            for n in nodes:
                wait_for_port(ports.evmrpc_port(custom_ethermint.base_port(n)))
            assert_debug_tx(0, blk_end != target_height)
            assert_debug_tx(1)
        finally:
            for proc in procs:
                proc.terminate()
                proc.wait()
            procs = []
