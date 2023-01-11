import time

from .utils import CONTRACTS, KEYS, deploy_contract, send_transaction


def test_message_call(ethermint):
    "stress test the evm by doing message calls as much as possible"
    w3 = ethermint.w3
    contract, _ = deploy_contract(
        w3,
        CONTRACTS["TestMessageCall"],
        key=KEYS["community"],
    )
    iterations = 6500
    tx = contract.functions.test(iterations).build_transaction()

    begin = time.time()
    tx["gas"] = w3.eth.estimate_gas(tx)
    diff = time.time() - begin
    print("diff: ", diff)
    assert diff < 5  # should finish in reasonable time

    receipt = send_transaction(w3, tx, KEYS["community"])
    assert receipt.status == 1, "shouldn't fail"
    assert len(receipt.logs) == iterations
