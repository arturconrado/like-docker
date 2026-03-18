import { AlertCircle, Gauge, PlayCircle, Shield, Sparkles } from 'lucide-react'

import { KpiCard } from './KpiCard'
import { ModeBadge, RiskBadge, StatusBadge } from './Badges'
import { WORKLOAD_EXAMPLES, type WorkloadExample } from '../lib/constants'
import { commandLabel, formatDuration, modeLabel } from '../lib/format'
import type { HostCapabilities, Workload } from '../types'

interface DashboardViewProps {
  workloads: Workload[]
  summaryLines: string[]
  runtimeMode?: string
  runtimeHealthy: boolean
  uptimeMs?: number
  capabilities?: HostCapabilities
  onSelect: (workload: Workload) => void
  onSeedDemo: () => void
  onRunExample: (example: WorkloadExample) => void
  runningExampleId?: string | null
  seedLoading: boolean
}

export function DashboardView({
  workloads,
  summaryLines,
  runtimeMode,
  runtimeHealthy,
  uptimeMs,
  capabilities,
  onSelect,
  onSeedDemo,
  onRunExample,
  runningExampleId,
  seedLoading,
}: DashboardViewProps) {
  const running = workloads.filter((item) => item.status === 'Running').length
  const flagged = workloads.filter((item) => item.riskLevel === 'Review' || item.riskLevel === 'Risky').length
  const avgDuration =
    workloads.filter((item) => item.durationMs > 0).reduce((acc, item) => acc + item.durationMs, 0) /
    Math.max(workloads.filter((item) => item.durationMs > 0).length, 1)

  const highlighted = workloads.slice(0, 4)
  const risky = workloads.filter((item) => item.riskLevel !== 'Safe').slice(0, 3)

  return (
    <section className="space-y-4">
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <KpiCard label="Total Workloads" value={String(workloads.length)} hint="Visão consolidada de execuções" icon={<Gauge className="size-4" />} />
        <KpiCard label="Ativas" value={String(running)} hint="Execuções em andamento" icon={<PlayCircle className="size-4" />} />
        <KpiCard label="Sinalizadas" value={String(flagged)} hint="Workloads com revisão recomendada" icon={<AlertCircle className="size-4" />} />
        <KpiCard
          label="Tempo Médio"
          value={formatDuration(avgDuration) || '—'}
          hint="Benchmark das últimas execuções"
          icon={<Shield className="size-4" />}
        />
      </div>

      <article className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
        <h3 className="text-sm font-semibold uppercase tracking-[0.16em] text-zinc-300">Painel de Saúde</h3>
        <div className="mt-3 grid gap-3 md:grid-cols-4">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/70 px-3 py-2">
            <p className="text-xs text-zinc-500">Estado Runtime</p>
            <p className={runtimeHealthy ? 'text-sm font-semibold text-emerald-200' : 'text-sm font-semibold text-rose-200'}>
              {runtimeHealthy ? 'Operacional' : 'Instável'}
            </p>
          </div>
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/70 px-3 py-2">
            <p className="text-xs text-zinc-500">Modo padrão</p>
            <p className="text-sm font-semibold text-zinc-100">{modeLabel(runtimeMode)}</p>
          </div>
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/70 px-3 py-2">
            <p className="text-xs text-zinc-500">Container Linux</p>
            <p className={capabilities?.supportsContainers ? 'text-sm font-semibold text-cyan-200' : 'text-sm font-semibold text-amber-200'}>
              {capabilities?.supportsContainers ? 'Disponível' : 'Fallback ativo'}
            </p>
          </div>
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/70 px-3 py-2">
            <p className="text-xs text-zinc-500">Uptime API</p>
            <p className="text-sm font-semibold text-zinc-100">{formatDuration(uptimeMs ?? 0)}</p>
          </div>
        </div>
      </article>

      {capabilities && (
        <article className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
          <h3 className="text-sm font-semibold uppercase tracking-[0.16em] text-zinc-300">Capabilities do Host</h3>
          <div className="mt-3 grid gap-3 md:grid-cols-5 text-xs">
            <CapabilityCard label="OS" value={capabilities.os} tone="zinc" />
            <CapabilityCard label="Namespaces" value={capabilities.supportsNamespaces ? 'Sim' : 'Não'} tone={capabilities.supportsNamespaces ? 'emerald' : 'amber'} />
            <CapabilityCard label="Pivot Root" value={capabilities.supportsPivotRoot ? 'Sim' : 'Não'} tone={capabilities.supportsPivotRoot ? 'emerald' : 'amber'} />
            <CapabilityCard label="Rootfs" value={capabilities.rootfsAvailable ? 'Disponível' : 'Ausente'} tone={capabilities.rootfsAvailable ? 'emerald' : 'amber'} />
            <CapabilityCard label="Recomendado" value={modeLabel(capabilities.recommendedMode)} tone="cyan" />
          </div>
          {capabilities.notes.length > 0 && (
            <ul className="mt-3 space-y-2 text-sm text-zinc-300">
              {capabilities.notes.map((note) => (
                <li key={note} className="rounded-lg border border-zinc-800 bg-zinc-900/60 px-3 py-2">
                  {note}
                </li>
              ))}
            </ul>
          )}
        </article>
      )}

      <div className="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
        <article className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-zinc-100">Workloads em destaque</h3>
            <button
              type="button"
              onClick={onSeedDemo}
              disabled={seedLoading}
              className="rounded-lg border border-zinc-700 px-3 py-2 text-xs font-semibold text-zinc-200 transition hover:bg-zinc-800 disabled:opacity-60"
            >
              {seedLoading ? 'Carregando demo...' : 'Carregar dados de demonstração'}
            </button>
          </div>

          {highlighted.length === 0 ? (
            <p className="text-sm text-zinc-400">Sem workloads no momento.</p>
          ) : (
            <ul className="space-y-3">
              {highlighted.map((workload) => (
                <li key={workload.id} className="rounded-2xl border border-zinc-800 bg-zinc-900/70 p-4">
                  <button type="button" onClick={() => onSelect(workload)} className="w-full text-left">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <p className="font-semibold text-zinc-100">{workload.name}</p>
                      <div className="flex gap-2">
                        <StatusBadge status={workload.status} />
                        <ModeBadge mode={workload.mode} />
                        <RiskBadge risk={workload.riskLevel} />
                      </div>
                    </div>
                    <p className="mt-2 font-mono text-xs text-zinc-500">{commandLabel(workload.command, workload.args)}</p>
                    <p className="mt-2 text-sm text-zinc-300">{workload.summary}</p>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </article>

        <article className="space-y-4">
          <section className="rounded-3xl border border-cyan-400/20 bg-cyan-500/10 p-5">
            <div className="mb-3 flex items-center gap-2 text-cyan-100">
              <Sparkles className="size-4" />
              <h3 className="text-sm font-semibold uppercase tracking-[0.16em]">Resumo Executivo Global</h3>
            </div>
            <ul className="space-y-2 text-sm text-cyan-50/90">
              {(summaryLines.length > 0 ? summaryLines : ['Sem leitura executiva disponível no momento.']).map((line) => (
                <li key={line} className="rounded-lg border border-cyan-300/20 bg-cyan-500/10 px-3 py-2">
                  {line}
                </li>
              ))}
            </ul>
          </section>

          <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
            <h3 className="text-sm font-semibold uppercase tracking-[0.16em] text-zinc-300">Workloads Sinalizadas</h3>
            {risky.length === 0 ? (
              <p className="mt-3 text-sm text-zinc-400">Nenhuma workload sinalizada por risco neste momento.</p>
            ) : (
              <ul className="mt-3 space-y-2">
                {risky.map((workload) => (
                  <li key={workload.id} className="rounded-xl border border-zinc-800 bg-zinc-900/70 px-3 py-2">
                    <button type="button" className="w-full text-left" onClick={() => onSelect(workload)}>
                      <p className="text-sm font-medium text-zinc-100">{workload.name}</p>
                      <p className="mt-1 text-xs text-zinc-400">{workload.suggestedAction}</p>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </article>
      </div>

      <article className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
        <h3 className="text-lg font-semibold text-zinc-100">Biblioteca de Exemplos</h3>
        <p className="mt-1 text-sm text-zinc-400">Executar exemplo cria uma workload pronta para demonstração de runtime e fallback.</p>

        <div className="mt-4 grid gap-3 lg:grid-cols-2">
          {WORKLOAD_EXAMPLES.map((example) => (
            <section key={example.id} className="rounded-2xl border border-zinc-800 bg-zinc-900/60 p-4">
              <div className="flex items-center justify-between gap-3">
                <h4 className="text-sm font-semibold text-zinc-100">{example.name}</h4>
                <span className="rounded-full border border-zinc-700 px-2 py-1 text-[11px] uppercase tracking-[0.16em] text-zinc-300">
                  {modeLabel(example.preferredMode)}
                </span>
              </div>
              <p className="mt-2 text-xs text-zinc-400">{example.objective}</p>
              <p className="mt-2 font-mono text-xs text-zinc-500">{commandLabel(example.command, example.args)}</p>
              <button
                type="button"
                onClick={() => onRunExample(example)}
                disabled={seedLoading || Boolean(runningExampleId)}
                className="mt-3 rounded-lg border border-cyan-400/30 bg-cyan-500/10 px-3 py-2 text-xs font-semibold text-cyan-100 transition hover:bg-cyan-500/20 disabled:opacity-50"
              >
                {runningExampleId === example.id ? 'Executando exemplo...' : 'Executar exemplo'}
              </button>
            </section>
          ))}
        </div>
      </article>
    </section>
  )
}

function CapabilityCard({
  label,
  value,
  tone,
}: {
  label: string
  value: string
  tone: 'zinc' | 'emerald' | 'amber' | 'cyan'
}) {
  const toneClass: Record<typeof tone, string> = {
    zinc: 'border-zinc-800 bg-zinc-900/70 text-zinc-200',
    emerald: 'border-emerald-400/30 bg-emerald-500/10 text-emerald-200',
    amber: 'border-amber-400/30 bg-amber-500/10 text-amber-200',
    cyan: 'border-cyan-400/30 bg-cyan-500/10 text-cyan-200',
  }

  return (
    <div className={`rounded-xl border px-3 py-2 ${toneClass[tone]}`}>
      <p className="text-[11px] uppercase tracking-[0.16em] opacity-80">{label}</p>
      <p className="mt-1 text-sm font-semibold">{value}</p>
    </div>
  )
}
