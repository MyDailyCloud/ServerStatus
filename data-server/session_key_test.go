package main

import "testing"

func TestDeriveSessionKeyPrefersExisting(t *testing.T) {
	info := &SystemInfo{
		SessionID: "explicit-session",
		Hostname:  "host-a",
	}

	got := deriveSessionKey(info, "proj-1", "10.0.0.1:1234")
	if got != "explicit-session" {
		t.Fatalf("expected existing sessionID to be used, got %s", got)
	}
}

func TestDeriveSessionKeyDeterministic(t *testing.T) {
	base := &SystemInfo{
		Hostname: "host-a",
		OS: OSInfo{
			Platform:     "linux",
			Version:      "6.1",
			Architecture: "amd64",
		},
		CPU: CPUInfo{CoreCount: 4},
		Memory: MemInfo{
			Total: 8192,
		},
		Disk: DiskInfo{
			Total: 1024,
		},
	}

	id1 := deriveSessionKey(base, "proj-1", "10.0.0.1:1234")
	id2 := deriveSessionKey(base, "proj-1", "10.0.0.1:9999") // 端口不同但IP相同
	if id1 != id2 {
		t.Fatalf("expected deterministic session key, got %s and %s", id1, id2)
	}

	// 修改hostname或来源IP应产生不同ID
	hostB := *base
	hostB.Hostname = "host-b"
	id3 := deriveSessionKey(&hostB, "proj-1", "10.0.0.1:1234")
	if id3 == id1 {
		t.Fatalf("expected different session key for different hostname, got same %s", id3)
	}

	id4 := deriveSessionKey(base, "proj-1", "10.0.0.2:1234") // 不同来源IP
	if id4 == id1 {
		t.Fatalf("expected different session key for different source IP, got same %s", id4)
	}
}

