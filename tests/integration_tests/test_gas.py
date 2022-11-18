from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    w3_wait_for_new_blocks,
)


def test_gas_eth_tx(geth, ethermint):
    tx_value = 10

    # send a transaction with geth
    geth_gas_price = geth.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": geth_gas_price}
    geth_reciept = send_transaction(geth.w3, tx, KEYS["validator"])

    # send an equivalent transaction with ethermint
    ethermint_gas_price = ethermint.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": ethermint_gas_price}
    ethermint_reciept = send_transaction(ethermint.w3, tx, KEYS["validator"])

    # ensure that the gasUsed is equivalent
    assert geth_reciept.gasUsed == ethermint_reciept.gasUsed

def test_gas_deployment(geth, ethermint):
     # deploy an identical contract on geth and ethermint
     # ensure that the gasUsed is equivalent
    _, geth_contract_reciept = deploy_contract(
        geth.w3,
        CONTRACTS["TestERC20A"])
    _, ethermint_contract_reciept = deploy_contract(
        ethermint.w3,
        CONTRACTS["TestERC20A"])
    assert geth_contract_reciept.gasUsed == ethermint_contract_reciept.gasUsed

def test_gas_call(geth, ethermint):
    function_input = 10

    # deploy an identical contract on geth and ethermint
    # ensure that the contract has a function which consumes non-trivial gas
    geth_contract, _ = deploy_contract(
        geth.w3,
        CONTRACTS["BurnGas"])
    ethermint_contract, _ = deploy_contract(
        ethermint.w3,
        CONTRACTS["BurnGas"])

    # call the contract locally (eth_call) and compare gas estimates
    geth_estimated_gas = geth_contract.functions.burnGas(function_input).estimate_gas()
    ethermint_estimated_gas = ethermint_contract.functions.burnGas(function_input).estimate_gas()
    assert geth_estimated_gas == ethermint_estimated_gas
