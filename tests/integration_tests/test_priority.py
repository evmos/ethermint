from .network import Ethermint
from .utils import KEYS, sign_transaction

PRIORITY_REDUCTION = 1000000


def effective_gas_price(tx, base_fee):
    if "maxFeePerGas" in tx:
        # dynamic fee tx
        return min(base_fee + tx["maxPriorityFeePerGas"], tx["maxFeePerGas"])
    else:
        # legacy tx
        return tx["gasPrice"]


def tx_priority(tx, base_fee):
    if "maxFeePerGas" in tx:
        # dynamic fee tx
        return (
            min(tx["maxPriorityFeePerGas"], tx["maxFeePerGas"] - base_fee)
            // PRIORITY_REDUCTION
        )
    else:
        # legacy tx
        return (tx["gasPrice"] - base_fee) // PRIORITY_REDUCTION


def test_priority(ethermint: Ethermint):
    """
    test priorities of different tx types
    """
    w3 = ethermint.w3
    amount = 10000
    base_fee = w3.eth.get_block("latest").baseFeePerGas

    # [ ( sender, tx ), ... ]
    # use different senders to avoid nonce conflicts
    test_cases = [
        (
            "validator",
            {
                "to": "0x0000000000000000000000000000000000000000",
                "value": amount,
                "gas": 21000,
                "maxFeePerGas": base_fee + PRIORITY_REDUCTION * 6,
                "maxPriorityFeePerGas": 0,
            },
        ),
        (
            "community",
            {
                "to": "0x0000000000000000000000000000000000000000",
                "value": amount,
                "gas": 21000,
                "gasPrice": base_fee + PRIORITY_REDUCTION * 2,
            },
        ),
        (
            "signer2",
            {
                "to": "0x0000000000000000000000000000000000000000",
                "value": amount,
                "gasPrice": base_fee + PRIORITY_REDUCTION * 4,
                "accessList": [
                    {
                        "address": "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae",
                        "storageKeys": (
                            "0x00000000000000000000000000000000000000000000000000000000"
                            "00000003",
                            "0x00000000000000000000000000000000000000000000000000000000"
                            "00000007",
                        ),
                    }
                ],
            },
        ),
        (
            "signer1",
            {
                "to": "0x0000000000000000000000000000000000000000",
                "value": amount,
                "gas": 21000,
                "maxFeePerGas": base_fee + PRIORITY_REDUCTION * 6,
                "maxPriorityFeePerGas": PRIORITY_REDUCTION * 6,
            },
        ),
    ]

    # test cases are ordered by priority
    priorities = [tx_priority(tx, base_fee) for _, tx in test_cases]
    assert all(a < b for a, b in zip(priorities, priorities[1:]))

    signed = [sign_transaction(w3, tx, key=KEYS[sender]) for sender, tx in test_cases]
    # send the txs from low priority to high,
    # but the later sent txs should be included earlier.
    txhashes = [w3.eth.send_raw_transaction(tx.rawTransaction) for tx in signed]

    receipts = [w3.eth.wait_for_transaction_receipt(txhash) for txhash in txhashes]
    print(receipts)
    assert all(receipt.status == 1 for receipt in receipts), "expect all txs success"

    # the later txs should be included earlier because of higher priority
    # FIXME there's some non-deterministics due to mempool logic
    assert all(included_earlier(r2, r1) for r1, r2 in zip(receipts, receipts[1:]))


def included_earlier(receipt1, receipt2):
    "returns true if receipt1 is earlier than receipt2"
    if receipt1.blockNumber < receipt2.blockNumber:
        return True
    elif receipt1.blockNumber == receipt2.blockNumber:
        return receipt1.transactionIndex < receipt2.transactionIndex
    else:
        return False
