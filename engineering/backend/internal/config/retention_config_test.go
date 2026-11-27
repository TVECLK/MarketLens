package config

import (
	"strings"
	"testing"
)

func TestRetentionConfigValidate(t *testing.T) {
	good := RetentionConfig{RawYears: 3, TimeZone: "Asia/Colombo", BatchSize: 5000}
	tests := []struct {
		name    string
		mutate  func(*RetentionConfig)
		wantErr string // substring; empty means valid
	}{
		{"purge off", func(c *RetentionConfig) {}, ""},
		{"purge on", func(c *RetentionConfig) { c.PurgeEnabled = true }, ""},
		{"no raw years", func(c *RetentionConfig) { c.RawYears = 0 }, "RETENTION_RAW_YEARS"},
		{"no batch size", func(c *RetentionConfig) { c.BatchSize = 0 }, "RETENTION_BATCH_SIZE"},
		{"unknown time zone", func(c *RetentionConfig) { c.TimeZone = "Mars/Olympus" }, "RETENTION_TZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := good
			tt.mutate(&cfg)
			err := cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() = %v, want an error mentioning %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig_RetentionDefaultsAndPurgeSwitch(t *testing.T) {
	for _, k := range []string{"RETENTION_RAW_YEARS", "RETENTION_TZ", "RETENTION_BATCH_SIZE", "PURGE_ENABLED"} {
		t.Setenv(k, "")
	}
	c := LoadConfig().Retention
	if c.RawYears != 3 || c.TimeZone != "Asia/Colombo" || c.BatchSize != 5000 || c.PurgeEnabled {
		t.Errorf("defaults = %+v, want 3 years, Asia/Colombo, batch 5000, purge off", c)
	}

	// Anything but exactly "true" leaves the purge off.
	for _, v := range []string{"1", "yes", "TRUE", "True", "on"} {
		t.Setenv("PURGE_ENABLED", v)
		if LoadConfig().Retention.PurgeEnabled {
			t.Errorf("PURGE_ENABLED=%q turned the purge on", v)
		}
	}
	t.Setenv("PURGE_ENABLED", "true")
	if !LoadConfig().Retention.PurgeEnabled {
		t.Error("PURGE_ENABLED=true did not turn the purge on")
	}
}
