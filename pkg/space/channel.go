package space

import "strings"

const SpaceChannelPrefix = "s"

// BuildChannelID 构建 Space 内的 channel_id
// 个人空间返回原始 peerID
func BuildChannelID(spaceID, peerID string) string {
	if spaceID == "" {
		return peerID
	}
	return SpaceChannelPrefix + spaceID + "_" + peerID
}

// ParseChannelID 从 channel_id 解析 space_id 和 peer_id
func ParseChannelID(channelID string) (spaceID, peerID string) {
	if !strings.HasPrefix(channelID, SpaceChannelPrefix) {
		return "", channelID
	}
	// 格式: s{spaceId}_{peerId}
	rest := channelID[len(SpaceChannelPrefix):]
	idx := strings.Index(rest, "_")
	if idx < 0 {
		return "", channelID
	}
	return rest[:idx], rest[idx+1:]
}
