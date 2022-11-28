import pytest

from .utils import ADDRS, CONTRACTS, KEYS, deploy_contract, send_transaction, w3_wait_for_new_blocks


def test_gas_eth_tx(geth, ethermint):
    tx_value = 10

    # send a transaction with geth
    geth_gas_price = geth.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": geth_gas_price}
    geth_receipt = send_transaction(geth.w3, tx, KEYS["validator"])

    # send an equivalent transaction with ethermint
    ethermint_gas_price = ethermint.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": ethermint_gas_price}
    ethermint_receipt = send_transaction(ethermint.w3, tx, KEYS["validator"])

    # ensure that the gasUsed is equivalent
    assert geth_receipt.gasUsed == ethermint_receipt.gasUsed


def test_gas_deployment(geth, ethermint):
    # deploy an identical contract on geth and ethermint
    # ensure that the gasUsed is equivalent
    _, geth_contract_receipt = deploy_contract(
        geth.w3,
        CONTRACTS["TestERC20A"])
    _, ethermint_contract_receipt = deploy_contract(
        ethermint.w3,
        CONTRACTS["TestERC20A"])
    assert geth_contract_receipt.gasUsed == ethermint_contract_receipt.gasUsed


def test_block_gas_limit(ethermint):
    tx_value = 10
    
    # get the block gas limit from the latest block
    w3_wait_for_new_blocks(ethermint.w3, 5)
    block = ethermint.w3.eth.get_block("latest")
    exceededGasLimit = block.gasLimit + 100

    # send a transaction exceeding the block gas limit
    ethermint_gas_price = ethermint.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gas": exceededGasLimit, "gasPrice": ethermint_gas_price}

    # expect an error due to the block gas limit
    with pytest.raises(Exception):
        send_transaction(ethermint.w3, tx, KEYS["validator"])
