#!/bin/bash
set -ex

export ROOT_DIR=/opt/git
export REPO=$ROOT_DIR/battleref.git
export WEB_USER=webserver
export BRANCH=master

sudo mkdir -p "$ROOT_DIR"
sudo chown "$WEB_USER:$WEB_USER" "$ROOT_DIR"

sudo -E su "$WEB_USER" <<'EOF'
set -ex
cd "$ROOT_DIR"
[[ -e "$REPO" ]] || git clone --bare git://github.com/GlenKelley/battleref.git "$REPO"
cd "$REPO"
git show "$BRANCH:scripts/post-receive" > "$REPO/hooks/post-receive"
EOF

