export type WorkloadStatus = 'Pending' | 'Running' | 'Completed' | 'Failed' | 'Stopped'

export type RiskLevel = 'Safe' | 'Review' | 'Risky'

export type RuntimeMode = 'demo' | 'processo-local' | 'container-linux' | 'namespace-runtime'

export interface RuntimeMetadata {
  engine: string
  isolated: boolean
  rootfs?: string
  containerHostname?: string
  mainPid?: number
  workloadType?: string
  port?: number
  dataDir?: string
  readinessState?: string
}

export interface Workload {
  id: string
  name: string
  command: string
  args: string[]
  workloadType: string
  requestedMode: RuntimeMode
  summary: string
  status: WorkloadStatus
  riskLevel: RiskLevel
  aiInsights: string[]
  suggestedAction: string
  startedAt: string | null
  finishedAt: string | null
  durationMs: number
  exitCode: number | null
  logs: string[]
  mode: RuntimeMode
  fallbackApplied: boolean
  fallbackReason?: string
  runtime: RuntimeMetadata
  createdAt: string
}

export interface EventItem {
  id: string
  type: string
  workloadId?: string
  message: string
  severity: 'info' | 'warn' | 'error'
  createdAt: string
}

export interface HealthResponse {
  status: 'ok'
  runtimeMode: RuntimeMode
  uptimeMs: number
  timestamp: string
}

export interface HostCapabilities {
  os: string
  supportsProcessLocal: boolean
  supportsContainers: boolean
  supportsNamespaces: boolean
  supportsPivotRoot: boolean
  rootfsAvailable: boolean
  rootfsPath?: string
  hasRootPrivileges: boolean
  postgresLocalAvailable: boolean
  postgresContainerAvailable: boolean
  supportsPostgresDemo: boolean
  recommendedMode: RuntimeMode
  notes: string[]
}

export interface DashboardSummaryResponse {
  lines: string[]
}

export interface CreateWorkloadPayload {
  command: string
  args?: string[]
  mode: RuntimeMode
  requestedMode?: RuntimeMode
  name?: string
  workloadType?: string
  port?: number
  dataDir?: string
  fallbackReasonHint?: string
}

export interface DemoDefinition {
  id: string
  name: string
  description: string
  objective: string
  preferredMode: RuntimeMode
  workloadType: string
  complexity: string
  requiredCapabilities: string[]
  expectedSignals: string[]
  tags: string[]
  icon: string
  port?: number
  dataDir?: string
}

export interface DemoRunResponse {
  demo: DemoDefinition
  workload: Workload
  executionModeUsed: RuntimeMode
  fallbackApplied: boolean
  fallbackReason?: string
}

export interface DemoValidation {
  demoId: string
  workloadId?: string
  success: boolean
  executionModeUsed: RuntimeMode
  fallbackApplied: boolean
  signals: string[]
  summaryLines: string[]
}
