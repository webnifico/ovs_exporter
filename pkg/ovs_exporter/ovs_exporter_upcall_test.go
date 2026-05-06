package ovs_exporter

import "testing"

func TestParseUpcallShowMetrics(t *testing.T) {
	output := `system@ovs-system:
  flows         : (current 300) (avg 300) (max 295560) (limit 175500)
  dump duration : 28ms
  ufid enabled : true

  762: (keys 30)
  763: (keys 34)
`

	metrics, err := parseUpcallShowMetrics(output)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("expected 1 datapath metric set, got %d", len(metrics))
	}

	got := metrics[0]
	if got.datapath != "system@ovs-system" {
		t.Fatalf("expected datapath system@ovs-system, got %s", got.datapath)
	}
	if got.flowsCurrent != 300 || got.flowsAvg != 300 || got.flowsMax != 295560 || got.flowsLimit != 175500 {
		t.Fatalf("unexpected flow counters: %+v", got)
	}
	if got.dumpDurationMs != 28 {
		t.Fatalf("expected dump duration 28ms, got %v", got.dumpDurationMs)
	}
	if got.ufidEnabled != 1 {
		t.Fatalf("expected ufid enabled to be 1, got %v", got.ufidEnabled)
	}
	if got.handlerKeys["762"] != 30 || got.handlerKeys["763"] != 34 {
		t.Fatalf("unexpected handler keys: %+v", got.handlerKeys)
	}
}

func TestParseUpcallShowMetricsNoDatapath(t *testing.T) {
	_, err := parseUpcallShowMetrics("flows         : (current 300) (avg 300) (max 295560) (limit 175500)")
	if err == nil {
		t.Fatal("expected parsing error when no datapath header is present")
	}
}
