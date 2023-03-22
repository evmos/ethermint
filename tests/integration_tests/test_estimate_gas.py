from .utils import ADDRS, CONTRACTS, deploy_contract


def test_revert(ethermint):
    w3 = ethermint.w3
    call = w3.provider.make_request
    validator = ADDRS["validator"]
    erc20, _ = deploy_contract(
        w3,
        CONTRACTS["TestERC20A"],
    )
    method = "eth_estimateGas"
    # revert methods
    for data in ["0x9ffb86a5"]:
        params = {"from": validator, "to": erc20.address, "data": data}
        error = call(method, [params])["error"]
        assert error["code"] == 3
        assert error["message"] == "execution reverted: Function has been reverted"
        assert error["data"] == "0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a46756e6374696f6e20686173206265656e207265766572746564000000000000"  # noqa: E501
