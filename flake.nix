{
  description = "NAIS CLI";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs, }:
    let
      version = builtins.substring 0 8
        (self.lastModifiedDate or self.lastModified or "19700101");
      withSystem = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      withPkgs = callback:
        withSystem (system:
          callback (import nixpkgs {
            inherit system;
            overlays = [
              (final: prev: {
                go = (prev.go.overrideAttrs {
                  version = "1.23.6";
                  src = prev.fetchurl {
                    url = "https://go.dev/dl/go1.23.6.src.tar.gz";
                    hash =
                      "sha256-A5xbBOZSedrO7opvcecL0Fz1uAF4K293xuGeLtBREiI=";
                  };
                });
              })
            ];
          }));
    in {
      packages = withPkgs (pkgs: rec {
        nais = pkgs.buildGoModule.override { go = pkgs.go; } {
          pname = "nais-cli";
          inherit version;
          src = ./.;
          vendorHash = "sha256-kr9VARylPRphfrBEf8KgY1mCV+a1lwQpT/lpur1T3tQ=";

          postInstall = ''
            mv $out/bin/cli $out/bin/nais
          '';
        };
        default = nais;
      });

      devShells = withPkgs (pkgs: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            nodejs_20
            nodePackages.prettier
          ];
        };
      });

      formatter = withPkgs (pkgs: pkgs.nixfmt-rfc-style);
    };
}
