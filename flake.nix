{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs @ {
    nixpkgs,
    flake-parts,
    flake-utils,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = flake-utils.lib.defaultSystems;
      perSystem = {pkgs, ...}: {
        packages.default = pkgs.buildGoModule {
          version = "1.0";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-D3sraCsi7VF//q7K4ZWw9JEEbQ3Cs94SYkTf9nQ4NW8=";
        };
      };
    };
}
