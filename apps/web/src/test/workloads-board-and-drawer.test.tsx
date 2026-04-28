import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { WorkloadDrawer } from '../components/WorkloadDrawer'
import { WorkloadsBoard } from '../components/WorkloadsBoard'
import type { Workload } from '../types'

const sampleWorkload: Workload = {
  id: 'wk_1',
  name: 'filesystem-inspection',
  command: 'ls',
  args: ['-la'],
  workloadType: 'Environment',
  requestedMode: 'processo-local',
  summary: 'Workload de inspeção detalhada do sistema de arquivos com metadados.',
  status: 'Completed',
  riskLevel: 'Safe',
  aiInsights: ['Insight 1', 'Insight 2'],
  suggestedAction: 'Seguro remover após validação.',
  startedAt: new Date().toISOString(),
  finishedAt: new Date().toISOString(),
  durationMs: 120,
  exitCode: 0,
  logs: ['[stdout] total 16'],
  mode: 'processo-local',
  fallbackApplied: false,
  fallbackReason: '',
  runtime: {
    engine: 'local-process-engine',
    isolated: false,
    workloadType: 'Environment',
    rootfs: '',
    containerHostname: '',
    mainPid: 0,
    pivotRootApplied: true,
    cgroupPath: '/sys/fs/cgroup/minidock/wk_1',
    cgroupVersion: 'v2',
    port: 0,
    dataDir: '',
    readinessState: '',
    modeUsed: '',
  },
  createdAt: new Date().toISOString(),
}

describe('Lista e drawer de workloads', () => {
  it('renderiza lista e dispara callback de detalhes', async () => {
    const user = userEvent.setup()
    const onSelect = vi.fn()

    render(
      <WorkloadsBoard
        workloads={[sampleWorkload]}
        loading={false}
        onSelect={onSelect}
        onStop={() => undefined}
        onRemove={() => undefined}
      />,
    )

    await user.click(screen.getByRole('button', { name: 'Detalhes' }))

    expect(onSelect).toHaveBeenCalledTimes(1)
    expect(screen.getByText('filesystem-inspection')).toBeInTheDocument()
  })

  it('renderiza drawer com logs e insights', () => {
    render(
      <WorkloadDrawer
        workload={sampleWorkload}
        stopping={false}
        removing={false}
        onClose={() => undefined}
        onStop={() => undefined}
        onRemove={() => undefined}
      />,
    )

    expect(screen.getByText('Resumo Normalizado')).toBeInTheDocument()
    expect(screen.getByText('Insights Executivos')).toBeInTheDocument()
    expect(screen.getByText('[stdout] total 16')).toBeInTheDocument()
    expect(screen.getByText('Próxima ação:')).toBeInTheDocument()
    expect(screen.getByText('Aplicado')).toBeInTheDocument()
    expect(screen.getByText('v2')).toBeInTheDocument()
  })
})
