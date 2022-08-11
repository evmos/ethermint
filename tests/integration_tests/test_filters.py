from web3 import Web3

import pytest
from .utils import (
    CONTRACTS,
    deploy_contract,
    send_transaction,
    send_successful_transaction,
)

def test_pending_transaction_filter(cluster):
    w3: Web3 = cluster.w3
    flt = w3.eth.filter("pending")

    # without tx
    assert flt.get_new_entries() == [] # GetFilterChanges

    # with tx
    txhash = send_successful_transaction(w3)
    assert txhash in flt.get_new_entries()

    # without new txs since last call
    assert flt.get_new_entries() == []


def test_block_filter(cluster):
    w3: Web3 = cluster.w3
    flt = w3.eth.filter("latest")

    # without tx
    assert flt.get_new_entries() == []

    # with tx
    send_successful_transaction(w3)
    blocks = flt.get_new_entries()
    assert len(blocks) >= 1

    # without new txs since last call
    assert flt.get_new_entries() == []


def test_event_log_filter_by_contract(cluster):
    w3: Web3 = cluster.w3
    myContract = deploy_contract(w3, CONTRACTS["Greeter"])
    assert myContract.caller.greet() == "Hello"

    # Create new filter from contract
    current_height = hex(w3.eth.get_block_number())
    flt = myContract.events.ChangeGreeting.createFilter(
        fromBlock=current_height
    )

    # without tx
    assert flt.get_new_entries() == [] # GetFilterChanges
    assert flt.get_all_entries() == [] # GetFilterLogs

    # with tx
    tx = myContract.functions.setGreeting("world").buildTransaction()
    tx_receipt = send_transaction(w3, tx)
    assert tx_receipt.status == 1

    log = myContract.events.ChangeGreeting().processReceipt(tx_receipt)[0]
    assert log["event"] == "ChangeGreeting"

    new_entries = flt.get_new_entries()
    assert len(new_entries) == 1
    assert new_entries[0] == log
    assert myContract.caller.greet() == "world"

    # without new txs since last call
    assert flt.get_new_entries() == []
    assert flt.get_all_entries() == new_entries

    # Uninstall
    assert w3.eth.uninstall_filter(flt.filter_id) == True
    assert w3.eth.uninstall_filter(flt.filter_id) == False
    with pytest.raises(Exception):
        flt.get_all_entries()


def test_event_log_filter_by_address(cluster):
    w3: Web3 = cluster.w3

    myContract = deploy_contract(w3, CONTRACTS["Greeter"])
    assert myContract.caller.greet() == "Hello"

    flt = w3.eth.filter({"address": myContract.address})
    flt2 = w3.eth.filter({"address": "0x0000000000000000000000000000000000000000"})

    # without tx
    assert flt.get_new_entries() == [] # GetFilterChanges
    assert flt.get_all_entries() == [] # GetFilterLogs

    # with tx
    tx = myContract.functions.setGreeting("world").buildTransaction()
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1

    assert len(flt.get_new_entries()) == 1
    assert len(flt2.get_new_entries()) == 0


def test_get_logs(cluster):
    w3: Web3 = cluster.w3

    myContract = deploy_contract(w3, CONTRACTS["Greeter"])

    # without tx
    assert w3.eth.get_logs({"address": myContract.address}) == []
    assert w3.eth.get_logs({"address": "0x0000000000000000000000000000000000000000"}) == []

    # with tx
    tx = myContract.functions.setGreeting("world").buildTransaction()
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1

    assert len(w3.eth.get_logs({"address": myContract.address})) == 1
    assert len(w3.eth.get_logs({"address": "0x0000000000000000000000000000000000000000"})) == 0