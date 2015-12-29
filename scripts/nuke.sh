#! /bin/bash

#
# Destroys the battleref backend to a server
# 
# Author: Glen Kelley
#

function failure {
  echo -e "\033[1;31m$@\033[0;30m" 
}
function header {
  echo -e "\033[1;32m$@\033[0;30m" 
}

function usage {
  if [[ -n "$@" ]]; then
    failure "$@"
  fi
  echo ""
  echo "$0 [-h host] [-v]"
  echo ""
  echo "Uninstalls battleref and users from a remote host. WARNING THIS IS DESTRUCTIVE."
  echo ""
  echo "        -v                   Verbose output"
  echo "        -h                   Prints this message"
  echo "" 
  exit 1
}

HOST="ec2-user@api.akusete.com"
SHUTDOWN_PORT=8080
while getopts "h:p:v?" opt; do
  case $opt in
    h ) HOST="$OPTARG" ;;
    p ) SHUTDOWN_PORT="$OPTARG" ;;
    v ) VERBOSE=TRUE ;;
    ? ) usage
  esac
done

set -e

if [[ -z "$HOST" ]] ; then
  usage "Error: You must define a host."
fi

if [[ -n "$VERBOSE" ]] ; then
	set -x
fi

echo -e "\033[1;31m\033[5;31mWARNING\033[0;30m\033[1;30m: This will completely remove battleref from $HOST, and delete the git and webserver directories.\033[0;30m"
read -p "Are you sure you want to continue[y/N]: " VERIFY
if [[ "$VERIFY" != "y" ]]; then
  failure "Aborting."
  exit 1
fi

header "Connecting to $HOST"

GIT_USER=git
WEBSERVER_USER=webserver
SUBMISSION_DIR=/opt/battleref

ssh -T $HOST <<EOF
set -e
if [[ -n "$VERBOSE" ]] ; then
	set -x
fi

function failure {
  echo -e "\033[1;31m\$@\033[0;30m" 
}
function header {
  echo -e "\033[1;32m\$@\033[0;30m" 
}

header "Checking if webserver user exitsts."
if id -u "$WEBSERVER_USER" > /dev/null 2>&1 ; then
  echo "Account $WEBSERVER_USER exists."
  WEBSERVER_HOME=\`getent passwd $WEBSERVER_USER | cut -d: -f6\`
  if sudo -u "$WEBSERVER_USER" test -d "\$WEBSERVER_HOME/.battleref" ; then
    echo "Battleref installation exists."
    echo "Shutting down server."
    echo "Shutdown by $0 to install" | sudo -u "$WEBSERVER_USER" tee \$WEBSERVER_HOME/.battleref/.shutdown
    if curl localhost:$SHUTDOWN_PORT/api > /dev/null 2>&1 ; then
      echo "Server is running."
      curl -X POST localhost:$SHUTDOWN_PORT/shutdown > /dev/null 2>&1 | true
      echo "Waiting for server to shutdown."
      ATTEMPTS=0
      IS_SHUTDOWN=
      while [[ -z "\$IS_SHUTDOWN" ]] && [[ "\$ATTEMPTS" -le 10 ]] && sleep 1; do
        if ! curl localhost:$SHUTDOWN_PORT/api > /dev/null 2>&1 ; then
	  echo "SHUTDOWN"
          IS_SHUTDOWN=TRUE
	fi
	ATTEMPTS=\$((ATTEMPTS + 1))
     done
     if [[ -n "\$IS_SHUTDOWN" ]] ; then
       echo "Server has been shutdown."
     else
      	failure "Failed to shut server down."
        exit 1
      fi
    else
      echo "Server is not running."
    fi
  fi
fi

rm -f ~/.ssh/webserver*
if id -u webserver > /dev/null 2>&1; then
  sudo pkill -u webserver | true
  sudo userdel webserver
fi
sudo rm -r /home/webserver

rm -f ~/.ssh/git*
if id -u git > /dev/null 2>&1; then
  sudo pkill -u git | true
  sudo userdel git
fi
sudo rm -r /home/git
sudo rm -rf /opt/battleref

header "Success."
