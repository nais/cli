{
  description = "Kafkarator";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { nixpkgs, ... }:
    let
      goOverlay = final: prev: let
        version = "1.25.6";
        newerGoVersion = prev.go.overrideAttrs (old: {
          inherit version;
          src = prev.fetchurl {
            url = "https://go.dev/dl/go${version}.src.tar.gz";
            hash = "sha256-WMv3ceRNdt5vVtGeM7d9dFoeSJNAkih15GWFuXXCsFk=";
          };
        });
        nixpkgsVersion = prev.go.version;
        newVersionNotInNixpkgs = -1 == builtins.compareVersions nixpkgsVersion version;
      in {
        go = if newVersionNotInNixpkgs then newerGoVersion else prev.go;
        buildGoModule = prev.buildGoModule.override { go = final.go; };
      };
      # helpers
      withSystem = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      withPkgs = f:
        withSystem (system:
          f (import nixpkgs {
            inherit system;
            overlays = [ goOverlay ];
          }));
    in {
      devShells = withPkgs (pkgs: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            gnumake
            go
            golangci-lint-langserver
            gopls
            python3
            python3Packages.python-lsp-server
            black
          ];
        };
      });
      formatter = withPkgs (pkgs: pkgs.nixfmt-rfc-style);
    };
}
