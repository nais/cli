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
    withSystem = nixpkgs.lib.genAttrs ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];
    withPkgs = callback:
      withSystem (
        system:
          callback
          (import nixpkgs {
            inherit system;
            overlays = [goOverlay];
          })
      );
  in {
    packages = withPkgs (
      pkgs: rec {
        nais = pkgs.buildGoModule {
          pname = "nais";
          inherit version;
          src = ./.;
          vendorHash = "sha256-1LuU26iUzWJYKjRR8Gz1ZeSQDHOzW1tQphBFTCiQO+0=";
          postInstall = ''
            mv $out/bin/cli $out/bin/nais
          '';
        };
        default = nais;
      }
    );

    devShells = withPkgs (pkgs: {
      default = pkgs.mkShell {
        buildInputs = with pkgs; [go gopls gotools go-tools];
      };
    });

    formatter = withPkgs (pkgs: pkgs.nixfmt-rfc-style);
  };
}
