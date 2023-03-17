let 
  pkgs = import ../../../nix { };
  current = pkgs.callPackage ../../../. { };
  patched = current.overrideAttrs (oldAttrs: rec {
    patches = oldAttrs.patches or [ ] ++ [
      ./cache-access-list-ethermintd.patch
    ];
  });
in
pkgs.linkFarm "cache-access-list-ethermintd" [
  { name = "genesis"; path = patched; }
  { name = "integration-test-patch"; path = current; }
]