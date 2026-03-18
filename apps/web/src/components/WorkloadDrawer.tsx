import { AnimatePresence, motion } from 'framer-motion'
import { PauseCircle, Trash2, X } from 'lucide-react'

import { ModeBadge, RiskBadge, StatusBadge } from './Badges'
import { commandLabel, formatDateTime, formatDuration, modeLabel } from '../lib/format'
import type { Workload } from '../types'

interface WorkloadDrawerProps {
  workload: Workload | null
  stopping: boolean
  removing: boolean
  onClose: () => void
  onStop: (workload: Workload) => void
  onRemove: (workload: Workload) => void
}

export function WorkloadDrawer({ workload, stopping, removing, onClose, onStop, onRemove }: WorkloadDrawerProps) {
  return (
    <AnimatePresence>
      {workload && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-40 bg-black/60"
          />
          <motion.aside
            initial={{ x: 420, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            exit={{ x: 420, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 180, damping: 24 }}
            className="fixed right-0 top-0 z-50 h-full w-full max-w-xl overflow-y-auto border-l border-zinc-800 bg-zinc-950/98 p-6"
          >
            <div className="mb-5 flex items-start justify-between gap-3">
              <div>
                <h3 className="text-xl font-semibold text-zinc-100">{workload.name}</h3>
                <p className="mt-1 font-mono text-xs text-zinc-400">{commandLabel(workload.command, workload.args)}</p>
              </div>
              <button type="button" onClick={onClose} className="rounded-lg border border-zinc-700 p-2 text-zinc-400 hover:text-zinc-100">
                <X className="size-4" />
              </button>
            </div>

            <div className="mb-5 flex flex-wrap gap-2">
              <StatusBadge status={workload.status} />
              <ModeBadge mode={workload.mode} />
              <RiskBadge risk={workload.riskLevel} />
            </div>

            <section className="rounded-2xl border border-zinc-800 bg-zinc-900/70 p-4">
              <h4 className="text-sm font-semibold text-zinc-100">Resumo Normalizado</h4>
              <p className="mt-2 text-sm text-zinc-300">{workload.summary}</p>

              <dl className="mt-4 grid gap-2 text-xs text-zinc-400 sm:grid-cols-2">
                <div>
                  <dt className="text-zinc-500">Início</dt>
                  <dd className="text-zinc-200">{formatDateTime(workload.startedAt)}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Fim</dt>
                  <dd className="text-zinc-200">{formatDateTime(workload.finishedAt)}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Duração</dt>
                  <dd className="text-zinc-200">{formatDuration(workload.durationMs)}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Exit code</dt>
                  <dd className="text-zinc-200">{workload.exitCode === null ? '—' : workload.exitCode}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Modo solicitado</dt>
                  <dd className="text-zinc-200">{modeLabel(workload.requestedMode)}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Tipo</dt>
                  <dd className="text-zinc-200">{workload.workloadType || workload.runtime.workloadType || 'Runtime'}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Modo efetivo</dt>
                  <dd className="text-zinc-200">{modeLabel(workload.mode)}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Rootfs</dt>
                  <dd className="truncate text-zinc-200">{workload.runtime.rootfs || '—'}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Hostname</dt>
                  <dd className="text-zinc-200">{workload.runtime.containerHostname || '—'}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">PID principal</dt>
                  <dd className="text-zinc-200">{(workload.runtime.mainPid ?? 0) > 0 ? workload.runtime.mainPid : '—'}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Porta</dt>
                  <dd className="text-zinc-200">{(workload.runtime.port ?? 0) > 0 ? workload.runtime.port : '—'}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Data dir</dt>
                  <dd className="truncate text-zinc-200">{workload.runtime.dataDir || '—'}</dd>
                </div>
                <div>
                  <dt className="text-zinc-500">Readiness</dt>
                  <dd className="text-zinc-200">{workload.runtime.readinessState || '—'}</dd>
                </div>
              </dl>

              {workload.fallbackApplied && (
                <div className="mt-3 rounded-lg border border-amber-400/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
                  <strong className="font-semibold">Fallback aplicado:</strong>{' '}
                  {workload.fallbackReason || 'Host sem suporte para isolamento avançado; processo-local utilizado.'}
                </div>
              )}
            </section>

            <section className="mt-5 rounded-2xl border border-zinc-800 bg-zinc-900/70 p-4">
              <h4 className="text-sm font-semibold text-zinc-100">Insights Executivos</h4>
              <ul className="mt-3 space-y-2 text-sm text-zinc-300">
                {workload.aiInsights.map((insight) => (
                  <li key={insight} className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2">
                    {insight}
                  </li>
                ))}
              </ul>
              <div className="mt-3 rounded-lg border border-cyan-400/25 bg-cyan-500/10 px-3 py-2 text-sm text-cyan-100">
                <strong className="font-semibold">Próxima ação:</strong> {workload.suggestedAction}
              </div>
            </section>

            <section className="mt-5 rounded-2xl border border-zinc-800 bg-zinc-900/70 p-4">
              <h4 className="text-sm font-semibold text-zinc-100">Logs</h4>
              <div className="mt-3 max-h-64 overflow-auto rounded-lg border border-zinc-800 bg-black/50 p-3">
                {workload.logs.length === 0 ? (
                  <p className="text-xs text-zinc-500">Sem logs até o momento.</p>
                ) : (
                  <pre className="space-y-1 font-mono text-xs text-zinc-300">
                    {workload.logs.map((line, index) => (
                      <div key={`${workload.id}-log-${index}`}>{line}</div>
                    ))}
                  </pre>
                )}
              </div>
            </section>

            <div className="mt-5 flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => onStop(workload)}
                disabled={stopping || (workload.status !== 'Running' && workload.status !== 'Pending')}
                className="inline-flex items-center gap-2 rounded-lg border border-amber-400/30 px-4 py-2 text-sm font-semibold text-amber-200 enabled:hover:bg-amber-500/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <PauseCircle className="size-4" />
                Stop
              </button>
              <button
                type="button"
                onClick={() => onRemove(workload)}
                disabled={removing}
                className="inline-flex items-center gap-2 rounded-lg border border-rose-400/35 px-4 py-2 text-sm font-semibold text-rose-200 hover:bg-rose-500/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <Trash2 className="size-4" />
                Remove
              </button>
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  )
}
