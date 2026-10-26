package metrics

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

type fakeBandwidth map[protocol.ID]metrics.Stats

func (f fakeBandwidth) NetworkStatsByProtocol() map[protocol.ID]metrics.Stats {
	return f
}

// TestCollectBandwidth checks that the totals are counters by protocol and
// direction, and that traffic without a protocol is labeled unknown.
func TestCollectBandwidth(t *testing.T) {
	const name = "awl_bootstrap_p2p_bandwidth_bytes_total"
	SetBandwidthStats(fakeBandwidth{
		"/awl/kad/1.0.0": {TotalIn: 10, TotalOut: 20, RateIn: 1, RateOut: 2},
		"":               {TotalIn: 3, TotalOut: 4},
	})
	t.Cleanup(func() { scrapeBandwidth.Store(nil) })

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	got := make(map[string]float64)
	for _, f := range families {
		if f.GetName() != name {
			continue
		}
		if f.GetType().String() != "COUNTER" {
			t.Errorf("%s has type %v, want counter", name, f.GetType())
		}
		for _, m := range f.GetMetric() {
			labels := make(map[string]string)
			for _, l := range m.GetLabel() {
				labels[l.GetName()] = l.GetValue()
			}
			got[labels["protocol"]+" "+labels["direction"]] = m.GetCounter().GetValue()
		}
	}
	want := map[string]float64{
		"/awl/kad/1.0.0 in":  10,
		"/awl/kad/1.0.0 out": 20,
		"unknown in":         3,
		"unknown out":        4,
	}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s{%s} = %v, want %v", name, k, got[k], v)
		}
	}
}
