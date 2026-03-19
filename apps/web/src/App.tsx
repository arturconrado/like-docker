import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertTriangle, RefreshCcw } from 'lucide-react'

import {
  createWorkload,
  deleteWorkload,
  getEvents,
  getCapabilities,
  getHealth,
  getSummary,
  getWorkload,
  listWorkloads,
  seedDemo,
  stopWorkload,
} from './api/client'
import { ActivityFeed } from './components/ActivityFeed'
import { BackendUnavailablePanel } from './components/BackendUnavailablePanel'
import { CreateWorkloadModal } from './components/CreateWorkloadModal'
import { DashboardView } from './components/DashboardView'
import { DemosView } from './components/DemosView'
import { HeaderBar } from './components/HeaderBar'
import { InsightsView } from './components/InsightsView'
import { Sidebar } from './components/Sidebar'
import { WorkloadDrawer } from './components/WorkloadDrawer'
import { WorkloadsBoard } from './components/WorkloadsBoard'
import { useLiveUpdates } from './hooks/useLiveUpdates'
import type { WorkloadExample } from './lib/constants'
import type { NavigationKey } from './lib/constants'
import type { RuntimeMode, Workload } from './types'

function App() {
  const queryClient = useQueryClient()

  const [activeNav, setActiveNav] = useState<NavigationKey>('dashboard')
  const [createOpen, setCreateOpen] = useState(false)
  const [selectedWorkloadId, setSelectedWorkloadId] = useState<string | null>(null)
  const [runningExampleId, setRunningExampleId] = useState<string | null>(null)

  const liveMode = useLiveUpdates(queryClient)
  const pollingFallback = liveMode === 'polling'

  const healthQuery = useQuery({
    queryKey: ['health'],
    queryFn: getHealth,
    refetchInterval: 12000,
    retry: 1,
  })

  const capabilitiesQuery = useQuery({
    queryKey: ['capabilities'],
    queryFn: getCapabilities,
    refetchInterval: 25000,
    retry: 1,
  })

  const workloadsQuery = useQuery({
    queryKey: ['workloads'],
    queryFn: listWorkloads,
    refetchInterval: pollingFallback ? 4000 : false,
    retry: 2,
  })

  const eventsQuery = useQuery({
    queryKey: ['events'],
    queryFn: getEvents,
    refetchInterval: pollingFallback ? 5000 : 20000,
    retry: 2,
  })

  const summaryQuery = useQuery({
    queryKey: ['summary'],
    queryFn: getSummary,
    refetchInterval: pollingFallback ? 6000 : 18000,
    retry: 2,
  })

  const selectedQuery = useQuery({
    queryKey: ['workload', selectedWorkloadId],
    queryFn: () => getWorkload(selectedWorkloadId as string),
    enabled: Boolean(selectedWorkloadId),
    refetchInterval: (query) =>
      query.state.data && ['Pending', 'Preparing', 'Starting', 'Running'].includes(query.state.data.status) ? 1500 : false,
  })

  const createMutation = useMutation({
    mutationFn: createWorkload,
    onSuccess: (created) => {
      setCreateOpen(false)
      setSelectedWorkloadId(created.id)
      refreshOperationalQueries(queryClient)
    },
  })

  const stopMutation = useMutation({
    mutationFn: stopWorkload,
    onSuccess: () => {
      refreshOperationalQueries(queryClient)
    },
  })

  const removeMutation = useMutation({
    mutationFn: deleteWorkload,
    onSuccess: () => {
      refreshOperationalQueries(queryClient)
      if (selectedWorkloadId) {
        setSelectedWorkloadId(null)
      }
    },
  })

  const seedMutation = useMutation({
    mutationFn: seedDemo,
    onSuccess: () => {
      refreshOperationalQueries(queryClient)
    },
  })

  const workloads = useMemo(() => workloadsQuery.data ?? [], [workloadsQuery.data])
  const events = useMemo(() => eventsQuery.data ?? [], [eventsQuery.data])
  const summaryLines = useMemo(() => summaryQuery.data?.lines ?? [], [summaryQuery.data?.lines])

  const selectedFromList = useMemo(
    () => workloads.find((item) => item.id === selectedWorkloadId) ?? null,
    [workloads, selectedWorkloadId],
  )
  const selectedWorkload = selectedQuery.data ?? selectedFromList

  const backendUnavailable = workloadsQuery.isError && healthQuery.isError
  const hasPartialError =
    !backendUnavailable &&
    (workloadsQuery.isError || eventsQuery.isError || summaryQuery.isError || capabilitiesQuery.isError)

  const handleRetry = () => {
    queryClient.invalidateQueries()
  }

  const handleStop = (workload: Workload) => {
    stopMutation.mutate(workload.id)
  }

  const handleRemove = (workload: Workload) => {
    const confirmed = window.confirm(`Remover workload ${workload.name}?`)
    if (!confirmed) return

    removeMutation.mutate(workload.id)
  }

  const handleCreate = (payload: { command: string; args?: string[]; mode: RuntimeMode; name?: string }) => {
    createMutation.mutate(payload)
  }

  const handleRunExample = (example: WorkloadExample) => {
    setRunningExampleId(example.id)
    createMutation.mutate(
      {
        name: example.name,
        command: example.command,
        args: example.args,
        mode: example.preferredMode,
      },
      {
        onSettled: () => setRunningExampleId(null),
      },
    )
  }

  const content = () => {
    if (backendUnavailable) {
      return <BackendUnavailablePanel onRetry={handleRetry} />
    }

    if (activeNav === 'dashboard') {
      return (
        <DashboardView
          workloads={workloads}
          summaryLines={summaryLines}
          runtimeMode={healthQuery.data?.runtimeMode}
          runtimeHealthy={!healthQuery.isError && healthQuery.data?.status === 'ok'}
          uptimeMs={healthQuery.data?.uptimeMs}
          capabilities={capabilitiesQuery.data}
          onSelect={(workload) => setSelectedWorkloadId(workload.id)}
          onSeedDemo={() => seedMutation.mutate()}
          onRunExample={handleRunExample}
          runningExampleId={runningExampleId}
          seedLoading={seedMutation.isPending}
        />
      )
    }

    if (activeNav === 'workloads') {
      return (
        <WorkloadsBoard
          workloads={workloads}
          loading={workloadsQuery.isLoading && !workloadsQuery.data}
          onSelect={(workload) => setSelectedWorkloadId(workload.id)}
          onStop={handleStop}
          onRemove={handleRemove}
        />
      )
    }

    if (activeNav === 'demos') {
      return (
        <DemosView
          capabilities={capabilitiesQuery.data}
          workloads={workloads}
          stopping={stopMutation.isPending}
          removing={removeMutation.isPending}
          onSelectWorkload={(workload) => setSelectedWorkloadId(workload.id)}
          onStopWorkload={handleStop}
          onRemoveWorkload={handleRemove}
        />
      )
    }

    if (activeNav === 'insights') {
      return (
        <InsightsView
          summaryLines={summaryLines}
          workloads={workloads}
          onSelect={(workload) => setSelectedWorkloadId(workload.id)}
        />
      )
    }

    return <ActivityFeed events={events} loading={eventsQuery.isLoading && !eventsQuery.data} />
  }

  return (
    <div className="min-h-screen px-4 py-6 md:px-6">
      <div className="mx-auto max-w-[1600px]">
        <div className="grid gap-4 xl:grid-cols-[280px_1fr]">
          <Sidebar active={activeNav} onChange={setActiveNav} />

          <main className="space-y-4">
            <HeaderBar
              health={healthQuery.data}
              capabilities={capabilitiesQuery.data}
              healthError={healthQuery.isError}
              liveMode={liveMode}
              onCreate={() => setCreateOpen(true)}
            />

            {hasPartialError && (
              <section className="flex items-center justify-between rounded-2xl border border-amber-400/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
                <span className="inline-flex items-center gap-2">
                  <AlertTriangle className="size-4" />
                  Alguns painéis estão temporariamente desatualizados.
                </span>
                <button
                  type="button"
                  onClick={handleRetry}
                  className="inline-flex items-center gap-2 rounded-lg border border-amber-300/40 px-3 py-1.5 text-xs font-semibold text-amber-50 hover:bg-amber-500/20"
                >
                  <RefreshCcw className="size-3.5" />
                  Recarregar
                </button>
              </section>
            )}

            {content()}
          </main>
        </div>
      </div>

      {createOpen && (
        <CreateWorkloadModal
          open={createOpen}
          submitting={createMutation.isPending}
          capabilities={capabilitiesQuery.data}
          onClose={() => setCreateOpen(false)}
          onSubmit={handleCreate}
        />
      )}

      <WorkloadDrawer
        workload={selectedWorkload}
        stopping={stopMutation.isPending}
        removing={removeMutation.isPending}
        onClose={() => setSelectedWorkloadId(null)}
        onStop={handleStop}
        onRemove={handleRemove}
      />
    </div>
  )
}

function refreshOperationalQueries(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: ['health'] })
  queryClient.invalidateQueries({ queryKey: ['capabilities'] })
  queryClient.invalidateQueries({ queryKey: ['workloads'] })
  queryClient.invalidateQueries({ queryKey: ['events'] })
  queryClient.invalidateQueries({ queryKey: ['summary'] })
}

export default App
