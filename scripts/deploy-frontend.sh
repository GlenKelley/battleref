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
DEST_DIR=~/Sites/battlecode
if [[ -n "$WEB_BUILD_DEST" ]]; then
  DEST_DIR="$WEB_BUILD_DEST"
fi

echo "Copying web files to ${DEST_DIR}"
mkdir -p ${DEST_DIR}
rm -rf ${DEST_DIR}/*

#Simple no substitution
#cp -r "${GIT_ROOT}/internal/web/." "$DEST/." 

SOURCE_DIR="${GIT_ROOT}/internal/web"
REPLACE_EXT="html js"
find "$SOURCE_DIR" | while read FILE ; do
	DEST_FILE="${FILE/$SOURCE_DIR/$DEST_DIR}"
	if [[ -d "$FILE" ]] ; then
		mkdir -p "$DEST_FILE"
	elif [[ -f "$FILE" ]] ; then
		COPIED=
		for EXT in $REPLACE_EXT ; do
			if [[ "$FILE" == *\.$EXT ]] ; then
				sed "s/localhost:8080/api.akusete.com:8080/g" "$FILE" > "$DEST_FILE"
				#chmod --reference "$FILE" "$DEST_FILE"
				COPIED="TRUE"
				break
			fi
		done
		if [[ -z "$COPIED" ]] ; then
			cp "$FILE" "$DEST_FILE"
		fi
	else
		exit 1
	fi
done

echo "Success."
