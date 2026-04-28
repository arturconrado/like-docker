package minidock

import "time"

type WorkloadStatus string

const (
	StatusPending   WorkloadStatus = "Pending"
	StatusPreparing WorkloadStatus = "Preparing"
	StatusStarting  WorkloadStatus = "Starting"
	StatusRunning   WorkloadStatus = "Running"
	StatusCompleted WorkloadStatus = "Completed"
	StatusFailed    WorkloadStatus = "Failed"
	StatusStopped   WorkloadStatus = "Stopped"
)

type RiskLevel string

const (
	RiskSafe   RiskLevel = "Safe"
	RiskReview RiskLevel = "Review"
	RiskRisky  RiskLevel = "Risky"
)

type RuntimeMode string

const (
	ModeDemo             RuntimeMode = "demo"
	ModeProcessLocal     RuntimeMode = "processo-local"
	ModeContainerLinux   RuntimeMode = "container-linux"
	ModeNamespaceRuntime RuntimeMode = "namespace-runtime" // alias legado
)

type PostgresDemoMode string

const (
	PostgresModeProcessLocalReal PostgresDemoMode = "processo-local-real"
	PostgresModeContainerLinux   PostgresDemoMode = "container-linux"
	PostgresModeDemo             PostgresDemoMode = "demo"
)

type PostgresBinaryPaths struct {
	Initdb    string `json:"initdb,omitempty"`
	Postgres  string `json:"postgres,omitempty"`
	PGIsReady string `json:"pgIsready,omitempty"`
}

type RuntimeMetadata struct {
	Engine            string `json:"engine"`
	Isolated          bool   `json:"isolated"`
	Rootfs            string `json:"rootfs,omitempty"`
	ContainerHostname string `json:"containerHostname,omitempty"`
	MainPID           int    `json:"mainPid,omitempty"`
	PivotRootApplied  bool   `json:"pivotRootApplied,omitempty"`
	CgroupPath        string `json:"cgroupPath,omitempty"`
	CgroupVersion     string `json:"cgroupVersion,omitempty"`
	WorkloadType      string `json:"workloadType,omitempty"`
	Port              int    `json:"port,omitempty"`
	DataDir           string `json:"dataDir,omitempty"`
	ReadinessState    string `json:"readinessState,omitempty"`
	ModeUsed          string `json:"modeUsed,omitempty"`
}

type Workload struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Command         string          `json:"command"`
	Args            []string        `json:"args"`
	WorkloadType    string          `json:"workloadType"`
	RequestedMode   RuntimeMode     `json:"requestedMode"`
	Summary         string          `json:"summary"`
	Status          WorkloadStatus  `json:"status"`
	RiskLevel       RiskLevel       `json:"riskLevel"`
	AIInsights      []string        `json:"aiInsights"`
	SuggestedAction string          `json:"suggestedAction"`
	StartedAt       *time.Time      `json:"startedAt"`
	FinishedAt      *time.Time      `json:"finishedAt"`
	DurationMs      int64           `json:"durationMs"`
	ExitCode        *int            `json:"exitCode"`
	Logs            []string        `json:"logs"`
	Mode            RuntimeMode     `json:"mode"`
	FallbackApplied bool            `json:"fallbackApplied"`
	FallbackReason  string          `json:"fallbackReason,omitempty"`
	Runtime         RuntimeMetadata `json:"runtime"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type EventSeverity string

const (
	SeverityInfo  EventSeverity = "info"
	SeverityWarn  EventSeverity = "warn"
	SeverityError EventSeverity = "error"
)

type Event struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	WorkloadID string        `json:"workloadId,omitempty"`
	Message    string        `json:"message"`
	Severity   EventSeverity `json:"severity"`
	CreatedAt  time.Time     `json:"createdAt"`
}

type CreateWorkloadRequest struct {
	Command            string      `json:"command"`
	Args               []string    `json:"args,omitempty"`
	Mode               RuntimeMode `json:"mode"`
	RequestedMode      RuntimeMode `json:"requestedMode,omitempty"`
	Name               string      `json:"name"`
	WorkloadType       string      `json:"workloadType,omitempty"`
	Port               int         `json:"port,omitempty"`
	DataDir            string      `json:"dataDir,omitempty"`
	FallbackReasonHint string      `json:"fallbackReasonHint,omitempty"`
}

type HealthResponse struct {
	Status      string      `json:"status"`
	RuntimeMode RuntimeMode `json:"runtimeMode"`
	UptimeMs    int64       `json:"uptimeMs"`
	Timestamp   time.Time   `json:"timestamp"`
}

type DashboardSummary struct {
	Lines []string `json:"lines"`
}

type DemoDefinition struct {
	ID                   string      `json:"id"`
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	Objective            string      `json:"objective"`
	PreferredMode        RuntimeMode `json:"preferredMode"`
	WorkloadType         string      `json:"workloadType"`
	Complexity           string      `json:"complexity"`
	RequiredCapabilities []string    `json:"requiredCapabilities"`
	ExpectedSignals      []string    `json:"expectedSignals"`
	Tags                 []string    `json:"tags"`
	Icon                 string      `json:"icon"`
	Port                 int         `json:"port,omitempty"`
	DataDir              string      `json:"dataDir,omitempty"`
}

type DemoRunResponse struct {
	Demo              DemoDefinition `json:"demo"`
	Workload          Workload       `json:"workload"`
	ExecutionModeUsed RuntimeMode    `json:"executionModeUsed"`
	FallbackApplied   bool           `json:"fallbackApplied"`
	FallbackReason    string         `json:"fallbackReason,omitempty"`
}

type DemoValidation struct {
	DemoID            string      `json:"demoId"`
	WorkloadID        string      `json:"workloadId,omitempty"`
	Success           bool        `json:"success"`
	ExecutionModeUsed RuntimeMode `json:"executionModeUsed"`
	FallbackApplied   bool        `json:"fallbackApplied"`
	Signals           []string    `json:"signals"`
	SummaryLines      []string    `json:"summaryLines"`
}
