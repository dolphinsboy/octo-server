#!/usr/bin/env bash
set -o pipefail
BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
FAIL=0

declare -a scripts=(test-auth.sh test-messaging.sh test-bot.sh test-group.sh)
for script in "${scripts[@]}"; do
  echo "\n===== ${script} ====="
  if ! "${BASE_DIR}/${script}"; then
    FAIL=1
  fi
done
exit ${FAIL}
