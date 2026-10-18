package metrics

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Grafana dashboard queries only metrics the collector serves, and every
// one of them is on it.
func TestGrafanaDashboardUsesTheMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../deploy/grafana/resticgui-dashboard.json")
	require.NoError(t, err)
	var dash struct {
		UID    string `json:"uid"`
		Panels []struct {
			Title   string `json:"title"`
			Targets []struct {
				Expr string `json:"expr"`
			} `json:"targets"`
		} `json:"panels"`
	}
	require.NoError(t, json.Unmarshal(raw, &dash))
	assert.Equal(t, "resticgui", dash.UID)

	served := map[string]bool{}
	descs := make(chan *prometheus.Desc, 32)
	Collector{}.Describe(descs)
	close(descs)
	name := regexp.MustCompile(`fqName: "([a-z_]+)"`)
	for d := range descs {
		served[name.FindStringSubmatch(d.String())[1]] = true
	}

	used := map[string]bool{}
	metric := regexp.MustCompile(`resticgui_[a-z_]+`)
	for _, p := range dash.Panels {
		for _, q := range p.Targets {
			require.NotEmpty(t, q.Expr, "panel %q", p.Title)
			for _, m := range metric.FindAllString(q.Expr, -1) {
				assert.True(t, served[m], "panel %q queries %s, which is not served", p.Title, m)
				used[m] = true
			}
		}
	}
	for m := range served {
		assert.True(t, used[m], "%s is on no panel", m)
	}
}
