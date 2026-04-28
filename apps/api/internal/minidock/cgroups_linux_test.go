//go:build linux

package minidock

import "testing"

func TestNormalizeCPUMax(t *testing.T) {
	if got := normalizeCPUMax(""); got != "200000 100000" {
		t.Fatalf("esperado fallback cpu.max, recebido %s", got)
	}
	if got := normalizeCPUMax("50000"); got != "50000 100000" {
		t.Fatalf("esperado normalização com período padrão, recebido %s", got)
	}
	if got := normalizeCPUMax("80000 200000 999"); got != "80000 200000" {
		t.Fatalf("esperado apenas quota/período, recebido %s", got)
	}
}

func TestParseCPUMaxForV1(t *testing.T) {
	quota, period := parseCPUMaxForV1("75000 150000")
	if quota != "75000" || period != "150000" {
		t.Fatalf("quota/período inesperados: %s %s", quota, period)
	}
}

func TestSanitizeCgroupGroupName(t *testing.T) {
	if got := sanitizeCgroupGroupName("wk_1/demo"); got != "wk_1-demo" {
		t.Fatalf("nome sanitizado inesperado: %s", got)
	}
}
