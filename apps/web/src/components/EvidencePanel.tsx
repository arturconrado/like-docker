import { commandLabel, modeLabel, postgresModeLabel } from '../lib/format'
import type { DemoDefinition, DemoValidation, Workload } from '../types'

interface EvidencePanelProps {
  demo: DemoDefinition
  workload?: Workload | null
  validation?: DemoValidation | null
  validationLoading?: boolean
}

export function EvidencePanel({ demo, workload, validation, validationLoading }: EvidencePanelProps) {
  const observedSignals = validation?.signals?.length
    ? validation.signals
    : workload
      ? [
          `Modo efetivo: ${modeLabel(workload.mode)}`,
          `Modo usado: ${postgresModeLabel(workload.runtime.modeUsed || workload.mode)}`,
          `Status atual: ${workload.status}`,
          workload.fallbackApplied ? 'Fallback aplicado por limitação do host.' : 'Execução sem fallback.',
          `${workload.logs.length} linha(s) de log coletadas.`,
          (workload.runtime.port ?? 0) > 0 ? `Porta observada: ${workload.runtime.port}` : 'Porta ainda não publicada.',
          workload.runtime.dataDir ? `Data directory observado: ${workload.runtime.dataDir}` : 'Data directory ainda não publicado.',
          workload.runtime.readinessState ? `Readiness: ${workload.runtime.readinessState}` : 'Readiness ainda não publicado.',
        ]
      : []

  return (
    <section className="rounded-2xl border border-zinc-800 bg-zinc-900/55 p-4">
      <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-zinc-300">Evidências Técnicas</h4>

      {workload && (
        <div className="mt-3 rounded-xl border border-zinc-800 bg-zinc-950/65 px-3 py-2">
          <p className="text-xs uppercase tracking-[0.16em] text-zinc-500">Comando executado</p>
          <p className="mt-1 font-mono text-xs text-zinc-300">{commandLabel(workload.command, workload.args)}</p>
        </div>
      )}

      <div className="mt-3 grid gap-3 lg:grid-cols-2">
        <div>
          <p className="text-xs uppercase tracking-[0.16em] text-zinc-500">Sinais esperados</p>
          <ul className="mt-2 space-y-1.5 text-sm text-zinc-300">
            {demo.expectedSignals.map((signal) => (
              <li key={signal} className="rounded-lg border border-zinc-800 bg-zinc-950/65 px-3 py-2">
                {signal}
              </li>
            ))}
          </ul>
        </div>

        <div>
          <p className="text-xs uppercase tracking-[0.16em] text-zinc-500">Sinais observados</p>
          {validationLoading ? (
            <p className="mt-2 text-sm text-zinc-400">Validando sinais da demonstração...</p>
          ) : observedSignals.length === 0 ? (
            <p className="mt-2 text-sm text-zinc-400">Execute a demonstração para gerar evidências.</p>
          ) : (
            <ul className="mt-2 space-y-1.5 text-sm text-zinc-300">
              {observedSignals.map((signal) => (
                <li key={signal} className="rounded-lg border border-zinc-800 bg-zinc-950/65 px-3 py-2">
                  {signal}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {workload && workload.logs.length > 0 && (
        <div className="mt-3 rounded-xl border border-zinc-800 bg-black/50 p-3">
          <p className="text-[11px] uppercase tracking-[0.16em] text-zinc-500">Logs recentes</p>
          <pre className="mt-2 space-y-1 font-mono text-xs text-zinc-300">
            {workload.logs.slice(-6).map((line, index) => (
              <div key={`${workload.id}-evidence-log-${index}`}>{line}</div>
            ))}
          </pre>
        </div>
      )}
    </section>
  )
}
