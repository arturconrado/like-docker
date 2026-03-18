package minidock

import (
	"fmt"
	"strings"
)

var destructivePatterns = []string{
	"rm -rf",
	"mkfs",
	"dd if=",
	"shutdown",
	"reboot",
	":(){",
	"chmod -R 777 /",
}

func SmartName(command string, args []string) string {
	normalized := NormalizeCommand(command, args)

	switch normalized {
	case "echo hello":
		return "output-check"
	case "ls", "ls -la":
		return "filesystem-inspection"
	case "pwd":
		return "runtime-validation"
	case "sleep 10", "sleep 60":
		return "background-worker"
	case "/bin/sh":
		return "shell-session"
	case "/bin/sh -c hostname && pwd && ls /":
		return "container-identity-check"
	case "/bin/sh -c echo hello from container":
		return "hello-container"
	case "/bin/sh -c echo starting && sleep 10 && echo finished":
		return "controlled-sleep"
	case "/bin/sh -c uname -a && id && ps":
		return "runtime-diagnostics"
	case "/bin/sh -c hostname && uname -a && ps":
		return "runtime-diagnostics"
	case "/bin/sh -c ls -la / && ls -la /bin":
		return "rootfs-inspection"
	case "/bin/sh -c echo fallback validation":
		return "fallback-demo"
	case "minidock-postgres-demo":
		return "postgres-demo"
	default:
		if strings.Contains(normalized, "sleep") {
			return "diagnostics-run"
		}
		return "runtime-task"
	}
}

func ExecutiveSummary(command string, args []string) string {
	normalized := NormalizeCommand(command, args)

	switch normalized {
	case "echo hello":
		return "Rotina de verificação de saída para validação rápida do runtime."
	case "ls":
		return "Workload de inspeção básica do sistema de arquivos local."
	case "ls -la":
		return "Workload de inspeção detalhada do sistema de arquivos com metadados."
	case "pwd":
		return "Validação de contexto de execução e diretório ativo do processo."
	case "sleep 10":
		return "Workload temporária de espera curta para teste de ciclo de vida."
	case "sleep 60":
		return "Workload temporária de validação ociosa com janela de observação estendida."
	case "/bin/sh":
		return "Sessão de diagnóstico shell não interativa para inspeção controlada."
	case "/bin/sh -c hostname && pwd && ls /":
		return "Inspeção de identidade e filesystem em ambiente isolado para validar container-linux."
	case "/bin/sh -c echo hello from container":
		return "Teste simples de execução e logs em runtime isolado ou fallback local."
	case "/bin/sh -c echo starting && sleep 10 && echo finished":
		return "Validação de transição Running -> Completed com janela para interrupção manual."
	case "/bin/sh -c uname -a && id && ps":
		return "Diagnóstico do ambiente para evidenciar contexto de isolamento e processo principal."
	case "/bin/sh -c hostname && uname -a && ps":
		return "Diagnóstico técnico com hostname, kernel e processos para validar modo efetivo."
	case "/bin/sh -c ls -la / && ls -la /bin":
		return "Inspeção do rootfs para comprovar montagem e estrutura interna do ambiente."
	case "/bin/sh -c echo fallback validation":
		return "Workload de validação de fallback automático quando container-linux não está disponível."
	case "minidock-postgres-demo":
		return "Demonstração de workload stateful com inicialização de PostgreSQL, readiness e evidências operacionais."
	default:
		return "Execução local para diagnóstico operacional em ambiente controlado."
	}
}

func RiskClassification(command string, args []string) RiskLevel {
	normalized := strings.ToLower(NormalizeCommand(command, args))

	for _, pattern := range destructivePatterns {
		if strings.Contains(normalized, pattern) {
			return RiskRisky
		}
	}

	safeCommands := map[string]struct{}{
		"echo hello":                           {},
		"ls":                                   {},
		"ls -la":                               {},
		"pwd":                                  {},
		"sleep 10":                             {},
		"sleep 60":                             {},
		"/bin/sh -c hostname && pwd && ls /":   {},
		"/bin/sh -c echo hello from container": {},
		"/bin/sh -c echo starting && sleep 10 && echo finished": {},
		"/bin/sh -c uname -a && id && ps":                       {},
		"/bin/sh -c hostname && uname -a && ps":                 {},
		"/bin/sh -c ls -la / && ls -la /bin":                    {},
		"/bin/sh -c echo fallback validation":                   {},
		"minidock-postgres-demo":                                {},
	}
	if _, ok := safeCommands[normalized]; ok {
		return RiskSafe
	}

	if normalized == "/bin/sh" || strings.HasPrefix(normalized, "/bin/sh -c") || strings.Contains(normalized, " sh ") {
		return RiskReview
	}

	return RiskReview
}

func InsightsFor(w Workload) []string {
	insights := make([]string, 0, 6)

	switch w.RiskLevel {
	case RiskSafe:
		insights = append(insights, "Esta workload foi classificada como de baixo risco operacional.")
	case RiskReview:
		insights = append(insights, "Esta workload merece revisão humana antes de repetição em escala.")
	case RiskRisky:
		insights = append(insights, "Foi detectado um padrão potencialmente destrutivo na linha de comando.")
	}

	if w.Mode == ModeContainerLinux {
		insights = append(insights, "Esta workload foi executada em isolamento Linux com rootfs dedicado.")
	}
	if w.FallbackApplied {
		insights = append(insights, "O host atual não suportou isolamento avançado; a execução foi redirecionada para processo local.")
	}
	if isDatabaseWorkload(w.WorkloadType) {
		insights = append(insights, "A workload PostgreSQL demonstra suporte a cenário stateful com observabilidade operacional.")
		if strings.EqualFold(w.Runtime.ReadinessState, "ready") {
			insights = append(insights, "Foram observados sinais consistentes de readiness do serviço de banco.")
		}
		if w.Runtime.Port > 0 {
			insights = append(insights, fmt.Sprintf("Porta operacional registrada: %d.", w.Runtime.Port))
		}
	}

	switch w.Status {
	case StatusRunning:
		insights = append(insights, "A execução está ativa e pode ser monitorada em tempo real pelos logs.")
	case StatusCompleted:
		insights = append(insights, "A execução foi concluída e os resultados já podem ser consolidados.")
	case StatusFailed:
		insights = append(insights, "A execução falhou e recomenda-se revisão do comando e contexto local.")
	case StatusStopped:
		insights = append(insights, "A execução foi interrompida manualmente para controle operacional.")
	default:
		insights = append(insights, "A workload foi registrada e aguarda ciclo completo de execução.")
	}

	normalized := NormalizeCommand(w.Command, w.Args)
	if strings.Contains(normalized, "sleep") {
		insights = append(insights, "O processo aparenta servir para validação temporal e testes de observabilidade.")
	}
	if normalized == "minidock-postgres-demo" {
		insights = append(insights, "Esta demonstração reforça a completude do produto com workload de banco de dados.")
	}
	if strings.Contains(normalized, "ls") || normalized == "pwd" {
		insights = append(insights, "A finalidade sugere inspeção diagnóstica de ambiente e estrutura local.")
	}

	if w.DurationMs > 0 {
		insights = append(insights, fmt.Sprintf("Duração observada: %.2fs, útil para benchmark de rotina.", float64(w.DurationMs)/1000))
	}

	if len(insights) < 2 {
		insights = append(insights, "A execução mantém perfil compatível com um cenário de demonstração técnica.")
	}
	if len(insights) > 4 {
		insights = insights[:4]
	}

	return insights
}

func SuggestedActionFor(w Workload) string {
	switch {
	case w.RiskLevel == RiskRisky:
		return "Revisar intenção e comando antes de qualquer nova execução."
	case isDatabaseWorkload(w.WorkloadType) && w.Status == StatusRunning:
		return "Acompanhar logs de readiness; encerrar após validar porta, data dir e status operacional."
	case isDatabaseWorkload(w.WorkloadType) && w.Status == StatusCompleted:
		return "Consolidar evidências de inicialização do PostgreSQL e remover workload ao finalizar a apresentação."
	case w.FallbackApplied:
		return "Validar capabilities do host e repetir em Linux compatível para demonstrar isolamento real."
	case w.Mode == ModeContainerLinux && w.Status == StatusCompleted:
		return "Consolidar evidências do rootfs/hostname isolados e remover workload após a demo."
	case w.RiskLevel == RiskReview:
		return "Validar contexto manualmente antes de reexecutar esta workload."
	case w.Status == StatusRunning:
		return "Monitorar por instantes e encerrar quando a validação for concluída."
	case w.Status == StatusCompleted:
		return "Seguro remover após consolidar os resultados da execução."
	case w.Status == StatusFailed:
		return "Ajustar comando e reexecutar com acompanhamento de logs."
	default:
		return "Manter apenas para depuração temporária e remover ao final da análise."
	}
}

func GlobalExecutiveSummary(workloads []Workload) []string {
	if len(workloads) == 0 {
		return []string{
			"Nenhuma workload ativa no momento.",
			"Inicie uma execução para habilitar leitura executiva de risco e desempenho.",
		}
	}

	var reviewCount, riskyCount, runningCount int
	var totalDuration int64
	var completedWithDuration int64
	diagnosticCount := 0
	containerCount := 0
	fallbackCount := 0
	databaseCount := 0

	for _, w := range workloads {
		if w.RiskLevel == RiskReview {
			reviewCount++
		}
		if w.RiskLevel == RiskRisky {
			riskyCount++
		}
		if w.Status == StatusRunning {
			runningCount++
		}
		if w.DurationMs > 0 {
			totalDuration += w.DurationMs
			completedWithDuration++
		}
		if w.Mode == ModeContainerLinux {
			containerCount++
		}
		if w.FallbackApplied {
			fallbackCount++
		}
		if isDatabaseWorkload(w.WorkloadType) {
			databaseCount++
		}

		normalized := NormalizeCommand(w.Command, w.Args)
		if strings.Contains(normalized, "ls") || strings.Contains(normalized, "pwd") || strings.Contains(normalized, "echo") {
			diagnosticCount++
		}
	}

	lines := []string{}
	if reviewCount > 0 {
		lines = append(lines, fmt.Sprintf("%d workloads requerem revisão manual.", reviewCount))
	}
	if riskyCount > 0 {
		lines = append(lines, fmt.Sprintf("%d workload(s) foram classificadas como arriscadas.", riskyCount))
	}
	if runningCount > 0 {
		lines = append(lines, fmt.Sprintf("%d execução(ões) estão ativas neste instante.", runningCount))
	}
	if containerCount > 0 {
		lines = append(lines, fmt.Sprintf("%d workload(s) rodaram em isolamento Linux real (container-linux).", containerCount))
	}
	if fallbackCount > 0 {
		lines = append(lines, fmt.Sprintf("Fallback automático foi aplicado em %d workload(s).", fallbackCount))
	}
	if databaseCount > 0 {
		lines = append(lines, fmt.Sprintf("%d workload(s) stateful de banco de dados foram processadas na sessão.", databaseCount))
	}
	if diagnosticCount >= len(workloads)/2 {
		lines = append(lines, "A maioria das execuções mantém caráter diagnóstico e de validação local.")
	}
	if completedWithDuration > 0 {
		avg := float64(totalDuration) / float64(completedWithDuration) / 1000
		lines = append(lines, fmt.Sprintf("Tempo médio recente de execução: %.1fs.", avg))
	}

	if len(lines) == 0 {
		lines = append(lines, "Perfil operacional estável, sem alertas relevantes no momento.")
	}
	if len(lines) > 4 {
		lines = lines[:4]
	}

	return lines
}

func NormalizeCommand(command string, args []string) string {
	joined := strings.TrimSpace(strings.Join(append([]string{strings.TrimSpace(command)}, args...), " "))
	return strings.Join(strings.Fields(joined), " ")
}
