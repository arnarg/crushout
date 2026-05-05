{
  lib,
  buildGoApplication,
}:
let
  version = "0.3.0";
in
buildGoApplication {
  inherit version;
  pname = "crushout";

  src = builtins.filterSource (
    path: type:
    type == "directory"
    || baseNameOf path == "go.mod"
    || baseNameOf path == "go.sum"
    || lib.hasSuffix ".go" path
  ) ./.;

  modules = ./gomod2nix.toml;

  subPackages = [ "cmd/crushout" ];

  CGO_ENABLED = "0";

  meta = {
    mainProgram = "crushout";
  };
}
