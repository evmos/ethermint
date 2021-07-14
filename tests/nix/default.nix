{ sources ? import ./sources.nix, system ? builtins.currentSystem }:

import sources.nixpkgs {
  overlays = [
    (_: pkgs: {
      ethermint = pkgs.buildGoModule rec {
        name = "ethermint";
        src = pkgs.nix-gitignore.gitignoreSource [] ../../.;
        subPackages = [ "./cmd/ethermintd" ];
        vendorSha256 = sha256:0zv5z5nwldb91hwzrsf1f6b2z7icpirmqby2zkbwn8m6kpavzjhq;
        doCheck = false;
      };
      pystarport = pkgs.poetry2nix.mkPoetryApplication {
        projectDir = sources.pystarport;
        src = sources.pystarport;
      };
    })
  ];
  config = { };
  inherit system;
}
