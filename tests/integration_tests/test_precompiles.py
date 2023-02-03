import pytest

from .utils import ADDRS, CONTRACTS, deploy_contract, eth_to_bech32, send_transaction


@pytest.mark.parametrize("suffix", [""])
def test_precompiles(ethermint, suffix):
    w3 = ethermint.w3
    addr = ADDRS["validator"]
    amount = 100
    contract, _ = deploy_contract(w3, CONTRACTS["TestBank"])
    data = {"from": addr}
    tx = contract.functions["nativeMint" + suffix](amount).build_transaction(data)
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1, "expect success"

    # query balance through contract
    balance_of = getattr(contract.caller, "nativeBalanceOf" + suffix)
    assert balance_of(addr) == amount
    # query balance through cosmos rpc
    cli = ethermint.cosmos_cli()
    assert cli.balance(eth_to_bech32(addr), "evm/" + contract.address) == amount

    # test exception revert
    tx = contract.functions["nativeMintRevert" + suffix](amount).build_transaction(
        {"from": addr, "gas": 210000}
    )
    receipt = send_transaction(w3, tx)
    assert receipt.status == 0, "expect failure"

    # check balance don't change
    assert balance_of(addr) == amount
    # query balance through cosmos rpc
    cli = ethermint.cosmos_cli()
    assert cli.balance(eth_to_bech32(addr), "evm/" + contract.address) == amount
