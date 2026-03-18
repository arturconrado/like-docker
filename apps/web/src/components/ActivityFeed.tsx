import { cn } from '../lib/cn'
import { formatDateTime } from '../lib/format'
import type { EventItem } from '../types'

interface ActivityFeedProps {
  events: EventItem[]
  loading: boolean
}

const severityStyle: Record<EventItem['severity'], string> = {
  info: 'border-cyan-400/30 bg-cyan-500/10 text-cyan-100',
  warn: 'border-amber-400/30 bg-amber-500/10 text-amber-100',
  error: 'border-rose-400/30 bg-rose-500/10 text-rose-100',
}

export function ActivityFeed({ events, loading }: ActivityFeedProps) {
  if (loading) {
    return (
      <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-6">
        <p className="text-sm text-zinc-400">Carregando atividade...</p>
      </section>
    )
  }

  if (events.length === 0) {
    return (
      <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-6">
        <p className="text-sm text-zinc-400">Nenhum evento disponível.</p>
      </section>
    )
  }

  return (
    <section className="rounded-3xl border border-zinc-800 bg-zinc-950/65 p-4">
      <ul className="space-y-2">
        {events
          .slice()
          .reverse()
          .slice(0, 60)
          .map((event) => (
            <li
              key={event.id}
              className={cn('rounded-2xl border px-4 py-3 text-sm', severityStyle[event.severity])}
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="font-medium">{event.message}</p>
                <span className="text-xs opacity-80">{formatDateTime(event.createdAt)}</span>
              </div>
              <p className="mt-1 text-xs uppercase tracking-[0.12em] opacity-70">{event.type}</p>
            </li>
          ))}
      </ul>
    </section>
  )
}
