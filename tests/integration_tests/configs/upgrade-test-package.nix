let
  pkgs = import ../../../nix { };
  fetchEthermint = rev: builtins.fetchTarball "https://github.com/evmos/ethermint/archive/${rev}.tar.gz";
  released = pkgs.buildGo118Module rec {
    name = "ethermintd";
    # the commit before https://github.com/evmos/ethermint/pull/943
    src = fetchEthermint "8866ae0ffd67a104e9d1cf4e50fba8391dda6c45";
    subPackages = [ "cmd/ethermintd" ];
    vendorSha256 = "sha256-oDtMamNlwe/393fZd+RNtRy6ipWpusbco8Xg1ZuKWYw=";
    doCheck = false;
  };
  current = pkgs.callPackage ../../../. { };
in
pkgs.linkFarm "upgrade-test-package" [
  { name = "genesis"; path = released; }
  { name = "integration-test-upgrade"; path = current; }
]
