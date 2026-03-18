//go:build !linux

package minidock

import (
	"context"
	"fmt"
)

type LinuxContainerEngine struct {
	rootfsPath string
}

func NewLinuxContainerEngine(rootfsPath string) *LinuxContainerEngine {
	return &LinuxContainerEngine{rootfsPath: rootfsPath}
}

func (e *LinuxContainerEngine) Mode() RuntimeMode {
	return ModeContainerLinux
}

func (e *LinuxContainerEngine) Create(workload Workload) (*RuntimeHandle, error) {
	return &RuntimeHandle{
		WorkloadID: workload.ID,
		Command:    workload.Command,
		Args:       cloneStringSlice(workload.Args),
		Rootfs:     e.rootfsPath,
	}, nil
}

func (e *LinuxContainerEngine) Start(_ context.Context, _ *RuntimeHandle, _ RuntimeHooks) RuntimeExecutionResult {
	return RuntimeExecutionResult{
		Status:     StatusFailed,
		ExitCode:   1,
		ExtraLog:   "container-linux indisponível fora de hosts Linux.",
		EventType:  "workload_failed",
		EventLabel: "Workload falhou",
		EventLevel: SeverityError,
		FinishedErr: fmt.Errorf(
			"container-linux requer Linux",
		),
	}
}

func (e *LinuxContainerEngine) Stop(_ *RuntimeHandle) error {
	return nil
}

func (e *LinuxContainerEngine) Remove(_ *RuntimeHandle) error {
	return nil
}

func (e *LinuxContainerEngine) Logs(_ *RuntimeHandle) []string {
	return []string{}
}

func (e *LinuxContainerEngine) Inspect(handle *RuntimeHandle) RuntimeInspect {
	return RuntimeInspect{
		Engine:   "linux-container-engine-stub",
		Isolated: false,
		Rootfs:   handle.Rootfs,
	}
}
