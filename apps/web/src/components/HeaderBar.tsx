import { CircleCheck, Loader2, PlusCircle, Radio, ServerCog } from 'lucide-react'
import type { ReactNode } from 'react'

import { cn } from '../lib/cn'
import { modeLabel } from '../lib/format'
import type { HealthResponse, HostCapabilities } from '../types'

interface HeaderBarProps {
  health?: HealthResponse
  capabilities?: HostCapabilities
  healthError: boolean
  liveMode: 'connecting' | 'live' | 'polling'
  onCreate: () => void
}

export function HeaderBar({ health, capabilities, healthError, liveMode, onCreate }: HeaderBarProps) {
  const healthLabel = healthError ? 'Indisponível' : health?.status === 'ok' ? 'Operacional' : 'Aguardando'

  return (
    <header className="rounded-3xl border border-zinc-800/80 bg-zinc-950/65 p-5 backdrop-blur-xl">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-zinc-100">MiniDock Executive AI</h2>
          <p className="mt-1 text-sm text-zinc-400">Inteligência de Runtime para Liderança Técnica</p>
        </div>

        <button
          type="button"
          onClick={onCreate}
          className="inline-flex items-center gap-2 rounded-xl border border-cyan-300/40 bg-cyan-500/20 px-4 py-2.5 text-sm font-semibold text-cyan-100 transition hover:bg-cyan-500/30"
        >
          <PlusCircle className="size-4" />
          Nova Workload
        </button>
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        <Badge icon={<ServerCog className="size-3.5" />} label="Ambiente" value={`Local · ${capabilities?.os ?? '—'}`} tone="zinc" />
        <Badge
          icon={healthError ? <Radio className="size-3.5" /> : <CircleCheck className="size-3.5" />}
          label="Runtime"
          value={healthLabel}
          tone={healthError ? 'rose' : 'emerald'}
        />
        <Badge icon={<Loader2 className={cn('size-3.5', liveMode === 'connecting' && 'animate-spin')} />} label="Sync" value={syncLabel(liveMode)} tone={liveMode === 'polling' ? 'amber' : 'cyan'} />
        <Badge icon={<Radio className="size-3.5" />} label="Modo" value={modeLabel(health?.runtimeMode)} tone="zinc" />
        <Badge
          icon={<Radio className="size-3.5" />}
          label="Container Linux"
          value={capabilities?.supportsContainers ? 'Disponível' : 'Fallback'}
          tone={capabilities?.supportsContainers ? 'cyan' : 'amber'}
        />
      </div>
    </header>
  )
}

function syncLabel(mode: 'connecting' | 'live' | 'polling') {
  if (mode === 'live') return 'SSE ativo'
  if (mode === 'polling') return 'Polling fallback'
  return 'Conectando'
}

function Badge({
  icon,
  label,
  value,
  tone,
}: {
  icon: ReactNode
  label: string
  value: string
  tone: 'zinc' | 'emerald' | 'rose' | 'amber' | 'cyan'
}) {
  const toneClass: Record<typeof tone, string> = {
    zinc: 'border-zinc-700 bg-zinc-900 text-zinc-300',
    emerald: 'border-emerald-400/40 bg-emerald-500/15 text-emerald-200',
    rose: 'border-rose-400/40 bg-rose-500/15 text-rose-200',
    amber: 'border-amber-400/40 bg-amber-500/15 text-amber-200',
    cyan: 'border-cyan-400/40 bg-cyan-500/15 text-cyan-200',
  }

  return (
    <span className={cn('inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs', toneClass[tone])}>
      {icon}
      <span className="uppercase tracking-[0.16em] text-[10px] text-zinc-400">{label}</span>
      <span className="font-semibold">{value}</span>
    </span>
  )
}
