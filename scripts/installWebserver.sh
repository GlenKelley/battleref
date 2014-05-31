#!/bin/bash
set -ex

# add user webserver
# install a battleref server

WEB_USER=webserver
GIT_USER=git
ADMIN_KEY=/home/$GIT_USER/.ssh/webserver_rsa
WEB_KEY=/home/$WEB_USER/.ssh/webserver_rsa
WEB_HOME=/home/$WEB_USER

id -u "$WEB_USER" &>/dev/null 2>&1 || sudo adduser "$WEB_USER"

sudo mkdir -m 0700 -p "$WEB_HOME/.ssh"
sudo cp -n "$ADMIN_KEY" "$WEB_KEY" || true
sudo cp -n ~/.ssh/authorized_keys "$WEB_HOME/.ssh/authorized_keys" || true
sudo chown -R "$WEB_USER:$WEB_USER" "$WEB_HOME/.ssh"

sudo -E su "$WEB_USER" <<'EOF'
	set -exv
	cd
	mkdir -p golib
	grep "GOPATH" ~/.bashrc || echo "export GOPATH=$HOME/golib" >> ~/.bashrc
	. ~/.bashrc
	go get "github.com/mattn/go-sqlite3"
	go install "github.com/mattn/go-sqlite3"
EOF
