import { Lightbulb, ShieldAlert } from 'lucide-react'

import { ModeBadge, RiskBadge, StatusBadge } from './Badges'
import { commandLabel } from '../lib/format'
import type { Workload } from '../types'

interface InsightsViewProps {
  summaryLines: string[]
  workloads: Workload[]
  onSelect: (workload: Workload) => void
}

export function InsightsView({ summaryLines, workloads, onSelect }: InsightsViewProps) {
  const candidates = workloads.filter((workload) => workload.riskLevel !== 'Safe').slice(0, 6)

  return (
    <section className="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
      <article className="rounded-3xl border border-cyan-400/20 bg-cyan-500/10 p-5">
        <div className="mb-3 flex items-center gap-2 text-cyan-100">
          <Lightbulb className="size-4" />
          <h3 className="text-sm font-semibold uppercase tracking-[0.16em]">Leitura executiva</h3>
        </div>
        <ul className="space-y-2 text-sm text-cyan-50/95">
          {(summaryLines.length > 0 ? summaryLines : ['Sem leitura executiva disponível.']).map((line) => (
            <li key={line} className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">
              {line}
            </li>
          ))}
        </ul>
      </article>

      <article className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
        <div className="mb-3 flex items-center gap-2 text-zinc-200">
          <ShieldAlert className="size-4" />
          <h3 className="text-sm font-semibold uppercase tracking-[0.16em]">Fila de revisão</h3>
        </div>

        {candidates.length === 0 ? (
          <p className="text-sm text-zinc-400">Nenhuma workload com alerta de risco.</p>
        ) : (
          <ul className="space-y-2">
            {candidates.map((workload) => (
              <li key={workload.id} className="rounded-xl border border-zinc-800 bg-zinc-900/70 p-3">
                <button type="button" onClick={() => onSelect(workload)} className="w-full text-left">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <p className="text-sm font-semibold text-zinc-100">{workload.name}</p>
                    <div className="flex gap-2">
                      <StatusBadge status={workload.status} />
                      <ModeBadge mode={workload.mode} />
                      <RiskBadge risk={workload.riskLevel} />
                    </div>
                  </div>
                  <p className="mt-1 font-mono text-xs text-zinc-500">{commandLabel(workload.command, workload.args)}</p>
                  <p className="mt-2 text-xs text-zinc-300">{workload.suggestedAction}</p>
                </button>
              </li>
            ))}
          </ul>
        )}
      </article>
    </section>
  )
}
