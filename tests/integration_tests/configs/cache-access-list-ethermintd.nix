{ pkgs ? import ../../../nix { } }:
let ethermintd = (pkgs.callPackage ../../../. { });
in
ethermintd.overrideAttrs (oldAttrs: {
  patches = oldAttrs.patches or [ ] ++ [
    ./cache-access-list-ethermintd.patch
  ];
})
