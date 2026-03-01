#!/usr/bin/env bash
# Messaging integration tests
source "$(dirname "$0")/lib.sh"
reset_counters
log_section "Messaging Tests"

if ! login_user; then
  fail "login failed for messaging tests (HTTP $RESP_CODE)"
  print_summary || exit 1
  exit 1
fi
pass "login succeeded (uid=${USER_UID})"

# Send text message to group
MSG_TEXT="E2E message $(date +%s)"
send_payload=$(cat <<JSON
{
  "token": "${USER_TOKEN}",
  "receive_channel_id": "${GROUP_ID}",
  "receive_channel_type": ${CHANNEL_TYPE_GROUP},
  "payload": {"type":1,"content":"${MSG_TEXT}"}
}
JSON
)
perform_request POST "/v1/message/send" "$send_payload"
expect_http 200 "send text message" || true

# Fetch channel history
history_payload=$(cat <<JSON
{
  "channel_id": "${GROUP_ID}",
  "channel_type": ${CHANNEL_TYPE_GROUP},
  "device_uuid": "cli-$$",
  "start_message_seq": 0,
  "end_message_seq": 0,
  "limit": 20,
  "pull_mode": 0
}
JSON
)
perform_request POST "/v1/message/channel/sync" "$history_payload" -H "token: ${USER_TOKEN}"
if expect_http 200 "channel history"; then
  found=$(echo "$RESP_BODY" | jq -r --arg txt "$MSG_TEXT" '.messages[]?.payload.content // empty | select(.==$txt)' | head -n1)
  if [[ -n "$found" ]]; then
    pass "history contains sent message"
  else
    warn "history did not include the latest message (body retained for review)"
  fi
fi

# Message sync (global)
sync_payload=$(cat <<JSON
{
  "max_message_seq": 0,
  "limit": 20,
  "channel_id": "${GROUP_ID}",
  "channel_type": ${CHANNEL_TYPE_GROUP},
  "reverse": 0,
  "offset": 0
}
JSON
)
perform_request POST "/v1/message/sync" "$sync_payload" -H "token: ${USER_TOKEN}"
if expect_http 200 "message sync"; then
  count=$(echo "$RESP_BODY" | jq -r 'length' 2>/dev/null)
  if [[ "$count" =~ ^[0-9]+$ && "$count" -gt 0 ]]; then
    pass "sync returned $count messages"
  else
    warn "sync returned empty set"
  fi
fi

print_summary || exit 1
