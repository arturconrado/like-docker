import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { PlayCircle } from 'lucide-react'

import { getDemos, runDemo, validateDemo } from '../api/client'
import { CapabilityPanel } from './CapabilityPanel'
import { DemoCatalog } from './DemoCatalog'
import { DemoRunnerDrawer } from './DemoRunnerDrawer'
import type { DemoDefinition, HostCapabilities, Workload } from '../types'

interface DemosViewProps {
  capabilities?: HostCapabilities
  workloads: Workload[]
  stopping: boolean
  removing: boolean
  onSelectWorkload: (workload: Workload) => void
  onStopWorkload: (workload: Workload) => void
  onRemoveWorkload: (workload: Workload) => void
}

export function DemosView({
  capabilities,
  workloads,
  stopping,
  removing,
  onSelectWorkload,
  onStopWorkload,
  onRemoveWorkload,
}: DemosViewProps) {
  const queryClient = useQueryClient()
  const [runnerOpen, setRunnerOpen] = useState(false)
  const [activeDemoID, setActiveDemoID] = useState<string | null>(null)
  const [activeWorkloadID, setActiveWorkloadID] = useState<string | null>(null)

  const demosQuery = useQuery({
    queryKey: ['demos'],
    queryFn: getDemos,
    retry: 2,
  })

  const runMutation = useMutation({
    mutationFn: runDemo,
    onSuccess: (result) => {
      setActiveDemoID(result.demo.id)
      setActiveWorkloadID(result.workload.id)
      refreshOperationalQueries(queryClient)
    },
  })

  const orderedDemos = useMemo(() => {
    const items = demosQuery.data ?? []
    return [...items].sort((a, b) => {
      if (a.id === 'postgres-demo') return -1
      if (b.id === 'postgres-demo') return 1
      return a.name.localeCompare(b.name)
    })
  }, [demosQuery.data])

  const primaryDemo = useMemo(
    () => orderedDemos.find((item) => item.id === 'postgres-demo') ?? null,
    [orderedDemos],
  )

  const resolvedDemoID = activeDemoID ?? primaryDemo?.id ?? orderedDemos[0]?.id ?? null
  const activeDemo = useMemo(
    () => orderedDemos.find((item) => item.id === resolvedDemoID) ?? null,
    [orderedDemos, resolvedDemoID],
  )
  const activeWorkload = useMemo(
    () => workloads.find((item) => item.id === activeWorkloadID) ?? null,
    [workloads, activeWorkloadID],
  )

  const validationQuery = useQuery({
    queryKey: ['demo-validate', resolvedDemoID, activeWorkloadID, activeWorkload?.status],
    queryFn: () => validateDemo(resolvedDemoID as string),
    enabled: runnerOpen && Boolean(resolvedDemoID),
    refetchInterval:
      activeWorkload && ['Pending', 'Preparing', 'Starting', 'Running'].includes(activeWorkload.status) ? 2500 : false,
    retry: 1,
  })

  const runningDemoID = runMutation.isPending ? runMutation.variables ?? null : null

  const handleRunDemo = (demo: DemoDefinition) => {
    setActiveDemoID(demo.id)
    setRunnerOpen(true)
    runMutation.mutate(demo.id)
  }

  return (
    <section className="space-y-4">
      <article className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-5">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.18em] text-cyan-300/80">Demonstrações</p>
            <h3 className="mt-1 text-2xl font-semibold text-zinc-100">Experiência Guiada de Valor</h3>
            <p className="mt-2 max-w-3xl text-sm text-zinc-400">
              Execute exemplos prontos com um clique, acompanhe preparação e execução em etapas e colete
              evidências técnicas para apresentação executiva do MiniDock.
            </p>
            <p className="mt-2 text-xs text-zinc-500">
              Sem conhecimento prévio: siga os passos, observe os badges e confirme os sinais de sucesso.
            </p>
          </div>
          <button
            type="button"
            onClick={() => activeDemo && handleRunDemo(activeDemo)}
            disabled={!activeDemo || runMutation.isPending}
            className="inline-flex items-center gap-2 rounded-xl border border-cyan-300/40 bg-cyan-500/15 px-4 py-2 text-sm font-semibold text-cyan-100 hover:bg-cyan-500/25 disabled:opacity-55"
          >
            <PlayCircle className="size-4" />
            Executar demo selecionada
          </button>
        </div>
      </article>

      <article className="rounded-3xl border border-cyan-400/30 bg-gradient-to-br from-cyan-500/12 via-zinc-950/70 to-zinc-950/70 p-5">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="max-w-3xl">
            <p className="text-xs uppercase tracking-[0.18em] text-cyan-200/90">Demonstração Principal</p>
            <h4 className="mt-1 text-2xl font-semibold text-cyan-50">PostgreSQL Demo</h4>
            <p className="mt-2 text-sm text-cyan-50/90">
              Esta jornada prova que a plataforma opera workloads stateful com prioridade para binários locais reais
              do PostgreSQL em Linux, preservando fallback elegante quando necessário.
            </p>
            <div className="mt-3 grid gap-2 text-xs text-cyan-100/90 sm:grid-cols-2">
              <span className="rounded-lg border border-cyan-300/30 bg-cyan-500/10 px-3 py-2">1. O ambiente valida Linux, binários, diretório temporário e porta livre.</span>
              <span className="rounded-lg border border-cyan-300/30 bg-cyan-500/10 px-3 py-2">2. O cluster sobe com `initdb`, `postgres` e logs reais em tempo real.</span>
              <span className="rounded-lg border border-cyan-300/30 bg-cyan-500/10 px-3 py-2">3. Evidências mostram porta, data dir, PID e readiness.</span>
              <span className="rounded-lg border border-cyan-300/30 bg-cyan-500/10 px-3 py-2">4. Se faltar capacidade no host, o fallback fica explícito na jornada.</span>
            </div>
          </div>
          <button
            type="button"
            onClick={() => primaryDemo && handleRunDemo(primaryDemo)}
            disabled={!primaryDemo || runMutation.isPending}
            className="inline-flex items-center gap-2 rounded-xl border border-cyan-200/40 bg-cyan-500/20 px-4 py-2.5 text-sm font-semibold text-cyan-50 hover:bg-cyan-500/30 disabled:opacity-55"
          >
            <PlayCircle className="size-4" />
            Executar PostgreSQL Demo
          </button>
        </div>
      </article>

      <CapabilityPanel capabilities={capabilities} selectedDemo={activeDemo} />

      {demosQuery.isLoading && (
        <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-6">
          <p className="text-sm text-zinc-400">Carregando catálogo de demonstrações...</p>
        </section>
      )}

      {demosQuery.isError && (
        <section className="rounded-3xl border border-rose-400/30 bg-rose-500/10 p-6">
          <p className="text-sm text-rose-100">Não foi possível carregar o catálogo de demos no backend.</p>
        </section>
      )}

      {!demosQuery.isLoading && !demosQuery.isError && (
        <DemoCatalog demos={orderedDemos} runningDemoId={runningDemoID} onRun={handleRunDemo} />
      )}

      {runMutation.isError && (
        <section className="rounded-2xl border border-amber-400/35 bg-amber-500/10 p-4 text-sm text-amber-100">
          Falha ao executar demonstração. Tente novamente ou valide capabilities do host.
        </section>
      )}

      <DemoRunnerDrawer
        open={runnerOpen}
        demo={activeDemo}
        workload={activeWorkload}
        capabilities={capabilities}
        validation={validationQuery.data}
        validationLoading={validationQuery.isLoading || validationQuery.isFetching}
        runningDemo={runMutation.isPending}
        stopping={stopping}
        removing={removing}
        onClose={() => setRunnerOpen(false)}
        onRerun={handleRunDemo}
        onStop={onStopWorkload}
        onRemove={onRemoveWorkload}
        onOpenDetails={onSelectWorkload}
      />
    </section>
  )
}

function refreshOperationalQueries(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: ['health'] })
  queryClient.invalidateQueries({ queryKey: ['capabilities'] })
  queryClient.invalidateQueries({ queryKey: ['workloads'] })
  queryClient.invalidateQueries({ queryKey: ['events'] })
  queryClient.invalidateQueries({ queryKey: ['summary'] })
}
