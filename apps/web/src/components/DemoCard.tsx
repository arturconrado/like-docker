import { PlayCircle } from 'lucide-react'

import { modeLabel } from '../lib/format'
import type { DemoDefinition } from '../types'

interface DemoCardProps {
  demo: DemoDefinition
  running: boolean
  isPrimary?: boolean
  onRun: (demo: DemoDefinition) => void
}

export function DemoCard({ demo, running, isPrimary, onRun }: DemoCardProps) {
  return (
    <article
      className={`rounded-3xl border bg-zinc-950/70 p-5 transition hover:bg-zinc-950/90 ${
        isPrimary
          ? 'border-cyan-400/35 shadow-[0_0_0_1px_rgba(56,189,248,0.25)_inset] hover:border-cyan-300/55'
          : 'border-zinc-800 hover:border-cyan-400/35'
      }`}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.18em] text-zinc-500">{demo.workloadType}</p>
          <h3 className="mt-1 text-lg font-semibold text-zinc-100">{demo.name}</h3>
        </div>
        <div className="flex gap-2">
          {isPrimary && (
            <span className="rounded-full border border-cyan-300/45 bg-cyan-500/15 px-2.5 py-1 text-[11px] uppercase tracking-[0.16em] text-cyan-100">
              Principal
            </span>
          )}
          <span className="rounded-full border border-zinc-700 px-2.5 py-1 text-[11px] uppercase tracking-[0.16em] text-zinc-300">
            {demo.complexity}
          </span>
        </div>
      </div>

      <p className="mt-3 text-sm text-zinc-300">{demo.description}</p>
      <p className="mt-2 text-xs text-zinc-500">Objetivo: {demo.objective}</p>

      <div className="mt-4 flex flex-wrap gap-2">
        <span className="rounded-full border border-cyan-400/35 bg-cyan-500/12 px-2.5 py-1 text-[11px] uppercase tracking-[0.14em] text-cyan-100">
          {modeLabel(demo.preferredMode)}
        </span>
        {demo.tags.slice(0, 3).map((tag) => (
          <span key={tag} className="rounded-full border border-zinc-700 px-2.5 py-1 text-[11px] uppercase tracking-[0.14em] text-zinc-300">
            {tag}
          </span>
        ))}
      </div>

      <button
        type="button"
        onClick={() => onRun(demo)}
        disabled={running}
        className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cyan-300/40 bg-cyan-500/15 px-3.5 py-2 text-xs font-semibold text-cyan-100 transition hover:bg-cyan-500/25 disabled:opacity-55"
      >
        <PlayCircle className="size-4" />
        {running ? 'Executando demonstração...' : 'Executar demonstração'}
      </button>
    </article>
  )
}
