import { render, screen } from '@testing-library/react'

import { RiskBadge, StatusBadge } from '../components/Badges'
import { BackendUnavailablePanel } from '../components/BackendUnavailablePanel'

describe('Badges e painel de indisponibilidade', () => {
  it('renderiza badges de status e risco', () => {
    render(
      <div>
        <StatusBadge status="Running" />
        <RiskBadge risk="Review" />
      </div>,
    )

    expect(screen.getByText('Running')).toBeInTheDocument()
    expect(screen.getByText('Review')).toBeInTheDocument()
  })

  it('renderiza painel de backend indisponível', () => {
    render(<BackendUnavailablePanel onRetry={() => undefined} />)

    expect(screen.getByText('Backend indisponível')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Tentar novamente' })).toBeInTheDocument()
  })
})
