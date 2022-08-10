from web3 import Web3

from .utils import ADDRS, CONTRACTS, deploy_contract, send_transaction, sign_transaction

def test_pending_transaction_filter(ethermint, geth):
  w3_eth: Web3 = ethermint.w3
  w3_geth: Web3 = geth.w3

  flt_eth = w3_eth.eth.filter("pending")
  flt_geth = w3_geth.eth.filter("pending")

  # without tx
  assert flt_eth.get_new_entries() == []
  assert flt_eth.get_new_entries() == flt_geth.get_new_entries()

  # with tx
  txhash1, receipt1 = send_transaction(w3_eth)
  txhash2, receipt2 = send_transaction(w3_geth)

  assert receipt1, receipt2
  assert txhash1, txhash2
  assert txhash1 in flt_eth.get_new_entries()
  assert txhash2 in flt_geth.get_new_entries()

def send_transaction(w3):
  signed = sign_transaction(w3, {"to": ADDRS["community"], "value": 1000})
  txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
  receipt = w3.eth.wait_for_transaction_receipt(txhash)
  assert receipt.status == 1

  return txhash, receipt
