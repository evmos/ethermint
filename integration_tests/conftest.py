import pytest

from .network import setup_ethermint, setup_geth


def pytest_configure(config):
    config.addinivalue_line("markers", "slow: marks tests as slow")


@pytest.fixture(scope="session")
def suspend_capture(pytestconfig):
    """
    used to pause in testing

    Example:
    ```
    def test_simple(suspend_capture):
        with suspend_capture:
            # read user input
            print(input())
    ```
    """

    class SuspendGuard:
        def __init__(self):
            self.capmanager = pytestconfig.pluginmanager.getplugin("capturemanager")

        def __enter__(self):
            self.capmanager.suspend_global_capture(in_=True)

        def __exit__(self, _1, _2, _3):
            self.capmanager.resume_global_capture()

    yield SuspendGuard()


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
    elif provider == "cronos-ws":
        ethemrint_ws = ethermint.copy()
        ethermint_ws.use_websocket()
        yield ethermint_ws
    else:
        raise NotImplementedError
