#!/bin/bash
set -ex

# create keys if necessary
# add user git
# create git repo /opt/battleref
# install giolite

ROOT_KEY=.ssh/gitolite_rsa
ADMIN_KEY=.ssh/webserver_rsa
GIT_USER=git
REPO=/opt/battleref
GITOLITE=git://github.com/sitaramc/gitolite

which git || sudo yum install git
which gcc || sudo yum install gcc
perl -e 'user Data::Dumper' || sudo yum install perl-Data-Dumper
which go || sudo yum install go
sudo yum install curl libcurl

[[ -n `diff "$ROOT_KEY" "$ADMIN_KEY"` ]] || ( echo "the root and admin keys must be different" && exit 1 )

id -u "$GIT_USER" &>/dev/null 2>&1 || sudo adduser "$GIT_USER"

mkdir -p "$REPO"
chown "$GIT_USER:$GIT_USER" "$REPO"

sudo -E su "$GIT_USER" <<'EOF'
	set -exv
	
	cd
	[[ -e gitolite ]] || git clone "$GITOLITE"
	mkdir -p bin
	gitolite/install -ln

	mkdir -m 0700 -p .ssh
	[[ -f "$ROOT_KEY"  ]] || ssh-keygen -f "$ROOT_KEY" -P "" -C "gitolite key"
	[[ -f "$ADMIN_KEY" ]] || ssh-keygen -f "$ADMIN_KEY" -P "" -C "webserver key"

	cat "$ROOT_KEY" > .ssh/authorized_keys
	chmod 0644 .ssh/authorized_keys

	rm -f repositories
	ln -s "$REPO" repositories

	bin/gitolite setup -pk `basename "$ADMIN_KEY.pub"`
EOF
