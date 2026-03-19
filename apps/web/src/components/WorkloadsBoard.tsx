import { Eye, PauseCircle, Trash2 } from 'lucide-react'

import { ModeBadge, RiskBadge, StatusBadge } from './Badges'
import { commandLabel, formatDateTime, formatDuration, modeLabel } from '../lib/format'
import type { Workload } from '../types'

interface WorkloadsBoardProps {
  workloads: Workload[]
  loading: boolean
  onSelect: (workload: Workload) => void
  onStop: (workload: Workload) => void
  onRemove: (workload: Workload) => void
}

export function WorkloadsBoard({ workloads, loading, onSelect, onStop, onRemove }: WorkloadsBoardProps) {
  if (loading) {
    return (
      <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-6">
        <p className="text-sm text-zinc-400">Carregando workloads...</p>
      </section>
    )
  }

  if (workloads.length === 0) {
    return (
      <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-6">
        <h3 className="text-lg font-semibold text-zinc-100">Nenhuma workload registrada</h3>
        <p className="mt-2 text-sm text-zinc-400">Crie uma nova workload ou carregue dados de demonstração.</p>
      </section>
    )
  }

  return (
    <section className="space-y-3">
      {workloads.map((workload) => (
        <article
          key={workload.id}
          className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5 backdrop-blur"
        >
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h3 className="text-lg font-semibold text-zinc-100">{workload.name}</h3>
              <p className="mt-1 font-mono text-xs text-zinc-400">{commandLabel(workload.command, workload.args)}</p>
              <p className="mt-3 max-w-3xl text-sm text-zinc-300">{workload.summary}</p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={workload.status} />
              <ModeBadge mode={workload.mode} />
              <RiskBadge risk={workload.riskLevel} />
            </div>
          </div>

          <div className="mt-4 grid gap-2 text-xs text-zinc-400 md:grid-cols-4">
            <span>
              <strong className="text-zinc-200">Duração:</strong> {formatDuration(workload.durationMs)}
            </span>
            <span>
              <strong className="text-zinc-200">Início:</strong> {formatDateTime(workload.startedAt)}
            </span>
            <span>
              <strong className="text-zinc-200">Modo:</strong> {modeLabel(workload.mode)}
            </span>
            <span>
              <strong className="text-zinc-200">Exit code:</strong>{' '}
              {workload.exitCode === null ? '—' : workload.exitCode}
            </span>
            <span>
              <strong className="text-zinc-200">Tipo:</strong> {workload.workloadType || workload.runtime.workloadType || 'Runtime'}
            </span>
          </div>

          <div className="mt-4 flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => onSelect(workload)}
              className="inline-flex items-center gap-2 rounded-lg border border-zinc-700 px-3 py-2 text-xs font-semibold text-zinc-200 transition hover:bg-zinc-800"
            >
              <Eye className="size-4" />
              Detalhes
            </button>
            <button
              type="button"
              onClick={() => onStop(workload)}
              disabled={!['Pending', 'Preparing', 'Starting', 'Running'].includes(workload.status)}
              className="inline-flex items-center gap-2 rounded-lg border border-amber-400/30 px-3 py-2 text-xs font-semibold text-amber-200 transition enabled:hover:bg-amber-500/20 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <PauseCircle className="size-4" />
              Stop
            </button>
            <button
              type="button"
              onClick={() => onRemove(workload)}
              className="inline-flex items-center gap-2 rounded-lg border border-rose-400/35 px-3 py-2 text-xs font-semibold text-rose-200 transition hover:bg-rose-500/20"
            >
              <Trash2 className="size-4" />
              Remove
            </button>
          </div>
        </article>
      ))}
    </section>
  )
}
