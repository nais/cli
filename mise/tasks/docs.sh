#!/usr/bin/env bash
#MISE description="Generate and serve the docs locally using Jekyll"
#MISE depends=["generate:docs"]
set -euo pipefail

if ! command -v bundle &>/dev/null; then
  echo "Error: 'bundle' not found. Install it with 'gem install bundler'." >&2
  exit 1
fi

cd docs
bundle install
bundle exec jekyll serve