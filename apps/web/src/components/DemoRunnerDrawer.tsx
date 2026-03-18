import { AnimatePresence, motion } from 'framer-motion'
import { Eye, PauseCircle, PlayCircle, Trash2, X } from 'lucide-react'

import { ModeBadge, RiskBadge, StatusBadge } from './Badges'
import { CapabilityPanel } from './CapabilityPanel'
import { DemoProgressStepper, type DemoStep } from './DemoProgressStepper'
import { DemoSummaryPanel } from './DemoSummaryPanel'
import { EvidencePanel } from './EvidencePanel'
import { PostgresStatusCard } from './PostgresStatusCard'
import { commandLabel, modeLabel } from '../lib/format'
import type { DemoDefinition, DemoValidation, HostCapabilities, Workload } from '../types'

interface DemoRunnerDrawerProps {
  open: boolean
  demo?: DemoDefinition | null
  workload?: Workload | null
  capabilities?: HostCapabilities
  validation?: DemoValidation | null
  validationLoading?: boolean
  runningDemo: boolean
  stopping: boolean
  removing: boolean
  onClose: () => void
  onRerun: (demo: DemoDefinition) => void
  onStop: (workload: Workload) => void
  onRemove: (workload: Workload) => void
  onOpenDetails: (workload: Workload) => void
}

export function DemoRunnerDrawer({
  open,
  demo,
  workload,
  capabilities,
  validation,
  validationLoading,
  runningDemo,
  stopping,
  removing,
  onClose,
  onRerun,
  onStop,
  onRemove,
  onOpenDetails,
}: DemoRunnerDrawerProps) {
  const hasWorkload = Boolean(workload)
  const isPending = workload?.status === 'Pending'
  const isRunning = workload?.status === 'Running'
  const isFinished = Boolean(workload && ['Completed', 'Failed', 'Stopped'].includes(workload.status))
  const validatedForCurrent =
    validation && (!workload?.id || !validation.workloadId || validation.workloadId === workload.id) ? validation : null

  const steps = computeSteps({
    hasDemo: Boolean(demo),
    capabilitiesReady: Boolean(capabilities),
    runningDemo,
    hasWorkload,
    isPending,
    isRunning,
    isFinished,
    hasValidation: Boolean(validatedForCurrent),
  })
  const progress =
    (steps.reduce((acc, step) => acc + (step.status === 'done' ? 1 : step.status === 'active' ? 0.5 : 0), 0) /
      Math.max(steps.length, 1)) *
    100

  return (
    <AnimatePresence>
      {open && demo && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-40 bg-black/65"
          />

          <motion.aside
            initial={{ x: 480, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            exit={{ x: 480, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 180, damping: 24 }}
            className="fixed right-0 top-0 z-50 h-full w-full max-w-3xl overflow-y-auto border-l border-zinc-800 bg-zinc-950/98 p-6"
          >
            <div className="mb-5 flex items-start justify-between gap-3">
              <div>
                <p className="text-xs uppercase tracking-[0.18em] text-cyan-300/80">Jornada Guiada</p>
                <h3 className="mt-1 text-2xl font-semibold text-zinc-100">{demo.name}</h3>
                <p className="mt-1 text-sm text-zinc-400">{demo.description}</p>
              </div>
              <button type="button" onClick={onClose} className="rounded-lg border border-zinc-700 p-2 text-zinc-400 hover:text-zinc-100">
                <X className="size-4" />
              </button>
            </div>

            <section className="rounded-2xl border border-cyan-400/25 bg-cyan-500/10 p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-semibold text-cyan-100">Etapa 1 — Seleção e contexto</p>
                <span className="rounded-full border border-cyan-300/35 px-2.5 py-1 text-[11px] uppercase tracking-[0.15em] text-cyan-100">
                  Preferencial: {modeLabel(demo.preferredMode)}
                </span>
              </div>
              <p className="mt-2 text-sm text-cyan-50/90">{demo.objective}</p>
              <p className="mt-2 text-xs text-cyan-200/80">Validações previstas: {demo.expectedSignals.join(' • ')}</p>
            </section>

            {demo.id === 'postgres-demo' && (
              <section className="mt-4 rounded-2xl border border-cyan-400/25 bg-cyan-500/8 p-4">
                <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-cyan-100">
                  Por Que Esta Demo Mostra Maturidade
                </h4>
                <p className="mt-2 text-sm text-cyan-50/90">
                  Não é necessário saber PostgreSQL: acompanhe apenas os sinais abaixo para comprovar uma operação
                  stateful completa.
                </p>
                <div className="mt-3 grid gap-2 text-xs text-cyan-100/90 sm:grid-cols-2">
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">Sinal 1: logs de inicialização consistentes.</span>
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">Sinal 2: status operacional Running/Completed.</span>
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">Sinal 3: metadados de porta e data directory.</span>
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">Sinal 4: readiness pronto para conexões.</span>
                </div>
              </section>
            )}

            <section className="mt-4 rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
              <div className="mb-2 flex items-center justify-between">
                <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Progresso da Demonstração</h4>
                <span className="text-xs text-zinc-400">{Math.round(progress)}%</span>
              </div>
              <div className="mb-3 h-1.5 overflow-hidden rounded-full bg-zinc-800">
                <div className="h-full rounded-full bg-cyan-400/80 transition-all" style={{ width: `${progress}%` }} />
              </div>
              <DemoProgressStepper steps={steps} />
            </section>

            <section className="mt-4">
              <CapabilityPanel capabilities={capabilities} selectedDemo={demo} />
            </section>

            <section className="mt-4 rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
              <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Etapa 4 — Execução</h4>
              {!workload ? (
                <p className="mt-2 text-sm text-zinc-400">Aguardando início da workload da demonstração.</p>
              ) : (
                <>
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    <StatusBadge status={workload.status} />
                    <ModeBadge mode={workload.mode} />
                    <RiskBadge risk={workload.riskLevel} />
                    <span className="rounded-full border border-zinc-700 px-2.5 py-1 text-[11px] uppercase tracking-[0.14em] text-zinc-300">
                      {workload.workloadType || 'Runtime'}
                    </span>
                  </div>
                  <p className="mt-2 font-mono text-xs text-zinc-400">{commandLabel(workload.command, workload.args)}</p>
                  {workload.fallbackApplied && (
                    <div className="mt-2 rounded-lg border border-amber-400/35 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
                      {workload.fallbackReason || 'Fallback aplicado para preservar executabilidade local.'}
                    </div>
                  )}
                </>
              )}
            </section>

            <section className="mt-4">
              <EvidencePanel demo={demo} workload={workload} validation={validatedForCurrent} validationLoading={validationLoading} />
            </section>

            {demo.id === 'postgres-demo' && (
              <section className="mt-4">
                <PostgresStatusCard workload={workload} />
              </section>
            )}

            <section className="mt-4">
              <DemoSummaryPanel demo={demo} workload={workload} validation={validatedForCurrent} capabilities={capabilities} />
            </section>

            <section className="mt-5 rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
              <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Etapa 6 — Encerramento opcional</h4>
              <div className="mt-3 flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => onRerun(demo)}
                  disabled={runningDemo}
                  className="inline-flex items-center gap-2 rounded-lg border border-cyan-300/40 bg-cyan-500/15 px-3 py-2 text-xs font-semibold text-cyan-100 hover:bg-cyan-500/25 disabled:opacity-55"
                >
                  <PlayCircle className="size-4" />
                  Reexecutar
                </button>
                {workload && (
                  <>
                    <button
                      type="button"
                      onClick={() => onStop(workload)}
                      disabled={stopping || (workload.status !== 'Running' && workload.status !== 'Pending')}
                      className="inline-flex items-center gap-2 rounded-lg border border-amber-400/35 px-3 py-2 text-xs font-semibold text-amber-100 enabled:hover:bg-amber-500/20 disabled:opacity-55"
                    >
                      <PauseCircle className="size-4" />
                      Stop
                    </button>
                    <button
                      type="button"
                      onClick={() => onRemove(workload)}
                      disabled={removing}
                      className="inline-flex items-center gap-2 rounded-lg border border-rose-400/35 px-3 py-2 text-xs font-semibold text-rose-100 hover:bg-rose-500/20 disabled:opacity-55"
                    >
                      <Trash2 className="size-4" />
                      Remove
                    </button>
                    <button
                      type="button"
                      onClick={() => onOpenDetails(workload)}
                      className="inline-flex items-center gap-2 rounded-lg border border-zinc-700 px-3 py-2 text-xs font-semibold text-zinc-200 hover:bg-zinc-800"
                    >
                      <Eye className="size-4" />
                      Abrir detalhes
                    </button>
                  </>
                )}
              </div>
            </section>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  )
}

function computeSteps({
  hasDemo,
  capabilitiesReady,
  runningDemo,
  hasWorkload,
  isPending,
  isRunning,
  isFinished,
  hasValidation,
}: {
  hasDemo: boolean
  capabilitiesReady: boolean
  runningDemo: boolean
  hasWorkload: boolean
  isPending: boolean
  isRunning: boolean
  isFinished: boolean
  hasValidation: boolean
}): DemoStep[] {
  return [
    {
      id: 'selection',
      title: 'Etapa 1 — Seleção e contexto',
      description: 'Nome da demo, objetivo e modo preferencial.',
      status: hasDemo ? 'done' : 'pending',
    },
    {
      id: 'environment',
      title: 'Etapa 2 — Verificação de ambiente',
      description: 'Capabilities do host e modo recomendado.',
      status: capabilitiesReady ? 'done' : hasDemo ? 'active' : 'pending',
    },
    {
      id: 'preparation',
      title: 'Etapa 3 — Preparação',
      description: 'Criação da workload e contexto de execução.',
      status: hasWorkload && !isPending ? 'done' : runningDemo || isPending || hasDemo ? 'active' : 'pending',
    },
    {
      id: 'execution',
      title: 'Etapa 4 — Execução',
      description: 'Status em tempo real e logs.',
      status: isFinished ? 'done' : isRunning ? 'active' : hasWorkload ? 'active' : 'pending',
    },
    {
      id: 'validation',
      title: 'Etapa 5 — Validação',
      description: 'Evidências técnicas, sinais e insights.',
      status: hasValidation ? 'done' : isFinished ? 'active' : 'pending',
    },
    {
      id: 'closure',
      title: 'Etapa 6 — Encerramento opcional',
      description: 'Stop, remove, reexecução e detalhes.',
      status: hasWorkload ? 'active' : 'pending',
    },
  ]
}
