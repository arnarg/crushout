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

        checks = {
          e2e-crush = pkgs.runCommand "crushout-e2e-crush" {
          nativeBuildInputs = [ pkgs.jq ];
        } ''
          export HOME=/tmp
          mkdir -p /tmp/crushout-root
          bash ${./tests/e2e/run.sh} ${self.packages.${system}.crushout}/bin/crushout ${./tests/e2e/cases_crush.jsonl}
          touch $out
        '';

          e2e-claude = pkgs.runCommand "crushout-e2e-claude" {
          nativeBuildInputs = [ pkgs.jq ];
        } ''
          export HOME=/tmp
          mkdir -p /tmp/crushout-root
          bash ${./tests/e2e/run.sh} ${self.packages.${system}.crushout}/bin/crushout ${./tests/e2e/cases_claude.jsonl}
          touch $out
        '';
        };

        devShells.default = pkgs.mkShellNoCC {
          packages = [
            gomod2nix.legacyPackages.${system}.gomod2nix
          ];
        };
      }
    ));
}
