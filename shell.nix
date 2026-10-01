{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
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
}
