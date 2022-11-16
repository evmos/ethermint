from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    send_successful_transaction,
    w3_wait_for_block,
    w3_wait_for_new_blocks,
)

def test_setup_geth(geth):
    w3 = geth.w3
    w3_wait_for_block(w3, 5)

    # test utilities
    send_successful_transaction(w3)
    print("successfully sent transaction")
    deploy_contract(w3, CONTRACTS["TestERC20A"])
    print("successfully deployed contract")

    # ensure blocks are still being produced
    w3_wait_for_new_blocks(w3, 5)