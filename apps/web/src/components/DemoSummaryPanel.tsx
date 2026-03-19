import { modeLabel, postgresModeLabel } from '../lib/format'
import type { DemoDefinition, DemoValidation, HostCapabilities, Workload } from '../types'

interface DemoSummaryPanelProps {
  demo: DemoDefinition
  workload?: Workload | null
  validation?: DemoValidation | null
  capabilities?: HostCapabilities
}

export function DemoSummaryPanel({ demo, workload, validation, capabilities }: DemoSummaryPanelProps) {
  const lines = validation?.summaryLines?.length
    ? validation.summaryLines
    : buildFallbackSummary(demo, workload, capabilities)

  return (
    <section className="rounded-2xl border border-emerald-400/25 bg-emerald-500/8 p-4">
      <h4 className="text-sm font-semibold uppercase tracking-[0.15em] text-emerald-100">Resumo Executivo da Demonstração</h4>
      <ul className="mt-3 space-y-2 text-sm text-emerald-50/95">
        {lines.map((line) => (
          <li key={line} className="rounded-lg border border-emerald-300/20 bg-emerald-500/10 px-3 py-2">
            {line}
          </li>
        ))}
      </ul>
    </section>
  )
}

function buildFallbackSummary(
  demo: DemoDefinition,
  workload?: Workload | null,
  capabilities?: HostCapabilities,
): string[] {
  const lines: string[] = []
  lines.push('Esta demonstração valida a capacidade de execução isolada e fallback controlado da ferramenta.')

  if (workload) {
    lines.push(`Modo efetivo observado: ${modeLabel(workload.mode)}.`)
    if (workload.runtime.modeUsed) {
      lines.push(`Modo usado pela demo PostgreSQL: ${postgresModeLabel(workload.runtime.modeUsed)}.`)
    }
    if (workload.fallbackApplied) {
      lines.push('O host atual não suportou o modo avançado; fallback foi aplicado para preservar a UX.')
    }
  } else if (capabilities && !capabilities.canRunPostgresDemo) {
    lines.push(`O host atual não suporta o caminho real; modo recomendado: ${postgresModeLabel(capabilities.recommendedPostgresMode)}.`)
  }

  if (demo.id === 'postgres-demo') {
    lines.push('A workload PostgreSQL demonstra suporte a cenário stateful com observabilidade operacional.')
    lines.push('O modo real foi priorizado para maximizar fidelidade técnica da demonstração.')
  }

  return lines.slice(0, 4)
}
