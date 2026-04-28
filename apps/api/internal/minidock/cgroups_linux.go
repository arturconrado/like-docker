//go:build linux

package minidock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const cgroupRoot = "/sys/fs/cgroup"

func DetectCgroupSupport() CgroupSupport {
	if !directoryExists(cgroupRoot) {
		return CgroupSupport{
			Supported: false,
			Version:   CgroupVersionNone,
			Notes:     []string{"Filesystem de cgroup não encontrado em /sys/fs/cgroup."},
		}
	}

	if fileExists(filepath.Join(cgroupRoot, "cgroup.controllers")) {
		return CgroupSupport{
			Supported: true,
			Version:   CgroupVersionV2,
			Notes:     []string{"Cgroup v2 detectado no host Linux."},
		}
	}

	controllers := availableV1Controllers()
	if len(controllers) > 0 {
		return CgroupSupport{
			Supported: true,
			Version:   CgroupVersionV1,
			Notes: []string{
				fmt.Sprintf("Cgroup v1 detectado com controladores: %s.", strings.Join(controllers, ", ")),
			},
		}
	}

	return CgroupSupport{
		Supported: false,
		Version:   CgroupVersionNone,
		Notes:     []string{"Cgroups não disponíveis para uso no host atual."},
	}
}

func AttachWorkloadCgroup(workloadID string, pid int, limits CgroupLimits) (*WorkloadCgroup, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("pid inválido para cgroup: %d", pid)
	}

	support := DetectCgroupSupport()
	switch support.Version {
	case CgroupVersionV2:
		return attachCgroupV2(workloadID, pid, limits)
	case CgroupVersionV1:
		return attachCgroupV1(workloadID, pid, limits)
	default:
		return nil, nil
	}
}

func attachCgroupV2(workloadID string, pid int, limits CgroupLimits) (*WorkloadCgroup, error) {
	groupName := sanitizeCgroupGroupName(workloadID)
	basePath := filepath.Join(cgroupRoot, "minidock")
	groupPath := filepath.Join(basePath, groupName)

	if err := os.MkdirAll(groupPath, 0o755); err != nil {
		return nil, fmt.Errorf("falha ao criar cgroup v2 em %s: %w", groupPath, err)
	}

	if err := writeCgroupLimit(filepath.Join(groupPath, "pids.max"), limits.PidsMax); err != nil {
		return nil, err
	}
	if err := writeCgroupLimit(filepath.Join(groupPath, "memory.max"), limits.MemoryMax); err != nil {
		return nil, err
	}
	if err := writeCgroupLimit(filepath.Join(groupPath, "cpu.max"), normalizeCPUMax(limits.CPUMax)); err != nil {
		return nil, err
	}

	if err := addPIDToCgroup(groupPath, pid); err != nil {
		return nil, err
	}

	return &WorkloadCgroup{
		Path:    groupPath,
		Version: CgroupVersionV2,
		cleanupFn: func() error {
			return os.Remove(groupPath)
		},
	}, nil
}

func attachCgroupV1(workloadID string, pid int, limits CgroupLimits) (*WorkloadCgroup, error) {
	groupName := sanitizeCgroupGroupName(workloadID)
	createdPaths := make([]string, 0, 3)
	cleanupPaths := make([]string, 0, 3)

	ensureController := func(controller string) (string, bool, error) {
		controllerRoot := filepath.Join(cgroupRoot, controller)
		if !directoryExists(controllerRoot) {
			return "", false, nil
		}
		groupPath := filepath.Join(controllerRoot, "minidock", groupName)
		if err := os.MkdirAll(groupPath, 0o755); err != nil {
			return "", false, err
		}
		createdPaths = append(createdPaths, groupPath)
		cleanupPaths = append(cleanupPaths, groupPath)
		return groupPath, true, nil
	}

	pidsPath, hasPids, err := ensureController("pids")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cgroup pids v1: %w", err)
	}
	if hasPids {
		if err := writeCgroupLimit(filepath.Join(pidsPath, "pids.max"), limits.PidsMax); err != nil {
			return nil, err
		}
		if err := addPIDToCgroup(pidsPath, pid); err != nil {
			return nil, err
		}
	}

	memoryPath, hasMemory, err := ensureController("memory")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cgroup memory v1: %w", err)
	}
	if hasMemory {
		if err := writeCgroupLimit(filepath.Join(memoryPath, "memory.limit_in_bytes"), limits.MemoryMax); err != nil {
			return nil, err
		}
		if err := addPIDToCgroup(memoryPath, pid); err != nil {
			return nil, err
		}
	}

	cpuPath, hasCPU, err := ensureController("cpu")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cgroup cpu v1: %w", err)
	}
	if hasCPU {
		quota, period := parseCPUMaxForV1(limits.CPUMax)
		if err := writeCgroupLimit(filepath.Join(cpuPath, "cpu.cfs_quota_us"), quota); err != nil {
			return nil, err
		}
		if err := writeCgroupLimit(filepath.Join(cpuPath, "cpu.cfs_period_us"), period); err != nil {
			return nil, err
		}
		if err := addPIDToCgroup(cpuPath, pid); err != nil {
			return nil, err
		}
	}

	if len(createdPaths) == 0 {
		return nil, errors.New("nenhum controlador cgroup v1 disponível para minidock")
	}

	return &WorkloadCgroup{
		Path:    strings.Join(createdPaths, " | "),
		Version: CgroupVersionV1,
		cleanupFn: func() error {
			for _, path := range cleanupPaths {
				_ = os.Remove(path)
			}
			return nil
		},
	}, nil
}

func availableV1Controllers() []string {
	controllers := make([]string, 0, 3)
	for _, controller := range []string{"pids", "memory", "cpu"} {
		if directoryExists(filepath.Join(cgroupRoot, controller)) {
			controllers = append(controllers, controller)
		}
	}
	return controllers
}

func sanitizeCgroupGroupName(workloadID string) string {
	trimmed := strings.TrimSpace(workloadID)
	if trimmed == "" {
		return "wk-unknown"
	}
	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('-')
	}
	value := b.String()
	if value == "" {
		return "wk-unknown"
	}
	return value
}

func writeCgroupLimit(path, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("falha ao validar arquivo de cgroup %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(trimmed), 0o644); err != nil {
		return fmt.Errorf("falha ao gravar limite de cgroup em %s: %w", path, err)
	}
	return nil
}

func addPIDToCgroup(path string, pid int) error {
	pidValue := strconv.Itoa(pid)
	paths := []string{
		filepath.Join(path, "cgroup.procs"),
		filepath.Join(path, "tasks"),
	}
	for _, candidate := range paths {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		if err := os.WriteFile(candidate, []byte(pidValue), 0o644); err != nil {
			return fmt.Errorf("falha ao associar pid %d ao cgroup em %s: %w", pid, candidate, err)
		}
		return nil
	}
	return fmt.Errorf("arquivo cgroup.procs/tasks não encontrado para cgroup %s", path)
}

func normalizeCPUMax(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "200000 100000"
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 1 {
		return parts[0] + " 100000"
	}
	return parts[0] + " " + parts[1]
}

func parseCPUMaxForV1(value string) (string, string) {
	normalized := normalizeCPUMax(value)
	parts := strings.Fields(normalized)
	if len(parts) < 2 {
		return "200000", "100000"
	}
	return parts[0], parts[1]
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
