#!/usr/bin/env sh
DEBUG=${DEBUG:-false}
[ $DEBUG = true ] && set -x

DIR="$(dirname $0)"
LOG_FILE="$DIR/logs/info.log"

mkdir -p $(dirname "$LOG_FILE")

if [ $DEBUG = true ]; then
  "$DIR/src/photoframe.sh"
else
  "$DIR/src/photoframe.sh" >> "$LOG_FILE" 2>&1 &
fi

