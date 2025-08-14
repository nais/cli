{
  pkgs,
  pkg,
  ...
}: let
in
  pkgs.dockerTools.buildImage {
    name = "nais-cli-docker";
    config = {
      Cmd = ["${pkg}/bin/nais"];
    };
  }
