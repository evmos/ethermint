from typing import NamedTuple

from eth_typing import HexAddress
from eth_utils import abi
from hexbytes import HexBytes

from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    eth_to_bech32,
    send_transaction,
)


def get_balance(cli, addr, denom):
    return cli.balance(eth_to_bech32(addr), denom)


def test_call(ethermint):
    w3 = ethermint.w3
    addr = ADDRS["validator"]
    amount = 100
    contract, _ = deploy_contract(w3, CONTRACTS["TestBank"])
    tx = contract.functions.nativeMint(amount).build_transaction({"from": addr})
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1, "expect success"

    # query balance through contract
    assert contract.caller.nativeBalanceOf(addr) == amount
    # query balance through cosmos rpc
    cli = ethermint.cosmos_cli()
    denom = "evm/" + contract.address
    assert get_balance(cli, addr, denom) == amount

    # test exception revert
    tx = contract.functions.nativeMintRevert(amount).build_transaction(
        {"from": addr, "gas": 210000}
    )
    receipt = send_transaction(w3, tx)
    assert receipt.status == 0, "expect failure"

    # check balance don't change
    assert contract.caller.nativeBalanceOf(addr) == amount
    # query balance through cosmos rpc
    cli = ethermint.cosmos_cli()
    assert get_balance(cli, addr, denom) == amount


def test_delegate(ethermint):
    w3 = ethermint.w3
    addr = ADDRS["validator"]
    amount = 100
    _, res = deploy_contract(w3, CONTRACTS["TestBank"])
    bank = res["contractAddress"]
    contract, _ = deploy_contract(w3, CONTRACTS["TestBankDelegate"])
    data = {"from": addr}
    tx = contract.functions.nativeMint(bank, amount).build_transaction(data)
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1, "expect success"

    # query balance through contract
    assert contract.caller.nativeBalanceOf(bank, addr) == amount
    # query balance through cosmos rpc
    cli = ethermint.cosmos_cli()
    denom = "evm/" + contract.address
    assert get_balance(cli, addr, denom) == amount

    # test exception revert
    tx = contract.functions.nativeMintRevert(bank, amount).build_transaction(
        {"from": addr, "gas": 210000}
    )
    receipt = send_transaction(w3, tx)
    assert receipt.status == 0, "expect failure"

    # check balance don't change
    assert contract.caller.nativeBalanceOf(bank, addr) == amount
    # query balance through cosmos rpc
    cli = ethermint.cosmos_cli()
    assert get_balance(cli, addr, denom) == amount


class Params(NamedTuple):
    address: HexAddress
    amount: int


def test_transfer(ethermint):
    w3 = ethermint.w3
    amount = 500000000000000000000
    name = CONTRACTS["TestTransfer"]
    contract, res = deploy_contract(w3, name, (), KEYS["validator"], amount)
    contract_address = res["contractAddress"]
    denom = "aphoton"
    addr = ADDRS["signer2"]
    recipient = ADDRS["signer1"]
    cli = ethermint.cosmos_cli()
    recipient_balance = get_balance(cli, recipient, denom)
    contract_balance = get_balance(cli, contract_address, denom)
    success = [1000, 20000]
    fail = [300000, amount * 2]
    params = [Params(recipient, val) for val in (success + fail)]
    tx = contract.functions.recursiveTransfer(params).build_transaction({"from": addr})
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1
    assert receipt.logs[0]["topics"] == [
        HexBytes(abi.event_signature_to_log_topic("Result(bool)")),
        HexBytes(b"\x00" * 32),
    ]
    expect_diff = sum(success)
    assert recipient_balance + expect_diff == get_balance(cli, recipient, denom)
    assert contract_balance - expect_diff == get_balance(cli, contract_address, denom)
