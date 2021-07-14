{ system ? builtins.currentSystem, pkgs ? import ./nix { inherit system; } }:
{
  inherit pkgs;
  scripts = (pkgs.callPackage ./nix/scripts.nix
    {
      config = {
        ethermint-config = ./scripts/ethermint-config.yaml;
        geth-genesis = ./scripts/geth-genesis.json;
      };
    }).scripts-env;
}
