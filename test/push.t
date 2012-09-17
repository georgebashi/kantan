#!/bin/sh

test_description='create a basic repo and push it'

. ./common.sh
. ./sharness.sh

start_kantan

test_expect_success 'set up repo' 'mkdir basic-repo &&
	cd basic-repo &&
	git init . &&
	echo "buildpack: git://github.com/georgebashi/buildpack-test.git" > .kantan.yml &&
	mkdir bin &&
	echo "#!/usr/bin/env bash" > bin/test &&
	echo "echo \"hello world\"" >> bin/test &&
	chmod +x bin/test &&
	git add .kantan.yml &&
	git add bin/test &&
	git commit -m "basic repo test"'

test_expect_success 'push repo' 'git push http://localhost:9090/projects/basic-repo/repo master 2>&1 | grep "hello world"'

stop_kantan
test_done
