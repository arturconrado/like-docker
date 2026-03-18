import { ShieldAlert, ShieldCheck, ShieldQuestion } from 'lucide-react'

import { cn } from '../lib/cn'
import { modeLabel } from '../lib/format'
import type { RiskLevel, RuntimeMode, WorkloadStatus } from '../types'

const statusStyles: Record<WorkloadStatus, string> = {
  Pending: 'bg-zinc-700/70 text-zinc-100 border-zinc-500/60',
  Running: 'bg-cyan-500/20 text-cyan-200 border-cyan-400/50',
  Completed: 'bg-emerald-500/20 text-emerald-200 border-emerald-400/50',
  Failed: 'bg-rose-500/20 text-rose-100 border-rose-400/50',
  Stopped: 'bg-amber-500/20 text-amber-100 border-amber-400/50',
}

const riskStyles: Record<RiskLevel, string> = {
  Safe: 'bg-emerald-500/20 text-emerald-200 border-emerald-400/50',
  Review: 'bg-amber-500/20 text-amber-100 border-amber-400/50',
  Risky: 'bg-rose-500/20 text-rose-100 border-rose-400/50',
}

const modeStyles: Record<'container-linux' | 'processo-local' | 'demo' | 'namespace-runtime', string> = {
  'container-linux': 'bg-cyan-500/20 text-cyan-100 border-cyan-400/50',
  'processo-local': 'bg-zinc-700/80 text-zinc-100 border-zinc-500/60',
  demo: 'bg-indigo-500/20 text-indigo-100 border-indigo-400/50',
  'namespace-runtime': 'bg-cyan-500/20 text-cyan-100 border-cyan-400/50',
}

export function StatusBadge({ status }: { status: WorkloadStatus }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold tracking-wide',
        statusStyles[status],
      )}
    >
      {status}
    </span>
  )
}

export function RiskBadge({ risk }: { risk: RiskLevel }) {
  const icon =
    risk === 'Safe' ? (
      <ShieldCheck className="size-3.5" />
    ) : risk === 'Review' ? (
      <ShieldQuestion className="size-3.5" />
    ) : (
      <ShieldAlert className="size-3.5" />
    )

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-semibold tracking-wide',
        riskStyles[risk],
      )}
    >
      {icon}
      {risk}
    </span>
  )
}

export function ModeBadge({ mode }: { mode: RuntimeMode }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold tracking-wide',
        modeStyles[mode] ?? modeStyles['processo-local'],
      )}
    >
      {modeLabel(mode)}
    </span>
  )
}
