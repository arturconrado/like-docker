package minidock

import "testing"

func TestSmartName(t *testing.T) {
	tests := []struct {
		command  string
		args     []string
		expected string
	}{
		{command: "echo", args: []string{"hello"}, expected: "output-check"},
		{command: "ls", args: []string{"-la"}, expected: "filesystem-inspection"},
		{command: "/bin/sh", expected: "shell-session"},
	}

	for _, tc := range tests {
		if got := SmartName(tc.command, tc.args); got != tc.expected {
			t.Fatalf("SmartName(%q) = %q, esperado %q", NormalizeCommand(tc.command, tc.args), got, tc.expected)
		}
	}
}

func TestExecutiveSummary(t *testing.T) {
	summary := ExecutiveSummary("sleep", []string{"60"})
	if summary == "" {
		t.Fatal("resumo executivo não pode ser vazio")
	}
}

func TestRiskClassification(t *testing.T) {
	if risk := RiskClassification("echo", []string{"hello"}); risk != RiskSafe {
		t.Fatalf("risk esperado Safe, recebido %s", risk)
	}
	if risk := RiskClassification("/bin/sh", nil); risk != RiskReview {
		t.Fatalf("risk esperado Review, recebido %s", risk)
	}
	if risk := RiskClassification("rm", []string{"-rf", "/"}); risk != RiskRisky {
		t.Fatalf("risk esperado Risky, recebido %s", risk)
	}
}

func TestInsightsAndSuggestedAction(t *testing.T) {
	w := Workload{
		Command:    "sleep",
		Args:       []string{"10"},
		RiskLevel:  RiskSafe,
		Status:     StatusCompleted,
		DurationMs: 10000,
	}
	insights := InsightsFor(w)
	if len(insights) < 2 || len(insights) > 4 {
		t.Fatalf("insights devem ter entre 2 e 4 itens, recebido %d", len(insights))
	}
	action := SuggestedActionFor(w)
	if action == "" {
		t.Fatal("ação sugerida não pode ser vazia")
	}
}

func TestGlobalExecutiveSummary(t *testing.T) {
	workloads := []Workload{
		{RiskLevel: RiskReview, Status: StatusRunning, Command: "/bin/sh"},
		{RiskLevel: RiskSafe, Status: StatusCompleted, DurationMs: 1200, Command: "ls", Args: []string{"-la"}},
	}
	lines := GlobalExecutiveSummary(workloads)
	if len(lines) == 0 {
		t.Fatal("resumo global não pode ser vazio")
	}
}
