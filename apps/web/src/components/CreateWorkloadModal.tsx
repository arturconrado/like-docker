import { useMemo, useState } from 'react'
import { Loader2, X } from 'lucide-react'

import { COMMAND_PRESETS, MODE_OPTIONS } from '../lib/constants'
import { modeLabel } from '../lib/format'
import type { HostCapabilities, RuntimeMode } from '../types'

interface CreateWorkloadModalProps {
  open: boolean
  submitting: boolean
  capabilities?: HostCapabilities
  onClose: () => void
  onSubmit: (payload: { command: string; args?: string[]; mode: RuntimeMode; name?: string }) => void
}

export function CreateWorkloadModal({
  open,
  submitting,
  capabilities,
  onClose,
  onSubmit,
}: CreateWorkloadModalProps) {
  const [presetId, setPresetId] = useState(COMMAND_PRESETS[0]?.id ?? '')
  const [mode, setMode] = useState<RuntimeMode>(capabilities?.recommendedMode ?? 'processo-local')
  const [name, setName] = useState('')

  const selectedPreset = useMemo(
    () => COMMAND_PRESETS.find((item) => item.id === presetId) ?? COMMAND_PRESETS[0],
    [presetId],
  )
  const selectedMode = useMemo(() => MODE_OPTIONS.find((item) => item.value === mode), [mode])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4 backdrop-blur-sm">
      <div className="w-full max-w-xl rounded-3xl border border-zinc-700 bg-zinc-950 p-6 shadow-2xl shadow-black/40">
        <div className="mb-5 flex items-start justify-between gap-3">
          <div>
            <h3 className="text-xl font-semibold text-zinc-100">Nova Workload</h3>
            <p className="mt-1 text-sm text-zinc-400">Crie uma execução controlada com heurísticas executivas locais.</p>
          </div>
          <button type="button" onClick={onClose} className="rounded-lg border border-zinc-700 p-2 text-zinc-400 hover:text-zinc-100">
            <X className="size-4" />
          </button>
        </div>

        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault()
            onSubmit({
              command: selectedPreset.command,
              args: selectedPreset.args,
              mode,
              name: name.trim() || undefined,
            })
          }}
        >
          <label className="block space-y-2">
            <span className="text-xs uppercase tracking-[0.16em] text-zinc-400">Preset de comando</span>
            <select
              value={presetId}
              onChange={(event) => setPresetId(event.target.value)}
              className="w-full rounded-xl border border-zinc-700 bg-zinc-900 px-3 py-2.5 text-sm text-zinc-100 focus:border-cyan-400 focus:outline-none"
            >
              {COMMAND_PRESETS.map((option) => (
                <option key={option.id} value={option.id}>
                  {option.label}
                </option>
              ))}
            </select>
            <p className="text-xs text-zinc-500">{selectedPreset.description}</p>
            <p className="font-mono text-xs text-zinc-500">
              {selectedPreset.command} {selectedPreset.args.join(' ')}
            </p>
          </label>

          <label className="block space-y-2">
            <span className="text-xs uppercase tracking-[0.16em] text-zinc-400">Modo</span>
            <select
              value={mode}
              onChange={(event) => setMode(event.target.value as RuntimeMode)}
              className="w-full rounded-xl border border-zinc-700 bg-zinc-900 px-3 py-2.5 text-sm text-zinc-100 focus:border-cyan-400 focus:outline-none"
            >
              {MODE_OPTIONS.map((option) => {
                const unavailableContainer = option.value === 'container-linux' && !capabilities?.supportsContainers
                return (
                  <option key={option.value} value={option.value} disabled={unavailableContainer}>
                    {option.label}
                    {unavailableContainer ? ' (indisponível neste host)' : ''}
                  </option>
                )
              })}
            </select>
            <p className="text-xs text-zinc-500">{selectedMode?.helper}</p>
            {mode === 'container-linux' && !capabilities?.supportsContainers && (
              <div className="rounded-lg border border-amber-400/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
                container-linux depende de Linux + root + rootfs. O backend aplicará fallback para processo-local.
              </div>
            )}
            {capabilities && (
              <p className="text-xs text-zinc-500">
                Modo recomendado no host atual: <strong className="text-zinc-300">{modeLabel(capabilities.recommendedMode)}</strong>
              </p>
            )}
          </label>

          <label className="block space-y-2">
            <span className="text-xs uppercase tracking-[0.16em] text-zinc-400">Nome opcional</span>
            <input
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Deixe vazio para sugestão inteligente"
              className="w-full rounded-xl border border-zinc-700 bg-zinc-900 px-3 py-2.5 text-sm text-zinc-100 placeholder:text-zinc-500 focus:border-cyan-400 focus:outline-none"
            />
          </label>

          <div className="flex flex-wrap items-center justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-zinc-700 px-4 py-2 text-sm font-semibold text-zinc-300 hover:bg-zinc-800"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="inline-flex items-center gap-2 rounded-lg border border-cyan-300/40 bg-cyan-500/20 px-4 py-2 text-sm font-semibold text-cyan-100 transition hover:bg-cyan-500/30 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {submitting && <Loader2 className="size-4 animate-spin" />}
              Criar workload
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
