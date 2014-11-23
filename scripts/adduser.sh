#! /bin/bash

#
# Creates a new user account
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0 -u username"
  echo ""
  echo "Creates a new user account."
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

if id -u "$NEW_USER" > /dev/null 2>&1 ; then
  echo "$NEW_USER already exists. Exiting"
  exit 1
fi

if [[ "$OSTYPE" == "linux-gnu" ]]; then # Linux
	sudo adduser "$NEW_USER"
elif [[ "$OSTYPE" == "darwin"* ]]; then # Mac OSX
	MAXID=$(dscl . -list /Users UniqueID | awk '{print $2}' | sort -ug | tail -1)
	USERID=$((MAXID+1))
	if [[ -z "$USERID" ]] ; then
		echo "Error: Empty User ID"
		exit 1
	fi

	MAXID=$(dscl . -list /Users PrimaryGroupID | awk '{print $2}' | sort -ug | tail -1)
	GROUPID=$((MAXID+1))
	if [[ -z "$GROUPID" ]] ; then
		echo "Error: Empty Group ID"
		exit 1
	fi

	USER_HOME="/Users/$NEW_USER"
	sudo dscl . create "$USER_HOME"
	sudo dscl . create "$USER_HOME" UserShell /bin/bash
	sudo dscl . create "$USER_HOME" RealName "$NEW_USER"
	sudo dscl . create "$USER_HOME" UniqueID "$USERID"
	sudo dscl . create "$USER_HOME" PrimaryGroupID "$GROUPID"
	sudo dscl . create "$USER_HOME" NFSHomeDirectory "$USER_HOME"
	sudo mkdir "$USER_HOME"
	sudo chown "${NEW_USER}" "$USER_HOME"
fi


