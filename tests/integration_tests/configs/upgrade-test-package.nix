let
  pkgs = import ../../../nix { };
  fetchEthermint = rev: builtins.fetchTarball "https://github.com/mmsqe/ethermint/archive/${rev}.tar.gz";
  released = pkgs.buildGo118Module rec {
    name = "ethermintd";
    src = fetchEthermint "0e4d41eef5f7983f84206b1bb07e88ed3a9cd44b";
    subPackages = [ "cmd/ethermintd" ];
    vendorSha256 = "sha256-cQAol54b6hNzsA4Q3MP9mTqFWM1MvR5uMPrYpaoj3SY=";
    doCheck = false;
  };
  current = pkgs.callPackage ../../../. { };
in
pkgs.linkFarm "upgrade-test-package" [
  { name = "genesis"; path = released; }
  { name = "integration-test-upgrade"; path = current; }
]
