package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const thirtySeconds = "30s"

func TestValidateRequestTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		scrapeInterval string
		requestTimeout string
		wantTimeout    string
		wantErr        string
	}{
		{name: "shorter than scrape interval", scrapeInterval: "1m", requestTimeout: thirtySeconds, wantTimeout: thirtySeconds},
		{name: "equal to scrape interval", scrapeInterval: thirtySeconds, requestTimeout: thirtySeconds, wantTimeout: thirtySeconds},
		{
			name:           "longer than scrape interval gets clamped",
			scrapeInterval: "15s",
			requestTimeout: thirtySeconds,
			wantTimeout:    "15s",
		},
		{
			name:           "zero timeout",
			scrapeInterval: "1m",
			requestTimeout: "0s",
			wantErr:        "ntopng requestTimeout must be greater than zero",
		},
		{
			name:           "negative timeout",
			scrapeInterval: "1m",
			requestTimeout: "-5s",
			wantErr:        "ntopng requestTimeout must be greater than zero",
		},
		{
			name:           "invalid timeout",
			scrapeInterval: "1m",
			requestTimeout: "slow",
			wantErr:        "was not able to parse configured request timeout: slow - time: invalid duration \"slow\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := Config{
				Ntopng: ntopng{
					AuthMethod:      "none",
					ScrapeInterval:  tt.scrapeInterval,
					RequestTimeout:  tt.requestTimeout,
					ScrapeTargets:   []string{AllScrape},
					ParallelWorkers: DefaultParallelWorkers,
				},
				Host: host{InterfacesToMonitor: []string{"eth0"}},
				Metric: metric{
					Serve: metricServe{IP: "0.0.0.0"},
				},
			}

			err := config.validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.wantTimeout, config.Ntopng.RequestTimeout)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestDefaultRequestTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		scrapeInterval time.Duration
		want           time.Duration
	}{
		{name: "long interval capped at default", scrapeInterval: time.Minute, want: DefaultRequestTimeout},
		{name: "short interval wins", scrapeInterval: 15 * time.Second, want: 15 * time.Second},
		{name: "equal interval", scrapeInterval: DefaultRequestTimeout, want: DefaultRequestTimeout},
		{name: "zero interval falls back to default", scrapeInterval: 0, want: DefaultRequestTimeout},
		{name: "negative interval falls back to default", scrapeInterval: -time.Second, want: DefaultRequestTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, defaultRequestTimeout(tt.scrapeInterval))
		})
	}
}
