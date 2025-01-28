#/bin/bash

# tinygo is required to run this script.
# https://tinygo.org/

set -e

cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" ./wasm_exec.js
tinygo build -o main.wasm -target wasm ./main.go
