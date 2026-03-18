import { Activity, BarChart3, BrainCircuit, FlaskConical, Layers3 } from 'lucide-react'
import type { ComponentType } from 'react'

import { cn } from '../lib/cn'
import { NAV_ITEMS, type NavigationKey } from '../lib/constants'

const ICONS: Record<NavigationKey, ComponentType<{ className?: string }>> = {
  dashboard: BarChart3,
  demos: FlaskConical,
  workloads: Layers3,
  insights: BrainCircuit,
  activity: Activity,
}

interface SidebarProps {
  active: NavigationKey
  onChange: (key: NavigationKey) => void
}

export function Sidebar({ active, onChange }: SidebarProps) {
  return (
    <aside className="rounded-3xl border border-zinc-800/80 bg-zinc-950/65 p-4 backdrop-blur-xl">
      <div className="mb-6 rounded-2xl border border-cyan-400/30 bg-cyan-500/10 p-4">
        <p className="text-xs uppercase tracking-[0.2em] text-cyan-200/80">MiniDock Executive AI</p>
        <h1 className="mt-1 text-lg font-bold text-zinc-100">Runtime Intelligence</h1>
      </div>

      <nav className="space-y-2">
        {NAV_ITEMS.map((item) => {
          const Icon = ICONS[item.key]
          const isActive = item.key === active
          return (
            <button
              key={item.key}
              type="button"
              onClick={() => onChange(item.key)}
              className={cn(
                'group flex w-full items-center gap-3 rounded-xl border px-3 py-2.5 text-left text-sm font-medium transition',
                isActive
                  ? 'border-cyan-400/50 bg-cyan-500/15 text-cyan-100'
                  : 'border-transparent text-zinc-400 hover:border-zinc-700 hover:bg-zinc-900/70 hover:text-zinc-100',
              )}
            >
              <Icon className={cn('size-4', isActive ? 'text-cyan-200' : 'text-zinc-500 group-hover:text-zinc-200')} />
              <span>{item.label}</span>
            </button>
          )
        })}
      </nav>

      <div className="mt-8 rounded-2xl border border-zinc-800 bg-zinc-900/70 p-4">
        <p className="text-xs uppercase tracking-[0.2em] text-zinc-500">Diretriz</p>
        <p className="mt-2 text-sm text-zinc-300">
          Priorize executabilidade local e clareza executiva para demo técnica.
        </p>
      </div>
    </aside>
  )
}
