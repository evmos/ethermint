import pytest
from pathlib import Path
from .network import setup_ethermint, setup_geth, setup_custom_ethermint

@pytest.fixture(scope="session")
def ethermint(tmp_path_factory):
    path = tmp_path_factory.mktemp("ethermint")
    yield from setup_ethermint(path, 26650)


@pytest.fixture(scope="session")
def geth(tmp_path_factory):
    path = tmp_path_factory.mktemp("geth")
    yield from setup_geth(path, 8545)

@pytest.fixture(scope="session")
def pruned(request, tmp_path_factory):
    """start-ethermint
    params: enable_auto_deployment
    """
    yield from setup_custom_ethermint(
        tmp_path_factory.mktemp("pruned"),
        26900,
        Path(__file__).parent / "configs/pruned_node.jsonnet",
    )

@pytest.fixture(scope="session", params=["ethermint", "geth", "pruned", "ethermint-ws"])
def cluster(request, ethermint, geth, pruned):
    """
    run on both ethermint and geth
    """
    provider = request.param
    if provider == "ethermint":
        yield ethermint
    elif provider == "geth":
        yield geth
    elif provider == "pruned":
        yield pruned
    elif provider == "ethermint-ws":
        ethermint_ws = ethermint.copy()
        ethermint_ws.use_websocket()
        yield ethermint_ws
    else:
        raise NotImplementedError
