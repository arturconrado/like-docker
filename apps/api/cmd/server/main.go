package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"minidock-executive-ai/apps/api/internal/minidock"
)

func main() {
	if minidock.HandleContainerInitCommand() {
		return
	}

	mode := minidock.RuntimeMode(strings.TrimSpace(getEnv("MINIDOCK_RUNTIME_MODE", string(minidock.ModeProcessLocal))))
	manager := minidock.NewManager(mode)

	if strings.ToLower(getEnv("MINIDOCK_SEED_DEMO", "true")) != "false" {
		manager.SeedDemo(false)
	}

	server := minidock.NewAPIServer(manager)
	port := getEnv("API_PORT", "8080")

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("MiniDock API ativa em http://localhost:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("falha ao iniciar API: %v", err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("erro ao encerrar API: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
