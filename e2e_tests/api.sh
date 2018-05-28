#!/bin/bash

api_get_file() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s localhost:1066/mybase/master/config.yml 2>&1`
    status="$?"

    expect_status='0'
    expect='hello!'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_get_file_500() {
    cfgd --custom-backend=./e2e_tests/scripts/always_exit_1.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s localhost:1066/mybase/master/config.yml 2>&1`
    status="$?"

    expect_status='0'
    expect='500 Internal Server Error'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_get_file_404() {
    cfgd --custom-backend=./e2e_tests/scripts/always_no.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s localhost:1066/mybase/master/config.yml 2>&1`
    status="$?"

    expect_status='0'
    expect='404 Content Not Found'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_get_file_400_case_1() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s localhost:1066/mybase/master 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_get_file_400_case_2() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s localhost:1066/mybase 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_get_file_400_case_3() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s localhost:1066 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_put_file() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null > cfgd_test_logs 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output="$(curl -s -X POST localhost:1066/mybase/master --data-binary @cfgd_test_put_file)"
    status="$?"

    expect_status='0'
    expect='201 Content Created'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_put_file_400_case_1() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s -X POST localhost:1066/mybase/master/ 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_put_file_400_case_2() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s -X POST localhost:1066/mybase/master/some_file.yml 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_put_file_400_case_3() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s -X POST localhost:1066/mybase 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_put_file_400_case_4() {
    cfgd --custom-backend=./e2e_tests/scripts/always_hello.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s -X POST localhost:1066 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

api_put_file_400_case_5() {
    cfgd --custom-backend=./e2e_tests/scripts/always_exit_1.sh >/dev/null 2>&1 &
    cfgd_pid="$!"

    sleep 0.1s

    output=`curl -s -X POST localhost:1066/mybase/master 2>&1`
    status="$?"

    expect_status='0'
    expect='400 Bad Request'

    kill "$cfgd_pid"
    wait "$cfgd_pid" >/dev/null 2>&1
}

tests=(
    "api_get_file"
    "api_get_file_500"
    "api_get_file_404"
    "api_get_file_400_case_1"
    "api_get_file_400_case_2"
    "api_get_file_400_case_3"
#    "api_put_file"
    "api_put_file_400_case_1"
    "api_put_file_400_case_2"
    "api_put_file_400_case_3"
    "api_put_file_400_case_4"
    "api_put_file_400_case_5"
)
