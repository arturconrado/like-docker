import { AlertTriangle, RefreshCcw } from 'lucide-react'
import { Component, type ReactNode } from 'react'

interface AppErrorBoundaryProps {
  children: ReactNode
}

interface AppErrorBoundaryState {
  hasError: boolean
  message: string
}

export class AppErrorBoundary extends Component<AppErrorBoundaryProps, AppErrorBoundaryState> {
  constructor(props: AppErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, message: '' }
  }

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || 'Falha inesperada no frontend.',
    }
  }

  componentDidCatch(error: Error) {
    // Mantém rastreio no console para diagnóstico local
    console.error('MiniDock frontend error:', error)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen grid place-items-center px-4">
          <div className="w-full max-w-xl rounded-3xl border border-rose-500/30 bg-rose-950/20 p-7 text-rose-50">
            <div className="mb-4 inline-flex rounded-2xl border border-rose-400/40 bg-rose-500/20 p-3">
              <AlertTriangle className="size-5" />
            </div>
            <h1 className="text-xl font-semibold">Falha de renderização</h1>
            <p className="mt-2 text-sm text-rose-100/90">
              O app encontrou um erro inesperado. Recarregue a página para retomar a sessão.
            </p>
            <pre className="mt-3 overflow-auto rounded-xl border border-rose-400/20 bg-black/30 p-3 text-xs text-rose-100/80">
              {this.state.message}
            </pre>
            <button
              type="button"
              onClick={() => window.location.reload()}
              className="mt-4 inline-flex items-center gap-2 rounded-lg border border-rose-300/40 bg-rose-500/20 px-4 py-2 text-sm font-semibold text-rose-100 hover:bg-rose-500/30"
            >
              <RefreshCcw className="size-4" />
              Recarregar
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
