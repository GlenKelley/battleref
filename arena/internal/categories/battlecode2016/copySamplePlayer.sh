#! /bin/bash

#
# Copy the sample player template to a new directory. 
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0"
  echo ""
  echo "	-d repo_dir 
  echo "	-n name 
  echo "" 
  exit 1
}

function error {
	echo "$@" >&2
}

while getopts "?d:n:" opt; do
  case $opt in
    d) REPO_DIR="$OPTARG" ;;
    n) NAME="$OPTARG" ;;
    ? ) usage
  esac
done
set -ex

if [[ -z "$REPO_DIR" ]] ; then
  error "Error: You must define the player repository."
  usage
elif [[ ! -d "$REPO_DIR" ]] ; then
  error "Error: $REPO_DIR is not a file."
  usage
fi

if [[ -z "$NAME" ]] ; then
  error "Error: You must define the player name."
  usage
fi

SOURCE_DIR="$(dirname $0)/sampleplayer"

find "$SOURCE_DIR" | while read FILE ; do
	DEST="${FILE/$SOURCE_DIR/$REPO_DIR}"
	DEST="${DEST/PLAYER_NAMESPACE/$NAME}"
	if [[ -d "$FILE" ]] ; then
		mkdir -p "$DEST"
	elif [[ -f "$FILE" ]] ; then
		sed "s/PLAYER_NAMESPACE/$NAME/" "$FILE" > "$DEST"
	else
		exit 1
	fi
done


