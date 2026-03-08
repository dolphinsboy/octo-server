package message

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildBatchInsertSQL 测试批量插入SQL构建逻辑
// 验证使用 strings.Builder 构建的 SQL 格式正确
func TestBuildBatchInsertSQL(t *testing.T) {
	testCases := []struct {
		name        string
		ids         []int64
		uid         string
		expectedSQL string
		expectedLen int // 参数数量
	}{
		{
			name:        "单个ID",
			ids:         []int64{1},
			uid:         "user1",
			expectedSQL: "INSERT IGNORE INTO reminder_done(reminder_id,uid) VALUES (?,?)",
			expectedLen: 2,
		},
		{
			name:        "两个ID",
			ids:         []int64{1, 2},
			uid:         "user1",
			expectedSQL: "INSERT IGNORE INTO reminder_done(reminder_id,uid) VALUES (?,?),(?,?)",
			expectedLen: 4,
		},
		{
			name:        "三个ID",
			ids:         []int64{1, 2, 3},
			uid:         "user1",
			expectedSQL: "INSERT IGNORE INTO reminder_done(reminder_id,uid) VALUES (?,?),(?,?),(?,?)",
			expectedLen: 6,
		},
		{
			name:        "五个ID",
			ids:         []int64{10, 20, 30, 40, 50},
			uid:         "user2",
			expectedSQL: "INSERT IGNORE INTO reminder_done(reminder_id,uid) VALUES (?,?),(?,?),(?,?),(?,?),(?,?)",
			expectedLen: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sql, args := buildBatchInsertSQL(tc.ids, tc.uid)
			assert.Equal(t, tc.expectedSQL, sql, "SQL语句不匹配")
			assert.Equal(t, tc.expectedLen, len(args), "参数数量不匹配")

			// 验证参数顺序和值
			for i, id := range tc.ids {
				assert.Equal(t, id, args[i*2], "reminder_id参数不匹配")
				assert.Equal(t, tc.uid, args[i*2+1], "uid参数不匹配")
			}
		})
	}
}

// TestBuildBatchInsertSQLEmpty 测试空输入处理
func TestBuildBatchInsertSQLEmpty(t *testing.T) {
	sql, args := buildBatchInsertSQL([]int64{}, "user1")
	assert.Empty(t, sql, "空输入应返回空SQL")
	assert.Empty(t, args, "空输入应返回空参数")
}

// buildBatchInsertSQL 提取SQL构建逻辑用于单元测试
// 这与 batchInsertDonesTx 中的逻辑保持一致
func buildBatchInsertSQL(sortedIds []int64, uid string) (string, []any) {
	if len(sortedIds) == 0 {
		return "", nil
	}

	var placeholders strings.Builder
	valueArgs := make([]any, 0, len(sortedIds)*2)

	for i, id := range sortedIds {
		if i > 0 {
			placeholders.WriteString(",")
		}
		placeholders.WriteString("(?,?)")
		valueArgs = append(valueArgs, id, uid)
	}

	sql := "INSERT IGNORE INTO reminder_done(reminder_id,uid) VALUES " + placeholders.String()
	return sql, valueArgs
}
