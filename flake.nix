{
  description = "Python dev shell with uv (NixOS-friendly)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go 
            pkgs.cmake
            pkgs.gcc
          ];

          shellHook = ''
            yay go
          '';
        };
      }
    );
}

