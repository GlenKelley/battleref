#! /bin/bash

#
# Starts a battleref server
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0"
  echo ""
  echo "	-r battlecode tar	The tar file of the battlecode application"
  echo "	-p player1		The name of the first player"
  echo "	-P player2		The name of the second player"
  echo "	-c commit1		The git commit hash of the first player repo"
  echo "	-C commit2		The git commit hash of the second player repo"
  echo "	-g git_url		The server connection string"
  echo "	-m map			The name of the map"
  echo "	-M map_file		The file of the map content"
  echo "        -R capture_replay       Returns the replay of the match"
  echo "" 
  exit 1
}

while getopts "?r:p:P:c:C:g:u:m:M:R" opt; do
  case $opt in
    r) BATTLECODE_TAR="$OPTARG" ;;
    p) PLAYER1="$OPTARG" ;;
    P) PLAYER2="$OPTARG" ;;
    c) COMMIT1="$OPTARG" ;;
    C) COMMIT2="$OPTARG" ;;
    g) GIT_URL="$OPTARG" ;;
    m) MAP="$OPTARG" ;;
    M) MAP_FILE="$OPTARG" ;;
    R) CAPTURE_REPLAY=TRUE ;;
    ? ) usage
  esac
done
set -e

if [[ -z "$BATTLECODE_TAR" ]] ; then
  echo "Error: You must define the battlecode root directory."
  usage
elif [[ ! -f "$BATTLECODE_TAR" ]] ; then
  echo "Error: $BATTLECODE_TAR is not a file."
  usage
fi

if [[ -z "$PLAYER1" ]] ; then
  echo "Error: You must define player1."
  usage
fi
if [[ -z "$PLAYER2" ]] ; then
  echo "Error: You must define player2."
  usage
fi
if [[ -z "$COMMIT1" ]] ; then
  echo "Error: You must define commit1."
  usage
fi
if [[ -z "$COMMIT2" ]] ; then
  echo "Error: You must define commit2."
  usage
fi
if [[ -z "$GIT_URL" ]] ; then
  echo "error: you must define a git url."
  usage
fi
if [[ -z "$MAP" ]] ; then
  echo "Error: You must define a map."
  usage
fi
if [[ -z "$MAP_FILE" ]] ; then
  echo "Error: You must define a map file."
  usage
elif [[ ! -f "$MAP_FILE" ]] ; then
  echo "Error: $MAP_FILE is not a file."
  usage
fi

DIR=`mktemp -d -t battlecode`

tar -xf "$BATTLECODE_TAR" -C "$DIR/"

#Create teams directory
mkdir -p "$DIR/teams/"

#Create maps directory
mkdir -p "$DIR/maps/"
cp "$MAP_FILE" "$DIR/maps/"

function fixPackage {
	find . -type f -name "*.java" -maxdepth 1 | while read FILE
	do 
		perl -i -p -e "s/package \w+;/package $1;/" $FILE
	done

}

#Create team1
git clone "${GIT_URL}${PLAYER1}.git" "$DIR/teams/player1"
pushd "$DIR/teams/$PLAYER1" >/dev/null
git checkout "player1"
fixPackage $PLAYER1
popd >/dev/null

#Create team2
git clone "${GIT_URL}${PLAYER2}.git" "$DIR/teams/player2"
pushd "$DIR/teams/$PLAYER2" >/dev/null
git checkout "$COMMIT2"
fixPackage player2
popd >/dev/null

echo "Match [$PLAYER1:$COMMIT1] vs [$PLAYER2:$COMMIT2] on $MAP" >&2

pushd "$DIR" >/dev/null
cp bc.conf.template bc.conf
echo "# Headless settings" >> bc.conf
echo "bc.game.maps=${MAP}.xml" >> bc.conf
echo "bc.game.team-a=player1" >> bc.conf
echo "bc.game.team-b=player2" >> bc.conf
echo "bc.game.save-file=match.rms" >> bc.conf

ant file -Dc=bc.conf > output.log 2> error.log
#cat error.log 2>&1

WINNER=`grep "\[java\] \[server\]" output.log | perl -i -n -e 'if(/player\d \((\w)\) wins/){print "$1"}'`
REASON=`grep "Reason:" output.log | perl -i -n -e '
	if(/Reason: ([^\n]+)/){
		if ($1 eq "The winning team won by getting a lot of milk.") { print "VICTORY" }
		if ($1 eq "The winning team won on tiebreakers.") { print "TIE" }
	}'`
REPLAY=
if [[ -n "$CAPTURE_REPLAY" ]] ; then
	REPLAY=`cat match.rms | base64`
fi

echo "{\"winner\":\"$WINNER\",\"reason\":\"$REASON\",\"replay\":\"$REPLAY\"}"

popd > /dev/null

#Cleanup
rm -rf $DIR

