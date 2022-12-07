{ pkgs ? import ../../../nix { } }:
let entangled = (pkgs.callPackage ../../../. { });
in
entangled.overrideAttrs (oldAttrs: {
  patches = oldAttrs.patches or [ ] ++ [
    ./broken-entangled.patch
  ];
})
