#!/bin/bash

cfgsnap_build_flat() {
    output="$(diff <($GOBIN/cfgsnap build $GOPATH/src/github.com/Confbase/cfgd/test_data/cfgsnap/test_snapshot/ 2>&1) <(cat $GOPATH/src/github.com/Confbase/cfgd/test_data/cfgsnap/test_snapshot.snap))"
    status="$?"

    expect_status='0'
    expect=''
}

cfgsnap_build_deep() {
    output="$(diff <($GOBIN/cfgsnap build $GOPATH/src/github.com/Confbase/cfgd/test_data/cfgsnap/test_deep_snapshot/ 2>&1) <(cat $GOPATH/src/github.com/Confbase/cfgd/test_data/cfgsnap/test_deep_snapshot.snap))"
    status="$?"

    expect_status='0'
    expect=''
}

cfgsnap_build_invalid() {
    output="$($GOBIN/cfgsnap build 2>&1)"
    status="$?"

    expect_status='255'
    expect='Error: requires at least 1 arg(s), only received 0
Usage:
  cfgsnap build [flags]

Flags:
  -h, --help         help for build
  -n, --no-dirname   omit dirnames in snap keys (default true)

requires at least 1 arg(s), only received 0'
}

tests=(
    "cfgsnap_build_flat"
    "cfgsnap_build_deep"
    "cfgsnap_build_invalid"
)
