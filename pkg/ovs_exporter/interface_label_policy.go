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
	"strings"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	interfaceAttributeOptions     = "options"
	interfaceAttributeStatus      = "status"
	interfaceAttributeExternalIDs = "external_ids"
)

// secretKeyParts are whole underscore-separated words that name credential
// material. Tunnel identifiers such as "key" and "in_key" are not in this set.
var secretKeyParts = map[string]struct{}{
	"psk":         {},
	"password":    {},
	"passwd":      {},
	"passphrase":  {},
	"secret":      {},
	"token":       {},
	"credential":  {},
	"credentials": {},
	"privkey":     {},
}

// allowedInterfaceAttributeKeys is the published set for each interface map.
// A key that is absent here is omitted. A secret key is omitted even if it is
// later added by mistake.
var allowedInterfaceAttributeKeys = map[string]map[string]struct{}{
	interfaceAttributeOptions: {
		"remote_ip":       {},
		"local_ip":        {},
		"in_key":          {},
		"out_key":         {},
		"key":             {},
		"tos":             {},
		"ttl":             {},
		"df_default":      {},
		"csum":            {},
		"packet_type":     {},
		"dst_port":        {},
		"egress_pkt_mark": {},
		"erspan_idx":      {},
		"erspan_ver":      {},
		"erspan_dir":      {},
		"erspan_hwid":     {},
		"peer":            {},
		"payload_type":    {},
		"remote_name":     {},
	},
	interfaceAttributeStatus: {
		"driver_name":                 {},
		"driver_version":              {},
		"firmware_version":            {},
		"source":                      {},
		"bus_info":                    {},
		"tunnel_egress_iface":         {},
		"tunnel_egress_iface_carrier": {},
		"pci_vendor_id":               {},
		"pci_device_id":               {},
		"port_no":                     {},
		"numa_id":                     {},
	},
	interfaceAttributeExternalIDs: {
		"iface-id":         {},
		"iface-status":     {},
		"attached-mac":     {},
		"vm-id":            {},
		"ovn-installed":    {},
		"ovn-installed-ts": {},
	},
}

// interfaceLabelDecision reports whether an interface map entry may become a
// Prometheus label. The value is never logged: it may be a secret.
func interfaceLabelDecision(class, key, value string) (bool, string) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
		return false, "empty key or value"
	}
	if isSecretInterfaceKey(key) {
		return false, "key names a secret"
	}
	allowed, ok := allowedInterfaceAttributeKeys[class]
	if !ok {
		return false, "unknown attribute class"
	}
	if _, ok := allowed[strings.ToLower(strings.TrimSpace(key))]; !ok {
		return false, "key is not on the allow-list"
	}
	return true, ""
}

func isSecretInterfaceKey(key string) bool {
	normalised := strings.ToLower(strings.TrimSpace(key))
	normalised = strings.ReplaceAll(normalised, "-", "_")
	if strings.Contains(normalised, "private_key") || strings.Contains(normalised, "pre_shared_key") {
		return true
	}
	for _, part := range strings.Split(normalised, "_") {
		if _, ok := secretKeyParts[part]; ok {
			return true
		}
	}
	return false
}

// appendInterfaceMapMetric publishes one sample per allow-listed entry.
func (e *Exporter) appendInterfaceMapMetric(desc *prometheus.Desc, systemID, uuid, class string, entries map[string]string) {
	for key, value := range entries {
		permitted, reason := interfaceLabelDecision(class, key, value)
		if !permitted {
			level.Debug(e.logger).Log(
				"msg", "omitted interface attribute from metrics",
				"system_id", systemID,
				"class", class,
				"key", key,
				"reason", reason,
			)
			continue
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			1,
			systemID,
			uuid,
			key,
			value,
		))
	}
}
