package search

import (
	"testing"

	"github.com/Mininglamp-OSS/octo-lib/common"
	"github.com/Mininglamp-OSS/octo-lib/config"
	"github.com/Mininglamp-OSS/octo-server/modules/group"
	"github.com/stretchr/testify/assert"
)

func TestCollectChannelIDs_ThreadMessage(t *testing.T) {
	tests := []struct {
		name            string
		messages        []*config.MessageResp
		expectGroupIDs  []string
		expectUIDs      []string
		expectFromUIDs  []string
		expectThreadMap map[string]string
	}{
		{
			name: "private_message",
			messages: []*config.MessageResp{
				{ChannelID: "uid_a", ChannelType: common.ChannelTypePerson.Uint8(), FromUID: "uid_b"},
			},
			expectGroupIDs:  []string{},
			expectUIDs:      []string{"uid_a"},
			expectFromUIDs:  []string{"uid_b"},
			expectThreadMap: map[string]string{},
		},
		{
			name: "group_message",
			messages: []*config.MessageResp{
				{ChannelID: "group123", ChannelType: common.ChannelTypeGroup.Uint8(), FromUID: "uid_a"},
			},
			expectGroupIDs:  []string{"group123"},
			expectUIDs:      []string{},
			expectFromUIDs:  []string{"uid_a"},
			expectThreadMap: map[string]string{},
		},
		{
			name: "thread_message_extracts_parent_group",
			messages: []*config.MessageResp{
				{ChannelID: "group123____2044239261124792320", ChannelType: common.ChannelTypeCommunityTopic.Uint8(), FromUID: "uid_a"},
			},
			expectGroupIDs:  []string{"group123"},
			expectUIDs:      []string{},
			expectFromUIDs:  []string{"uid_a"},
			expectThreadMap: map[string]string{"group123____2044239261124792320": "group123"},
		},
		{
			name: "thread_invalid_format_skipped",
			messages: []*config.MessageResp{
				{ChannelID: "no_separator", ChannelType: common.ChannelTypeCommunityTopic.Uint8(), FromUID: "uid_a"},
			},
			expectGroupIDs:  []string{},
			expectUIDs:      []string{},
			expectFromUIDs:  []string{"uid_a"},
			expectThreadMap: map[string]string{},
		},
		{
			name: "mixed_messages",
			messages: []*config.MessageResp{
				{ChannelID: "uid_x", ChannelType: common.ChannelTypePerson.Uint8(), FromUID: "uid_y"},
				{ChannelID: "grp1", ChannelType: common.ChannelTypeGroup.Uint8(), FromUID: "uid_z"},
				{ChannelID: "grp2____20441234", ChannelType: common.ChannelTypeCommunityTopic.Uint8(), FromUID: "uid_w"},
			},
			expectGroupIDs:  []string{"grp1", "grp2"},
			expectUIDs:      []string{"uid_x"},
			expectFromUIDs:  []string{"uid_y", "uid_z", "uid_w"},
			expectThreadMap: map[string]string{"grp2____20441234": "grp2"},
		},
		{
			name:            "empty_messages",
			messages:        []*config.MessageResp{},
			expectGroupIDs:  []string{},
			expectUIDs:      []string{},
			expectFromUIDs:  []string{},
			expectThreadMap: map[string]string{},
		},
		{
			name: "from_uid_empty_not_collected",
			messages: []*config.MessageResp{
				{ChannelID: "uid_a", ChannelType: common.ChannelTypePerson.Uint8(), FromUID: ""},
			},
			expectGroupIDs:  []string{},
			expectUIDs:      []string{"uid_a"},
			expectFromUIDs:  []string{},
			expectThreadMap: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupIDs, uids, fromUIDs, threadMap := collectChannelIDs(tt.messages)
			assert.Equal(t, tt.expectGroupIDs, groupIDs)
			assert.Equal(t, tt.expectUIDs, uids)
			assert.Equal(t, tt.expectFromUIDs, fromUIDs)
			assert.Equal(t, tt.expectThreadMap, threadMap)
		})
	}
}

func TestBuildThreadChannel(t *testing.T) {
	groups := []*group.GroupResp{
		{GroupNo: "grp1", Name: "开发群", Remark: "dev team"},
		{GroupNo: "grp2", Name: "测试群", Remark: ""},
	}

	tests := []struct {
		name          string
		channelID     string
		parentGroupNo string
		groups        []*group.GroupResp
		expectNil     bool
		expectID      string
		expectType    uint8
		expectName    string
	}{
		{
			name:          "thread_with_known_parent_group",
			channelID:     "grp1____2044239261124792320",
			parentGroupNo: "grp1",
			groups:        groups,
			expectNil:     false,
			expectID:      "grp1____2044239261124792320",
			expectType:    common.ChannelTypeCommunityTopic.Uint8(),
			expectName:    "开发群",
		},
		{
			name:          "thread_with_unknown_parent_group",
			channelID:     "unknown____2044239261124792320",
			parentGroupNo: "unknown",
			groups:        groups,
			expectNil:     true,
		},
		{
			name:          "thread_with_empty_groups",
			channelID:     "grp1____2044239261124792320",
			parentGroupNo: "grp1",
			groups:        nil,
			expectNil:     true,
		},
		{
			name:          "thread_html_escape_in_name",
			channelID:     "grp3____2044239261124792320",
			parentGroupNo: "grp3",
			groups:        []*group.GroupResp{{GroupNo: "grp3", Name: "<script>alert</script>", Remark: "a&b"}},
			expectNil:     false,
			expectID:      "grp3____2044239261124792320",
			expectType:    common.ChannelTypeCommunityTopic.Uint8(),
			expectName:    "&lt;script&gt;alert&lt;/script&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildThreadChannel(tt.channelID, tt.parentGroupNo, tt.groups)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectID, result.ChannelID)
				assert.Equal(t, tt.expectType, result.ChannelType)
				assert.Equal(t, tt.expectName, result.ChannelName)
			}
		})
	}
}
