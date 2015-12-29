#! /bin/bash

#
# Deploys the battleref backend to a server
# 
# Author: Glen Kelley
#

function header {
  echo -e "\033[1;32m$@\033[0;30m" 
}

function usage {
  if [[ -n "$@" ]]; then
    header "$@"
  fi
  echo ""
  echo "$0 -r repo-url [-h host] [-e env] [-v] [-h]"
  echo ""
  echo "Deploys a battleref server to a remote host. Requires sudo"
  echo ""
  echo "        -r repo-url          The go package of the battleref source code to install on the server"
  echo "        -h host-url          The connection url (user@hostname) of the target server"
  echo "        -e environment       The environment to use when running the webserver"
  echo "        -v                   Verbose output"
  echo "        -h                   Prints this message"
  echo "" 
  exit 1
}

REPO="github.com/GlenKelley/battleref"
HOST="ec2-user@api.akusete.com"
ENV="prod"
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

if [[ -z "$HOST" ]] ; then
  usage "Error: You must define a host."
fi

if [[ -z "$REPO" ]] ; then
  usage "Error: You must define a repo."
fi

ENVIRONMENTS="dev\nprod"
if [[ -z "$ENV" ]] ; then
  usage "Error: You must define an environment."
elif ! ( echo -e "$ENVIRONMENTS" | grep "$ENV" > /dev/null ) ; then
  usage "Error: Environment $ENV is not one of the following:\n$ENVIRONMENTS"
fi

if [[ -n "$VERBOSE" ]] ; then
	set -x
fi

echo "Connecting to $HOST"

GIT_USER=git
WEBSERVER_USER=webserver
SUBMISSION_DIR=/opt/battleref

ssh -T $HOST <<EOF
set -e
if [[ -n "$VERBOSE" ]] ; then
	set -x
fi

function header {
  echo -e "\033[1;32m\$@\033[0;30m" 
}

# Checking if webserver user exitsts.
if id -u "$WEBSERVER_USER" > /dev/null 2>&1 ; then
  echo "Account $WEBSERVER_USER exists."
  WEBSERVER_HOME=\`getent passwd $WEBSERVER_USER | cut -d: -f6\`
  if sudo -u "$WEBSERVER_USER" test -d "\$WEBSERVER_HOME/.battleref" ; then
    echo "Shutting down existing server."
    echo "Shutdown by $0 to install" | sudo -u "$WEBSERVER_USER" tee \$WEBSERVER_HOME/.battleref/.shutdown
    curl localhost:$SHUTDOWN_PORT/shutdown > /dev/null 2>&1 | true
    # TODO:(gkelley) verify application has been shutdown.
  fi
fi

header "Starting: Installing applications."

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
go version

#TODO: sudo yum install curl libcurl

header "Finished: Installing applications."

header "Starting: Creating key pairs."

if [[ ! -f .ssh/git ]] ; then
  ssh-keygen -f .ssh/git -P "" -C "battleref git ssh key"
fi

if [[ ! -f .ssh/webserver ]] ; then
  ssh-keygen -f .ssh/webserver -P "" -C "battleref webserver key"
fi

header "Finished: Creating key pairs."

header "Starting: Installing gitolite."

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

header "Creating ${SUBMISSION_DIR}."
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

function header {
  echo -e "\033[1;32m\$@\033[0;30m" 
}

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

header "Running gitolite setup."
bin/gitolite setup -pk .ssh/webserver.pub

exit

header "Finished: Installing gitolite." 

header "Starting: Installing webserver."

if ! id -u "$WEBSERVER_USER" > /dev/null 2>&1 ; then
   echo "Creating $WEBSERVER_USER user."
   sudo adduser "$WEBSERVER_USER"
fi


header "Finished: Installing webserver."

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

function header {
  echo -e "\033[1;32m\$@\033[0;30m" 
}

mkdir -p golib

if ! grep "GOPATH=\$HOME/golib" ~/.bashrc > /dev/null ; then
  echo "GOPATH=\$HOME/golib" >> ~/.bashrc
fi
export GOPATH=\$HOME/golib

go get "github.com/mattn/go-sqlite3"
go install "github.com/mattn/go-sqlite3"

go get -u "$REPO"
go build -o ~/bin/startBattlerefServer "$REPO"

PROMPT="Generated by battleref installer"
START_SERVER_CRON_LINE="*/1 * * * * \$GOPATH/src/$REPO/scripts/restartServer.sh -e $ENV -d $REPO # \$PROMPT"
(
	crontab -l | grep -v "\$PROMPT" | cat
	echo "\$START_SERVER_CRON_LINE" 
) | crontab -

# Remove shutdown lock file.
rm -f ~/.battleref/.shutdown

exit

EOF

header "Success."
