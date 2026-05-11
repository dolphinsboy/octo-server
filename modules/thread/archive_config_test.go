package thread

import (
	"os"
	"testing"
	"time"
)

func TestLoadArchiveConfig(t *testing.T) {
	// 用例之间互不污染：把所有相关 env 都清空，子用例显式 setenv 自己关心的那些。
	envs := []string{
		envArchiveEnabled,
		envArchiveDays,
		envArchiveInterval,
		envArchiveBatchSize,
		envArchiveBatchSleep,
	}

	day := 24 * time.Hour

	tests := []struct {
		name string
		env  map[string]string
		want ArchiveConfig
	}{
		{
			name: "all unset uses defaults but disabled",
			env:  map[string]string{},
			want: ArchiveConfig{
				Enabled:    false,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "enabled=true uses defaults",
			env: map[string]string{
				envArchiveEnabled: "true",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "enabled=1 also enables (truthy variant)",
			env: map[string]string{
				envArchiveEnabled: "1",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "enabled=false stays disabled",
			env: map[string]string{
				envArchiveEnabled: "false",
			},
			want: ArchiveConfig{
				Enabled:    false,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "custom valid values",
			env: map[string]string{
				envArchiveEnabled:    "true",
				envArchiveDays:       "7",
				envArchiveInterval:   "30m",
				envArchiveBatchSize:  "200",
				envArchiveBatchSleep: "50ms",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  7 * day,
				Interval:   30 * time.Minute,
				BatchSize:  200,
				BatchSleep: 50 * time.Millisecond,
			},
		},
		{
			name: "days=0 disables threshold (cron stays off via Enabled gate)",
			env: map[string]string{
				envArchiveEnabled: "true",
				envArchiveDays:    "0",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  0,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "invalid values fall back to defaults",
			env: map[string]string{
				envArchiveEnabled:    "true",
				envArchiveDays:       "abc",
				envArchiveInterval:   "not-a-duration",
				envArchiveBatchSize:  "-5",
				envArchiveBatchSleep: "garbage",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "negative days fall back to default",
			env: map[string]string{
				envArchiveEnabled: "true",
				envArchiveDays:    "-3",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "oversized batch_size is capped at 5000",
			env: map[string]string{
				envArchiveEnabled:   "true",
				envArchiveBatchSize: "999999",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  5000,
				BatchSleep: 100 * time.Millisecond,
			},
		},
		{
			name: "zero batch_sleep is valid (no pause between batches)",
			env: map[string]string{
				envArchiveEnabled:    "true",
				envArchiveBatchSleep: "0",
			},
			want: ArchiveConfig{
				Enabled:    true,
				Threshold:  3 * day,
				Interval:   time.Hour,
				BatchSize:  500,
				BatchSleep: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, e := range envs {
				_ = os.Unsetenv(e)
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got := LoadArchiveConfig()
			if got != tt.want {
				t.Errorf("LoadArchiveConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
