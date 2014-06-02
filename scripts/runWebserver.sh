#!/bin/bash
set -x

REPO=/opt/git/battleref.git
LOG_DIR=~/.battleref/log
RLOG=$LOG_DIR/runloop.log
mkdir -p "$LOG_DIR"

until false; do
	export BATTLEREF_DB=~/.battleref/db.sqlite3

	RUN_DIR=`mktemp -d --suffix battleref`
	git clone "$REPO" "$RUN_DIR"

	cd "$RUN_DIR"
	EXIT_STATUS=go run "$RUN_DIR/battleref/app/main.go" "$RUN_DIR/battleref/app/config.json" >> "$LOG_DIR/out.log" 2>> "$LOG_DIR/err.log"
	rm -rf "$RUN_DIR"

	[[ $EXIT_STATUS -eq 0 ]] || exit $EXIT_STATUS
    echo "BattleCodeServer terminated with exit code $EXIT_STATUS. Restarting ..." >>"$RLOG"

	sleep 1
done
