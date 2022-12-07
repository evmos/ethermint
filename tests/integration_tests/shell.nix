{ system ? builtins.currentSystem, pkgs ? import ../../nix { inherit system; } }:
pkgs.mkShell {
  buildInputs = [
    pkgs.jq
    (pkgs.callPackage ../../. { }) # entangled
    pkgs.start-scripts
    pkgs.go-ethereum
    pkgs.cosmovisor
    pkgs.nodejs
    pkgs.test-env
  ];
  shellHook = ''
    . ${../../scripts/.env}
  '';
}
