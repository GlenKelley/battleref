#! /bin/bash

#
# Installs the gitolite test user
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0 -r repo-url -h host"
  echo ""
  echo "Installs a gitolite test user account."
  echo ""
  echo "        -u test_user_account    This account will be used to install gitolite"
  echo "        -v                      Enable verbose debugging."
  echo "" 
  exit 1
}

TEST_USER="git-test"
while getopts "h:u:v" opt; do
  case $opt in
    u ) TEST_USER="$OPTARG" ;;
    v ) VERBOSE=TRUE ;;
    ? ) usage
  esac
done

set -e

DIR=`dirname $0`

if [[ -z "$TEST_USER" ]] ; then
  echo "Error: You must define a test user."
  usage
elif ! echo "$TEST_USER" | grep "test" >/dev/null ; then 
  echo "Error: $TEST_USER does not contain 'test'. This script will destroy user data."
  usage
fi

if [[ -z "$VERBOSE" ]] ; then
	set -x
fi

if ! id -u "$TEST_USER" > /dev/null 2>&1 ; then
   echo "Creating $TEST_USER user."
   "$DIR/adduser.sh" -u "$TEST_USER"
fi

echo "Starting: Creating key pairs."

pushd ~ > /dev/null
if [[ ! -f .ssh/git ]] ; then
  ssh-keygen -f .ssh/git -P "" -C "battleref git ssh key"
fi

if [[ ! -f .ssh/webserver ]] ; then
  ssh-keygen -f .ssh/webserver -P "" -C "battleref webserver key"
fi
popd > /dev/null

echo "Finished: Creating key pairs."

echo "Starting: Installing gitolite."

GIT_HOME=
if [[ "$OSTYPE" == "linux-gnu" ]]; then # Linux
	GIT_HOME=`getent passwd $GIT_HOME | cut -d: -f6`
elif [[ "$OSTYPE" == "darwin"* ]]; then # Mac OSX
	GIT_HOME=/Users/$TEST_USER
fi

sudo -u "$TEST_USER" mkdir -p -m 0700 $GIT_HOME/.ssh

sudo cp ~/.ssh/git.pub $GIT_HOME/.ssh/authorized_keys
sudo chmod 0655 $GIT_HOME/.ssh/authorized_keys
sudo cp ~/.ssh/webserver.pub $GIT_HOME/.ssh/ 
sudo chown -R "$GIT_USER" $GIT_HOME/.ssh/

sudo su "$TEST_USER" <<EOF
set -e
cd

if [[ ! -d gitolite ]] ; then
  git clone git://github.com/sitaramc/gitolite gitolite
#else
#  pushd gitolite > /dev/null
#  git pull
#  popd > /dev/null
fi

rm -rf .gitolite

mkdir -p bin
gitolite/install -ln

mkdir -p repositories

echo "Running gitolite setup."
bin/gitolite setup -pk .ssh/webserver.pub

exit
EOF

echo "Finished: Installing gitolite." 

echo "Success."
