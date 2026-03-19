import { AnimatePresence, motion } from 'framer-motion'
import { Eye, PauseCircle, PlayCircle, Trash2, X } from 'lucide-react'

import { ModeBadge, RiskBadge, StatusBadge } from './Badges'
import { CapabilityPanel } from './CapabilityPanel'
import { DemoProgressStepper, type DemoStep } from './DemoProgressStepper'
import { DemoSummaryPanel } from './DemoSummaryPanel'
import { EvidencePanel } from './EvidencePanel'
import { PostgresStatusCard } from './PostgresStatusCard'
import { commandLabel, modeLabel, postgresModeLabel, postgresOperationalStatus } from '../lib/format'
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
  const isPreparing = workload?.status === 'Pending' || workload?.status === 'Preparing'
  const isStarting = workload?.status === 'Starting' || workload?.runtime.readinessState === 'starting'
  const isRunning = workload?.status === 'Running'
  const isFinished = Boolean(workload && ['Completed', 'Failed', 'Stopped'].includes(workload.status))
  const validatedForCurrent =
    validation && (!workload?.id || !validation.workloadId || validation.workloadId === workload.id) ? validation : null

  const steps = computeSteps({
    hasDemo: Boolean(demo),
    capabilitiesReady: Boolean(capabilities),
    runningDemo,
    hasWorkload,
    isPreparing,
    isStarting,
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
                <p className="text-sm font-semibold text-cyan-100">Etapa 1 — Verificação do ambiente</p>
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
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">1. Verificação do ambiente confirma Linux, binários e porta disponível.</span>
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">2. Preparação do cluster evidencia `initdb` e `PGDATA` temporário.</span>
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">3. Inicialização do servidor publica porta, PID e logs reais.</span>
                  <span className="rounded-lg border border-cyan-300/25 bg-cyan-500/10 px-3 py-2">4. `pg_isready` e `readiness` confirmam workload pronta para conexões.</span>
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
              <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Etapa 5 — Workload em execução</h4>
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
                  <div className="mt-2 grid gap-2 text-xs text-zinc-300 sm:grid-cols-3">
                    <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 px-3 py-2">
                      <p className="uppercase tracking-[0.14em] text-zinc-500">Modo usado</p>
                      <p className="mt-1 text-sm font-semibold text-zinc-100">
                        {postgresModeLabel(workload.runtime.modeUsed || workload.mode)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 px-3 py-2">
                      <p className="uppercase tracking-[0.14em] text-zinc-500">Status operacional</p>
                      <p className="mt-1 text-sm font-semibold text-zinc-100">
                        {postgresOperationalStatus(workload.status, workload.runtime.readinessState)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 px-3 py-2">
                      <p className="uppercase tracking-[0.14em] text-zinc-500">Readiness</p>
                      <p className="mt-1 text-sm font-semibold text-zinc-100">{workload.runtime.readinessState || 'pending'}</p>
                    </div>
                  </div>
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
              <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Ações</h4>
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
                      disabled={stopping || !['Pending', 'Preparing', 'Starting', 'Running'].includes(workload.status)}
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
  isPreparing,
  isStarting,
  isRunning,
  isFinished,
  hasValidation,
}: {
  hasDemo: boolean
  capabilitiesReady: boolean
  runningDemo: boolean
  hasWorkload: boolean
  isPreparing: boolean
  isStarting: boolean
  isRunning: boolean
  isFinished: boolean
  hasValidation: boolean
}): DemoStep[] {
  return [
    {
      id: 'environment',
      title: '1. Verificação do ambiente',
      description: 'Linux, binários PostgreSQL, diretório temporário e porta livre.',
      status: capabilitiesReady ? 'done' : hasDemo ? 'active' : 'pending',
    },
    {
      id: 'cluster',
      title: '2. Preparação do cluster',
      description: 'Criação do PGDATA temporário e execução do initdb.',
      status: isStarting || isRunning || isFinished ? 'done' : runningDemo || isPreparing ? 'active' : capabilitiesReady ? 'pending' : 'pending',
    },
    {
      id: 'server',
      title: '3. Inicialização do servidor',
      description: 'Subida do processo postgres, PID principal e porta dinâmica.',
      status: isRunning || isFinished ? 'done' : isStarting ? 'active' : hasWorkload ? 'pending' : 'pending',
    },
    {
      id: 'readiness',
      title: '4. Validação de readiness',
      description: 'Confirmação por pg_isready e sinais nos logs.',
      status: hasValidation ? 'done' : isRunning || isFinished ? 'active' : isStarting ? 'active' : 'pending',
    },
    {
      id: 'running',
      title: '5. Workload em execução',
      description: 'Status operacional, logs reais, evidências e ações.',
      status: isRunning || isFinished ? 'done' : hasWorkload ? 'active' : 'pending',
    },
  ]
}
