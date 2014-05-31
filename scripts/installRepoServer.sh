#!/bin/bash
set -ex

GIT_DIR=/opt/git/
REPO=$GIT_DIR/battleref.git
WEB_USER=webserver
BRANCH=master

mkdir -p "$GIT_DIR"
chown "$WEB_USER:$WEB_USER" "$GIT_DIR"

sudo -E su "$WEB_USER" <<'EOF'
[[ -e "$REPO" ]] || git clone --bare https://github.com/GlenKelley/battlecode.git "$REPO"

cd "$REPO"
git show "$MASTER:scripts/post-receive" > "$REPO/.git/hooks/post-receive"
EOF
