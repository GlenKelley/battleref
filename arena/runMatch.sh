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
  echo "	-p player1		The git repo of the first player"
  echo "	-P player2		The git repo of the second player"
  echo "	-c commit1		The git commit hash of the first player repo"
  echo "	-C commit2		The git commit hash of the second player repo"
  echo "	-m map			The name of the map"
  echo "	-M map_file		The file of the map content"
  echo "        -R capture_replay       Returns the replay of the match"
  echo "" 
  exit 1
}

function error {
	echo "$@" >&2
}

while getopts "?r:p:P:c:C:u:m:M:R" opt; do
  case $opt in
    r) BATTLECODE_TAR="$OPTARG" ;;
    p) PLAYER1="$OPTARG" ;;
    P) PLAYER2="$OPTARG" ;;
    c) COMMIT1="$OPTARG" ;;
    C) COMMIT2="$OPTARG" ;;
    m) MAP="$OPTARG" ;;
    M) MAP_FILE="$OPTARG" ;;
    R) CAPTURE_REPLAY=TRUE ;;
    ? ) usage
  esac
done
set -ex

if [[ -z "$BATTLECODE_TAR" ]] ; then
  error "Error: You must define the battlecode root directory."
  usage
elif [[ ! -f "$BATTLECODE_TAR" ]] ; then
  error "Error: $BATTLECODE_TAR is not a file."
  usage
fi

if [[ -z "$PLAYER1" ]] ; then
  error "Error: You must define player1."
  usage
fi
if [[ -z "$PLAYER2" ]] ; then
  error "Error: You must define player2."
  usage
fi
if [[ -z "$COMMIT1" ]] ; then
  error "Error: You must define commit1."
  usage
fi
if [[ -z "$COMMIT2" ]] ; then
  error "Error: You must define commit2."
  usage
fi
if [[ -z "$MAP" ]] ; then
  error "Error: You must define a map."
  usage
fi
if [[ -z "$MAP_FILE" ]] ; then
  error "Error: You must define a map file."
  usage
elif [[ ! -f "$MAP_FILE" ]] ; then
  error "Error: $MAP_FILE is not a file."
  usage
fi

BATTLECODE_DIR=$(mktemp -d -t battlecode)
function cleanup {
	rm -rf "$BATTLECODE_DIR"
}
trap cleanup EXIT

LOG=~/.battleref/arena.log
echo "`date` START" >> $LOG

tar -xf "$BATTLECODE_TAR" -C "$BATTLECODE_DIR/"
#Create teams directory
mkdir -p "$BATTLECODE_DIR/teams/"

#Create maps directory
mkdir -p "$BATTLECODE_DIR/maps/"
cp "$MAP_FILE" "$BATTLECODE_DIR/maps/${MAP}.xml"

function createRepo {
	REPO_URL=$1
	REPO_NAME=$2
	COMMIT=$3
	REPO_PATH=$BATTLECODE_DIR/teams/$REPO_NAME
	git clone "$REPO_URL" "$REPO_PATH"
	pushd "$REPO_PATH" >/dev/null
	if ! git checkout "$COMMIT" 2>error.log ; then 
		cat error.log >&2
	fi
	popd >/dev/null
}

#Create teams
NAME1=`basename "${PLAYER1%.git}" | sed 's/^.*://'`
NAME2=`basename "${PLAYER2%.git}" | sed 's/^.*://'`
#NAME1=`basename "${PLAYER1%.git}"`
#NAME2=`basename "${PLAYER2%.git}"`
createRepo "$PLAYER1" "$NAME1" "$COMMIT1"

if [[ "$PLAYER1" = "$PLAYER2" ]] ; then
  if [[ "$COMMIT1" != "$COMMIT2" ]] ; then
    error "Error: Unable to play different commits of the same player."
    exit 1
  fi
else
  createRepo "$PLAYER2" "$NAME2" "$COMMIT2"
fi

error "Match [$NAME1:$COMMIT1] vs [$NAME2:$COMMIT2] on $MAP" >&2

pushd "$BATTLECODE_DIR" >/dev/null
cp bc.conf.template bc.conf
echo "# Headless settings" >> bc.conf
echo "bc.game.maps=${MAP}.xml" >> bc.conf
echo "bc.game.team-a=${NAME1}" >> bc.conf
echo "bc.game.team-b=${NAME2}" >> bc.conf
echo "bc.game.save-file=match.rms" >> bc.conf

echo "`date` RUN MATCH" >> $LOG
if ! ant file -Dc=bc.conf 2> error.log | tee -a $LOG > output.log ; then
	cat output.log error.log >&2
	exit 1
fi

if grep "~~~~~~~ERROR~~~~~~~" output.log >/dev/null ; then
	cat output.log error.log >&2
	exit 1 
fi
cat output.log error.log >&2

WINNER=`grep "\[java\] \[server\]" output.log | perl -i -n -e 'if(/ \((\w)\) wins/){print "$1"}'`
REASON=`grep "Reason:" output.log | perl -i -n -e '
	if(/Reason: ([^\n]+)/){
		if ($1 eq "The winning team won by getting a lot of milk.") { print "VICTORY" }
		if ($1 eq "The winning team won on tiebreakers.") { print "TIE" }
		if ($1 =~ /Team (A|B) won by default./) { print "TIE" }
	}'`
REPLAY=
if [[ -n "$CAPTURE_REPLAY" ]] ; then
	REPLAY=`cat match.rms | base64`
fi

echo -n "{\"winner\":\"$WINNER\",\"reason\":\"$REASON\",\"replay\":\"$REPLAY\"}"

popd > /dev/null

