#! /bin/bash

#
# Remove user account
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0 -u username"
  echo ""
  echo "Removes a user account."
  echo ""
  echo "        -u username		The name of the user account." 
  echo "        -v                      Enable verbose debugging."
  echo "" 
  exit 1
}

while getopts "h:u:v" opt; do
  case $opt in
    u ) NEW_USER="$OPTARG" ;;
    v ) VERBOSE=TRUE ;;
    ? ) usage
  esac
done

set -e

if [[ -z "$NEW_USER" ]] ; then
  echo "Error: You must define a user."
  usage
fi

if [[ -n "$VERBOSE" ]] ; then
	set -x
fi

if ! id -u "$NEW_USER" > /dev/null 2>&1 ; then
  echo "$NEW_USER is not an account. Exiting"
  exit 1
fi

if [[ "$OSTYPE" == "linux-gnu" ]]; then # Linux
	sudo userdel -r "$USER_HOME"
elif [[ "$OSTYPE" == "darwin"* ]]; then # Mac OSX
	USER_HOME="/Users/$NEW_USER"
	sudo dscl . -delete "$USER_HOME"
	sudo rm -rf "$USER_HOME"
fi


