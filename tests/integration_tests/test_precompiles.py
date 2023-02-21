import pytest
import web3

from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    eth_to_bech32,
    module_address,
    send_transaction,
)


def get_balance(cli, addr, denom):
    return cli.balance(eth_to_bech32(addr), denom)


def test_call(ethermint):
    w3 = ethermint.w3
    cli = ethermint.cosmos_cli()
    addr = ADDRS["signer1"]
    keys = KEYS["signer1"]
    contract, _ = deploy_contract(w3, CONTRACTS["TestBank"], (), keys)
    denom = "evm/" + contract.address

    def assert_balance(tx, expect_status, amt):
        balance = get_balance(cli, addr, denom)
        assert balance == contract.caller.nativeBalanceOf(addr)
        crc20_balance = contract.caller.balanceOf(addr)
        receipt = send_transaction(w3, tx, keys)
        assert receipt.status == expect_status
        balance += amt
        assert balance == get_balance(cli, addr, denom)
        assert balance == contract.caller.nativeBalanceOf(addr)
        assert crc20_balance - amt == contract.caller.balanceOf(addr)

    # test mint
    amt1 = 100
    data = {"from": addr}
    tx = contract.functions.moveToNative(amt1).build_transaction(data)
    assert_balance(tx, 1, amt1)

    # test exception revert
    tx = contract.functions.moveToNativeRevert(amt1).build_transaction(
        {"from": addr, "gas": 210000}
    )
    assert_balance(tx, 0, 0)

    # test burn
    amt2 = 50
    tx = contract.functions.moveFromNative(amt2).build_transaction(data)
    assert_balance(tx, 1, -amt2)

    # test transfer
    amt3 = 10
    addr2 = ADDRS["signer2"]
    tx = contract.functions.nativeTransfer(addr2, amt3).build_transaction(data)
    balance = get_balance(cli, addr, denom)
    assert balance == contract.caller.nativeBalanceOf(addr)
    crc20_balance = contract.caller.balanceOf(addr)

    balance2 = get_balance(cli, addr2, denom)
    assert balance2 == contract.caller.nativeBalanceOf(addr2)
    crc20_balance2 = contract.caller.balanceOf(addr2)

    receipt = send_transaction(w3, tx, keys)
    assert receipt.status == 1

    balance -= amt3
    assert balance == get_balance(cli, addr, denom)
    assert balance == contract.caller.nativeBalanceOf(addr)
    assert crc20_balance - amt3 == contract.caller.balanceOf(addr)

    balance2 += amt3
    assert balance2 == get_balance(cli, addr2, denom)
    assert balance2 == contract.caller.nativeBalanceOf(addr2)
    assert crc20_balance2 + amt3 == contract.caller.balanceOf(addr2)

    # test transfer to blocked address
    recipient = module_address("evm")
    amt4 = 20
    with pytest.raises(web3.exceptions.ContractLogicError):
        contract.functions.nativeTransfer(recipient, amt4).build_transaction(data)


@pytest.mark.parametrize("suffix", ["Delegate", "Static"])
def test_readonly_call(ethermint, suffix):
    w3 = ethermint.w3
    cli = ethermint.cosmos_cli()
    addr = ADDRS["signer1"]
    keys = KEYS["signer1"]
    contract, _ = deploy_contract(w3, CONTRACTS["TestBank"], (), keys)
    denom = "evm/" + contract.address
    native_balance_of = getattr(contract.caller, "nativeBalanceOf" + suffix)

    def get_balances():
        return [
            contract.caller.balanceOf(addr),
            native_balance_of(addr),
            get_balance(cli, addr, denom),
        ]

    balances = get_balances()
    # test mint
    amt1 = 100
    data = {"from": addr}
    with pytest.raises(web3.exceptions.ContractLogicError):
        contract.functions["moveToNative" + suffix](amt1).build_transaction(data)

    # test exception revert
    tx = contract.functions["moveToNativeRevert" + suffix](amt1).build_transaction(
        {"from": addr, "gas": 210000}
    )
    receipt = send_transaction(w3, tx, keys)
    assert receipt.status == 0

    # test burn
    amt2 = 50
    with pytest.raises(web3.exceptions.ContractLogicError):
        contract.functions["moveFromNative" + suffix](amt2).build_transaction(data)

    # test transfer
    amt3 = 10
    addr2 = ADDRS["signer2"]
    native_transfer = contract.functions["nativeTransfer" + suffix]
    with pytest.raises(web3.exceptions.ContractLogicError):
        native_transfer(addr2, amt3).build_transaction(data)

    # balance no change
    assert balances == get_balances()


def test_nested(ethermint):
    w3 = ethermint.w3
    addr = ADDRS["validator"]
    amount = 100
    _, res = deploy_contract(w3, CONTRACTS["TestBank"])
    contract, _ = deploy_contract(w3, CONTRACTS["TestBankCaller"])
    data = {"from": addr}
    tx = contract.functions.mint(res["contractAddress"], amount).build_transaction(data)
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1, "expect success"
    assert contract.caller.getLastState() == 1
    denom = "evm/" + contract.address
    assert get_balance(ethermint.cosmos_cli(), addr, denom) == 0
