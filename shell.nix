{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.gopls
    pkgs.docker-compose
    pkgs.sqlc
    pkgs.go-migrate
    pkgs.go-task
    pkgs.pkg-config
    pkgs.zeromq
  ];
}