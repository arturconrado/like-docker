//go:build !linux

package minidock

func DetectCgroupSupport() CgroupSupport {
	return CgroupSupport{
		Supported: false,
		Version:   CgroupVersionNone,
		Notes:     []string{"Cgroups indisponíveis fora de hosts Linux."},
	}
}

func AttachWorkloadCgroup(_ string, _ int, _ CgroupLimits) (*WorkloadCgroup, error) {
	return nil, nil
}
