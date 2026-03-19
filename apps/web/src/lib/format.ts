export function formatDateTime(value?: string | null) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'

  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(date)
}

export function formatDuration(durationMs?: number | null) {
  if (!durationMs || durationMs <= 0) return '—'

  const seconds = Math.round(durationMs / 1000)
  if (seconds < 60) return `${seconds}s`

  const minutes = Math.floor(seconds / 60)
  const remainder = seconds % 60
  return `${minutes}m ${remainder}s`
}

export function commandLabel(command: string, args?: readonly string[] | null) {
  const safeArgs = Array.isArray(args) ? args : []
  return [command, ...safeArgs].join(' ').trim()
}

export function modeLabel(mode?: string | null) {
  if (!mode) return '—'
  if (mode === 'container-linux') return 'Container Linux'
  if (mode === 'processo-local') return 'Processo Local'
  if (mode === 'demo') return 'Demo'
  if (mode === 'namespace-runtime') return 'Container Linux (alias)'
  return mode
}

export function postgresModeLabel(mode?: string | null) {
  if (!mode) return '—'
  if (mode === 'processo-local-real') return 'Processo local real'
  if (mode === 'container-linux') return 'Container Linux'
  if (mode === 'demo') return 'Demo'
  return modeLabel(mode)
}

export function postgresOperationalStatus(status?: string | null, readinessState?: string | null) {
  if (!status) return '—'
  if (status === 'Preparing') return 'Preparing'
  if (status === 'Starting') return 'Starting'
  if (status === 'Pending' && readinessState === 'preparing') return 'Preparing'
  if (status === 'Running' && readinessState === 'starting') return 'Starting'
  if (status === 'Running' && readinessState === 'ready') return 'Running'
  return status
}

export function titleCase(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1)
}
