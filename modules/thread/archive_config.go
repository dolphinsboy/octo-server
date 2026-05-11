package thread

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// ArchiveConfig 子区自动归档 worker 的运行参数。
type ArchiveConfig struct {
	// Enabled 总开关。disabled 时 worker 不启动 ticker。
	Enabled bool
	// Threshold 子区"陈旧"判定阈值（自 last_message_at 起算）。0 视为禁用。
	Threshold time.Duration
	// Interval 两次 cron tick 之间的间隔。<=0 视为禁用。
	Interval time.Duration
	// BatchSize 单次 UPDATE 的最大行数。<=0 时回退默认值。
	BatchSize int
	// BatchSleep 两次批之间的 sleep，给 DB 喘息。<0 视为 0。
	BatchSleep time.Duration
}

const (
	envArchiveEnabled    = "DM_THREAD_AUTO_ARCHIVE_ENABLED"
	envArchiveDays       = "DM_THREAD_AUTO_ARCHIVE_DAYS"
	envArchiveInterval   = "DM_THREAD_AUTO_ARCHIVE_INTERVAL"
	envArchiveBatchSize  = "DM_THREAD_AUTO_ARCHIVE_BATCH_SIZE"
	envArchiveBatchSleep = "DM_THREAD_AUTO_ARCHIVE_BATCH_SLEEP"

	defaultArchiveDays       = 3
	defaultArchiveInterval   = time.Hour
	defaultArchiveBatchSize  = 500
	defaultArchiveBatchSleep = 100 * time.Millisecond
	// maxArchiveBatchSize 防止运维误填巨数把单次 UPDATE 变成长事务，超出即截断到上限。
	// 5000 行单次 UPDATE 在 InnoDB 上的锁/binlog 体量仍可控。
	maxArchiveBatchSize = 5000
)

// LoadArchiveConfig 从环境变量装载配置。
// 错误 / 越界值一律回退默认值，避免运维误填导致 worker 行为失控。
// `DM_THREAD_AUTO_ARCHIVE_DAYS=0` 显式禁用阈值（threshold 归零），但 Enabled 仍可为 true。
func LoadArchiveConfig() ArchiveConfig {
	cfg := ArchiveConfig{
		Enabled:    parseBool(os.Getenv(envArchiveEnabled)),
		Threshold:  parseDays(os.Getenv(envArchiveDays), defaultArchiveDays),
		Interval:   parseDuration(os.Getenv(envArchiveInterval), defaultArchiveInterval),
		BatchSize:  parseBoundedPositiveInt(os.Getenv(envArchiveBatchSize), defaultArchiveBatchSize, maxArchiveBatchSize),
		BatchSleep: parseNonNegativeDuration(os.Getenv(envArchiveBatchSleep), defaultArchiveBatchSleep),
	}
	return cfg
}

func parseBool(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	return v == "true" || v == "1"
}

// parseDays 把字符串当"天数"解析。负数视为非法，回默认；0 是合法的"禁用"。
func parseDays(raw string, defaultDays int) time.Duration {
	if raw == "" {
		return time.Duration(defaultDays) * 24 * time.Hour
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return time.Duration(defaultDays) * 24 * time.Hour
	}
	return time.Duration(n) * 24 * time.Hour
}

func parseDuration(raw string, def time.Duration) time.Duration {
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func parseNonNegativeDuration(raw string, def time.Duration) time.Duration {
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < 0 {
		return def
	}
	return d
}

func parsePositiveInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// parseBoundedPositiveInt 解析正整数并截断到上限。<=0 或非法值回默认；>max 截到 max。
func parseBoundedPositiveInt(raw string, def, max int) int {
	n := parsePositiveInt(raw, def)
	if n > max {
		return max
	}
	return n
}
