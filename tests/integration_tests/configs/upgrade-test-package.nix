let
  pkgs = import ../../../nix { };
  fetchEthermint = rev: builtins.fetchTarball "https://github.com/evmos/ethermint/archive/${rev}.tar.gz";
  released = pkgs.buildGo118Module rec {
    name = "ethermintd";
    # the commit before https://github.com/evmos/ethermint/pull/943
    src = fetchEthermint "f21592ebfe74da7590eb42ed926dae970b2a9a3f";
    subPackages = [ "cmd/ethermintd" ];
    vendorSha256 = "sha256-ABm5t6R/u2S6pThGrgdsqe8n3fH5tIWw7a57kxJPbYw=";
    doCheck = false;
  };
  current = pkgs.callPackage ../../../. { };
in
pkgs.linkFarm "upgrade-test-package" [
  { name = "genesis"; path = released; }
  { name = "integration-test-upgrade"; path = current; }
]
