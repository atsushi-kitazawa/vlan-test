{
  description = "VLAN Sample Go Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
          ];

          shellHook = ''
            # ホスト側のGOROOTが残っているとバージョンの不整合エラーになるため解除
            unset GOROOT
            echo "========================================="
            echo "  VLAN Sample Go Development Environment"
            echo "  Go binary: $(which go)"
            echo "  Go version: $(go version)"
            echo "========================================="
          '';
        };
      }
    );
}
