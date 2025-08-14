{
  pkgs,
  vendorHash,
  src,
  ...
}:
pkgs.buildGo125Module {
  name = "nais-cli";
  inherit vendorHash;
  inherit src;
  env.CGO_ENABLED = 0;
  subPackages = ["."];
  postInstall = ''
    mv $out/bin/cli $out/bin/nais
  '';
}
