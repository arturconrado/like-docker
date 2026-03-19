import { postgresModeLabel, postgresOperationalStatus } from '../lib/format'
import type { Workload } from '../types'

interface PostgresStatusCardProps {
  workload?: Workload | null
}

export function PostgresStatusCard({ workload }: PostgresStatusCardProps) {
  if (!workload) {
    return (
      <section className="rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
        <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">PostgreSQL Demo</h4>
        <p className="mt-2 text-sm text-zinc-400">Execute a demonstração para visualizar estado operacional do PostgreSQL.</p>
      </section>
    )
  }

  const modeUsed = workload.runtime.modeUsed || workload.mode
  const status = postgresOperationalStatus(workload.status, workload.runtime.readinessState)

  return (
    <section className="rounded-2xl border border-cyan-400/25 bg-cyan-500/10 p-4">
      <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-cyan-100">PostgreSQL Status</h4>

      <div className="mt-3 grid gap-2 text-xs sm:grid-cols-2 lg:grid-cols-3">
        <StatusCell label="Tipo" value={workload.workloadType || 'Database'} />
        <StatusCell label="Modo usado" value={postgresModeLabel(modeUsed)} />
        <StatusCell label="Status" value={status} />
        <StatusCell label="Porta" value={(workload.runtime.port ?? 0) > 0 ? String(workload.runtime.port) : '—'} />
        <StatusCell label="Data Dir" value={workload.runtime.dataDir || '—'} />
        <StatusCell label="Readiness" value={workload.runtime.readinessState || 'pending'} />
        <StatusCell label="PID Principal" value={(workload.runtime.mainPid ?? 0) > 0 ? String(workload.runtime.mainPid) : '—'} />
        <StatusCell label="Hostname" value={workload.runtime.containerHostname || '—'} />
      </div>

      <div className="mt-3 grid gap-2 text-xs text-cyan-50/90 sm:grid-cols-3">
        <div className="rounded-xl border border-cyan-300/20 bg-cyan-500/8 px-3 py-2">1. Cluster preparado e evidenciado por `initdb`.</div>
        <div className="rounded-xl border border-cyan-300/20 bg-cyan-500/8 px-3 py-2">2. Servidor iniciado com porta dinâmica e PID principal.</div>
        <div className="rounded-xl border border-cyan-300/20 bg-cyan-500/8 px-3 py-2">3. `pg_isready` e `readiness` confirmam prontidão operacional.</div>
      </div>
    </section>
  )
}

function StatusCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-cyan-300/25 bg-cyan-500/5 px-3 py-2">
      <p className="uppercase tracking-[0.14em] text-cyan-200/70">{label}</p>
      <p className="mt-1 truncate text-sm font-semibold text-cyan-50">{value}</p>
    </div>
  )
}
