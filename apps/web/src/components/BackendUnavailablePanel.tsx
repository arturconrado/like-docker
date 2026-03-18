import { AlertTriangle, RefreshCcw } from 'lucide-react'

interface BackendUnavailablePanelProps {
  onRetry: () => void
}

export function BackendUnavailablePanel({ onRetry }: BackendUnavailablePanelProps) {
  return (
    <section className="rounded-3xl border border-rose-500/30 bg-rose-950/20 p-8 backdrop-blur">
      <div className="mb-4 inline-flex rounded-2xl border border-rose-500/40 bg-rose-500/20 p-3 text-rose-100">
        <AlertTriangle className="size-5" />
      </div>
      <h3 className="text-xl font-semibold text-rose-50">Backend indisponível</h3>
      <p className="mt-2 max-w-lg text-sm text-rose-100/80">
        Não foi possível conectar à API do MiniDock. Verifique se `apps/api` está em execução e tente novamente.
      </p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-5 inline-flex items-center gap-2 rounded-xl border border-rose-300/30 bg-rose-500/20 px-4 py-2 text-sm font-semibold text-rose-50 transition hover:bg-rose-500/30"
      >
        <RefreshCcw className="size-4" />
        Tentar novamente
      </button>
    </section>
  )
}
