package minidock

import (
	"os"
	"strings"
)

const (
	CgroupVersionNone = "none"
	CgroupVersionV1   = "v1"
	CgroupVersionV2   = "v2"
)

type CgroupSupport struct {
	Supported bool
	Version   string
	Notes     []string
}

type CgroupLimits struct {
	PidsMax   string
	MemoryMax string
	CPUMax    string
}

type WorkloadCgroup struct {
	Path      string
	Version   string
	cleanupFn func() error
}

func (c *WorkloadCgroup) Cleanup() error {
	if c == nil || c.cleanupFn == nil {
		return nil
	}
	return c.cleanupFn()
}

func LoadCgroupLimitsFromEnv() CgroupLimits {
	return CgroupLimits{
		PidsMax:   readEnvWithDefault("MINIDOCK_CGROUP_PIDS_MAX", "256"),
		MemoryMax: readEnvWithDefault("MINIDOCK_CGROUP_MEMORY_MAX", "1073741824"),
		CPUMax:    readEnvWithDefault("MINIDOCK_CGROUP_CPU_MAX", "200000 100000"),
	}
}

func readEnvWithDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
