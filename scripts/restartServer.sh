#! /bin/bash

#
# Starts a battleref server
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0 -e environment"
  echo ""
  echo "Restarts the battleref webserver application"
  echo ""
  echo "        -e environment       The environment to use when running the webserver"
  echo "" 
  exit 1
}

while getopts ":e:?" opt; do
  case $opt in
    e ) ENV="$OPTARG" ;;
    ? ) usage
  esac
done
set -e

if [[ -z "$ENV" ]] ; then
  echo "Error: You must define an environment."
  usage
fi

if ps -ef | grep "bash $0" > /dev/null ; then
	echo "Battleref server already running. Exiting Script."
	exit 0
if

DIR=`dirname $0`
pushdir $DIR > /dev/null
GIT_ROOT=`git rev-parse --show-toplevel`
popdir > /dev/null

BATTLEREF_DIR=~/.battleref

mkdir -p $BATTLEREF_DIR
cd $BATTLEREF_DIR
LOCK_FILE=$BATTLEREF_DIR/.shutdown
ENV_FILE="$GIT_ROOT/server.$ENV.properties"
LOG=$BATTLEREF_DIR/restart.log

while true ; do
	if [[ -f "$LOCK_FILE" ]] ; then
		echo `date` "Server was shutdown. Exiting." | tee -a $LOG
		cat $LOCKFILE | tee -a $LOG
		exit 0
	fi
	
	if [[ ! -f "$ENV_FILE" ]] ; then
		echo `date` "$ENV_FILE is not a file." | tee -a $LOG
		usage
	fi

	set +e	
	$BATTLEREF_DIR/runServer -e $ENV_FILE > $BATTLEREF_DIR/server.log 2> $BATTLEREF_DIR/error.log
	EXIT_STATUS=$?
	set -e
	echo `date` "Battleref server quit with exit status $EXIT_STATUS" | tee -a $LOG
done
