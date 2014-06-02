#!/bin/bash
set -ex

# create keys if necessary
# add user git
# create git repo /opt/battleref
# install giolite

export ROOT_KEY=.ssh/gitolite_rsa
export ADMIN_KEY=.ssh/webserver_rsa
export GIT_USER=git
export REPO=/opt/battleref
export GITOLITE=git://github.com/sitaramc/gitolite

which git 2>/dev/null || sudo yum install git
which gcc 2>/dev/null || sudo yum install gcc
perl -e 'user Data::Dumper' 2>/dev/null || sudo yum install perl-Data-Dumper
which go 2>/dev/null || sudo yum install go
sudo yum install curl libcurl

id -u "$GIT_USER" &>/dev/null 2>&1 || sudo adduser "$GIT_USER"

sudo mkdir -p "$REPO"
sudo chown "$GIT_USER:$GIT_USER" "$REPO"

sudo -E su "$GIT_USER" <<'EOF'
	set -exv

	cd
	[[ -e gitolite ]] || git clone "$GITOLITE" gitolite
	mkdir -p bin
	gitolite/install -ln

	mkdir -m 0700 -p .ssh
	[[ -f "$ROOT_KEY"  ]] || ssh-keygen -f "$ROOT_KEY" -P "" -C "gitolite key"
	[[ -f "$ADMIN_KEY" ]] || ssh-keygen -f "$ADMIN_KEY" -P "" -C "webserver key"

	cat "$ROOT_KEY.pub" > .ssh/authorized_keys
	chmod 0644 .ssh/authorized_keys

	rm -f repositories
	ln -s "$REPO" repositories

	bin/gitolite setup -pk "$ADMIN_KEY.pub"
EOF
