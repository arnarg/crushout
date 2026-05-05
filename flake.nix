{
  description = "crushout - crush hook that auto-approves read-only bash commands";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        callPackage = pkgs.callPackage;
      in
      {
        packages = {
          default = self.packages.${system}.crushout;
          crushout = callPackage ./. {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };
        };

        devShells.default = pkgs.mkShellNoCC {
          packages = [
            gomod2nix.legacyPackages.${system}.gomod2nix
          ];
        };
      }
    ));
}
