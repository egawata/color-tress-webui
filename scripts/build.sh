#/bin/bash

# tinygo is required to run this script.
# https://tinygo.org/

set -e

files=(
    colortress.html
    generating.gif
    noimage.png
)
dest="./build"

mkdir -p $dest
rm -f ${dest}/*

cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ${dest}/wasm_exec.js
GOOS=js GOARCH=wasm go build -o ${dest}/main.wasm main.go

#cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" ${dest}/wasm_exec.js
#cp ./my_wasm_exec.js ${dest}/wasm_exec.js
#tinygo build -o ${dest}/main.wasm -target wasm ./main.go

for file in "${files[@]}"; do
    cp $file $dest
done
