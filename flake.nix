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
          name = "janitor-scraper";
          version = "1.1.1";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-4VfMBmyE2vrhWxZoCRk2avM+iZXKJKsM/J/b7rOO0qo=";
        };
      };
    };
}
