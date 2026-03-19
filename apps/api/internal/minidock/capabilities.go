package minidock

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type HostCapabilities struct {
	OS                         string              `json:"os"`
	IsLinux                    bool                `json:"isLinux"`
	SupportsProcessLocal       bool                `json:"supportsProcessLocal"`
	SupportsContainers         bool                `json:"supportsContainers"`
	SupportsNamespaces         bool                `json:"supportsNamespaces"`
	SupportsPivotRoot          bool                `json:"supportsPivotRoot"`
	RootfsAvailable            bool                `json:"rootfsAvailable"`
	RootfsPath                 string              `json:"rootfsPath,omitempty"`
	HasRootPrivileges          bool                `json:"hasRootPrivileges"`
	PostgresLocalAvailable     bool                `json:"postgresLocalAvailable"`
	PostgresContainerAvailable bool                `json:"postgresContainerAvailable"`
	SupportsPostgresDemo       bool                `json:"supportsPostgresDemo"`
	RecommendedMode            RuntimeMode         `json:"recommendedMode"`
	PostgresBinariesAvailable  bool                `json:"postgresBinariesAvailable"`
	PostgresBinaryPaths        PostgresBinaryPaths `json:"postgresBinaryPaths"`
	CanCreateTempDir           bool                `json:"canCreateTempDir"`
	CanAllocatePort            bool                `json:"canAllocatePort"`
	CanRunPostgresDemo         bool                `json:"canRunPostgresDemo"`
	RecommendedPostgresMode    PostgresDemoMode    `json:"recommendedPostgresMode"`
	Notes                      []string            `json:"notes"`
}

func DetectHostCapabilities(rootfsHint string) HostCapabilities {
	osName := runtime.GOOS
	isLinux := osName == "linux"
	hasRoot := os.Geteuid() == 0

	rootfsPath := resolveRootfsPath(rootfsHint)
	rootfsAvailable := rootfsPath != "" && isUsableRootfs(rootfsPath)
	postgresBinaryPaths := detectPostgresBinaryPaths()
	postgresBinariesAvailable := postgresBinaryPaths.Initdb != "" &&
		postgresBinaryPaths.Postgres != "" &&
		postgresBinaryPaths.PGIsReady != ""
	canCreateTempDir := canCreateTemporaryDirectory()
	canAllocatePort := canAllocateLocalPort()
	canRunPostgresDemo := isLinux && !hasRoot && postgresBinariesAvailable && canCreateTempDir && canAllocatePort
	postgresLocalAvailable := postgresBinariesAvailable
	postgresContainerAvailable := hasRootfsBinary(rootfsPath, "postgres") && hasRootfsBinary(rootfsPath, "initdb")
	supportsPostgresDemo := true
	supportsNamespaces := isLinux &&
		fileExists("/proc/self/ns/mnt") &&
		fileExists("/proc/self/ns/pid") &&
		fileExists("/proc/self/ns/uts")
	supportsPivotRoot := isLinux && hasRoot && fileExists("/proc/self/ns/mnt")

	supportsContainers := isLinux && supportsNamespaces && rootfsAvailable && hasRoot

	notes := make([]string, 0, 4)
	if !isLinux {
		notes = append(notes, "Host não é Linux; container-linux fica indisponível e o sistema usa fallback automático.")
	}
	if hasRoot {
		notes = append(notes, "Processo atual está em root; initdb local do PostgreSQL exige usuário não-root para o caminho processo-local-real.")
	}
	if isLinux && !hasRoot {
		notes = append(notes, "Sem privilégios root para chroot/namespaces; container-linux requer sudo neste MVP.")
	}
	if isLinux && !supportsNamespaces {
		notes = append(notes, "Namespaces Linux mínimos não foram detectados no host.")
	}
	if !rootfsAvailable {
		notes = append(notes, "Rootfs de demonstração não disponível; use scripts/prepare-rootfs.sh para habilitar container-linux.")
	}
	if supportsContainers {
		notes = append(notes, "Host apto para execução container-linux em modo avançado.")
	}
	if postgresBinariesAvailable {
		notes = append(notes, "Binários locais de PostgreSQL detectados para execução processo-local.")
	}
	if postgresContainerAvailable {
		notes = append(notes, "Rootfs contém binários PostgreSQL para tentativa em container-linux.")
	}
	if !canCreateTempDir {
		notes = append(notes, "Não foi possível validar criação de diretório temporário para PGDATA local.")
	}
	if !canAllocatePort {
		notes = append(notes, "Não foi possível reservar porta TCP local para o PostgreSQL Demo.")
	}
	if !postgresBinariesAvailable && !postgresContainerAvailable {
		notes = append(notes, "PostgreSQL real não foi detectado; a demonstração PostgreSQL usará fallback para modo demo quando necessário.")
	}
	if canRunPostgresDemo {
		notes = append(notes, "O PostgreSQL Demo pode rodar em Linux real com binários locais do host.")
	}

	recommended := ModeProcessLocal
	if supportsContainers {
		recommended = ModeContainerLinux
	}
	recommendedPostgresMode := PostgresModeDemo
	switch {
	case canRunPostgresDemo:
		recommendedPostgresMode = PostgresModeProcessLocalReal
	case supportsContainers && postgresContainerAvailable:
		recommendedPostgresMode = PostgresModeContainerLinux
	default:
		recommendedPostgresMode = PostgresModeDemo
	}

	return HostCapabilities{
		OS:                         osName,
		IsLinux:                    isLinux,
		SupportsProcessLocal:       true,
		SupportsContainers:         supportsContainers,
		SupportsNamespaces:         supportsNamespaces,
		SupportsPivotRoot:          supportsPivotRoot,
		RootfsAvailable:            rootfsAvailable,
		RootfsPath:                 rootfsPath,
		HasRootPrivileges:          hasRoot,
		PostgresLocalAvailable:     postgresLocalAvailable,
		PostgresContainerAvailable: postgresContainerAvailable,
		SupportsPostgresDemo:       supportsPostgresDemo,
		RecommendedMode:            recommended,
		PostgresBinariesAvailable:  postgresBinariesAvailable,
		PostgresBinaryPaths:        postgresBinaryPaths,
		CanCreateTempDir:           canCreateTempDir,
		CanAllocatePort:            canAllocatePort,
		CanRunPostgresDemo:         canRunPostgresDemo,
		RecommendedPostgresMode:    recommendedPostgresMode,
		Notes:                      notes,
	}
}

func resolveRootfsPath(rootfsHint string) string {
	trimmed := strings.TrimSpace(rootfsHint)
	candidates := []string{}
	if trimmed != "" {
		candidates = append(candidates, trimmed)
	}

	candidates = append(candidates,
		"./examples/rootfs/demo",
		"./examples/rootfs/busybox",
		"../examples/rootfs/demo",
		"../examples/rootfs/busybox",
		"../../examples/rootfs/demo",
		"../../examples/rootfs/busybox",
	)

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if isUsableRootfs(abs) {
			return abs
		}
	}

	if trimmed == "" {
		return ""
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return trimmed
	}
	return abs
}

func isUsableRootfs(rootfsPath string) bool {
	if rootfsPath == "" {
		return false
	}
	info, err := os.Stat(rootfsPath)
	if err != nil || !info.IsDir() {
		return false
	}
	return fileExists(filepath.Join(rootfsPath, "bin", "sh"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func detectPostgresBinaryPaths() PostgresBinaryPaths {
	return PostgresBinaryPaths{
		Initdb:    lookupBinary("initdb"),
		Postgres:  lookupBinary("postgres"),
		PGIsReady: lookupBinary("pg_isready"),
	}
}

func lookupBinary(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func canCreateTemporaryDirectory() bool {
	dir, err := os.MkdirTemp("", "minidock-capabilities-*")
	if err != nil {
		return false
	}
	_ = os.RemoveAll(dir)
	return true
}

func canAllocateLocalPort() bool {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

func hasRootfsBinary(rootfsPath, name string) bool {
	if rootfsPath == "" {
		return false
	}
	candidates := []string{
		filepath.Join(rootfsPath, "usr", "bin", name),
		filepath.Join(rootfsPath, "bin", name),
	}
	for _, candidate := range candidates {
		if fileExists(candidate) {
			return true
		}
	}
	return false
}
