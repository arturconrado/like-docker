package minidock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type APIServer struct {
	manager *Manager
}

func NewAPIServer(manager *Manager) *APIServer {
	return &APIServer{manager: manager}
}

func (s *APIServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/workloads", s.handleWorkloads)
	mux.HandleFunc("/api/workloads/", s.handleWorkloadByID)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/stream", s.handleStream)
	mux.HandleFunc("/api/demo/seed", s.handleDemoSeed)
	mux.HandleFunc("/api/demos", s.handleDemos)
	mux.HandleFunc("/api/demos/", s.handleDemoByID)
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/capabilities", s.handleCapabilities)

	return withCORS(mux)
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	writeJSON(w, http.StatusOK, s.manager.Health())
}

func (s *APIServer) handleWorkloads(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.manager.ListWorkloads())
	case http.MethodPost:
		var req CreateWorkloadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "payload inválido")
			return
		}
		created, err := s.manager.CreateWorkload(req)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, ErrInvalidCommand) {
				status = http.StatusBadRequest
			}
			writeError(w, status, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
	}
}

func (s *APIServer) handleWorkloadByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/workloads/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeError(w, http.StatusNotFound, "workload não encontrada")
		return
	}

	if strings.HasSuffix(path, "/stop") {
		id := strings.TrimSuffix(path, "/stop")
		id = strings.Trim(id, "/")
		s.handleStopWorkload(w, r, id)
		return
	}
	if strings.HasSuffix(path, "/logs") {
		id := strings.TrimSuffix(path, "/logs")
		id = strings.Trim(id, "/")
		s.handleLogs(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		workload, ok := s.manager.GetWorkload(path)
		if !ok {
			writeError(w, http.StatusNotFound, "workload não encontrada")
			return
		}
		writeJSON(w, http.StatusOK, workload)
	case http.MethodDelete:
		if err := s.manager.DeleteWorkload(path); err != nil {
			if errors.Is(err, ErrWorkloadNotFound) {
				writeError(w, http.StatusNotFound, "workload não encontrada")
				return
			}
			writeError(w, http.StatusInternalServerError, "erro ao remover workload")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
	}
}

func (s *APIServer) handleStopWorkload(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	updated, err := s.manager.StopWorkload(id)
	if err != nil {
		if errors.Is(err, ErrWorkloadNotFound) {
			writeError(w, http.StatusNotFound, "workload não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao interromper workload")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *APIServer) handleLogs(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	logs, err := s.manager.GetLogs(id)
	if err != nil {
		if errors.Is(err, ErrWorkloadNotFound) {
			writeError(w, http.StatusNotFound, "workload não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao consultar logs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workloadId": id, "logs": logs})
}

func (s *APIServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	writeJSON(w, http.StatusOK, s.manager.ListEvents())
}

func (s *APIServer) handleDemoSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	workloads := s.manager.SeedDemo(true)
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "dados de demonstração carregados",
		"workloads": workloads,
	})
}

func (s *APIServer) handleDemos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	writeJSON(w, http.StatusOK, s.manager.ListDemos())
}

func (s *APIServer) handleDemoByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/demos/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeError(w, http.StatusNotFound, "demo não encontrada")
		return
	}

	if strings.HasSuffix(path, "/run") {
		id := strings.Trim(strings.TrimSuffix(path, "/run"), "/")
		s.handleRunDemo(w, r, id)
		return
	}
	if strings.HasSuffix(path, "/validate") {
		id := strings.Trim(strings.TrimSuffix(path, "/validate"), "/")
		s.handleValidateDemo(w, r, id)
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	demo, ok := s.manager.GetDemo(path)
	if !ok {
		writeError(w, http.StatusNotFound, "demo não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, demo)
}

func (s *APIServer) handleRunDemo(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	result, err := s.manager.RunDemo(id)
	if err != nil {
		if errors.Is(err, ErrDemoNotFound) {
			writeError(w, http.StatusNotFound, "demo não encontrada")
			return
		}
		if errors.Is(err, ErrInvalidCommand) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao executar demo")
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *APIServer) handleValidateDemo(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	result, err := s.manager.ValidateDemo(id)
	if err != nil {
		if errors.Is(err, ErrDemoNotFound) {
			writeError(w, http.StatusNotFound, "demo não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao validar demo")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *APIServer) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	writeJSON(w, http.StatusOK, s.manager.ExecutiveSummary())
}

func (s *APIServer) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	writeJSON(w, http.StatusOK, s.manager.Capabilities())
}

func (s *APIServer) handleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming não suportado")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := s.manager.SubscribeEvents()
	defer unsubscribe()

	_, _ = fmt.Fprintf(w, "event: ready\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, open := <-events:
			if !open {
				return
			}
			payload, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, payload)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "erro ao serializar resposta", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
