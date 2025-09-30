#!/usr/bin/env bash
#MISE description="Generate release information using git-cliff"
#MISE depends=["build", "completions"]
set -euo pipefail

binary="nais"
if [[ "$GOOS" == "windows" ]]; then
  binary="nais.exe"
  mv "bin/nais" "bin/$binary"

  if [[ -n "$SIGN_CERT" && -n "$SIGN_KEY" ]]; then
    sudo apt-get update
    sudo apt-get install --yes osslsigncode

    echo "$SIGN_CERT" > nais.crt
    echo "$SIGN_KEY" > nais.key

    osslsigncode sign -certs nais.crt -key nais.key -n "nais-cli" -i "https://docs.nais.io/cli" -verbose -in "bin/$binary" -out "bin/nais-signed"
    mv "bin/nais-signed" "bin/$binary"
  fi
fi

tar -zcf "nais-cli_${GOOS}_${GOARCH}.tgz" ./completions -C bin/ "$binary"
