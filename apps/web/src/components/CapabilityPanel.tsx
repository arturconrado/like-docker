import { modeLabel } from '../lib/format'
import type { DemoDefinition, HostCapabilities } from '../types'

interface CapabilityPanelProps {
  capabilities?: HostCapabilities
  selectedDemo?: DemoDefinition | null
}

export function CapabilityPanel({ capabilities, selectedDemo }: CapabilityPanelProps) {
  if (!capabilities) {
    return (
      <section className="rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
        <p className="text-sm text-zinc-400">Consultando capabilities do host...</p>
      </section>
    )
  }

  const checks = [
    { label: 'Processo local', ok: capabilities.supportsProcessLocal },
    { label: 'Container Linux', ok: capabilities.supportsContainers },
    { label: 'Rootfs demo', ok: capabilities.rootfsAvailable },
    { label: 'PostgreSQL local', ok: capabilities.postgresLocalAvailable },
    { label: 'PostgreSQL rootfs', ok: capabilities.postgresContainerAvailable },
  ]

  return (
    <section className="rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Verificação de Ambiente</h4>
        <span className="rounded-full border border-cyan-400/35 bg-cyan-500/12 px-2.5 py-1 text-[11px] uppercase tracking-[0.15em] text-cyan-100">
          Recomendado: {modeLabel(capabilities.recommendedMode)}
        </span>
      </div>

      <div className="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-5">
        {checks.map((check) => (
          <div
            key={check.label}
            className={`rounded-xl border px-3 py-2 text-xs ${check.ok ? 'border-emerald-300/30 bg-emerald-500/10 text-emerald-100' : 'border-amber-300/30 bg-amber-500/10 text-amber-100'}`}
          >
            <p className="uppercase tracking-[0.14em] opacity-90">{check.label}</p>
            <p className="mt-1 text-sm font-semibold">{check.ok ? 'Disponível' : 'Indisponível'}</p>
          </div>
        ))}
      </div>

      {selectedDemo && (
        <div className="mt-3 rounded-xl border border-zinc-800 bg-zinc-950/70 px-3 py-2">
          <p className="text-xs uppercase tracking-[0.16em] text-zinc-500">Demo selecionada</p>
          <p className="mt-1 text-sm text-zinc-200">{selectedDemo.name}</p>
          <p className="text-xs text-zinc-400">Capacidades esperadas: {selectedDemo.requiredCapabilities.join(', ')}</p>
        </div>
      )}

      {capabilities.notes.length > 0 && (
        <ul className="mt-3 space-y-2 text-xs text-zinc-300">
          {capabilities.notes.slice(0, 4).map((note) => (
            <li key={note} className="rounded-lg border border-zinc-800 bg-zinc-950/70 px-3 py-2">
              {note}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
