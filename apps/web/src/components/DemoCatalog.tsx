import { DemoCard } from './DemoCard'
import type { DemoDefinition } from '../types'

interface DemoCatalogProps {
  demos: DemoDefinition[]
  runningDemoId?: string | null
  onRun: (demo: DemoDefinition) => void
}

export function DemoCatalog({ demos, runningDemoId, onRun }: DemoCatalogProps) {
  if (demos.length === 0) {
    return (
      <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-6">
        <p className="text-sm text-zinc-400">Nenhuma demonstração disponível no momento.</p>
      </section>
    )
  }

  return (
    <section className="grid gap-4 xl:grid-cols-2">
      {demos.map((demo) => (
        <DemoCard
          key={demo.id}
          demo={demo}
          running={runningDemoId === demo.id}
          isPrimary={demo.id === 'postgres-demo'}
          onRun={onRun}
        />
      ))}
    </section>
  )
}
