#!/usr/bin/env bash
output() {
	key="$1"
	value="$2"

	if [[ -z "$GITHUB_OUTPUT" ]]; then
		echo "output: $key => $value"
		return
	fi

	if [[ -z "$key" ]]; then
		echo "output: missing key " >&2
	elif [[ -z "$value" ]]; then
		echo "output: missing value for key '$key'" >&2
	else
		{
			echo "$key<<EOF"
			echo "$value"
			echo "EOF"
		} >>"$GITHUB_OUTPUT"
	fi
}
