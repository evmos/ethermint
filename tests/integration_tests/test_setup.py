from .utils import (
    CONTRACTS,
    deploy_contract,
    send_successful_transaction,
    w3_wait_for_block,
    w3_wait_for_new_blocks,
)

def test_setup_geth(geth):
    w3 = geth.w3
    w3_wait_for_block(w3, 5)

    # test utilities
    send_successful_transaction(w3)
    deploy_contract(w3, CONTRACTS["TestERC20A"])

    # ensure blocks are still being produced
    w3_wait_for_new_blocks(w3, 5)