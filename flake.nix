{
  description = "Nix flake for nais-cli ";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
  };
  outputs = {
    nixpkgs,
    self,
    ...
  }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs {
      inherit system;
    };
  in {
    packages.${system} = rec {
      bin = pkgs.callPackage ./nix/package.nix {
        inherit pkgs;
        src = self;
        vendorHash = "sha256-mCEDP/koAqlZOetCRLUWQVy0HuPCo2KDLRL6kPP1tZ8=";
      };
      docker = pkgs.callPackage ./nix/docker.nix {pkg = bin;};
      default = docker;
    };
  };
}
