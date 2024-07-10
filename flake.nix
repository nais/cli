{
  description = "NAIS CLI";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = {
    self,
    nixpkgs,
  }: let
    version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
    goOverlay = final: prev: {
      go = prev.go.overrideAttrs (old: {
        version = "1.22.5";
        src = prev.fetchurl {
          url = "https://go.dev/dl/go1.22.5.src.tar.gz";
          hash = "sha256-rJxyPyJJaa7mJLw0/TTJ4T8qIS11xxyAfeZEu0bhEvY=";
        };
      });
    };
    each = callback:
      nixpkgs.lib.genAttrs ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"] (
        system: let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [goOverlay];
          };
        in (
          callback system pkgs
        )
      );
  in {
    packages = each (
      system: pkgs: rec {
        nais = pkgs.buildGoModule {
          pname = "nais";
          inherit version;
          src = ./.;
          vendorHash = "sha256-AgRQO3h7Atq4lnieTBohzrwrw0lRcbQi2cvpeol3owM=";
        };
        default = nais;
      }
    );

    devShells = each (system: pkgs: {
      default = pkgs.mkShell {
        buildInputs = with pkgs; [go gopls gotools go-tools];
      };
    });

    defaultPackage = each (system: _: self.packages.${system}.nais);
    formatter = each (_: pkgs: pkgs.nixfmt-rfc-style);
  };
}
