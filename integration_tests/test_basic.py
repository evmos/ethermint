import json
import time
from pathlib import Path

import pytest
import web3
from eth_bloom import BloomFilter
from eth_utils import abi, big_endian_to_int
from hexbytes import HexBytes
from pystarport import cluster, ports

def test_basic(cluster):
    w3 = cluster.w3
    assert w3.eth.chain_id == 9000