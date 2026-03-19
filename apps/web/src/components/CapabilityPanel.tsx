import { modeLabel, postgresModeLabel } from '../lib/format'
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
    { label: 'Linux', ok: capabilities.isLinux },
    { label: 'Processo local', ok: capabilities.supportsProcessLocal },
    { label: 'Container Linux', ok: capabilities.supportsContainers },
    { label: 'Rootfs demo', ok: capabilities.rootfsAvailable },
    { label: 'PostgreSQL binários', ok: capabilities.postgresBinariesAvailable },
    { label: 'PGDATA temporário', ok: capabilities.canCreateTempDir },
    { label: 'Porta livre', ok: capabilities.canAllocatePort },
    { label: 'PostgreSQL real', ok: capabilities.canRunPostgresDemo },
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

      <div className="mt-3 rounded-xl border border-cyan-400/20 bg-cyan-500/8 px-3 py-3 text-xs text-cyan-50">
        <p className="uppercase tracking-[0.14em] text-cyan-200/75">PostgreSQL Demo</p>
        <p className="mt-1 text-sm font-semibold text-cyan-50">
          Modo recomendado: {postgresModeLabel(capabilities.recommendedPostgresMode)}
        </p>
      </div>

      <div className="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
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

      {(capabilities.postgresBinaryPaths.initdb || capabilities.postgresBinaryPaths.postgres || capabilities.postgresBinaryPaths.pgIsready) && (
        <div className="mt-3 rounded-xl border border-zinc-800 bg-zinc-950/70 px-3 py-3 text-xs text-zinc-300">
          <p className="uppercase tracking-[0.16em] text-zinc-500">Binários PostgreSQL detectados</p>
          <p className="mt-2 font-mono text-[11px] text-zinc-400">initdb: {capabilities.postgresBinaryPaths.initdb || '—'}</p>
          <p className="mt-1 font-mono text-[11px] text-zinc-400">postgres: {capabilities.postgresBinaryPaths.postgres || '—'}</p>
          <p className="mt-1 font-mono text-[11px] text-zinc-400">pg_isready: {capabilities.postgresBinaryPaths.pgIsready || '—'}</p>
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
