#! /bin/bash

GIT_ROOT=`git rev-parse --show-toplevel`

function testPackage {
	pushd $1 >/dev/null
	go test
	popd >/dev/null
}

set -e
pushd $GIT_ROOT >/dev/null
PACKAGES="git arena tournament server ."
for PACKAGE_DIR in $PACKAGES
do 
	testPackage $PACKAGE_DIR
done
popd >/dev/null


