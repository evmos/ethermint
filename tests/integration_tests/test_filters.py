import pytest
from web3 import Web3
from eth_abi import abi

from .utils import (
    ADDRS,
    CONTRACTS,
    deploy_contract,
    send_successful_transaction,
    send_transaction,
)

# ChangeGreeting topic from Greeter contract calculated from event signature
CHANGE_GREETING_TOPIC = Web3.keccak(text="ChangeGreeting(address,string)")
# ERC-20 Transfer event topic
TRANSFER_TOPIC = Web3.keccak(text="Transfer(address,address,uint256)")


def test_pending_transaction_filter(cluster):
    w3: Web3 = cluster.w3
    flt = w3.eth.filter("pending")

    # without tx
    assert flt.get_new_entries() == []  # GetFilterChanges

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

    # check if the returned block hash is correct
    # getBlockByHash
    block = w3.eth.get_block(blocks[0])
    # block should exist
    assert block.hash == blocks[0]

    # without new txs since last call
    assert flt.get_new_entries() == []


def test_event_log_filter_by_contract(cluster):
    w3: Web3 = cluster.w3
    contract, _ = deploy_contract(w3, CONTRACTS["Greeter"])
    assert contract.caller.greet() == "Hello"

    # Create new filter from contract
    current_height = hex(w3.eth.get_block_number())
    flt = contract.events.ChangeGreeting.createFilter(fromBlock=current_height)

    # without tx
    assert flt.get_new_entries() == []  # GetFilterChanges
    assert flt.get_all_entries() == []  # GetFilterLogs

    # with tx
    tx = contract.functions.setGreeting("world").build_transaction()
    tx_receipt = send_transaction(w3, tx)
    assert tx_receipt.status == 1

    log = contract.events.ChangeGreeting().processReceipt(tx_receipt)[0]
    assert log["event"] == "ChangeGreeting"

    new_entries = flt.get_new_entries()
    assert len(new_entries) == 1
    assert new_entries[0] == log
    assert contract.caller.greet() == "world"

    # without new txs since last call
    assert flt.get_new_entries() == []
    assert flt.get_all_entries() == new_entries

    # Uninstall
    assert w3.eth.uninstall_filter(flt.filter_id)
    assert not w3.eth.uninstall_filter(flt.filter_id)
    with pytest.raises(Exception):
        flt.get_all_entries()


def test_event_log_filter_by_address(cluster):
    w3: Web3 = cluster.w3

    contract, _ = deploy_contract(w3, CONTRACTS["Greeter"])
    assert contract.caller.greet() == "Hello"

    flt = w3.eth.filter({"address": contract.address})
    flt2 = w3.eth.filter({"address": ADDRS["validator"]})

    # without tx
    assert flt.get_new_entries() == []  # GetFilterChanges
    assert flt.get_all_entries() == []  # GetFilterLogs

    # with tx
    tx = contract.functions.setGreeting("world").build_transaction()
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1

    assert len(flt.get_new_entries()) == 1
    assert len(flt2.get_new_entries()) == 0


def test_multiple_filters(cluster):
    w3: Web3 = cluster.w3

    contract, _ = deploy_contract(w3, CONTRACTS["Greeter"])
    # test the contract was deployed successfully
    assert contract.caller.greet() == "Hello"

    new_greeting = "hello, world"

    # calculate topic from event signature
    topic = CHANGE_GREETING_TOPIC
    # another topic not related to the contract deployed
    another_topic = TRANSFER_TOPIC

    test_cases = [
        {
            "name": "register multiple filters and check for updates",
            "filters": [
                {
                    "params": {"address": contract.address},
                    "exp_len": 1,
                },
                {
                    "params": {"topics": [topic.hex()]},
                    "exp_len": 1,
                },
                {
                    "params": {
                        "topics": [
                            topic.hex(),
                            another_topic.hex(),
                        ],  # 'with all topics' condition
                    },
                    "exp_len": 0,
                },
                {
                    "params": {
                        "topics": [
                            [topic.hex(), another_topic.hex()]
                        ],  # 'with any topic' condition
                    },
                    "exp_len": 1,
                },
                {
                    "params": {
                        "address": contract.address,
                        "topics": [[topic.hex(), another_topic.hex()]],
                    },
                    "exp_len": 1,
                },
                {
                    "params": {
                        "fromBlock": 1,
                        "toBlock": 2,
                        "address": contract.address,
                        "topics": [[topic.hex(), another_topic.hex()]],
                    },
                    "exp_len": 0,
                },
                {
                    "params": {
                        "fromBlock": 1,
                        "toBlock": "latest",
                        "address": contract.address,
                        "topics": [[topic.hex(), another_topic.hex()]],
                    },
                    "exp_len": 1,
                },
            ],
        },
        # {
        #     "name": "register some filters, then remove the last filter and check for updates again",
        #     "filters": [],
        # },
        {
            "name": "register more filters than allowed (default: 200) - should get error",
            "register_err": "error creating filter: max limit reached",
            "filters": make_filter_array(205),
        },
    ]

    for tc in test_cases:
        print("Case: {}".format(tc["name"]))

        # register the filters
        fltrs = []
        try:
            for flt in tc["filters"]:
                fltrs.append(w3.eth.filter(flt["params"]))
        except Exception as err:
            if tc["register_err"] is None:
                print(f"Unexpected {err=}, {type(err)=}")
                raise

            # if exception was expected when registering filters
            # the test is finished
            else:
                assert tc["register_err"] in str(err)
                continue

        # without tx: filters should not return any entries
        for flt in fltrs:
            assert flt.get_new_entries() == []  # GetFilterChanges

        # with tx
        tx = contract.functions.setGreeting(new_greeting).build_transaction()
        receipt = send_transaction(w3, tx)
        assert receipt.status == 1

        tx_block_num = w3.eth.block_number

        for i, flt in enumerate(fltrs):
            new_entries = flt.get_new_entries()  # GetFilterChanges
            assert len(new_entries) == tc["filters"][i]["exp_len"]

            if tc["filters"][i]["exp_len"] == 1:
                # check if the new_entries have valid information
                log = new_entries[0]
                assert log["address"] == contract.address
                assert_log_topics(log, [topic])
                assert_log_block(w3, log, tx_block_num)
                assert_change_greet_log_data(log, new_greeting)

        # on next call of GetFilterChanges, no entries should be found
        # because there were no new logs that meet the filters params
        for flt in fltrs:
            assert flt.get_new_entries() == []  # GetFilterChanges


def test_get_logs(cluster):
    w3: Web3 = cluster.w3

    # deploy greeter contract
    contract, _ = deploy_contract(w3, CONTRACTS["Greeter"])
    # test the contract was deployed successfully
    assert contract.caller.greet() == "Hello"

    # calculate topic from event signature
    topic = CHANGE_GREETING_TOPIC

    # another topic not related to the contract deployed
    another_topic = TRANSFER_TOPIC

    # without tx - logs should be empty
    assert w3.eth.get_logs({"address": contract.address}) == []
    assert w3.eth.get_logs({"address": ADDRS["validator"]}) == []

    # with tx
    # update greeting
    new_greeting = "hello, world"
    tx = contract.functions.setGreeting(new_greeting).build_transaction()
    receipt = send_transaction(w3, tx)
    assert receipt.status == 1

    tx_block_num = w3.eth.block_number

    test_cases = [
        {
            "name": "get logs by block range - tx block number is within the range",
            "logs": w3.eth.get_logs({"fromBlock": 1, "toBlock": tx_block_num}),
            "exp_log": True,
            "exp_len": None,  # there are other events not belonging to the contract within the block range specified
        },
        {
            "name": "get logs by block range - tx block number outside the range",
            "logs": w3.eth.get_logs({"fromBlock": 1, "toBlock": 2}),
            "exp_log": False,
            "exp_len": 0,
        },
        {
            "name": "get logs by contract address",
            "logs": w3.eth.get_logs({"address": contract.address}),
            "exp_log": True,
            "exp_len": 1,
        },
        {
            "name": "get logs by topic",
            "logs": w3.eth.get_logs({"topics": [topic.hex()]}),
            "exp_log": True,
            "exp_len": 1,
        },
        {
            "name": "get logs by incorrect topic - should not have logs",
            "logs": w3.eth.get_logs({"topics": [another_topic.hex()]}),
            "exp_log": False,
            "exp_len": 0,
        },
        {
            "name": "get logs by multiple topics ('with all' condition)",
            "logs": w3.eth.get_logs(
                {
                    "topics": [
                        topic.hex(),
                        another_topic.hex(),
                    ]
                }
            ),
            "exp_log": False,
            "exp_len": 0,
        },
        {
            "name": "get logs by multiple topics ('match any' condition)",
            "logs": w3.eth.get_logs({"topics": [[topic.hex(), another_topic.hex()]]}),
            "exp_log": True,
            "exp_len": 1,
        },
    ]

    for tc in test_cases:
        print("Case: {}".format(tc["name"]))

        # logs for validator address should remain empty
        assert len(w3.eth.get_logs({"address": ADDRS["validator"]})) == 0

        logs = tc["logs"]

        if tc["exp_len"] is not None:
            assert len(logs) == tc["exp_len"]

        if tc["exp_log"]:
            found_log = False

            for log in logs:
                if log["address"] == contract.address:
                    # for the current test cases,
                    # this event was emitted only once
                    # so one log from this contract should exist
                    # we check the flag to know it is not repeated
                    assert found_log == False

                    found_log = True

                    assert_log_topics(log, [topic])
                    assert_log_block(w3, log, tx_block_num)
                    assert_change_greet_log_data(log, new_greeting)

            assert found_log == True


def assert_log_topics(log, topics):
    assert len(log["topics"]) == len(topics)
    assert log["topics"] == topics


def assert_log_block(w3, log, tx_block_num):
    block_hash = log["blockHash"]
    # check if the returned block hash is correct
    # getBlockByHash
    block = w3.eth.get_block(block_hash)
    # block should exist
    assert block.hash == block_hash
    assert block.number == tx_block_num

    # check tx hash is correct
    tx_data = w3.eth.get_transaction(log["transactionHash"])
    assert tx_data["blockHash"] == block.hash


def assert_change_greet_log_data(log, new_greeting):
    # check event log data ('from' and 'value' fields)
    types = ["address", "string"]
    names = ["from", "value"]
    values = abi.decode_abi(types, log["data"])
    log_data = dict(zip(names, values))

    # the address stored in the data field may defer on lower/upper case characters
    # then, set all as uppercase for assertion
    assert log_data["from"].upper() == ADDRS["validator"].upper()
    assert log_data["value"] == new_greeting


def make_filter_array(array_len):
    filters = []
    for _ in range(array_len):
        filters.append(
            {
                "params": {"fromBlock": 1, "toBlock": "latest"},
                "exp_len": 1,
            },
        )
    return filters
