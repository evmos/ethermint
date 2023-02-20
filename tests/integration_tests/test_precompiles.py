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
    cli = ethermint.cosmos_cli()
    native_denom = "aphoton"
    sender = ADDRS["signer1"]
    keys = KEYS["signer1"]
    amount = 100
    contract, _ = deploy_contract(w3, CONTRACTS["TestBank"])
    denom = "evm/" + contract.address

    def assert_sender_balance(tx, expect_status, amount):
        balance = get_balance(cli, sender, native_denom)
        receipt = send_transaction(w3, tx, keys)
        assert receipt.status == expect_status
        fee = receipt["cumulativeGasUsed"] * receipt["effectiveGasPrice"]
        current = get_balance(cli, sender, native_denom)
        assert balance == current + fee + amount

    def assert_crc20_balance(address, amt):
        # query balance through contract
        assert contract.caller.nativeBalanceOf(address) == amt
        # query balance through cosmos rpc
        assert get_balance(cli, address, denom) == amt

    # test mint
    tx = contract.functions.nativeMint(amount).build_transaction({"from": sender})
    assert_sender_balance(tx, 1, amount)
    assert_crc20_balance(sender, amount)

    # test exception revert
    tx = contract.functions.nativeMintRevert(amount).build_transaction(
        {"from": sender, "gas": 210000}
    )
    assert_sender_balance(tx, 0, 0)
    # check balance don't change
    assert_crc20_balance(sender, amount)

    # test transfer
    recipient = ADDRS["signer2"]
    transfer_amt = 10
    recipient_balance = get_balance(cli, recipient, native_denom)
    tx = contract.functions.nativeTransfer(recipient, transfer_amt).build_transaction(
        {"from": sender}
    )
    assert_sender_balance(tx, 1, transfer_amt)
    assert_crc20_balance(sender, amount - transfer_amt)
    assert get_balance(cli, recipient, native_denom) == recipient_balance + transfer_amt
    assert_crc20_balance(recipient, transfer_amt)


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
