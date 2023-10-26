#!/bin/bash

files=(address_gen.go identifier_gen.go index_gen.go slot_identifier_gen.go unlock_ref_gen.go)

for file in "${files[@]}"
do
  echo "Generating $file..."
  go generate "$file"
done