import os

from eth_account import Account

from .utils import ADDRS, w3_wait_for_new_blocks


def test_get_transaction_count(ethermint_rpc_ws, geth):
    for p in [ethermint_rpc_ws, geth]:
        w3 = p.w3
        blk = hex(w3.eth.block_number)
        sender = ADDRS["validator"]

        # derive a new address
        account_path = "m/44'/60'/0'/0/1"
        mnemonic = os.getenv("COMMUNITY_MNEMONIC")
        receiver = (Account.from_mnemonic(mnemonic, account_path=account_path)).address
        n0 = w3.eth.get_transaction_count(receiver, blk)
        # ensure transaction send in new block
        w3_wait_for_new_blocks(w3, 1, sleep=0.1)
        txhash = w3.eth.send_transaction(
            {
                "from": sender,
                "to": receiver,
                "value": 1000,
            }
        )
        receipt = w3.eth.wait_for_transaction_receipt(txhash)
        assert receipt.status == 1
        [n1, n2] = [w3.eth.get_transaction_count(receiver, b) for b in [blk, "latest"]]
        assert n0 == n1
        assert n0 == n2
