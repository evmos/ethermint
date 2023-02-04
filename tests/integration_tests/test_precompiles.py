from .utils import ADDRS, CONTRACTS, deploy_contract, eth_to_bech32, send_transaction


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
    assert cli.balance(eth_to_bech32(addr), "evm/" + contract.address) == amount

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
    assert cli.balance(eth_to_bech32(addr), "evm/" + contract.address) == amount


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
    assert cli.balance(eth_to_bech32(addr), "evm/" + contract.address) == amount

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
    assert cli.balance(eth_to_bech32(addr), "evm/" + contract.address) == amount
