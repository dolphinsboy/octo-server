#!/usr/bin/env bash
# Group API integration tests
source "$(dirname "$0")/lib.sh"
reset_counters
log_section "Group Tests"

if ! login_user; then
  fail "login failed for group tests (HTTP $RESP_CODE)"
  print_summary || exit 1
  exit 1
fi
pass "login succeeded (uid=${USER_UID})"

# Group info
perform_request GET "/v1/groups/${GROUP_ID}" "" -H "token: ${USER_TOKEN}"
if expect_http 200 "group detail"; then
  group_no=$(json_get '.group_no // .groupNo // .id')
  if [[ -n "$group_no" && "$group_no" != "null" ]]; then
    pass "group detail returned id=${group_no}"
  else
    warn "group detail missing group_no field"
  fi
fi

# Group members
perform_request GET "/v1/groups/${GROUP_ID}/members?limit=50&page=1" "" -H "token: ${USER_TOKEN}"
if expect_http 200 "group members"; then
  member_count=$(echo "$RESP_BODY" | jq -r 'length' 2>/dev/null)
  if [[ "$member_count" =~ ^[0-9]+$ ]]; then
    pass "members count=${member_count}"
  else
    warn "members response not an array"
  fi
fi

print_summary || exit 1
