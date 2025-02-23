#/bin/bash

# tinygo is required to run this script.
# https://tinygo.org/

set -e

files=(
    main.wasm
    wasm_exec.js
    colortress.html
    generating.gif
    noimage.png
)
dest="./build"

cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" ./wasm_exec.js
tinygo build -o main.wasm -target wasm ./main.go

mkdir -p $dest
rm -f $dest/*

for file in "${files[@]}"; do
    cp $file $dest
done
