import { Check, Loader2 } from 'lucide-react'

import { cn } from '../lib/cn'

export interface DemoStep {
  id: string
  title: string
  description: string
  status: 'pending' | 'active' | 'done'
}

interface DemoProgressStepperProps {
  steps: DemoStep[]
}

export function DemoProgressStepper({ steps }: DemoProgressStepperProps) {
  return (
    <ol className="space-y-3">
      {steps.map((step) => (
        <li key={step.id} className="flex items-start gap-3 rounded-xl border border-zinc-800 bg-zinc-900/55 px-3 py-2.5">
          <span
            className={cn(
              'mt-0.5 inline-flex size-6 items-center justify-center rounded-full border text-[11px]',
              step.status === 'done' && 'border-emerald-300/50 bg-emerald-500/20 text-emerald-100',
              step.status === 'active' && 'border-cyan-300/50 bg-cyan-500/20 text-cyan-100',
              step.status === 'pending' && 'border-zinc-700 bg-zinc-900 text-zinc-500',
            )}
          >
            {step.status === 'done' ? <Check className="size-3.5" /> : step.status === 'active' ? <Loader2 className="size-3.5 animate-spin" /> : '•'}
          </span>

          <div className="min-w-0">
            <p className="text-sm font-semibold text-zinc-100">{step.title}</p>
            <p className="text-xs text-zinc-400">{step.description}</p>
          </div>
        </li>
      ))}
    </ol>
  )
}
