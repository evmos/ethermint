import configparser
import json
import os
import re
import socket
import subprocess
import sys
import time
from pathlib import Path

import bech32
import eth_utils
import rlp
import toml
from dateutil.parser import isoparse
from dotenv import load_dotenv
from eth_account import Account
from hexbytes import HexBytes
from pystarport import ledger
from web3._utils.method_formatters import receipt_formatter
from web3._utils.transactions import fill_nonce, fill_transaction_defaults
from web3.datastructures import AttributeDict

load_dotenv(Path(__file__).parent.parent.parent / "scripts/.env")
Account.enable_unaudited_hdwallet_features()
ACCOUNTS = {
    "validator": Account.from_mnemonic(os.getenv("VALIDATOR1_MNEMONIC")),
    "community": Account.from_mnemonic(os.getenv("COMMUNITY_MNEMONIC")),
    "signer1": Account.from_mnemonic(os.getenv("SIGNER1_MNEMONIC")),
    "signer2": Account.from_mnemonic(os.getenv("SIGNER2_MNEMONIC")),
}
KEYS = {name: account.key for name, account in ACCOUNTS.items()}
ADDRS = {name: account.address for name, account in ACCOUNTS.items()}
ETHERMINT_ADDRESS_PREFIX = "ethm"
TEST_CONTRACTS = {
    "TestERC20A": "TestERC20A.sol",
}


def contract_path(name, filename):
    return (
        Path(__file__).parent
        / "contracts/artifacts/"
        / filename
        / (name + ".json")
    )


CONTRACTS = {
    **{
        name: contract_path(name, filename) for name, filename in TEST_CONTRACTS.items()
    },
}


def wait_for_fn(name, fn, *, timeout=240, interval=1):
    for i in range(int(timeout / interval)):
        result = fn()
        print("check", name, result)
        if result:
            break
        time.sleep(interval)
    else:
        raise TimeoutError(f"wait for {name} timeout")


def wait_for_block(cli, height, timeout=240):
    for i in range(timeout * 2):
        try:
            status = cli.status()
        except AssertionError as e:
            print(f"get sync status failed: {e}", file=sys.stderr)
        else:
            current_height = int(status["SyncInfo"]["latest_block_height"])
            if current_height >= height:
                break
            print("current block height", current_height)
        time.sleep(0.5)
    else:
        raise TimeoutError(f"wait for block {height} timeout")


def wait_for_new_blocks(cli, n, sleep=0.5):
    begin_height = int((cli.status())["SyncInfo"]["latest_block_height"])
    while True:
        time.sleep(sleep)
        cur_height = int((cli.status())["SyncInfo"]["latest_block_height"])
        if cur_height - begin_height >= n:
            break


def wait_for_block_time(cli, t):
    print("wait for block time", t)
    while True:
        now = isoparse((cli.status())["SyncInfo"]["latest_block_time"])
        print("block time now:", now)
        if now >= t:
            break
        time.sleep(0.5)


def wait_for_port(port, host="127.0.0.1", timeout=40.0):
    start_time = time.perf_counter()
    while True:
        try:
            with socket.create_connection((host, port), timeout=timeout):
                break
        except OSError as ex:
            time.sleep(0.1)
            if time.perf_counter() - start_time >= timeout:
                raise TimeoutError(
                    "Waited too long for the port {} on host {} to start accepting "
                    "connections.".format(port, host)
                ) from ex


def wait_for_ipc(path, timeout=40.0):
    print("wait for unix socket", path, "to be available")
    start_time = time.perf_counter()
    while True:
        if os.path.exists(path):
            break
        time.sleep(0.1)
        if time.perf_counter() - start_time >= timeout:
            raise TimeoutError(
                "Waited too long for the unix socket {path} to be available"
            )


def w3_wait_for_block(w3, height, timeout=240):
    for i in range(timeout * 2):
        try:
            current_height = w3.eth.block_number
        except AssertionError as e:
            print(f"get current block number failed: {e}", file=sys.stderr)
        else:
            if current_height >= height:
                break
            print("current block height", current_height)
        time.sleep(0.5)
    else:
        raise TimeoutError(f"wait for block {height} timeout")


def w3_wait_for_new_blocks(w3, n):
    begin_height = w3.eth.block_number
    while True:
        time.sleep(0.5)
        cur_height = w3.eth.block_number
        if cur_height - begin_height >= n:
            break


def get_ledger():
    return ledger.Ledger()


def parse_events(logs):
    return {
        ev["type"]: {attr["key"]: attr["value"] for attr in ev["attributes"]}
        for ev in logs[0]["events"]
    }


_next_unique = 0


def gen_base_port(worker_index):
    global _next_unique
    base_port = 10000 + (worker_index * 10 + _next_unique) * 100
    _next_unique += 1
    return base_port


def sign_single_tx_with_options(cli, tx_file, singer_name, **options):
    return json.loads(
        cli.cosmos_cli(0).raw(
            "tx",
            "sign",
            tx_file,
            from_=singer_name,
            home=cli.cosmos_cli(0).data_dir,
            keyring_backend="test",
            chain_id=cli.cosmos_cli(0).chain_id,
            node=cli.cosmos_cli(0).node_rpc,
            **options,
        )
    )


def find_balance(balances, denom):
    "find a denom in the coin list, return the amount, if not exists, return 0"
    for balance in balances:
        if balance["denom"] == denom:
            return int(balance["amount"])
    return 0


class ContractAddress(rlp.Serializable):
    fields = [
        ("from", rlp.sedes.Binary()),
        ("nonce", rlp.sedes.big_endian_int),
    ]


def contract_address(addr, nonce):
    return eth_utils.to_checksum_address(
        eth_utils.to_hex(
            eth_utils.keccak(
                rlp.encode(ContractAddress(eth_utils.to_bytes(hexstr=addr), nonce))
            )[12:]
        )
    )


def decode_bech32(addr):
    _, bz = bech32.bech32_decode(addr)
    return HexBytes(bytes(bech32.convertbits(bz, 5, 8)))


def bech32_to_eth(addr):
    return decode_bech32(addr).hex()


def eth_to_bech32(addr, prefix=ETHERMINT_ADDRESS_PREFIX):
    bz = bech32.convertbits(HexBytes(addr), 8, 5)
    return bech32.bech32_encode(prefix, bz)


def add_ini_sections(inipath, sections):
    ini = configparser.RawConfigParser()
    ini.read(inipath)
    for name, value in sections.items():
        ini.add_section(name)
        ini[name].update(value)
    with inipath.open("w") as fp:
        ini.write(fp)


def supervisorctl(inipath, *args):
    subprocess.run(
        (sys.executable, "-msupervisor.supervisorctl", "-c", inipath, *args),
        check=True,
    )


def deploy_contract(w3, jsonfile, args=(), key=KEYS["validator"]):
    """
    deploy contract and return the deployed contract instance
    """
    acct = Account.from_key(key)
    info = json.loads(jsonfile.read_text())
    contract = w3.eth.contract(abi=info["abi"], bytecode=info["bytecode"])
    tx = contract.constructor(*args).buildTransaction({"from": acct.address})
    txreceipt = send_transaction(w3, tx, key)
    assert txreceipt.status == 1
    address = txreceipt.contractAddress
    return w3.eth.contract(address=address, abi=info["abi"])


def sign_transaction(w3, tx, key=KEYS["validator"]):
    "fill default fields and sign"
    acct = Account.from_key(key)
    tx["from"] = acct.address
    tx = fill_transaction_defaults(w3, tx)
    tx = fill_nonce(w3, tx)
    return acct.sign_transaction(tx)


def send_transaction(w3, tx, key=KEYS["validator"]):
    signed = sign_transaction(w3, tx, key)
    txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
    return w3.eth.wait_for_transaction_receipt(txhash)


def cronos_address_from_mnemonics(mnemonics, prefix=ETHERMINT_ADDRESS_PREFIX):
    "return cronos address from mnemonics"
    acct = Account.from_mnemonic(mnemonics)
    return eth_to_bech32(acct.address, prefix)


def send_to_cosmos(gravity_contract, token_contract, recipient, amount, key=None):
    """
    do approve and sendToCronos on ethereum side
    """
    acct = Account.from_key(key)
    txreceipt = send_transaction(
        token_contract.web3,
        token_contract.functions.approve(
            gravity_contract.address, amount
        ).buildTransaction({"from": acct.address}),
        key,
    )
    assert txreceipt.status == 1, "approve failed"

    return send_transaction(
        gravity_contract.web3,
        gravity_contract.functions.sendToCronos(
            token_contract.address, HexBytes(recipient), amount
        ).buildTransaction({"from": acct.address}),
        key,
    )


class InlineTable(dict, toml.decoder.InlineTableDict):
    "a hack to dump inline table with toml library"
    pass


def dump_toml(obj):
    return toml.dumps(obj, encoder=toml.TomlPreserveInlineDictEncoder())


class Contract:
    "General contract."

    def __init__(self, contract_path, private_key=KEYS["validator"], chain_id=777):
        self.chain_id = chain_id
        self.account = Account.from_key(private_key)
        self.address = self.account.address
        self.private_key = private_key
        with open(contract_path) as f:
            json_data = f.read()
            contract_json = json.loads(json_data)
        self.bytecode = contract_json["bytecode"]
        self.abi = contract_json["abi"]
        self.contract = None
        self.w3 = None

    def deploy(self, w3):
        "Deploy Greeter contract on `w3` and return the receipt."
        if self.contract is None:
            self.w3 = w3
            contract = self.w3.eth.contract(abi=self.abi, bytecode=self.bytecode)
            transaction = contract.constructor().buildTransaction(
                {"chainId": self.chain_id, "from": self.address}
            )
            receipt = send_transaction(self.w3, transaction, self.private_key)
            self.contract = self.w3.eth.contract(
                address=receipt.contractAddress, abi=self.abi
            )
            return receipt
        else:
            return receipt


class Greeter(Contract):
    "Greeter contract."

    def transfer(self, string):
        "Call contract on `w3` and return the receipt."
        transaction = self.contract.functions.setGreeting(string).buildTransaction(
            {
                "chainId": self.chain_id,
                "from": self.address,
            }
        )
        receipt = send_transaction(self.w3, transaction, self.private_key)
        assert string == self.contract.functions.greet().call()
        return receipt


class RevertTestContract(Contract):
    "RevertTestContract contract."

    def transfer(self, value):
        "Call contract on `w3` and return the receipt."
        transaction = self.contract.functions.transfer(value).buildTransaction(
            {
                "chainId": self.chain_id,
                "from": self.address,
                "gas": 100000,  # skip estimateGas error
            }
        )
        receipt = send_transaction(self.w3, transaction, self.private_key)
        return receipt


def modify_command_in_supervisor_config(ini: Path, fn, **kwargs):
    "replace the first node with the instrumented binary"
    ini.write_text(
        re.sub(
            r"^command = (cronosd .*$)",
            lambda m: f"command = {fn(m.group(1))}",
            ini.read_text(),
            flags=re.M,
            **kwargs,
        )
    )


def build_batch_tx(w3, cli, txs, key=KEYS["validator"]):
    "return cosmos batch tx and eth tx hashes"
    signed_txs = [sign_transaction(w3, tx, key) for tx in txs]
    tmp_txs = [cli.build_evm_tx(signed.rawTransaction.hex()) for signed in signed_txs]

    msgs = [tx["body"]["messages"][0] for tx in tmp_txs]
    fee = sum(int(tx["auth_info"]["fee"]["amount"][0]["amount"]) for tx in tmp_txs)
    gas_limit = sum(int(tx["auth_info"]["fee"]["gas_limit"]) for tx in tmp_txs)

    tx_hashes = [signed.hash for signed in signed_txs]

    # build batch cosmos tx
    return {
        "body": {
            "messages": msgs,
            "memo": "",
            "timeout_height": "0",
            "extension_options": [
                {"@type": "/ethermint.evm.v1.ExtensionOptionsEthereumTx"}
            ],
            "non_critical_extension_options": [],
        },
        "auth_info": {
            "signer_infos": [],
            "fee": {
                "amount": [{"denom": "basetcro", "amount": str(fee)}],
                "gas_limit": str(gas_limit),
                "payer": "",
                "granter": "",
            },
        },
        "signatures": [],
    }, tx_hashes


def get_receipts_by_block(w3, blk):
    if isinstance(blk, int):
        blk = hex(blk)
    rsp = w3.provider.make_request("cronos_getTransactionReceiptsByBlock", [blk])
    if "error" not in rsp:
        rsp["result"] = [
            AttributeDict(receipt_formatter(item)) for item in rsp["result"]
        ]
    return rsp
