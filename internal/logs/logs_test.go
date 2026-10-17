package Logs

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitlogs(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		want   zerolog.Level
	}{
		{"defaults to info", map[string]any{}, zerolog.InfoLevel},
		{"reads logs.level", map[string]any{"logs.level": 2}, zerolog.WarnLevel},
		{"logs.debug overrides level", map[string]any{"logs.debug": true, "logs.level": 3}, zerolog.DebugLevel},
		// regression: the keys used to be read as log.* so the config file was ignored and the level was 0 (debug)
		{"ignores legacy log.* keys", map[string]any{"log.debug": true}, zerolog.InfoLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			for k, v := range tt.config {
				viper.Set(k, v)
			}
			Initlogs()
			assert.Equal(t, tt.want, zerolog.GlobalLevel())
		})
	}
}
