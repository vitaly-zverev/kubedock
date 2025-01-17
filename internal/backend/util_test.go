package backend

import (
	"testing"
)

func TestToKubernetesValue(t *testing.T) {
	tests := []struct {
		in    string
		key   string
		value string
		name  string
	}{
		{in: "__-abc", key: "abc", value: "abc", name: "abc"},
		{in: "/a/b/c", key: "a/b/c", value: "abc", name: "abc"},
		{
			in:    "StrategicMars",
			key:   "StrategicMars",
			value: "StrategicMars",
			name:  "StrategicMars",
		},
		{
			in:    "2107007e-b7c8-df23-18fb-6a6f79726578",
			key:   "2107007e-b7c8-df23-18fb-6a6f79726578",
			value: "2107007e-b7c8-df23-18fb-6a6f79726578",
			name:  "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:    "0123456789012345678901234567890123456789012345678901234567890123456789",
			key:   "012345678901234567890123456789012345678901234567890123456789012",
			value: "012345678901234567890123456789012345678901234567890123456789012",
			name:  "012345678901234567890123456789012345678901234567890123456789012",
		},
		{
			in:    "StrategicMars-",
			key:   "StrategicMars",
			value: "StrategicMars",
			name:  "StrategicMars",
		},
		{
			in:    "StrategicMars/-",
			key:   "StrategicMars",
			value: "StrategicMars",
			name:  "StrategicMars",
		},
		{
			in:    "2107007e-b7c8-df23-18fb-6a6f79726578",
			key:   "2107007e-b7c8-df23-18fb-6a6f79726578",
			value: "2107007e-b7c8-df23-18fb-6a6f79726578",
			name:  "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:    "app.kubernetes.io/name",
			key:   "app.kubernetes.io/name",
			value: "app.kubernetes.ioname",
			name:  "appkubernetesioname",
		},
		{
			in:    "",
			key:   "",
			value: "",
			name:  "undef",
		},
	}

	for i, tst := range tests {
		kub := &instance{}
		key := kub.toKubernetesKey(tst.in)
		if key != tst.key {
			t.Errorf("failed test %d - expected key %s, but got %s", i, tst.key, key)
		}
		value := kub.toKubernetesValue(tst.in)
		if value != tst.value {
			t.Errorf("failed test %d - expected value %s, but got %s", i, tst.value, value)
		}
		name := kub.toKubernetesName(tst.in)
		if name != tst.name {
			t.Errorf("failed test %d - expected name %s, but got %s", i, tst.name, name)
		}
	}
}

func TestRandomPort(t *testing.T) {
	m := map[int]int{}
	kub := &instance{
		randomPorts: map[int]int{},
	}
	for i := 0; i < 100; i++ {
		p := kub.RandomPort()
		if p < 1024 {
			t.Errorf("Invalid random port %d", p)
			break
		}
		if _, ok := m[p]; ok {
			t.Errorf("Random port collision, port %d already provided", p)
			break
		}
		m[p] = p
	}
}
