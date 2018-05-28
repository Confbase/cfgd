#!/bin/bash

cfg-build-snap_flat() {
    output="$(diff <(cfgsnap build $GOPATH/src/github.com/Confbase/cfgd/cfg-build-snap/test_snapshot/ 2>&1) <(cat $GOPATH/src/github.com/Confbase/cfgd/cfg-build-snap/test_snapshot.snap))"
    status="$?"

    expect_status='0'
    expect=''
}

cfg-build-snap_deep() {
    output="$(diff <(cfgsnap build $GOPATH/src/github.com/Confbase/cfgd/cfg-build-snap/test_deep_snapshot/ 2>&1) <(cat $GOPATH/src/github.com/Confbase/cfgd/cfg-build-snap/test_deep_snapshot.snap))"
    status="$?"

    expect_status='0'
    expect=''
}

cfg-build-snap_invalid() {
    output="$(cfgsnap build 2>&1)"
    status="$?"

    expect_status='1'
    expect='usage: cfg-build-snap [flags] [files]

Flags:
    --no-dirname    omit dirname of targets'
}

tests=(
    "cfg-build-snap_flat"
    "cfg-build-snap_deep"
)
