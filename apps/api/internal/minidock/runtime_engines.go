package minidock

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type RuntimeInspect struct {
	Engine            string
	Isolated          bool
	Rootfs            string
	ContainerHostname string
	MainPID           int
	Port              int
	DataDir           string
	ReadinessState    string
	ModeUsed          string
}

type RuntimeHandle struct {
	WorkloadID        string
	Command           string
	Args              []string
	Rootfs            string
	ContainerHostname string
	MainPID           int
	Port              int
	DataDir           string
	ReadinessState    string
	ModeUsed          string
	OwnDataDir        bool
}

type RuntimeUpdate struct {
	Status          *WorkloadStatus
	Mode            *RuntimeMode
	Engine          *string
	Isolated        *bool
	Port            *int
	DataDir         *string
	ReadinessState  *string
	ModeUsed        *string
	FallbackApplied *bool
	FallbackReason  *string
}

type RuntimeHooks struct {
	OnLog           func(line string)
	OnMainPID       func(pid int)
	OnStatus        func(status WorkloadStatus)
	OnRuntimeUpdate func(update RuntimeUpdate)
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
	handle := &RuntimeHandle{
		WorkloadID: workload.ID,
		Command:    workload.Command,
		Args:       cloneStringSlice(workload.Args),
	}
	if workload.Command == "minidock-postgres-demo" {
		handle.Port = workload.Runtime.Port
		if handle.Port <= 0 {
			handle.Port = 55432
		}
		handle.DataDir = workload.Runtime.DataDir
		if strings.TrimSpace(handle.DataDir) == "" {
			handle.DataDir = "/tmp/minidock-postgres-demo-simulated"
		}
		handle.ReadinessState = workload.Runtime.ReadinessState
		handle.ModeUsed = string(PostgresModeDemo)
	}
	return handle, nil
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
		duration = 45 * time.Second
		mode := string(PostgresModeDemo)
		safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
			Status:         ptrWorkloadStatus(StatusPreparing),
			Mode:           ptrRuntimeMode(ModeDemo),
			Engine:         ptrString("demo-engine"),
			Isolated:       ptrBool(false),
			Port:           ptrInt(handle.Port),
			DataDir:        ptrString(handle.DataDir),
			ReadinessState: ptrString("preparing"),
			ModeUsed:       &mode,
		})
		safeStatus(hooks.OnStatus, StatusPreparing)
	}

	for _, line := range demoOutputFor(normalized) {
		safeLog(hooks.OnLog, line)
	}

	if normalized == "minidock-postgres-demo" {
		mode := string(PostgresModeDemo)
		safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
			Status:         ptrWorkloadStatus(StatusRunning),
			Mode:           ptrRuntimeMode(ModeDemo),
			Engine:         ptrString("demo-engine"),
			Isolated:       ptrBool(false),
			Port:           ptrInt(handle.Port),
			DataDir:        ptrString(handle.DataDir),
			ReadinessState: ptrString("ready"),
			ModeUsed:       &mode,
		})
		safeStatus(hooks.OnStatus, StatusRunning)
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

func (e *DemoEngine) Inspect(handle *RuntimeHandle) RuntimeInspect {
	if handle == nil {
		return RuntimeInspect{
			Engine:   "demo-engine",
			Isolated: false,
		}
	}
	return RuntimeInspect{
		Engine:         "demo-engine",
		Isolated:       false,
		Port:           handle.Port,
		DataDir:        handle.DataDir,
		ReadinessState: handle.ReadinessState,
		ModeUsed:       handle.ModeUsed,
	}
}

type LocalProcessEngine struct{}

func (e *LocalProcessEngine) Mode() RuntimeMode {
	return ModeProcessLocal
}

func (e *LocalProcessEngine) Create(workload Workload) (*RuntimeHandle, error) {
	handle := &RuntimeHandle{
		WorkloadID: workload.ID,
		Command:    workload.Command,
		Args:       cloneStringSlice(workload.Args),
	}
	if workload.Command == "minidock-postgres-demo" {
		dataDir := strings.TrimSpace(workload.Runtime.DataDir)
		if dataDir == "" {
			dir, err := os.MkdirTemp("", "minidock-postgres-*")
			if err != nil {
				return nil, fmt.Errorf("falha ao criar diretório temporário para PostgreSQL: %w", err)
			}
			dataDir = dir
			handle.OwnDataDir = true
		}
		port := workload.Runtime.Port
		if port <= 0 {
			detectedPort, err := findAvailableTCPPort()
			if err != nil {
				if handle.OwnDataDir {
					_ = os.RemoveAll(dataDir)
				}
				return nil, fmt.Errorf("falha ao reservar porta local para PostgreSQL: %w", err)
			}
			port = detectedPort
		}
		handle.DataDir = dataDir
		handle.Port = port
		handle.ReadinessState = workload.Runtime.ReadinessState
		handle.ModeUsed = workload.Runtime.ModeUsed
	}
	return handle, nil
}

func (e *LocalProcessEngine) Start(ctx context.Context, handle *RuntimeHandle, hooks RuntimeHooks) RuntimeExecutionResult {
	if handle.Command == "minidock-postgres-demo" {
		return e.startPostgresDemo(ctx, handle, hooks)
	}

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

func (e *LocalProcessEngine) Stop(handle *RuntimeHandle) error {
	if handle == nil || handle.MainPID <= 0 {
		return nil
	}
	if err := syscall.Kill(handle.MainPID, syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}

func (e *LocalProcessEngine) Remove(handle *RuntimeHandle) error {
	if handle != nil && handle.MainPID > 0 {
		_ = e.Stop(handle)
	}
	if handle != nil && handle.OwnDataDir && strings.TrimSpace(handle.DataDir) != "" {
		return os.RemoveAll(handle.DataDir)
	}
	return nil
}

func (e *LocalProcessEngine) Logs(_ *RuntimeHandle) []string {
	return []string{}
}

func (e *LocalProcessEngine) Inspect(handle *RuntimeHandle) RuntimeInspect {
	if handle == nil {
		return RuntimeInspect{
			Engine:   "local-process-engine",
			Isolated: false,
		}
	}
	return RuntimeInspect{
		Engine:         "local-process-engine",
		Isolated:       false,
		MainPID:        handle.MainPID,
		Port:           handle.Port,
		DataDir:        handle.DataDir,
		ReadinessState: handle.ReadinessState,
		ModeUsed:       handle.ModeUsed,
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

func safeStatus(onStatus func(WorkloadStatus), status WorkloadStatus) {
	if onStatus != nil {
		onStatus(status)
	}
}

func safeRuntimeUpdate(onUpdate func(RuntimeUpdate), update RuntimeUpdate) {
	if onUpdate != nil {
		onUpdate(update)
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
			"[stdout] [postgres-demo] verificação do ambiente: fallback demo ativo",
			"[stdout] [postgres-demo] preparing data directory /tmp/minidock-postgres-demo-simulated",
			"[stdout] [postgres-demo] initdb concluído com sucesso (demo)",
			"[stdout] [postgres-demo] postgres iniciado em modo demonstrativo",
			"[stdout] [postgres-demo] Serviço pronto para aceitar conexões (demo)",
			"[stdout] [postgres-demo] readiness: ready",
		}
	default:
		return []string{fmt.Sprintf("[demo] comando: %s", normalized)}
	}
}
