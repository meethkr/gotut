#!/usr/bin/env bash
# Start GoTut and serve it at http://127.0.0.1:8080
#
#   ./run.sh          run in the foreground (Ctrl-C to stop)
#   ./run.sh --bg     run in the background, logging to /tmp/gotut.log
#
# Editing anything under static/ only needs a page reload; changes to Go code or
# the curriculum need a restart.
set -euo pipefail
cd "$(dirname "$0")"

PORT="${PORT:-8080}"
if lsof -ti "tcp:$PORT" >/dev/null 2>&1; then
  echo "Something is already listening on port $PORT:"
  lsof -i "tcp:$PORT" | tail -n +2
  echo
  echo "Stop it with:  lsof -ti tcp:$PORT | xargs kill"
  exit 1
fi

if [[ "${1:-}" == "--bg" ]]; then
  (nohup go run . >/tmp/gotut.log 2>&1 </dev/null &)
  sleep 2
  if lsof -ti "tcp:$PORT" >/dev/null 2>&1; then
    echo "GoTut is running at http://127.0.0.1:$PORT  (log: /tmp/gotut.log)"
  else
    echo "Failed to start. Log:"
    cat /tmp/gotut.log
    exit 1
  fi
else
  exec go run .
fi
