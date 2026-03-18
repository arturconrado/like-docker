package minidock

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type RuntimeInspect struct {
	Engine            string
	Isolated          bool
	Rootfs            string
	ContainerHostname string
	MainPID           int
}

type RuntimeHandle struct {
	WorkloadID        string
	Command           string
	Args              []string
	Rootfs            string
	ContainerHostname string
	MainPID           int
}

type RuntimeHooks struct {
	OnLog     func(line string)
	OnMainPID func(pid int)
}

type RuntimeExecutionResult struct {
	Status      WorkloadStatus
	ExitCode    int
	ExtraLog    string
	EventType   string
	EventLabel  string
	EventLevel  EventSeverity
	FinishedErr error
}

type RuntimeEngine interface {
	Mode() RuntimeMode
	Create(workload Workload) (*RuntimeHandle, error)
	Start(ctx context.Context, handle *RuntimeHandle, hooks RuntimeHooks) RuntimeExecutionResult
	Stop(handle *RuntimeHandle) error
	Remove(handle *RuntimeHandle) error
	Logs(handle *RuntimeHandle) []string
	Inspect(handle *RuntimeHandle) RuntimeInspect
}

type DemoEngine struct{}

func (e *DemoEngine) Mode() RuntimeMode {
	return ModeDemo
}

func (e *DemoEngine) Create(workload Workload) (*RuntimeHandle, error) {
	return &RuntimeHandle{
		WorkloadID: workload.ID,
		Command:    workload.Command,
		Args:       cloneStringSlice(workload.Args),
	}, nil
}

func (e *DemoEngine) Start(ctx context.Context, handle *RuntimeHandle, hooks RuntimeHooks) RuntimeExecutionResult {
	normalized := NormalizeCommand(handle.Command, handle.Args)
	safeLog(hooks.OnLog, "[demo] execução simulada iniciada")

	duration := 700 * time.Millisecond
	if strings.Contains(normalized, "sleep 10") {
		duration = 4 * time.Second
	}
	if strings.Contains(normalized, "sleep 60") {
		duration = 6 * time.Second
	}
	if normalized == "minidock-postgres-demo" {
		duration = 8 * time.Second
	}

	for _, line := range demoOutputFor(normalized) {
		safeLog(hooks.OnLog, line)
	}

	select {
	case <-ctx.Done():
		return RuntimeExecutionResult{
			Status:     StatusStopped,
			ExitCode:   130,
			ExtraLog:   "Execução cancelada pelo usuário.",
			EventType:  "workload_stopped",
			EventLabel: "Workload interrompida",
			EventLevel: SeverityWarn,
		}
	case <-time.After(duration):
	}

	safeLog(hooks.OnLog, "[demo] rotina concluída com sucesso")
	return RuntimeExecutionResult{
		Status:     StatusCompleted,
		ExitCode:   0,
		EventType:  "workload_completed",
		EventLabel: "Workload concluída",
		EventLevel: SeverityInfo,
	}
}

func (e *DemoEngine) Stop(_ *RuntimeHandle) error {
	return nil
}

func (e *DemoEngine) Remove(_ *RuntimeHandle) error {
	return nil
}

func (e *DemoEngine) Logs(_ *RuntimeHandle) []string {
	return []string{}
}

func (e *DemoEngine) Inspect(_ *RuntimeHandle) RuntimeInspect {
	return RuntimeInspect{
		Engine:   "demo-engine",
		Isolated: false,
	}
}

type LocalProcessEngine struct{}

func (e *LocalProcessEngine) Mode() RuntimeMode {
	return ModeProcessLocal
}

func (e *LocalProcessEngine) Create(workload Workload) (*RuntimeHandle, error) {
	return &RuntimeHandle{
		WorkloadID: workload.ID,
		Command:    workload.Command,
		Args:       cloneStringSlice(workload.Args),
	}, nil
}

func (e *LocalProcessEngine) Start(ctx context.Context, handle *RuntimeHandle, hooks RuntimeHooks) RuntimeExecutionResult {
	command, args, preface := effectiveCommand(handle.Command, handle.Args)
	for _, line := range preface {
		safeLog(hooks.OnLog, line)
	}

	cmd := exec.CommandContext(ctx, command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return RuntimeExecutionResult{
			Status:      StatusFailed,
			ExitCode:    1,
			ExtraLog:    fmt.Sprintf("Falha ao criar pipe stdout: %v", err),
			EventType:   "workload_failed",
			EventLabel:  "Workload falhou",
			EventLevel:  SeverityError,
			FinishedErr: err,
		}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return RuntimeExecutionResult{
			Status:      StatusFailed,
			ExitCode:    1,
			ExtraLog:    fmt.Sprintf("Falha ao criar pipe stderr: %v", err),
			EventType:   "workload_failed",
			EventLabel:  "Workload falhou",
			EventLevel:  SeverityError,
			FinishedErr: err,
		}
	}

	if err := cmd.Start(); err != nil {
		return RuntimeExecutionResult{
			Status:      StatusFailed,
			ExitCode:    1,
			ExtraLog:    fmt.Sprintf("Falha ao iniciar processo: %v", err),
			EventType:   "workload_failed",
			EventLabel:  "Workload falhou",
			EventLevel:  SeverityError,
			FinishedErr: err,
		}
	}
	if cmd.Process != nil {
		handle.MainPID = cmd.Process.Pid
		safePID(hooks.OnMainPID, cmd.Process.Pid)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go collectPipe(&wg, stdout, "stdout", hooks.OnLog)
	go collectPipe(&wg, stderr, "stderr", hooks.OnLog)

	waitErr := cmd.Wait()
	wg.Wait()

	if ctx.Err() == context.Canceled {
		return RuntimeExecutionResult{
			Status:     StatusStopped,
			ExitCode:   130,
			ExtraLog:   "Execução cancelada pelo usuário.",
			EventType:  "workload_stopped",
			EventLabel: "Workload interrompida",
			EventLevel: SeverityWarn,
		}
	}

	if waitErr != nil {
		exitCode := extractExitCode(waitErr)
		return RuntimeExecutionResult{
			Status:      StatusFailed,
			ExitCode:    exitCode,
			ExtraLog:    fmt.Sprintf("Processo finalizou com erro: %v", waitErr),
			EventType:   "workload_failed",
			EventLabel:  "Workload falhou",
			EventLevel:  SeverityError,
			FinishedErr: waitErr,
		}
	}

	return RuntimeExecutionResult{
		Status:     StatusCompleted,
		ExitCode:   0,
		EventType:  "workload_completed",
		EventLabel: "Workload concluída",
		EventLevel: SeverityInfo,
	}
}

func (e *LocalProcessEngine) Stop(_ *RuntimeHandle) error {
	return nil
}

func (e *LocalProcessEngine) Remove(_ *RuntimeHandle) error {
	return nil
}

func (e *LocalProcessEngine) Logs(_ *RuntimeHandle) []string {
	return []string{}
}

func (e *LocalProcessEngine) Inspect(_ *RuntimeHandle) RuntimeInspect {
	return RuntimeInspect{
		Engine:   "local-process-engine",
		Isolated: false,
	}
}

func collectPipe(wg *sync.WaitGroup, reader io.Reader, stream string, onLog func(string)) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, 0, 1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		safeLog(onLog, fmt.Sprintf("[%s] %s", stream, scanner.Text()))
	}
}

func extractExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func safeLog(onLog func(string), line string) {
	if onLog != nil {
		onLog(line)
	}
}

func safePID(onPID func(int), pid int) {
	if onPID != nil {
		onPID(pid)
	}
}

func demoOutputFor(normalized string) []string {
	switch normalized {
	case "/bin/sh -c hostname && pwd && ls /":
		return []string{
			"[stdout] mdk-demo",
			"[stdout] /",
			"[stdout] bin",
			"[stdout] proc",
			"[stdout] tmp",
		}
	case "/bin/sh -c echo hello from container":
		return []string{"[stdout] hello from container"}
	case "/bin/sh -c echo starting && sleep 10 && echo finished":
		return []string{"[stdout] starting", "[stdout] finished"}
	case "/bin/sh -c uname -a && id && ps":
		return []string{"[stdout] Linux minidock-demo 6.x", "[stdout] uid=0(root) gid=0(root)", "[stdout] PID TTY TIME CMD"}
	case "/bin/sh -c hostname && uname -a && ps":
		return []string{"[stdout] mdk-demo", "[stdout] Linux minidock-demo 6.x", "[stdout] PID TTY TIME CMD"}
	case "/bin/sh -c ls -la / && ls -la /bin":
		return []string{"[stdout] total 16", "[stdout] drwxr-xr-x  3 root root 4096 /", "[stdout] -rwxr-xr-x  1 root root busybox"}
	case "/bin/sh -c echo fallback validation":
		return []string{"[stdout] fallback validation"}
	case "minidock-postgres-demo":
		return []string{
			"[stdout] [postgres-demo] preparing data directory /tmp/minidock-postgres-demo",
			"[stdout] [postgres-demo] PostgreSQL inicializado com sucesso (simulado)",
			"[stdout] [postgres-demo] Banco pronto para aceitar conexões (simulado)",
			"[stdout] [postgres-demo] readiness: ready",
		}
	default:
		return []string{fmt.Sprintf("[demo] comando: %s", normalized)}
	}
}
