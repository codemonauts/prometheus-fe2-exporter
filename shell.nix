{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    musl
    golangci-lint
    gopls
    goreleaser
    gotools
  ];
  shellHook = ''
  '';
}
