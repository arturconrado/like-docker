import type { ReactNode } from 'react'

interface KpiCardProps {
  label: string
  value: string
  hint: string
  icon: ReactNode
}

export function KpiCard({ label, value, hint, icon }: KpiCardProps) {
  return (
    <article className="rounded-2xl border border-zinc-800 bg-zinc-950/70 p-4">
      <div className="flex items-center justify-between">
        <span className="text-xs uppercase tracking-[0.18em] text-zinc-500">{label}</span>
        <span className="text-zinc-300">{icon}</span>
      </div>
      <p className="mt-3 text-3xl font-semibold text-zinc-100">{value}</p>
      <p className="mt-2 text-xs text-zinc-400">{hint}</p>
    </article>
  )
}
