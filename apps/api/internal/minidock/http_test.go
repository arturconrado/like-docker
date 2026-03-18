package minidock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPCreateListGetLogsAndEvents(t *testing.T) {
	server, manager := newTestServer()
	_ = manager

	created := createWorkload(t, server, map[string]any{
		"command": "echo hello",
		"mode":    string(ModeProcessLocal),
	})

	waitForStatus(t, server, created.ID, []WorkloadStatus{StatusCompleted, StatusFailed})

	resp, err := http.Get(server.URL + "/api/workloads")
	if err != nil {
		t.Fatalf("erro ao listar workloads: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado ao listar workloads: %d", resp.StatusCode)
	}

	var workloads []Workload
	if err := json.NewDecoder(resp.Body).Decode(&workloads); err != nil {
		t.Fatalf("erro ao decodificar workloads: %v", err)
	}
	if len(workloads) == 0 {
		t.Fatal("lista de workloads não deveria estar vazia")
	}

	logsResp, err := http.Get(server.URL + "/api/workloads/" + created.ID + "/logs")
	if err != nil {
		t.Fatalf("erro ao buscar logs: %v", err)
	}
	defer logsResp.Body.Close()
	if logsResp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado ao buscar logs: %d", logsResp.StatusCode)
	}

	var logsPayload struct {
		Logs []string `json:"logs"`
	}
	if err := json.NewDecoder(logsResp.Body).Decode(&logsPayload); err != nil {
		t.Fatalf("erro ao decodificar logs: %v", err)
	}
	if len(logsPayload.Logs) == 0 {
		t.Fatal("esperava logs para a workload echo hello")
	}

	eventsResp, err := http.Get(server.URL + "/api/events")
	if err != nil {
		t.Fatalf("erro ao buscar eventos: %v", err)
	}
	defer eventsResp.Body.Close()
	if eventsResp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado ao buscar eventos: %d", eventsResp.StatusCode)
	}

	var events []Event
	if err := json.NewDecoder(eventsResp.Body).Decode(&events); err != nil {
		t.Fatalf("erro ao decodificar eventos: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("esperava ao menos um evento")
	}
}

func TestHTTPStopAndDelete(t *testing.T) {
	server, _ := newTestServer()

	created := createWorkload(t, server, map[string]any{
		"command": "sleep 60",
		"mode":    string(ModeProcessLocal),
	})

	waitForStatus(t, server, created.ID, []WorkloadStatus{StatusRunning, StatusStopped, StatusCompleted})

	stopReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/workloads/"+created.ID+"/stop", nil)
	stopResp, err := http.DefaultClient.Do(stopReq)
	if err != nil {
		t.Fatalf("erro ao interromper workload: %v", err)
	}
	defer stopResp.Body.Close()
	if stopResp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado ao interromper workload: %d", stopResp.StatusCode)
	}

	waitForStatus(t, server, created.ID, []WorkloadStatus{StatusStopped})

	deleteReq, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/workloads/"+created.ID, nil)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("erro ao remover workload: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("status inesperado ao remover workload: %d", deleteResp.StatusCode)
	}

	getResp, err := http.Get(server.URL + "/api/workloads/" + created.ID)
	if err != nil {
		t.Fatalf("erro ao buscar workload removida: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("status esperado 404 para workload removida, recebido %d", getResp.StatusCode)
	}
}

func TestContainerModeFallback(t *testing.T) {
	t.Setenv("MINIDOCK_CONTAINER_ROOTFS", "/tmp/minidock-rootfs-nao-existe")
	server, _ := newTestServer()

	created := createWorkload(t, server, map[string]any{
		"command": "/bin/sh",
		"args":    []string{"-c", "echo fallback validation"},
		"mode":    string(ModeContainerLinux),
	})

	if created.RequestedMode != ModeContainerLinux {
		t.Fatalf("requested mode esperado container-linux, recebido %s", created.RequestedMode)
	}
	if created.Mode != ModeProcessLocal {
		t.Fatalf("modo deveria fazer fallback para processo-local, recebido %s", created.Mode)
	}
	if !created.FallbackApplied {
		t.Fatal("esperava fallback aplicado para container-linux sem rootfs disponível")
	}

	logs := waitForLogs(t, server, created.ID)
	foundFallback := false
	for _, line := range logs {
		if bytes.Contains(bytes.ToLower([]byte(line)), []byte("fallback")) {
			foundFallback = true
			break
		}
	}
	if !foundFallback {
		t.Fatal("esperava log indicando fallback de container-linux")
	}
}

func TestCapabilitiesEndpoint(t *testing.T) {
	server, _ := newTestServer()

	resp, err := http.Get(server.URL + "/api/capabilities")
	if err != nil {
		t.Fatalf("erro ao buscar capabilities: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado no endpoint de capabilities: %d", resp.StatusCode)
	}

	var caps HostCapabilities
	if err := json.NewDecoder(resp.Body).Decode(&caps); err != nil {
		t.Fatalf("erro ao decodificar capabilities: %v", err)
	}
	if caps.OS == "" {
		t.Fatal("campo os não pode ser vazio")
	}
	if caps.RecommendedMode == "" {
		t.Fatal("recommendedMode não pode ser vazio")
	}
}

func TestDemoSeedEndpoint(t *testing.T) {
	server, _ := newTestServer()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/demo/seed", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("erro ao chamar seed demo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado no seed demo: %d", resp.StatusCode)
	}

	var payload struct {
		Workloads []Workload `json:"workloads"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("erro ao decodificar seed demo: %v", err)
	}
	if len(payload.Workloads) < 4 {
		t.Fatalf("esperava ao menos 4 workloads de demo, recebido %d", len(payload.Workloads))
	}
}

func TestDemosEndpoints(t *testing.T) {
	server, _ := newTestServer()

	resp, err := http.Get(server.URL + "/api/demos")
	if err != nil {
		t.Fatalf("erro ao listar demos: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado ao listar demos: %d", resp.StatusCode)
	}

	var demos []DemoDefinition
	if err := json.NewDecoder(resp.Body).Decode(&demos); err != nil {
		t.Fatalf("erro ao decodificar demos: %v", err)
	}
	if len(demos) < 5 {
		t.Fatalf("esperava ao menos 5 demos, recebido %d", len(demos))
	}

	detailResp, err := http.Get(server.URL + "/api/demos/postgres-demo")
	if err != nil {
		t.Fatalf("erro ao buscar detalhes da demo postgres: %v", err)
	}
	defer detailResp.Body.Close()
	if detailResp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado ao detalhar demo postgres: %d", detailResp.StatusCode)
	}

	var detail DemoDefinition
	if err := json.NewDecoder(detailResp.Body).Decode(&detail); err != nil {
		t.Fatalf("erro ao decodificar detalhe da demo: %v", err)
	}
	if detail.ID != "postgres-demo" {
		t.Fatalf("demo esperada postgres-demo, recebido %s", detail.ID)
	}

	runReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/demos/postgres-demo/run", nil)
	runResp, err := http.DefaultClient.Do(runReq)
	if err != nil {
		t.Fatalf("erro ao executar demo postgres: %v", err)
	}
	defer runResp.Body.Close()
	if runResp.StatusCode != http.StatusCreated {
		t.Fatalf("status inesperado ao executar demo postgres: %d", runResp.StatusCode)
	}

	var runPayload DemoRunResponse
	if err := json.NewDecoder(runResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("erro ao decodificar execução da demo: %v", err)
	}
	if runPayload.Demo.ID != "postgres-demo" {
		t.Fatalf("demo de execução esperada postgres-demo, recebido %s", runPayload.Demo.ID)
	}
	if runPayload.Workload.ID == "" {
		t.Fatal("workload da execução da demo não pode ser vazia")
	}

	time.Sleep(150 * time.Millisecond)

	validateResp, err := http.Get(server.URL + "/api/demos/postgres-demo/validate")
	if err != nil {
		t.Fatalf("erro ao validar demo postgres: %v", err)
	}
	defer validateResp.Body.Close()
	if validateResp.StatusCode != http.StatusOK {
		t.Fatalf("status inesperado na validação da demo postgres: %d", validateResp.StatusCode)
	}

	var validation DemoValidation
	if err := json.NewDecoder(validateResp.Body).Decode(&validation); err != nil {
		t.Fatalf("erro ao decodificar validação da demo: %v", err)
	}
	if validation.DemoID != "postgres-demo" {
		t.Fatalf("demoId esperado postgres-demo, recebido %s", validation.DemoID)
	}
	if validation.WorkloadID == "" {
		t.Fatal("workloadId de validação não pode ser vazio")
	}
	if len(validation.Signals) == 0 {
		t.Fatal("validação deve retornar sinais")
	}
}

func newTestServer() (*httptest.Server, *Manager) {
	manager := NewManager(ModeProcessLocal)
	api := NewAPIServer(manager)
	server := httptest.NewServer(api.Handler())
	return server, manager
}

func createWorkload(t *testing.T, server *httptest.Server, body map[string]any) Workload {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("erro ao serializar payload: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/workloads", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("erro ao criar workload: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status inesperado ao criar workload: %d", resp.StatusCode)
	}

	var created Workload
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("erro ao decodificar workload criada: %v", err)
	}
	if created.ID == "" {
		t.Fatal("id da workload criada não pode ser vazio")
	}
	return created
}

func waitForStatus(t *testing.T, server *httptest.Server, id string, accepted []WorkloadStatus) Workload {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(server.URL + "/api/workloads/" + id)
		if err != nil {
			t.Fatalf("erro ao consultar workload: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			var item Workload
			if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
				resp.Body.Close()
				t.Fatalf("erro ao decodificar workload: %v", err)
			}
			resp.Body.Close()
			for _, status := range accepted {
				if item.Status == status {
					return item
				}
			}
		} else {
			resp.Body.Close()
		}
		time.Sleep(120 * time.Millisecond)
	}
	t.Fatalf("timeout aguardando status em %v", accepted)
	return Workload{}
}

func waitForLogs(t *testing.T, server *httptest.Server, id string) []string {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(server.URL + "/api/workloads/" + id + "/logs")
		if err != nil {
			t.Fatalf("erro ao consultar logs: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			var payload struct {
				Logs []string `json:"logs"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				resp.Body.Close()
				t.Fatalf("erro ao decodificar logs: %v", err)
			}
			resp.Body.Close()
			if len(payload.Logs) > 0 {
				return payload.Logs
			}
		} else {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("timeout aguardando logs")
	return nil
}
