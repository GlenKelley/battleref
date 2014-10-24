#! /bin/bash

#
# Deploys the battleref app to a server
# 
# Author: Glen Kelley
#

function usage {
  echo ""
  echo "$0 -k root-public-key -a admin-public-key -r repo-url -H host"
  echo ""
  echo "        -r repo-url          The URL of the battleref source code to install on the server"
  echo "        -H host-url          The connection url (user@hostname) of the target server"
  echo "" 
  exit 1
}

while getopts "ha:k:H:r:" opt; do
  case $opt in
    H ) HOST="$OPTARG" ;;
    r ) REPO="$OPTARG" ;;
    h ) usage
  esac
done

set -e

if [[ -z "$HOST" ]] ; then
  echo "Error: You must define a host."
  usage
fi

if [[ -z "$REPO" ]] ; then
  echo "Error: You must define a repo."
  usage
fi

echo "Connecting to $HOST"

GIT_USER=git
WEBSERVER_USER=webserver
SUBMISSION_DIR=/opt/battleref

ssh -T $HOST <<EOF
set -e
echo "Starting: Installing applications."

if ! which git > /dev/null 2>&1 ; then
  echo "Installing git."
  yes | sudo yum install git
fi

if ! which gcc > /dev/null 2>&1 ; then
  echo "Installing gcc."
  yes | sudo yum install gcc
fi

if ! perl -e 'use Data::Dumper' > /dev/null 2>&1 ; then
  echo "Installing perl-Data-Dumper."
  yes | sudo yum install perl-Data-Dumper
fi

if ! which go > /dev/null 2>&1 ; then
  echo "Installing Go."
  yes | sudo yum install go
fi 

#TODO: sudo yum install curl libcurl

echo "Finished: Installing applications."

echo "Starting: Creating key pairs."

if [[ ! -f .ssh/git ]] ; then
  ssh-keygen -f .ssh/git -P "" -C "battleref git ssh key"
fi

if [[ ! -f .ssh/webserver ]] ; then
  ssh-keygen -f .ssh/webserver -P "" -C "battleref webserver key"
fi

echo "Finished: Creating key pairs."

echo "Starting: Installing gitolite."

if ! id -u "$GIT_USER" > /dev/null 2>&1 ; then
   echo "Creating $GIT_USER user."
   sudo adduser "$GIT_USER"
fi

if [[ -e "$SUBMISSION_DIR" ]] ; then
  if [[ ! -d "$SUBMISSION_DIR" ]] ; then
    echo "$SUBMISSION_DIR exists and is not a directory."
    exit 1
  fi
  OWNER=\`stat -c %U "$SUBMISSION_DIR"\`

  if [[ "\$OWNER" != "$GIT_USER" ]] ; then
    echo "$SUBMISSION_DIR exists but has owner \${OWNER}."
    exit 1
  fi 
  echo "Removing existing ${SUBMISSION_DIR}."
  sudo rm -r "$SUBMISSION_DIR"
fi

echo "Creating ${SUBMISSION_DIR}."
sudo mkdir -p -m 755 "$SUBMISSION_DIR"
sudo chown "${GIT_USER}:${GIT_USER}" "$SUBMISSION_DIR"

GIT_HOME=\`getent passwd $GIT_USER | cut -d: -f6\`
sudo -u $GIT_USER mkdir -p -m 0700 "\${GIT_HOME}/.ssh"
sudo cp .ssh/git.pub \${GIT_HOME}/.ssh/authorized_keys
sudo chmod 0655 \${GIT_HOME}/.ssh/authorized_keys
sudo cp .ssh/webserver.pub \${GIT_HOME}/.ssh/ 
sudo chown -R "${GIT_USER}:${GIT_USER}" \${GIT_HOME}/.ssh/

sudo su "$GIT_USER"
set -e
cd

if [[ ! -d gitolite ]] ; then
  git clone git://github.com/sitaramc/gitolite gitolite
else
  pushd gitolite > /dev/null
  git pull
  popd > /dev/null
fi

rm -rf .gitolite

mkdir -p bin
gitolite/install -ln

rm -f repositories
ln -s "$SUBMISSION_DIR" repositories

echo "Running gitolite setup."
bin/gitolite setup -pk .ssh/webserver.pub

exit

echo "Finished: Installing gitolite." 

echo "Starting: Installing webserver."

if ! id -u "$WEBSERVER_USER" > /dev/null 2>&1 ; then
   echo "Creating $WEBSERVER_USER user."
   sudo adduser "$WEBSERVER_USER"
fi


echo "Finished: Installing webserver."

WEBSERVER_HOME=\`getent passwd $WEBSERVER_USER | cut -d: -f6\`
if [[ -z "\$WEBSERVER_HOME" ]] ; then
  echo "Error: No user $WEBSERVER_USER"
  exit 1
fi
sudo -u $WEBSERVER_USER mkdir -p -m 0700 "\${WEBSERVER_HOME}/.ssh"
sudo cp .ssh/webserver \${WEBSERVER_HOME}/.ssh/webserver
sudo chown -R "${WEBSERVER_USER:$WEBSERVER_USER}" \${WEBSERVER_HOME}/.ssh/

sudo su "$WEBSERVER_USER"
set -e
cd

mkdir -p golib

if ! grep "GOPATH=\$HOME/golib" ~/.bashrc > /dev/null ; then
  echo "GOPATH=\$HOME/golib" >> ~/.bashrc
fi
export GOPATH=\$HOME/golib

go get "github.com/mattn/go-sqlite3"
go install "github.com/mattn/go-sqlite3"

go get "$REPO"
go update "$REPO"

exit

EOF

echo "Success."
