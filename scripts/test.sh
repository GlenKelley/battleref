#! /bin/bash

GIT_ROOT=`git rev-parse --show-toplevel`

function testPackage {
	pushd $1 >/dev/null
	go test -v
	popd >/dev/null
}

set -e
pushd $GIT_ROOT >/dev/null
PACKAGES="simulator web git arena tournament server ."
for PACKAGE_DIR in $PACKAGES
do 
	testPackage $PACKAGE_DIR
done
popd >/dev/null


