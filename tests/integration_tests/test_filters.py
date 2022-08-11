from web3 import Web3

import pytest
from .utils import (
    ADDRS,
    CONTRACTS,
    deploy_contract,
    send_transaction,
    sign_transaction,
)

# def test_pending_transaction_filter(cluster):
#     w3: Web3 = cluster.w3
#     flt = w3.eth.filter("pending")

#     # without tx
#     assert flt.get_new_entries() == [] # GetFilterChanges

#     # with tx
#     txhash = send_successful_transaction(w3)
#     assert txhash in flt.get_new_entries()

#     # without new txs since last call
#     assert flt.get_new_entries() == []


# def test_block_filter(cluster):
#     w3: Web3 = cluster.w3
#     flt = w3.eth.filter("latest")

#     # without tx
#     assert flt.get_new_entries() == []

#     # with tx
#     send_successful_transaction(w3)
#     blocks = flt.get_new_entries()
#     assert len(blocks) >= 1

#     # without new txs since last call
#     assert flt.get_new_entries() == []


# # Event Logs

# #     # through contract
# #     event_filter = mycontract.events.myEvent.createFilter(fromBlock='latest', argument_filters={'arg1':10})

# #     # manual
# #     event_filter = w3.eth.filter({"address": contract_address})

# #     # with existing filter id
# #     existing_filter = w3.eth.filter(filter_id="0x0")


def test_event_log_filter_from_contract(cluster):
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
    # TODO how to get the addr?
    myContract = deploy_contract(w3, CONTRACTS["Greeter"])
    assert myContract.caller.greet() == "Hello"

    flt = w3.eth.filter({"address": addr})

    # without tx
    assert flt.get_new_entries() == [] # GetFilterChanges
    assert flt.get_all_entries() == [] # GetFilterLogs

    # with tx
    tx = myContract.functions.setGreeting("world").buildTransaction()
    tx_receipt = send_transaction(w3, tx)
    assert tx_receipt.status == 1

    new_entries = flt.get_new_entries()
    assert len(new_entries) == 1


# TODO replace with send_and_get_hash
def send_successful_transaction(w3, addr=ADDRS["community"]):
    signed = sign_transaction(w3, {"to": addr, "value": 1000})
    txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
    receipt = w3.eth.wait_for_transaction_receipt(txhash)
    assert receipt.status == 1
    return txhash



