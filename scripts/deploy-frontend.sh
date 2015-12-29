#! /bin/bash

#
# Deploys the battleref frontend to a server
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0 -r repo-url -h host"
  echo ""
  echo "Deploys the battleref front-end websever."
  echo ""
  echo "" 
  exit 1
}

SHUTDOWN_PORT=8080
while getopts "h:r:e:p:v?" opt; do
  case $opt in
    h ) HOST="$OPTARG" ;;
    r ) REPO="$OPTARG" ;;
    e ) ENV="$OPTARG" ;;
    p ) SHUTDOWN_PORT="$OPTARG" ;;
    v ) VERBOSE=TRUE ;;
    ? ) usage
  esac
done

set -e

if [[ -n "$VERBOSE" ]] ; then
	set -x
fi

GIT_ROOT="$(git rev-parse --show-toplevel)"
DEST=~/Sites

rm -rf ${DEST}/*
cp -r "${GIT_ROOT}/internal/web/." "$DEST/." 

echo "Success."
