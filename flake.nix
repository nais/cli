{
  description = "NAIS CLI";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs =
    { self, ... }@inputs:
    inputs.flake-utils.lib.eachSystem
      [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ]
      (
        system:
        let
          version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
          pkgs = import inputs.nixpkgs {
            localSystem = { inherit system; };
            overlays = [
              (
                final: prev:
                let
                  version = "1.23.6";
                  newerGoVersion = prev.go.overrideAttrs (old: {
                    inherit version;
                    src = prev.fetchurl {
                      url = "https://go.dev/dl/go${version}.src.tar.gz";
                      hash = "sha256-A5xbBOZSedrO7opvcecL0Fz1uAF4K293xuGeLtBREiI=";
                    };
                  });
                  nixpkgsVersion = prev.go.version;
                  newVersionNotInNixpkgs = -1 == builtins.compareVersions nixpkgsVersion version;
                in
                {
                  go = if newVersionNotInNixpkgs then newerGoVersion else prev.go;
                  buildGoModule = prev.buildGoModule.override { go = final.go; };
                }
              )
            ];
          };
        in
        {
          packages = rec {
            nais = pkgs.buildGoModule {
              pname = "nais-cli";
              inherit version;
              src = ./.;
              vendorHash = "sha256-PqaPtcH2Mc8YS4aV3ctM7o0+7yAk7IZ4wB3GBmhtIsQ=";
              postInstall = ''
                mv $out/bin/cli $out/bin/nais
              '';
            };
            default = nais;
          };

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              go-tools
              nodejs_20
              nodePackages.prettier
            ];
          };

          formatter = pkgs.nixfmt-rfc-style;
        }
      );
}
