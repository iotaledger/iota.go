#!/bin/bash

COMMIT=$1
if [ -z "$COMMIT" ]
then
    echo "ERROR: no commit hash given!"
    exit 1
fi

HIVE_MODULES=$(grep -E "^\sgithub.com/iotaledger/hive.go" "go.mod" | awk '{print $1}')
for dependency in $HIVE_MODULES
do
    echo "go get -u $dependency@$COMMIT..."
    go get -u "$dependency@$COMMIT" >/dev/null
done

# Run go mod tidy
echo "Running go mod tidy..."
go mod tidy >/dev/null
