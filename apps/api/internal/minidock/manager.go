package minidock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrWorkloadNotFound = errors.New("workload not found")
	ErrInvalidCommand   = errors.New("invalid command")
	ErrDemoNotFound     = errors.New("demo not found")
)

const eventBufferSize = 200

type internalWorkload struct {
	data          Workload
	cancel        context.CancelFunc
	engine        RuntimeEngine
	runtimeHandle *RuntimeHandle
}

type Manager struct {
	mu           sync.RWMutex
	workloads    map[string]*internalWorkload
	order        []string
	demoRuns     map[string]string
	events       []Event
	subscribers  map[chan Event]struct{}
	defaultMode  RuntimeMode
	capabilities HostCapabilities
	engines      map[RuntimeMode]RuntimeEngine
	startedAt    time.Time
	counter      uint64
}

func NewManager(defaultMode RuntimeMode) *Manager {
	rootfsHint := strings.TrimSpace(os.Getenv("MINIDOCK_CONTAINER_ROOTFS"))
	capabilities := DetectHostCapabilities(rootfsHint)

	manager := &Manager{
		workloads:    make(map[string]*internalWorkload),
		order:        make([]string, 0),
		demoRuns:     make(map[string]string),
		events:       make([]Event, 0, eventBufferSize),
		subscribers:  make(map[chan Event]struct{}),
		defaultMode:  ModeProcessLocal,
		capabilities: capabilities,
		engines: map[RuntimeMode]RuntimeEngine{
			ModeDemo:           &DemoEngine{},
			ModeProcessLocal:   &LocalProcessEngine{},
			ModeContainerLinux: NewLinuxContainerEngine(capabilities.RootfsPath),
		},
		startedAt: time.Now(),
	}

	if strings.TrimSpace(string(defaultMode)) == "" {
		defaultMode = capabilities.RecommendedMode
	}
	resolved, _, _ := manager.resolveMode(defaultMode)
	manager.defaultMode = resolved

	return manager
}

func (m *Manager) Health() HealthResponse {
	return HealthResponse{
		Status:      "ok",
		RuntimeMode: m.defaultMode,
		UptimeMs:    time.Since(m.startedAt).Milliseconds(),
		Timestamp:   time.Now(),
	}
}

func (m *Manager) Capabilities() HostCapabilities {
	caps := m.capabilities
	caps.Notes = cloneStringSlice(caps.Notes)
	return caps
}

func (m *Manager) CreateWorkload(req CreateWorkloadRequest) (Workload, error) {
	command, args, err := normalizeCreateCommand(req)
	if err != nil {
		return Workload{}, err
	}
	if err := validateAllowedCommand(command, args); err != nil {
		return Workload{}, err
	}

	targetMode := canonicalMode(req.Mode)
	requestedMode := canonicalMode(req.RequestedMode)
	if requestedMode == "" {
		requestedMode = targetMode
	}
	if requestedMode == "" {
		requestedMode = m.defaultMode
	}
	if targetMode == "" {
		targetMode = requestedMode
	}
	mode, fallbackReason, fallbackApplied := m.resolveMode(targetMode)
	if requestedMode != mode {
		fallbackApplied = true
	}
	if fallbackApplied && fallbackReason == "" {
		fallbackReason = fmt.Sprintf("[minidock] modo solicitado %s redirecionado para %s.", requestedMode, mode)
	}
	fallbackReason = appendFallbackReason(fallbackReason, req.FallbackReasonHint)

	id := m.nextID("wk")
	now := time.Now()
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = SmartName(command, args)
	}
	workloadType := inferWorkloadType(command, args)
	if trimmedType := strings.TrimSpace(req.WorkloadType); trimmedType != "" {
		workloadType = trimmedType
	}
	dataDir := strings.TrimSpace(req.DataDir)

	workload := Workload{
		ID:              id,
		Name:            name,
		Command:         command,
		Args:            cloneStringSlice(args),
		WorkloadType:    workloadType,
		RequestedMode:   requestedMode,
		Summary:         ExecutiveSummary(command, args),
		Status:          StatusPending,
		RiskLevel:       RiskClassification(command, args),
		Mode:            mode,
		FallbackApplied: fallbackApplied,
		FallbackReason:  fallbackReason,
		Logs:            []string{},
		Runtime: RuntimeMetadata{
			Engine:         string(mode) + "-engine",
			Isolated:       mode == ModeContainerLinux,
			WorkloadType:   workloadType,
			Port:           req.Port,
			DataDir:        dataDir,
			ReadinessState: initialReadinessState(workloadType),
		},
		CreatedAt: now,
	}

	if workload.Mode == ModeContainerLinux {
		workload.Runtime.Rootfs = m.capabilities.RootfsPath
		workload.Runtime.ContainerHostname = shortHostnameForWorkload(workload.ID)
	}

	engine := m.engineForMode(mode)
	runtimeHandle, err := engine.Create(workload)
	if err != nil {
		engine, runtimeHandle, workload = m.fallbackFromEngineCreateError(engine, workload, err)
		if runtimeHandle == nil {
			return Workload{}, fmt.Errorf("%w: falha ao preparar runtime: %v", ErrInvalidCommand, err)
		}
	}

	inspect := engine.Inspect(runtimeHandle)
	workload.Runtime.Engine = inspect.Engine
	workload.Runtime.Isolated = inspect.Isolated
	if inspect.Rootfs != "" {
		workload.Runtime.Rootfs = inspect.Rootfs
	}
	if inspect.ContainerHostname != "" {
		workload.Runtime.ContainerHostname = inspect.ContainerHostname
	}
	if inspect.MainPID > 0 {
		workload.Runtime.MainPID = inspect.MainPID
	}

	workload.AIInsights = InsightsFor(workload)
	workload.SuggestedAction = SuggestedActionFor(workload)

	m.mu.Lock()
	m.workloads[id] = &internalWorkload{data: workload, engine: engine, runtimeHandle: runtimeHandle}
	m.order = append([]string{id}, m.order...)
	m.mu.Unlock()

	m.addEvent("workload_created", workload.ID, fmt.Sprintf("Workload %s criada", workload.Name), SeverityInfo)
	if workload.RiskLevel != RiskSafe {
		m.addEvent("workload_flagged", workload.ID, "Workload sinalizada para revisão de risco", SeverityWarn)
	}
	if workload.FallbackApplied && workload.FallbackReason != "" {
		m.appendLog(workload.ID, workload.FallbackReason)
		m.addEvent("workload_fallback", workload.ID, "Fallback automático de runtime aplicado", SeverityWarn)
	}

	go m.runWorkload(workload.ID)

	return m.GetWorkloadCopy(workload.ID)
}

func (m *Manager) fallbackFromEngineCreateError(
	currentEngine RuntimeEngine,
	workload Workload,
	createErr error,
) (RuntimeEngine, *RuntimeHandle, Workload) {
	note := fmt.Sprintf("[minidock] falha ao preparar modo %s (%v).", workload.Mode, createErr)

	if workload.Mode != ModeProcessLocal {
		processEngine := m.engineForMode(ModeProcessLocal)
		workload.Mode = ModeProcessLocal
		workload.FallbackApplied = true
		workload.FallbackReason = appendFallbackReason(workload.FallbackReason, note+" fallback para processo-local.")
		workload.Runtime.Engine = "local-process-engine"
		workload.Runtime.Isolated = false
		workload.Runtime.Rootfs = ""
		workload.Runtime.ContainerHostname = ""
		workload.Runtime.MainPID = 0
		if handle, err := processEngine.Create(workload); err == nil {
			return processEngine, handle, workload
		}
	}

	if workload.Mode != ModeDemo {
		demoEngine := m.engineForMode(ModeDemo)
		workload.Mode = ModeDemo
		workload.FallbackApplied = true
		workload.FallbackReason = appendFallbackReason(workload.FallbackReason, note+" fallback para demo.")
		workload.Runtime.Engine = "demo-engine"
		workload.Runtime.Isolated = false
		workload.Runtime.Rootfs = ""
		workload.Runtime.ContainerHostname = ""
		workload.Runtime.MainPID = 0
		if handle, err := demoEngine.Create(workload); err == nil {
			return demoEngine, handle, workload
		}
	}

	return currentEngine, nil, workload
}

func appendFallbackReason(current, extra string) string {
	trimmedCurrent := strings.TrimSpace(current)
	trimmedExtra := strings.TrimSpace(extra)
	switch {
	case trimmedCurrent == "":
		return trimmedExtra
	case trimmedExtra == "":
		return trimmedCurrent
	default:
		return trimmedCurrent + " " + trimmedExtra
	}
}

func (m *Manager) runWorkload(id string) {
	_, ok := m.transitionToRunning(id)
	if !ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	engine, handle, ok := m.attachRuntime(id, cancel)
	if !ok {
		cancel()
		return
	}

	result := engine.Start(ctx, handle, RuntimeHooks{
		OnLog: func(line string) {
			m.appendLog(id, line)
		},
		OnMainPID: func(pid int) {
			m.updateMainPID(id, pid)
		},
	})
	m.finishWithRuntimeResult(id, result)
}

func (m *Manager) ListWorkloads() []Workload {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Workload, 0, len(m.order))
	for _, id := range m.order {
		if w, ok := m.workloads[id]; ok {
			result = append(result, cloneWorkload(w.data))
		}
	}
	return result
}

func (m *Manager) GetWorkload(id string) (Workload, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.workloads[id]
	if !ok {
		return Workload{}, false
	}
	return cloneWorkload(w.data), true
}

func (m *Manager) GetWorkloadCopy(id string) (Workload, error) {
	w, ok := m.GetWorkload(id)
	if !ok {
		return Workload{}, ErrWorkloadNotFound
	}
	return w, nil
}

func (m *Manager) StopWorkload(id string) (Workload, error) {
	var (
		cancel context.CancelFunc
		engine RuntimeEngine
		handle *RuntimeHandle
	)

	m.mu.Lock()
	w, ok := m.workloads[id]
	if !ok {
		m.mu.Unlock()
		return Workload{}, ErrWorkloadNotFound
	}
	if w.data.Status != StatusRunning && w.data.Status != StatusPending {
		copy := cloneWorkload(w.data)
		m.mu.Unlock()
		return copy, nil
	}
	cancel = w.cancel
	engine = w.engine
	handle = w.runtimeHandle
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if engine != nil {
		_ = engine.Stop(handle)
	}

	m.finishWithRuntimeResult(id, RuntimeExecutionResult{
		Status:     StatusStopped,
		ExitCode:   130,
		ExtraLog:   "Execução interrompida manualmente.",
		EventType:  "workload_stopped",
		EventLabel: "Workload interrompida",
		EventLevel: SeverityWarn,
	})
	return m.GetWorkloadCopy(id)
}

func (m *Manager) DeleteWorkload(id string) error {
	var (
		name   string
		cancel context.CancelFunc
		engine RuntimeEngine
		handle *RuntimeHandle
	)

	m.mu.Lock()
	w, ok := m.workloads[id]
	if !ok {
		m.mu.Unlock()
		return ErrWorkloadNotFound
	}
	name = w.data.Name
	cancel = w.cancel
	engine = w.engine
	handle = w.runtimeHandle
	delete(m.workloads, id)
	for i, value := range m.order {
		if value == id {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if engine != nil {
		_ = engine.Remove(handle)
	}

	m.addEvent("workload_removed", id, fmt.Sprintf("Workload %s removida", name), SeverityInfo)
	return nil
}

func (m *Manager) GetLogs(id string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.workloads[id]
	if !ok {
		return nil, ErrWorkloadNotFound
	}
	return cloneStringSlice(w.data.Logs), nil
}

func (m *Manager) ListEvents() []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	items := make([]Event, len(m.events))
	copy(items, m.events)
	return items
}

func (m *Manager) SubscribeEvents() (chan Event, func()) {
	ch := make(chan Event, 32)
	m.mu.Lock()
	m.subscribers[ch] = struct{}{}
	m.mu.Unlock()

	unsubscribe := func() {
		m.mu.Lock()
		if _, ok := m.subscribers[ch]; ok {
			delete(m.subscribers, ch)
			close(ch)
		}
		m.mu.Unlock()
	}
	return ch, unsubscribe
}

func (m *Manager) SeedDemo(force bool) []Workload {
	m.mu.Lock()
	if !force && len(m.workloads) > 0 {
		current := m.snapshotWorkloadsLocked()
		m.mu.Unlock()
		return current
	}

	for _, item := range m.workloads {
		if item.cancel != nil {
			item.cancel()
		}
	}

	m.workloads = make(map[string]*internalWorkload)
	m.order = []string{}
	m.demoRuns = make(map[string]string)

	now := time.Now()
	seed := []Workload{
		{
			ID:            m.nextID("wk"),
			Name:          "filesystem-inspection",
			Command:       "ls",
			Args:          []string{"-la"},
			WorkloadType:  "Environment",
			RequestedMode: ModeDemo,
			Summary:       ExecutiveSummary("ls", []string{"-la"}),
			Status:        StatusCompleted,
			RiskLevel:     RiskSafe,
			Mode:          ModeDemo,
			Runtime: RuntimeMetadata{
				Engine:       "demo-engine",
				Isolated:     false,
				WorkloadType: "Environment",
			},
			StartedAt: ptrTime(now.Add(-4 * time.Minute)),
			FinishedAt: ptrTime(now.Add(-4*time.Minute +
				120*time.Millisecond)),
			DurationMs: 120,
			ExitCode:   ptrInt(0),
			Logs: []string{
				"[stdout] total 16",
				"[stdout] drwxr-xr-x  6 dev  staff   192 Mar 17 16:58 .",
				"[stdout] -rw-r--r--  1 dev  staff  1024 README.md",
			},
			CreatedAt: now.Add(-5 * time.Minute),
		},
		{
			ID:            m.nextID("wk"),
			Name:          "shell-session",
			Command:       "/bin/sh",
			Args:          []string{},
			WorkloadType:  "Runtime",
			RequestedMode: ModeDemo,
			Summary:       ExecutiveSummary("/bin/sh", nil),
			Status:        StatusStopped,
			RiskLevel:     RiskReview,
			Mode:          ModeDemo,
			Runtime: RuntimeMetadata{
				Engine:       "demo-engine",
				Isolated:     false,
				WorkloadType: "Runtime",
			},
			StartedAt: ptrTime(now.Add(-3 * time.Minute)),
			FinishedAt: ptrTime(now.Add(-3*time.Minute +
				5*time.Second)),
			DurationMs: 5000,
			ExitCode:   ptrInt(130),
			Logs: []string{
				"[minidock] sessão /bin/sh executada em modo não interativo.",
				"[stdout] MiniDock shell diagnostic session (non-interactive)",
				"[stdout] /Users/dev/demo",
			},
			CreatedAt: now.Add(-4 * time.Minute),
		},
		{
			ID:            m.nextID("wk"),
			Name:          "runtime-validation",
			Command:       "sleep",
			Args:          []string{"60"},
			WorkloadType:  "Lifecycle",
			RequestedMode: ModeDemo,
			Summary:       ExecutiveSummary("sleep", []string{"60"}),
			Status:        StatusRunning,
			RiskLevel:     RiskSafe,
			Mode:          ModeDemo,
			Runtime: RuntimeMetadata{
				Engine:       "demo-engine",
				Isolated:     false,
				WorkloadType: "Lifecycle",
			},
			StartedAt: ptrTime(now.Add(-40 * time.Second)),
			Logs: []string{
				"[stdout] rotina de espera em execução...",
			},
			CreatedAt: now.Add(-45 * time.Second),
		},
		{
			ID:            m.nextID("wk"),
			Name:          "output-check",
			Command:       "echo",
			Args:          []string{"hello"},
			WorkloadType:  "Runtime",
			RequestedMode: ModeDemo,
			Summary:       ExecutiveSummary("echo", []string{"hello"}),
			Status:        StatusCompleted,
			RiskLevel:     RiskSafe,
			Mode:          ModeDemo,
			Runtime: RuntimeMetadata{
				Engine:       "demo-engine",
				Isolated:     false,
				WorkloadType: "Runtime",
			},
			StartedAt: ptrTime(now.Add(-2 * time.Minute)),
			FinishedAt: ptrTime(now.Add(-2*time.Minute +
				40*time.Millisecond)),
			DurationMs: 40,
			ExitCode:   ptrInt(0),
			Logs: []string{
				"[stdout] hello",
			},
			CreatedAt: now.Add(-2*time.Minute - 5*time.Second),
		},
	}

	for _, entry := range seed {
		entry.AIInsights = InsightsFor(entry)
		entry.SuggestedAction = SuggestedActionFor(entry)
		engine := m.engineForMode(entry.Mode)
		handle, _ := engine.Create(entry)
		m.workloads[entry.ID] = &internalWorkload{data: entry, engine: engine, runtimeHandle: handle}
		m.order = append(m.order, entry.ID)
	}
	m.order = reverseStrings(m.order)
	snapshot := m.snapshotWorkloadsLocked()
	m.mu.Unlock()

	for _, item := range snapshot {
		m.addEvent("workload_created", item.ID, fmt.Sprintf("Workload %s preparada para demonstração", item.Name), SeverityInfo)
		if item.RiskLevel != RiskSafe {
			m.addEvent("workload_flagged", item.ID, "Workload sinalizada para revisão de risco", SeverityWarn)
		}
	}
	m.addEvent("demo_seeded", "", "Base de demonstração carregada", SeverityInfo)

	return snapshot
}

func (m *Manager) ExecutiveSummary() DashboardSummary {
	return DashboardSummary{Lines: GlobalExecutiveSummary(m.ListWorkloads())}
}

func (m *Manager) transitionToRunning(id string) (Workload, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workloads[id]
	if !ok {
		return Workload{}, false
	}
	if w.data.Status != StatusPending {
		return cloneWorkload(w.data), false
	}
	now := time.Now()
	w.data.Status = StatusRunning
	w.data.StartedAt = &now
	if isDatabaseWorkload(w.data.WorkloadType) {
		w.data.Runtime.ReadinessState = "starting"
	}
	w.data.AIInsights = InsightsFor(w.data)
	w.data.SuggestedAction = SuggestedActionFor(w.data)
	go m.addEvent("workload_started", id, "Workload iniciada", SeverityInfo)
	return cloneWorkload(w.data), true
}

func (m *Manager) attachRuntime(id string, cancel context.CancelFunc) (RuntimeEngine, *RuntimeHandle, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workloads[id]
	if !ok {
		return nil, nil, false
	}
	w.cancel = cancel
	if w.engine == nil {
		w.engine = m.engineForMode(w.data.Mode)
	}
	if w.runtimeHandle == nil {
		handle, err := w.engine.Create(w.data)
		if err != nil {
			return nil, nil, false
		}
		w.runtimeHandle = handle
	}
	return w.engine, w.runtimeHandle, true
}

func (m *Manager) finishWithRuntimeResult(id string, result RuntimeExecutionResult) {
	m.mu.Lock()
	w, ok := m.workloads[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	if w.data.FinishedAt != nil {
		m.mu.Unlock()
		return
	}

	now := time.Now()
	w.data.Status = result.Status
	w.data.FinishedAt = &now
	if w.data.StartedAt != nil {
		w.data.DurationMs = now.Sub(*w.data.StartedAt).Milliseconds()
	}
	exitCode := result.ExitCode
	w.data.ExitCode = &exitCode
	if result.ExtraLog != "" {
		w.data.Logs = append(w.data.Logs, result.ExtraLog)
	}
	if w.runtimeHandle != nil && w.runtimeHandle.MainPID > 0 {
		w.data.Runtime.MainPID = w.runtimeHandle.MainPID
	}
	switch result.Status {
	case StatusCompleted:
		if isDatabaseWorkload(w.data.WorkloadType) && w.data.Runtime.ReadinessState == "starting" {
			w.data.Runtime.ReadinessState = "ready"
		}
	case StatusStopped:
		if isDatabaseWorkload(w.data.WorkloadType) {
			w.data.Runtime.ReadinessState = "stopped"
		}
	case StatusFailed:
		if isDatabaseWorkload(w.data.WorkloadType) {
			w.data.Runtime.ReadinessState = "failed"
		}
	}
	w.data.AIInsights = InsightsFor(w.data)
	w.data.SuggestedAction = SuggestedActionFor(w.data)
	w.cancel = nil
	m.mu.Unlock()

	eventType := result.EventType
	eventLabel := result.EventLabel
	eventLevel := result.EventLevel
	if eventType == "" || eventLabel == "" {
		eventType, eventLabel, eventLevel = defaultEventForStatus(result.Status)
	}
	m.addEvent(eventType, id, eventLabel, eventLevel)
}

func defaultEventForStatus(status WorkloadStatus) (string, string, EventSeverity) {
	switch status {
	case StatusCompleted:
		return "workload_completed", "Workload concluída", SeverityInfo
	case StatusStopped:
		return "workload_stopped", "Workload interrompida", SeverityWarn
	case StatusFailed:
		return "workload_failed", "Workload falhou", SeverityError
	default:
		return "workload_updated", "Workload atualizada", SeverityInfo
	}
}

func (m *Manager) updateMainPID(id string, pid int) {
	if pid <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workloads[id]
	if !ok {
		return
	}
	w.data.Runtime.MainPID = pid
	if w.runtimeHandle != nil {
		w.runtimeHandle.MainPID = pid
	}
}

func (m *Manager) appendLog(id, line string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workloads[id]
	if !ok {
		return
	}
	w.data.Logs = append(w.data.Logs, line)
	if isDatabaseWorkload(w.data.WorkloadType) {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "ready to accept connections") || strings.Contains(lower, "readiness: ready") {
			w.data.Runtime.ReadinessState = "ready"
		}
		if strings.Contains(lower, "starting") && w.data.Runtime.ReadinessState == "" {
			w.data.Runtime.ReadinessState = "starting"
		}
	}
}

func (m *Manager) addEvent(eventType, workloadID, message string, severity EventSeverity) {
	event := Event{
		ID:         m.nextID("evt"),
		Type:       eventType,
		WorkloadID: workloadID,
		Message:    message,
		Severity:   severity,
		CreatedAt:  time.Now(),
	}

	m.mu.Lock()
	m.events = append(m.events, event)
	if len(m.events) > eventBufferSize {
		m.events = m.events[len(m.events)-eventBufferSize:]
	}

	subscribers := make([]chan Event, 0, len(m.subscribers))
	for ch := range m.subscribers {
		subscribers = append(subscribers, ch)
	}
	m.mu.Unlock()

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (m *Manager) snapshotWorkloadsLocked() []Workload {
	result := make([]Workload, 0, len(m.order))
	for _, id := range m.order {
		if w, ok := m.workloads[id]; ok {
			result = append(result, cloneWorkload(w.data))
		}
	}
	return result
}

func (m *Manager) resolveMode(requested RuntimeMode) (RuntimeMode, string, bool) {
	mode := canonicalMode(requested)
	if mode == "" {
		mode = canonicalMode(m.defaultMode)
		if mode == "" {
			mode = m.capabilities.RecommendedMode
		}
	}

	switch mode {
	case ModeDemo:
		return ModeDemo, "", false
	case ModeProcessLocal:
		return ModeProcessLocal, "", false
	case ModeContainerLinux:
		if m.capabilities.SupportsContainers {
			return ModeContainerLinux, "", false
		}
		if m.capabilities.HasRootPrivileges {
			return ModeProcessLocal, "[minidock] container-linux indisponível neste host; fallback para processo-local.", true
		}
		return ModeProcessLocal, "[minidock] container-linux requer Linux + root + rootfs. Fallback para processo-local.", true
	default:
		return ModeProcessLocal, "[minidock] modo inválido recebido; fallback automático para processo-local.", true
	}
}

func (m *Manager) engineForMode(mode RuntimeMode) RuntimeEngine {
	resolved := canonicalMode(mode)
	if engine, ok := m.engines[resolved]; ok {
		return engine
	}
	if engine, ok := m.engines[ModeProcessLocal]; ok {
		return engine
	}
	return m.engines[ModeDemo]
}

func (m *Manager) nextID(prefix string) string {
	counter := atomic.AddUint64(&m.counter, 1)
	return fmt.Sprintf("%s_%d", prefix, counter)
}

func canonicalMode(mode RuntimeMode) RuntimeMode {
	trimmed := RuntimeMode(strings.TrimSpace(string(mode)))
	if trimmed == ModeNamespaceRuntime {
		return ModeContainerLinux
	}
	return trimmed
}

func normalizeCreateCommand(req CreateWorkloadRequest) (string, []string, error) {
	if len(req.Args) > 0 {
		command := strings.TrimSpace(req.Command)
		if command == "" {
			return "", nil, fmt.Errorf("%w: command é obrigatório", ErrInvalidCommand)
		}
		return command, cloneStringSlice(req.Args), nil
	}
	return splitCommand(req.Command)
}

func splitCommand(raw string) (string, []string, error) {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("%w: command é obrigatório", ErrInvalidCommand)
	}
	return parts[0], parts[1:], nil
}

func validateAllowedCommand(command string, args []string) error {
	normalized := NormalizeCommand(command, args)
	allowed := map[string]struct{}{
		"echo hello":                           {},
		"ls":                                   {},
		"ls -la":                               {},
		"pwd":                                  {},
		"sleep 10":                             {},
		"sleep 60":                             {},
		"/bin/sh":                              {},
		"/bin/sh -c hostname && pwd && ls /":   {},
		"/bin/sh -c echo hello from container": {},
		"/bin/sh -c echo starting && sleep 10 && echo finished": {},
		"/bin/sh -c uname -a && id && ps":                       {},
		"/bin/sh -c hostname && uname -a && ps":                 {},
		"/bin/sh -c ls -la / && ls -la /bin":                    {},
		"/bin/sh -c echo fallback validation":                   {},
		"minidock-postgres-demo":                                {},
	}
	if _, ok := allowed[normalized]; ok {
		return nil
	}
	return fmt.Errorf("%w: comando não suportado no MVP (%s)", ErrInvalidCommand, normalized)
}

func effectiveCommand(command string, args []string) (string, []string, []string) {
	if command == "/bin/sh" && len(args) == 0 {
		return "/bin/sh", []string{"-c", "echo 'MiniDock shell diagnostic session (non-interactive)'; pwd; ls"}, []string{"[minidock] sessão /bin/sh executada em modo não interativo."}
	}
	if command == "minidock-postgres-demo" {
		return "/bin/sh", []string{"-c", localPostgresDemoScript()}, []string{
			"[minidock] iniciando PostgreSQL demo em processo-local (com fallback automático).",
		}
	}
	return command, args, nil
}

func inferWorkloadType(command string, args []string) string {
	normalized := NormalizeCommand(command, args)
	switch normalized {
	case "minidock-postgres-demo":
		return "Database"
	case "/bin/sh -c echo starting && sleep 10 && echo finished", "sleep 10", "sleep 60":
		return "Lifecycle"
	case "/bin/sh -c hostname && uname -a && ps", "/bin/sh -c uname -a && id && ps":
		return "Diagnostics"
	case "/bin/sh -c ls -la / && ls -la /bin", "ls", "ls -la", "pwd":
		return "Environment"
	default:
		return "Runtime"
	}
}

func isDatabaseWorkload(workloadType string) bool {
	return strings.EqualFold(strings.TrimSpace(workloadType), "Database")
}

func initialReadinessState(workloadType string) string {
	if isDatabaseWorkload(workloadType) {
		return "pending"
	}
	return ""
}

func localPostgresDemoScript() string {
	return strings.Join([]string{
		"set -eu",
		"PORT=\"${MINIDOCK_POSTGRES_PORT:-55432}\"",
		"DATA_DIR=\"${MINIDOCK_POSTGRES_DATA_DIR:-/tmp/minidock-postgres-demo}\"",
		"echo \"[postgres-demo] preparing data directory ${DATA_DIR}\"",
		"if command -v initdb >/dev/null 2>&1 && command -v postgres >/dev/null 2>&1; then",
		"  mkdir -p \"${DATA_DIR}\"",
		"  if [ ! -f \"${DATA_DIR}/PG_VERSION\" ]; then",
		"    if ! initdb -D \"${DATA_DIR}\" -A trust >/tmp/minidock-initdb.log 2>&1; then",
		"      echo \"[postgres-demo] initdb failed; fallback para simulação.\"",
		"      cat /tmp/minidock-initdb.log || true",
		"      echo \"[postgres-demo] PostgreSQL inicializado com sucesso (simulado)\"",
		"      echo \"[postgres-demo] Banco pronto para aceitar conexões (simulado)\"",
		"      echo \"[postgres-demo] readiness: ready\"",
		"      sleep 45",
		"      exit 0",
		"    fi",
		"  fi",
		"  echo \"[postgres-demo] starting postgres on port ${PORT}\"",
		"  exec postgres -D \"${DATA_DIR}\" -p \"${PORT}\" -k \"${DATA_DIR}\"",
		"fi",
		"echo \"[postgres-demo] binários postgres/initdb não encontrados; fallback para demo.\"",
		"echo \"[postgres-demo] PostgreSQL inicializado com sucesso (simulado)\"",
		"echo \"[postgres-demo] Banco pronto para aceitar conexões (simulado)\"",
		"echo \"[postgres-demo] readiness: ready\"",
		"sleep 45",
	}, "; ")
}

func cloneWorkload(w Workload) Workload {
	copyWorkload := w
	copyWorkload.Args = cloneStringSlice(w.Args)
	copyWorkload.Logs = cloneStringSlice(w.Logs)
	copyWorkload.AIInsights = cloneStringSlice(w.AIInsights)
	if w.StartedAt != nil {
		value := *w.StartedAt
		copyWorkload.StartedAt = &value
	}
	if w.FinishedAt != nil {
		value := *w.FinishedAt
		copyWorkload.FinishedAt = &value
	}
	if w.ExitCode != nil {
		value := *w.ExitCode
		copyWorkload.ExitCode = &value
	}
	return copyWorkload
}

func cloneStringSlice(items []string) []string {
	cloned := slices.Clone(items)
	if cloned == nil {
		return []string{}
	}
	return cloned
}

func ptrTime(value time.Time) *time.Time {
	v := value
	return &v
}

func ptrInt(value int) *int {
	v := value
	return &v
}

func reverseStrings(items []string) []string {
	for i := 0; i < len(items)/2; i++ {
		j := len(items) - 1 - i
		items[i], items[j] = items[j], items[i]
	}
	return items
}
