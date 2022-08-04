import pytest

from .network import setup_ethermint, setup_geth


@pytest.fixture(scope="session")
def ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("ethermint")
    yield from setup_ethermint(path, 26650)


@pytest.fixture(scope="session")
def geth(tmp_path_factory):
    path = tmp_path_factory.mktemp("geth")
    yield from setup_geth(path, 8545)


@pytest.fixture(scope="session", params=["ethermint", "geth", "ethermint-ws"])
def cluster(request, ethermint, geth):
    """
    run on both ethermint and geth
    """
    provider = request.param
    if provider == "ethermint":
        yield ethermint
    elif provider == "geth":
        yield geth
    elif provider == "ethermint-ws":
        ethermint_ws = ethermint.copy()
        ethermint_ws.use_websocket()
        yield ethermint_ws
    else:
        raise NotImplementedError
