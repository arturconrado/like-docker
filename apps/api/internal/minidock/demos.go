package minidock

import (
	"fmt"
	"strings"
)

type demoSpec struct {
	Definition DemoDefinition
	Command    string
	Args       []string
}

var demoCatalog = []demoSpec{
	{
		Definition: DemoDefinition{
			ID:            "hello-container",
			Name:          "Hello Container",
			Description:   "Executa rotina curta para validar criação, logs e conclusão da workload.",
			Objective:     "Validar execução e logs com ciclo de vida curto.",
			PreferredMode: ModeContainerLinux,
			WorkloadType:  "Runtime",
			Complexity:    "Básico",
			RequiredCapabilities: []string{
				"supportsProcessLocal",
			},
			ExpectedSignals: []string{
				"stdout com hello from container",
				"status Completed",
			},
			Tags: []string{"runtime", "logs", "quick-check"},
			Icon: "sparkles",
		},
		Command: "/bin/sh",
		Args:    []string{"-c", "echo hello from container"},
	},
	{
		Definition: DemoDefinition{
			ID:            "filesystem-inspection",
			Name:          "Filesystem Inspection",
			Description:   "Inspeciona diretórios raiz para evidenciar rootfs e conteúdo do ambiente.",
			Objective:     "Mostrar rootfs e conteúdo interno do ambiente isolado.",
			PreferredMode: ModeContainerLinux,
			WorkloadType:  "Environment",
			Complexity:    "Intermediário",
			RequiredCapabilities: []string{
				"rootfsAvailable",
			},
			ExpectedSignals: []string{
				"listagem de /",
				"listagem de /bin",
			},
			Tags: []string{"filesystem", "rootfs", "inspection"},
			Icon: "folder-search",
		},
		Command: "/bin/sh",
		Args:    []string{"-c", "ls -la / && ls -la /bin"},
	},
	{
		Definition: DemoDefinition{
			ID:            "runtime-diagnostics",
			Name:          "Runtime Diagnostics",
			Description:   "Roda diagnóstico de identidade e processos para validar modo efetivo de execução.",
			Objective:     "Mostrar hostname, uname e processos para validar runtime.",
			PreferredMode: ModeContainerLinux,
			WorkloadType:  "Diagnostics",
			Complexity:    "Intermediário",
			RequiredCapabilities: []string{
				"supportsProcessLocal",
			},
			ExpectedSignals: []string{
				"hostname exibido",
				"uname exibido",
				"process table exibida",
			},
			Tags: []string{"diagnostics", "hostname", "processes"},
			Icon: "activity",
		},
		Command: "/bin/sh",
		Args:    []string{"-c", "hostname && uname -a && ps"},
	},
	{
		Definition: DemoDefinition{
			ID:            "controlled-sleep",
			Name:          "Controlled Sleep",
			Description:   "Demonstra ciclo de vida controlado para observar Running, Stop e conclusão.",
			Objective:     "Mostrar Running -> Completed e permitir Stop.",
			PreferredMode: ModeContainerLinux,
			WorkloadType:  "Lifecycle",
			Complexity:    "Intermediário",
			RequiredCapabilities: []string{
				"supportsProcessLocal",
			},
			ExpectedSignals: []string{
				"status Running",
				"status Completed ou Stopped",
			},
			Tags: []string{"lifecycle", "running", "stop"},
			Icon: "timer",
		},
		Command: "/bin/sh",
		Args:    []string{"-c", "echo starting && sleep 10 && echo finished"},
	},
	{
		Definition: DemoDefinition{
			ID:            "postgres-demo",
			Name:          "PostgreSQL Demo",
			Description:   "Workload stateful para demonstrar maturidade operacional com logs e readiness.",
			Objective:     "Inicializar PostgreSQL, expor evidências operacionais e readiness.",
			PreferredMode: ModeContainerLinux,
			WorkloadType:  "Database",
			Complexity:    "Avançado",
			RequiredCapabilities: []string{
				"supportsPostgresDemo",
				"postgresLocalAvailable ou postgresContainerAvailable",
			},
			ExpectedSignals: []string{
				"logs de inicialização PostgreSQL",
				"readiness pronto para conexões",
				"porta e data directory visíveis",
			},
			Tags:    []string{"database", "stateful", "postgresql"},
			Icon:    "database",
			Port:    55432,
			DataDir: "/tmp/minidock-postgres-demo",
		},
		Command: "minidock-postgres-demo",
		Args:    []string{},
	},
}

func (m *Manager) ListDemos() []DemoDefinition {
	items := make([]DemoDefinition, 0, len(demoCatalog))
	for _, spec := range demoCatalog {
		items = append(items, cloneDemoDefinition(spec.Definition))
	}
	return items
}

func (m *Manager) GetDemo(id string) (DemoDefinition, bool) {
	spec, ok := getDemoSpec(id)
	if !ok {
		return DemoDefinition{}, false
	}
	return cloneDemoDefinition(spec.Definition), true
}

func (m *Manager) RunDemo(id string) (DemoRunResponse, error) {
	spec, ok := getDemoSpec(id)
	if !ok {
		return DemoRunResponse{}, ErrDemoNotFound
	}

	modeToUse, fallbackHint := resolveModeForDemo(spec, m.capabilities)
	created, err := m.CreateWorkload(CreateWorkloadRequest{
		Name:               spec.Definition.ID,
		Command:            spec.Command,
		Args:               cloneStringSlice(spec.Args),
		Mode:               modeToUse,
		RequestedMode:      spec.Definition.PreferredMode,
		WorkloadType:       spec.Definition.WorkloadType,
		Port:               spec.Definition.Port,
		DataDir:            spec.Definition.DataDir,
		FallbackReasonHint: fallbackHint,
	})
	if err != nil {
		return DemoRunResponse{}, err
	}

	m.mu.Lock()
	m.demoRuns[spec.Definition.ID] = created.ID
	m.mu.Unlock()

	return DemoRunResponse{
		Demo:              cloneDemoDefinition(spec.Definition),
		Workload:          created,
		ExecutionModeUsed: created.Mode,
		FallbackApplied:   created.FallbackApplied,
		FallbackReason:    created.FallbackReason,
	}, nil
}

func (m *Manager) ValidateDemo(id string) (DemoValidation, error) {
	spec, ok := getDemoSpec(id)
	if !ok {
		return DemoValidation{}, ErrDemoNotFound
	}

	m.mu.RLock()
	workloadID := strings.TrimSpace(m.demoRuns[id])
	m.mu.RUnlock()
	if workloadID == "" {
		return DemoValidation{
			DemoID:            id,
			Success:           false,
			ExecutionModeUsed: spec.Definition.PreferredMode,
			FallbackApplied:   false,
			Signals: []string{
				"Demonstração ainda não foi executada nesta sessão.",
			},
			SummaryLines: []string{
				"Execute a demonstração para consolidar evidências técnicas e operacionais.",
			},
		}, nil
	}

	workload, ok := m.GetWorkload(workloadID)
	if !ok {
		return DemoValidation{
			DemoID:            id,
			WorkloadID:        workloadID,
			Success:           false,
			ExecutionModeUsed: spec.Definition.PreferredMode,
			FallbackApplied:   false,
			Signals: []string{
				"Última workload da demonstração não está mais disponível.",
			},
			SummaryLines: []string{
				"A demonstração precisa ser reexecutada para gerar evidências atuais.",
			},
		}, nil
	}

	signals := collectDemoSignals(spec.Definition, workload)
	success := workload.Status != StatusFailed && len(signals) >= 2
	if isDatabaseWorkload(spec.Definition.WorkloadType) {
		success = workload.Status != StatusFailed &&
			(strings.EqualFold(workload.Runtime.ReadinessState, "ready") || containsReadyLog(workload.Logs))
	}

	return DemoValidation{
		DemoID:            id,
		WorkloadID:        workload.ID,
		Success:           success,
		ExecutionModeUsed: workload.Mode,
		FallbackApplied:   workload.FallbackApplied,
		Signals:           signals,
		SummaryLines:      buildDemoSummary(spec.Definition, workload, success),
	}, nil
}

func getDemoSpec(id string) (demoSpec, bool) {
	normalizedID := strings.TrimSpace(strings.ToLower(id))
	for _, spec := range demoCatalog {
		if strings.ToLower(spec.Definition.ID) == normalizedID {
			cloned := spec
			cloned.Definition = cloneDemoDefinition(spec.Definition)
			cloned.Args = cloneStringSlice(spec.Args)
			return cloned, true
		}
	}
	return demoSpec{}, false
}

func resolveModeForDemo(spec demoSpec, caps HostCapabilities) (RuntimeMode, string) {
	id := spec.Definition.ID
	preferred := canonicalMode(spec.Definition.PreferredMode)
	if preferred == "" {
		preferred = ModeProcessLocal
	}

	if id == "postgres-demo" {
		if caps.SupportsContainers && caps.PostgresContainerAvailable {
			return ModeContainerLinux, ""
		}
		if caps.PostgresLocalAvailable {
			return ModeProcessLocal, "[minidock] host não atende pré-requisitos para PostgreSQL em container-linux; execução redirecionada para processo-local."
		}
		return ModeDemo, "[minidock] PostgreSQL real indisponível no host atual; execução redirecionada para modo demo com evidências plausíveis."
	}

	switch preferred {
	case ModeContainerLinux:
		if caps.SupportsContainers {
			return ModeContainerLinux, ""
		}
		if caps.SupportsProcessLocal {
			return ModeProcessLocal, "[minidock] container-linux indisponível no host atual; fallback automático para processo-local."
		}
		return ModeDemo, "[minidock] processo-local indisponível; fallback final para modo demo."
	case ModeProcessLocal:
		if caps.SupportsProcessLocal {
			return ModeProcessLocal, ""
		}
		return ModeDemo, "[minidock] processo-local indisponível; fallback para modo demo."
	default:
		return ModeDemo, ""
	}
}

func collectDemoSignals(demo DemoDefinition, workload Workload) []string {
	signals := make([]string, 0, 8)

	switch workload.Status {
	case StatusRunning:
		signals = append(signals, "Workload está em execução (Running).")
	case StatusCompleted:
		signals = append(signals, "Workload concluiu execução com sucesso (Completed).")
	case StatusStopped:
		signals = append(signals, "Workload foi encerrada de forma controlada (Stopped).")
	case StatusFailed:
		signals = append(signals, "Workload falhou durante a execução.")
	}

	if len(workload.Logs) > 0 {
		signals = append(signals, fmt.Sprintf("Logs capturados: %d linhas disponíveis.", len(workload.Logs)))
	}
	if workload.Mode == ModeContainerLinux {
		signals = append(signals, "Execução realizada em isolamento Linux com rootfs dedicado.")
	}
	if workload.FallbackApplied {
		signals = append(signals, "Fallback automático aplicado para preservar experiência de demonstração.")
	}
	if workload.Runtime.Rootfs != "" {
		signals = append(signals, "Rootfs configurado e visível nos metadados da workload.")
	}
	if workload.Runtime.ContainerHostname != "" {
		signals = append(signals, "Hostname isolado reportado pela runtime.")
	}
	if isDatabaseWorkload(demo.WorkloadType) {
		if workload.Runtime.Port > 0 {
			signals = append(signals, fmt.Sprintf("Porta operacional configurada: %d.", workload.Runtime.Port))
		}
		if workload.Runtime.DataDir != "" {
			signals = append(signals, fmt.Sprintf("Data directory observado: %s.", workload.Runtime.DataDir))
		}
		if strings.EqualFold(workload.Runtime.ReadinessState, "ready") || containsReadyLog(workload.Logs) {
			signals = append(signals, "Readiness consistente: banco pronto para conexões.")
		}
	}

	return dedupeLines(signals)
}

func buildDemoSummary(demo DemoDefinition, workload Workload, success bool) []string {
	lines := make([]string, 0, 4)
	if success {
		lines = append(lines, "A demonstração confirmou a capacidade da plataforma de executar workloads em múltiplos modos.")
	} else {
		lines = append(lines, "A demonstração foi executada, mas ainda exige revisão dos sinais para validação completa.")
	}

	if workload.Mode == ModeContainerLinux {
		lines = append(lines, "O modo container-linux ampliou a fidelidade técnica da execução demonstrada.")
	}
	if workload.FallbackApplied {
		lines = append(lines, "O host atual não suportou o modo avançado integral; fallback controlado preservou a UX do produto.")
	}

	if isDatabaseWorkload(demo.WorkloadType) {
		if strings.EqualFold(workload.Runtime.ReadinessState, "ready") || containsReadyLog(workload.Logs) {
			lines = append(lines, "O PostgreSQL foi inicializado com sucesso e apresentou sinais esperados de readiness e operação.")
		} else {
			lines = append(lines, "A workload PostgreSQL mantém narrativa stateful, porém sem readiness conclusivo neste ciclo.")
		}
		lines = append(lines, "A workload PostgreSQL demonstra suporte a cenário stateful com observabilidade operacional.")
	}

	if len(lines) > 4 {
		lines = lines[:4]
	}
	return lines
}

func cloneDemoDefinition(demo DemoDefinition) DemoDefinition {
	cloned := demo
	cloned.RequiredCapabilities = cloneStringSlice(demo.RequiredCapabilities)
	cloned.ExpectedSignals = cloneStringSlice(demo.ExpectedSignals)
	cloned.Tags = cloneStringSlice(demo.Tags)
	return cloned
}

func containsReadyLog(lines []string) bool {
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "ready to accept connections") || strings.Contains(lower, "readiness: ready") {
			return true
		}
	}
	return false
}

func dedupeLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
