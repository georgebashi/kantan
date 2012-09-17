#!/bin/sh

test_description='check kantan runs and responds to requests'

. ./common.sh
. ./sharness.sh

start_kantan

test_expect_success "curl / gets a 404" 'curl http://localhost:9090/ | grep 404'

stop_kantan
test_done

