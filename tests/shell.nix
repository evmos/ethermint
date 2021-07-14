with import ./. { };
pkgs.mkShell {
  buildInputs = [
    pkgs.ethermint
    pkgs.go-ethereum
    scripts
  ];
}
