package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TestRelayMetricsTracer checks that a closed circuit reaches both
// histograms: ours and the libp2p one behind the wrapper.
func TestRelayMetricsTracer(t *testing.T) {
	const (
		ours   = "awl_bootstrap_p2p_relay_circuit_duration_seconds"
		libp2p = "libp2p_relaysvc_connection_duration_seconds"
	)
	reg := prometheus.NewRegistry()
	tracer := RelayMetricsTracer(reg)
	oursBefore := sampleCount(t, prometheus.DefaultGatherer, ours)
	libp2pBefore := sampleCount(t, reg, libp2p)

	tracer.ConnectionClosed(2 * time.Minute)

	if got := sampleCount(t, prometheus.DefaultGatherer, ours) - oursBefore; got != 1 {
		t.Errorf("%s got %d samples, want 1", ours, got)
	}
	if got := sampleCount(t, reg, libp2p) - libp2pBefore; got != 1 {
		t.Errorf("%s got %d samples, want 1", libp2p, got)
	}
}

// sampleCount returns how many values the histogram name in g has observed.
func sampleCount(t *testing.T, g prometheus.Gatherer, name string) uint64 {
	t.Helper()
	families, err := g.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == name {
			return f.GetMetric()[0].GetHistogram().GetSampleCount()
		}
	}
	t.Fatalf("%s is not registered", name)
	return 0
}
