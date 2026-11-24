// Copyright 2018 Paul Greenberg (greenpau@outlook.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ovs_exporter

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestInterfaceLabelPolicyOmitsSecrets(t *testing.T) {
	published := publishAttributeKeys(t, interfaceAttributeOptions, map[string]string{
		"remote_ip":   "192.0.2.10",
		"local_ip":    "192.0.2.1",
		"key":         "flow",
		"in_key":      "10",
		"psk":         "fixture-psk-do-not-use",
		"password":    "fixture-password-do-not-use",
		"private_key": "/tmp/fixture-private-key",
		"PSK":         "fixture-psk-upper",
		"private-key": "/tmp/fixture-private-key-hyphen",
		"api_token":   "fixture-token",
		"unknown_opt": "keep-me-out",
		"":            "blank-key",
		"ttl":         "",
	})

	assertPublished(t, published, "remote_ip", "192.0.2.10")
	assertPublished(t, published, "local_ip", "192.0.2.1")
	assertPublished(t, published, "key", "flow")
	assertPublished(t, published, "in_key", "10")

	for _, secretKey := range []string{"psk", "password", "private_key", "PSK", "private-key", "api_token", "unknown_opt", "", "ttl"} {
		if _, ok := published[secretKey]; ok {
			t.Fatalf("published interface option %q", secretKey)
		}
	}
	for _, sample := range published {
		if sample.value == "fixture-psk-do-not-use" || sample.value == "fixture-password-do-not-use" {
			t.Fatalf("published secret value for key %q", sample.key)
		}
	}
}

func TestInterfaceLabelPolicyFiltersStatusAndExternalIDs(t *testing.T) {
	status := publishAttributeKeys(t, interfaceAttributeStatus, map[string]string{
		"driver_name":      "veth",
		"firmware_version": "1.0",
		"psk":              "fixture-psk-do-not-use",
		"internal_state":   "secret-looking",
	})
	assertPublished(t, status, "driver_name", "veth")
	assertPublished(t, status, "firmware_version", "1.0")
	if _, ok := status["psk"]; ok {
		t.Fatal("published secret status key")
	}
	if _, ok := status["internal_state"]; ok {
		t.Fatal("published status key that is not on the allow-list")
	}

	externalIDs := publishAttributeKeys(t, interfaceAttributeExternalIDs, map[string]string{
		"iface-id":     "vm-1",
		"attached-mac": "02:00:00:00:00:01",
		"password":     "fixture-external-password",
		"tenant-token": "fixture-token",
	})
	assertPublished(t, externalIDs, "iface-id", "vm-1")
	assertPublished(t, externalIDs, "attached-mac", "02:00:00:00:00:01")
	if _, ok := externalIDs["password"]; ok {
		t.Fatal("published secret external id")
	}
	if _, ok := externalIDs["tenant-token"]; ok {
		t.Fatal("published external id that names a secret")
	}
}

func TestInterfaceLabelPolicyEmptyInput(t *testing.T) {
	published := publishAttributeKeys(t, interfaceAttributeOptions, nil)
	if len(published) != 0 {
		t.Fatalf("nil map published %d metrics", len(published))
	}
	published = publishAttributeKeys(t, interfaceAttributeOptions, map[string]string{})
	if len(published) != 0 {
		t.Fatalf("empty map published %d metrics", len(published))
	}
}

func TestListenAddressIsLoopback(t *testing.T) {
	loopback := []string{"127.0.0.1:9475", "localhost:9475", "[::1]:9475"}
	for _, address := range loopback {
		if !ListenAddressIsLoopback(address) {
			t.Fatalf("%s should be treated as loopback", address)
		}
	}

	exposed := []string{"", ":9475", "0.0.0.0:9475", "[::]:9475", "192.0.2.10:9475", "127.0.0.1"}
	for _, address := range exposed {
		if ListenAddressIsLoopback(address) {
			t.Fatalf("%s should not be treated as loopback", address)
		}
	}
}

func TestDefaultListenAddressIsLoopback(t *testing.T) {
	if DefaultListenAddress != "127.0.0.1:9475" {
		t.Fatalf("default listen address = %q", DefaultListenAddress)
	}
	if !ListenAddressIsLoopback(DefaultListenAddress) {
		t.Fatal("default listen address is not loopback")
	}
}

type publishedLabel struct {
	key   string
	value string
}

func publishAttributeKeys(t *testing.T, class string, entries map[string]string) map[string]publishedLabel {
	t.Helper()
	exporter := &Exporter{logger: log.NewNopLogger()}
	exporter.appendInterfaceMapMetric(descForAttribute(t, class), "system-1", "uuid-1", class, entries)

	published := make(map[string]publishedLabel, len(exporter.metrics))
	for _, metric := range exporter.metrics {
		key := metricLabel(t, metric, "key")
		published[key] = publishedLabel{key: key, value: metricLabel(t, metric, "value")}
	}
	return published
}

func descForAttribute(t *testing.T, class string) *prometheus.Desc {
	t.Helper()
	switch class {
	case interfaceAttributeOptions:
		return interfaceOptionsKeyValuePair
	case interfaceAttributeStatus:
		return interfaceStatusKeyValuePair
	case interfaceAttributeExternalIDs:
		return interfaceExternalIdKeyValuePair
	default:
		t.Fatalf("unknown attribute class %q", class)
		return nil
	}
}

func assertPublished(t *testing.T, published map[string]publishedLabel, key, value string) {
	t.Helper()
	sample, ok := published[key]
	if !ok {
		t.Fatalf("expected published key %q", key)
	}
	if sample.value != value {
		t.Fatalf("key %q value = %q, want %q", key, sample.value, value)
	}
}

func metricLabel(t *testing.T, metric prometheus.Metric, name string) string {
	t.Helper()
	var sample dto.Metric
	if err := metric.Write(&sample); err != nil {
		t.Fatalf("write metric: %v", err)
	}
	for _, label := range sample.Label {
		if label.GetName() == name {
			return label.GetValue()
		}
	}
	t.Fatalf("metric has no label %q", name)
	return ""
}
