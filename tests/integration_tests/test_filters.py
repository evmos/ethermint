from web3 import Web3

from .utils import ADDRS, CONTRACTS, deploy_contract, send_transaction, sign_transaction

def test_pending_transaction_filter(ethermint, geth):
  w3_ethm: Web3 = ethermint.w3
  w3_geth: Web3 = geth.w3
  flt_eth = w3_ethm.eth.filter("pending")
  flt_geth = w3_geth.eth.filter("pending")

  # without tx
  assert flt_eth.get_new_entries() == []
  assert flt_eth.get_new_entries() == flt_geth.get_new_entries()

  # with tx
  txhash1, receipt1 = send_successful_transaction(w3_ethm)
  txhash2, receipt2 = send_successful_transaction(w3_geth)
  assert receipt1, receipt2
  assert txhash1, txhash2
  assert txhash1 in flt_eth.get_new_entries()
  assert txhash2 in flt_geth.get_new_entries()

def test_block_filter(ethermint, geth):
  w3_ethm: Web3 = ethermint.w3
  w3_geth: Web3 = geth.w3
  flt_eth = w3_ethm.eth.filter("latest")
  flt_geth = w3_geth.eth.filter("latest")

  # without tx
  assert flt_eth.get_new_entries() == []
  assert flt_eth.get_new_entries() == flt_geth.get_new_entries()

  # with tx
  send_successful_transaction(w3_ethm)
  send_successful_transaction(w3_geth)
  blocks1 = flt_eth.get_new_entries()
  blocks2 = flt_geth.get_new_entries()

  assert len(blocks1) >= 1
  assert len(blocks2) >= 1
  assert blocks1, blocks2

def send_successful_transaction(w3):
  signed = sign_transaction(w3, {"to": ADDRS["community"], "value": 1000})
  txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
  receipt = w3.eth.wait_for_transaction_receipt(txhash)
  assert receipt.status == 1
  return txhash, receipt

def test_event_log_filter(ethermint, geth):
  w3_ethm: Web3 = ethermint.w3
  w3_geth: Web3 = geth.w3

  contract1 = deploy_contract(w3_ethm, CONTRACTS["Greeter"])
  contract2 = deploy_contract(w3_geth, CONTRACTS["Greeter"])

  assert contract1.caller.greet() == "Hello"
  assert contract1.caller.greet() == contract2.caller.greet()

  current_height1 = hex(w3_ethm.eth.get_block_number())
  event_filter1 = contract1.events.ChangeGreeting.createFilter(
    fromBlock=current_height1
  )

  current_height2 = hex(w3_ethm.eth.get_block_number())
  event_filter2 = contract1.events.ChangeGreeting.createFilter(
    fromBlock=current_height2
  )

  tx1 = contract1.functions.setGreeting("world").buildTransaction()
  tx2 = contract2.functions.setGreeting("world").buildTransaction()
  tx_receipt1 = send_transaction(w3_ethm, tx1)
  tx_receipt2 = send_transaction(w3_geth, tx2)
  assert tx_receipt1.status == 1
  assert tx_receipt2.status == 1

  log1 = contract1.events.ChangeGreeting().processReceipt(tx_receipt1)[0]
  log2 = contract1.events.ChangeGreeting().processReceipt(tx_receipt1)[0]
  assert log1["event"] == log2["event"]

  new_entries1 = event_filter1.get_new_entries()
  new_entries2 = event_filter2.get_new_entries()
  assert len(new_entries1) == len(new_entries2)
  # print(f"get event: {new_entries}")
  assert new_entries1[0] == new_entries2[0]
  assert  contract1.caller.greet() == "world"
  assert  contract2.caller.greet() == "world"