import { render, screen } from '@testing-library/react'

import { CapabilityPanel } from '../components/CapabilityPanel'
import type { HostCapabilities } from '../types'

const capabilities: HostCapabilities = {
  os: 'linux',
  isLinux: true,
  supportsProcessLocal: true,
  supportsContainers: true,
  supportsNamespaces: true,
  supportsCgroups: true,
  cgroupVersion: 'v2',
  cgroupNotes: ['Cgroup v2 detectado no host Linux.'],
  supportsPivotRoot: true,
  rootfsAvailable: true,
  rootfsPath: '/tmp/rootfs',
  hasRootPrivileges: true,
  postgresLocalAvailable: true,
  postgresContainerAvailable: true,
  supportsPostgresDemo: true,
  recommendedMode: 'container-linux',
  postgresBinariesAvailable: true,
  postgresBinaryPaths: {
    initdb: '/usr/bin/initdb',
    postgres: '/usr/bin/postgres',
    pgIsready: '/usr/bin/pg_isready',
  },
  canCreateTempDir: true,
  canAllocatePort: true,
  canRunPostgresDemo: false,
  recommendedPostgresMode: 'container-linux',
  notes: ['Host apto para execução container-linux em modo avançado.'],
}

describe('CapabilityPanel', () => {
  it('renderiza sinais de pivot_root e cgroups', () => {
    render(<CapabilityPanel capabilities={capabilities} selectedDemo={null} />)

    expect(screen.getByText('Pivot Root')).toBeInTheDocument()
    expect(screen.getByText('Cgroups')).toBeInTheDocument()
    expect(screen.getByText(/Cgroup: v2/i)).toBeInTheDocument()
    expect(screen.getByText('Cgroup Notes')).toBeInTheDocument()
    expect(screen.getByText('Cgroup v2 detectado no host Linux.')).toBeInTheDocument()
  })
})
