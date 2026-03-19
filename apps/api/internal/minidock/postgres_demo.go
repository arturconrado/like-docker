package minidock

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const postgresDemoReadinessTimeout = 20 * time.Second

func (e *LocalProcessEngine) startPostgresDemo(ctx context.Context, handle *RuntimeHandle, hooks RuntimeHooks) RuntimeExecutionResult {
	caps := DetectHostCapabilities(strings.TrimSpace(os.Getenv("MINIDOCK_CONTAINER_ROOTFS")))
	realMode := string(PostgresModeProcessLocalReal)

	safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
		Status:         ptrWorkloadStatus(StatusPreparing),
		Port:           ptrInt(handle.Port),
		DataDir:        ptrString(handle.DataDir),
		ReadinessState: ptrString("preparing"),
		ModeUsed:       &realMode,
	})
	safeStatus(hooks.OnStatus, StatusPreparing)
	safeLog(hooks.OnLog, fmt.Sprintf("[postgres-demo] verificação do ambiente: os=%s linux=%t binarios=%t", caps.OS, caps.IsLinux, caps.PostgresBinariesAvailable))
	safeLog(hooks.OnLog, fmt.Sprintf("[postgres-demo] preparing data directory %s", handle.DataDir))

	if !caps.CanRunPostgresDemo {
		return e.runPostgresDemoFallback(ctx, handle, hooks, postgresFallbackReasonForCapabilities(caps))
	}

	if err := os.MkdirAll(handle.DataDir, 0o700); err != nil {
		return e.runPostgresDemoFallback(ctx, handle, hooks, fmt.Sprintf("[postgres-demo] falha ao preparar diretório de dados: %v. Fallback para modo demonstrativo.", err))
	}

	if !fileExists(filepath.Join(handle.DataDir, "PG_VERSION")) {
		safeLog(hooks.OnLog, "[postgres-demo] initializing PostgreSQL cluster with initdb")
		if err := runLoggedCommand(ctx, caps.PostgresBinaryPaths.Initdb, []string{"-D", handle.DataDir, "-A", "trust"}, hooks.OnLog); err != nil {
			if ctx.Err() == context.Canceled {
				return stoppedRuntimeResult("Execução cancelada pelo usuário.")
			}
			return e.runPostgresDemoFallback(ctx, handle, hooks, fmt.Sprintf("[postgres-demo] initdb falhou: %v. Fallback para modo demonstrativo.", err))
		}
		safeLog(hooks.OnLog, "[postgres-demo] initdb concluído com sucesso")
	} else {
		safeLog(hooks.OnLog, "[postgres-demo] cluster existente detectado; initdb reaproveitado")
	}

	safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
		Status:         ptrWorkloadStatus(StatusStarting),
		ReadinessState: ptrString("starting"),
		ModeUsed:       &realMode,
	})
	safeStatus(hooks.OnStatus, StatusStarting)

	args := []string{
		"-D", handle.DataDir,
		"-p", strconv.Itoa(handle.Port),
		"-c", "listen_addresses=127.0.0.1",
		"-c", "unix_socket_directories=" + handle.DataDir,
	}
	safeLog(hooks.OnLog, fmt.Sprintf("[postgres-demo] starting postgres on 127.0.0.1:%d", handle.Port))

	cmd := exec.Command(caps.PostgresBinaryPaths.Postgres, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return e.runPostgresDemoFallback(ctx, handle, hooks, fmt.Sprintf("[postgres-demo] falha ao criar pipe stdout: %v. Fallback para modo demonstrativo.", err))
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return e.runPostgresDemoFallback(ctx, handle, hooks, fmt.Sprintf("[postgres-demo] falha ao criar pipe stderr: %v. Fallback para modo demonstrativo.", err))
	}
	if err := cmd.Start(); err != nil {
		return e.runPostgresDemoFallback(ctx, handle, hooks, fmt.Sprintf("[postgres-demo] falha ao iniciar postgres: %v. Fallback para modo demonstrativo.", err))
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go collectPipe(&wg, stdout, "stdout", hooks.OnLog)
	go collectPipe(&wg, stderr, "stderr", hooks.OnLog)

	if cmd.Process != nil {
		handle.MainPID = cmd.Process.Pid
		safePID(hooks.OnMainPID, cmd.Process.Pid)
	}

	stopDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = terminateProcessGracefully(handle.MainPID, 5*time.Second)
		case <-stopDone:
		}
	}()

	if err := waitForPostgresReadiness(ctx, caps.PostgresBinaryPaths.PGIsReady, handle.Port); err != nil {
		close(stopDone)
		_ = terminateProcessGracefully(handle.MainPID, 5*time.Second)
		waitErr := cmd.Wait()
		wg.Wait()
		if ctx.Err() == context.Canceled {
			return stoppedRuntimeResult("Execução cancelada pelo usuário.")
		}
		if waitErr != nil {
			safeLog(hooks.OnLog, fmt.Sprintf("[postgres-demo] postgres encerrou durante readiness: %v", waitErr))
		}
		return e.runPostgresDemoFallback(ctx, handle, hooks, fmt.Sprintf("[postgres-demo] pg_isready não confirmou readiness: %v. Fallback para modo demonstrativo.", err))
	}

	safeLog(hooks.OnLog, "[postgres-demo] pg_isready retornou sucesso")
	safeLog(hooks.OnLog, "[postgres-demo] Serviço pronto para aceitar conexões")
	safeLog(hooks.OnLog, "[postgres-demo] readiness: ready")
	safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
		Status:         ptrWorkloadStatus(StatusRunning),
		Mode:           ptrRuntimeMode(ModeProcessLocal),
		Engine:         ptrString("local-process-engine"),
		Isolated:       ptrBool(false),
		Port:           ptrInt(handle.Port),
		DataDir:        ptrString(handle.DataDir),
		ReadinessState: ptrString("ready"),
		ModeUsed:       &realMode,
	})
	safeStatus(hooks.OnStatus, StatusRunning)

	waitErr := cmd.Wait()
	close(stopDone)
	wg.Wait()

	if ctx.Err() == context.Canceled {
		return stoppedRuntimeResult("Execução cancelada pelo usuário.")
	}
	if waitErr != nil {
		return RuntimeExecutionResult{
			Status:      StatusFailed,
			ExitCode:    extractExitCode(waitErr),
			ExtraLog:    fmt.Sprintf("Servidor PostgreSQL finalizou com erro: %v", waitErr),
			EventType:   "workload_failed",
			EventLabel:  "Workload falhou",
			EventLevel:  SeverityError,
			FinishedErr: waitErr,
		}
	}

	return RuntimeExecutionResult{
		Status:     StatusStopped,
		ExitCode:   0,
		ExtraLog:   "Servidor PostgreSQL encerrado sem erro.",
		EventType:  "workload_stopped",
		EventLabel: "Workload interrompida",
		EventLevel: SeverityWarn,
	}
}

func (e *LocalProcessEngine) runPostgresDemoFallback(
	ctx context.Context,
	handle *RuntimeHandle,
	hooks RuntimeHooks,
	reason string,
) RuntimeExecutionResult {
	demoMode := string(PostgresModeDemo)
	safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
		Status:          ptrWorkloadStatus(StatusStarting),
		Mode:            ptrRuntimeMode(ModeDemo),
		Engine:          ptrString("demo-engine"),
		Isolated:        ptrBool(false),
		Port:            ptrInt(handle.Port),
		DataDir:         ptrString(handle.DataDir),
		ReadinessState:  ptrString("starting"),
		ModeUsed:        &demoMode,
		FallbackApplied: ptrBool(true),
		FallbackReason:  ptrString(reason),
	})
	safeStatus(hooks.OnStatus, StatusStarting)
	if strings.TrimSpace(reason) != "" {
		safeLog(hooks.OnLog, reason)
	}
	safeLog(hooks.OnLog, "[postgres-demo] fallback demo ativo")
	safeLog(hooks.OnLog, "[postgres-demo] PostgreSQL inicializado com sucesso (demo)")
	safeLog(hooks.OnLog, "[postgres-demo] Serviço pronto para aceitar conexões (demo)")
	safeLog(hooks.OnLog, "[postgres-demo] Workload stateful em execução real no Linux: não")
	safeLog(hooks.OnLog, "[postgres-demo] readiness: ready")
	safeRuntimeUpdate(hooks.OnRuntimeUpdate, RuntimeUpdate{
		Status:         ptrWorkloadStatus(StatusRunning),
		Mode:           ptrRuntimeMode(ModeDemo),
		Engine:         ptrString("demo-engine"),
		Isolated:       ptrBool(false),
		ReadinessState: ptrString("ready"),
		ModeUsed:       &demoMode,
	})
	safeStatus(hooks.OnStatus, StatusRunning)

	select {
	case <-ctx.Done():
		return stoppedRuntimeResult("Execução cancelada pelo usuário.")
	case <-time.After(45 * time.Second):
		return RuntimeExecutionResult{
			Status:     StatusCompleted,
			ExitCode:   0,
			EventType:  "workload_completed",
			EventLabel: "Workload concluída",
			EventLevel: SeverityInfo,
		}
	}
}

func runLoggedCommand(ctx context.Context, command string, args []string, onLog func(string)) error {
	cmd := exec.CommandContext(ctx, command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go collectPipe(&wg, stdout, "stdout", onLog)
	go collectPipe(&wg, stderr, "stderr", onLog)
	waitErr := cmd.Wait()
	wg.Wait()
	return waitErr
}

func waitForPostgresReadiness(ctx context.Context, pgIsReadyPath string, port int) error {
	deadline := time.Now().Add(postgresDemoReadinessTimeout)
	var lastErr error

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		output, err := exec.CommandContext(
			probeCtx,
			pgIsReadyPath,
			"-h", "127.0.0.1",
			"-p", strconv.Itoa(port),
		).CombinedOutput()
		cancel()
		if err == nil {
			return nil
		}
		lastErr = fmt.Errorf("%v: %s", err, strings.TrimSpace(string(output)))
		time.Sleep(500 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = errors.New("timeout aguardando readiness")
	}
	return lastErr
}

func terminateProcessGracefully(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return nil
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if processExited(pid) {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}

func processExited(pid int) bool {
	err := syscall.Kill(pid, 0)
	return errors.Is(err, syscall.ESRCH)
}

func findAvailableTCPPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || addr.Port <= 0 {
		return 0, fmt.Errorf("porta dinâmica inválida")
	}
	return addr.Port, nil
}

func postgresFallbackReasonForCapabilities(caps HostCapabilities) string {
	switch {
	case !caps.IsLinux:
		return "[postgres-demo] host não é Linux. Fallback para modo demonstrativo."
	case caps.HasRootPrivileges:
		return "[postgres-demo] initdb local requer usuário não-root no host atual. Fallback para modo demonstrativo."
	case !caps.PostgresBinariesAvailable:
		return "[postgres-demo] binários initdb/postgres/pg_isready não estão disponíveis no host. Fallback para modo demonstrativo."
	case !caps.CanCreateTempDir:
		return "[postgres-demo] não foi possível preparar diretório temporário para PGDATA. Fallback para modo demonstrativo."
	case !caps.CanAllocatePort:
		return "[postgres-demo] não foi possível reservar porta TCP livre para o PostgreSQL Demo. Fallback para modo demonstrativo."
	default:
		return "[postgres-demo] ambiente não atende aos pré-requisitos do modo real. Fallback para modo demonstrativo."
	}
}

func ptrString(value string) *string {
	v := value
	return &v
}

func ptrBool(value bool) *bool {
	v := value
	return &v
}

func ptrRuntimeMode(value RuntimeMode) *RuntimeMode {
	v := value
	return &v
}

func ptrWorkloadStatus(value WorkloadStatus) *WorkloadStatus {
	v := value
	return &v
}

func stoppedRuntimeResult(message string) RuntimeExecutionResult {
	return RuntimeExecutionResult{
		Status:     StatusStopped,
		ExitCode:   130,
		ExtraLog:   message,
		EventType:  "workload_stopped",
		EventLabel: "Workload interrompida",
		EventLevel: SeverityWarn,
	}
}
