#!/usr/bin/env bash
#MISE description="Generate shell completions"
set -euo pipefail

rm -rf completions
mkdir completions
for sh in bash zsh fish powershell; do
	go run main.go completion "$sh" >"completions/nais.$sh"
done
