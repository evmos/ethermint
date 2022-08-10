from web3 import Web3

from .utils import ADDRS, CONTRACTS, deploy_contract, send_transaction, sign_transaction

def test_pending_transaction_filter(cluster):
  w3: Web3 = cluster.w3
  flt = w3.eth.filter("pending")
  assert flt.get_new_entries() == []

  txhash = send_successful_transaction(w3)
  assert txhash in flt.get_new_entries()

def test_block_filter(cluster):
  w3: Web3 = cluster.w3
  flt = w3.eth.filter("latest")
  assert flt.get_new_entries() == []

  send_successful_transaction(w3)
  blocks = flt.get_new_entries()
  assert len(blocks) >= 1

# TODO replace with send_and_get_hash
def send_successful_transaction(w3):
  signed = sign_transaction(w3, {"to": ADDRS["community"], "value": 1000})
  txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
  receipt = w3.eth.wait_for_transaction_receipt(txhash)
  assert receipt.status == 1
  return txhash

def test_event_log_filter(cluster):
  w3: Web3 = cluster.w3
  myContract = deploy_contract(w3, CONTRACTS["Greeter"])
  assert myContract.caller.greet() == "Hello"

  current_height = hex(w3.eth.get_block_number())
  event_filter = myContract.events.ChangeGreeting.createFilter(
    fromBlock=current_height
  )

  tx = myContract.functions.setGreeting("world").buildTransaction()
  tx_receipt = send_transaction(w3, tx)
  assert tx_receipt.status == 1

  log = myContract.events.ChangeGreeting().processReceipt(tx_receipt)[0]
  assert log["event"] == "ChangeGreeting"

  new_entries = event_filter.get_new_entries()
  assert len(new_entries) == 1
  assert new_entries[0] == log
  assert myContract.caller.greet() == "world"
