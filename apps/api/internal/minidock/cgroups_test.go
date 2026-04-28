package minidock

import "testing"

func TestLoadCgroupLimitsFromEnvDefaults(t *testing.T) {
	t.Setenv("MINIDOCK_CGROUP_PIDS_MAX", "")
	t.Setenv("MINIDOCK_CGROUP_MEMORY_MAX", "")
	t.Setenv("MINIDOCK_CGROUP_CPU_MAX", "")

	limits := LoadCgroupLimitsFromEnv()
	if limits.PidsMax != "256" {
		t.Fatalf("pids default esperado 256, recebido %s", limits.PidsMax)
	}
	if limits.MemoryMax != "1073741824" {
		t.Fatalf("memory default esperado 1073741824, recebido %s", limits.MemoryMax)
	}
	if limits.CPUMax != "200000 100000" {
		t.Fatalf("cpu default esperado \"200000 100000\", recebido %s", limits.CPUMax)
	}
}

func TestLoadCgroupLimitsFromEnvOverrides(t *testing.T) {
	t.Setenv("MINIDOCK_CGROUP_PIDS_MAX", "128")
	t.Setenv("MINIDOCK_CGROUP_MEMORY_MAX", "536870912")
	t.Setenv("MINIDOCK_CGROUP_CPU_MAX", "100000 100000")

	limits := LoadCgroupLimitsFromEnv()
	if limits.PidsMax != "128" {
		t.Fatalf("pids override esperado 128, recebido %s", limits.PidsMax)
	}
	if limits.MemoryMax != "536870912" {
		t.Fatalf("memory override esperado 536870912, recebido %s", limits.MemoryMax)
	}
	if limits.CPUMax != "100000 100000" {
		t.Fatalf("cpu override esperado \"100000 100000\", recebido %s", limits.CPUMax)
	}
}

func TestDetectCgroupSupportHasVersion(t *testing.T) {
	support := DetectCgroupSupport()
	if support.Version == "" {
		t.Fatal("version de cgroup não pode ser vazia")
	}
}
