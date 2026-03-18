import type { RuntimeMode } from '../types'

export interface CommandPreset {
  id: string
  label: string
  description: string
  command: string
  args: string[]
}

export const COMMAND_PRESETS: CommandPreset[] = [
  {
    id: 'output-check',
    label: 'Output check',
    description: 'Validação rápida de stdout.',
    command: 'echo',
    args: ['hello'],
  },
  {
    id: 'filesystem-inspection',
    label: 'Filesystem inspection',
    description: 'Inspeção detalhada local com metadados.',
    command: 'ls',
    args: ['-la'],
  },
  {
    id: 'runtime-validation',
    label: 'Runtime validation',
    description: 'Confere diretório corrente do runtime.',
    command: 'pwd',
    args: [],
  },
  {
    id: 'controlled-sleep',
    label: 'Controlled sleep',
    description: 'Fluxo com janela para stop manual.',
    command: '/bin/sh',
    args: ['-c', 'echo starting && sleep 10 && echo finished'],
  },
  {
    id: 'container-identity-check',
    label: 'Container identity check',
    description: 'Hostname + rootfs em modo isolado.',
    command: '/bin/sh',
    args: ['-c', 'hostname && pwd && ls /'],
  },
  {
    id: 'runtime-diagnostics',
    label: 'Runtime diagnostics',
    description: 'Diagnóstico do ambiente de execução.',
    command: '/bin/sh',
    args: ['-c', 'hostname && uname -a && ps'],
  },
]

export const MODE_OPTIONS: { value: RuntimeMode; label: string; helper: string }[] = [
  {
    value: 'container-linux',
    label: 'Container Linux',
    helper: 'Executa em isolamento Linux real com rootfs dedicado quando o host suportar.',
  },
  {
    value: 'processo-local',
    label: 'Processo Local',
    helper: 'Execução real via processo local com máxima compatibilidade de ambiente.',
  },
  {
    value: 'demo',
    label: 'Demo',
    helper: 'Execução simulada para storytelling rápido em apresentações.',
  },
]

export interface WorkloadExample {
  id: string
  name: string
  preferredMode: RuntimeMode
  command: string
  args: string[]
  objective: string
}

export const WORKLOAD_EXAMPLES: WorkloadExample[] = [
  {
    id: 'container-identity-check',
    name: 'container-identity-check',
    preferredMode: 'container-linux',
    command: '/bin/sh',
    args: ['-c', 'hostname && pwd && ls /'],
    objective: 'Mostrar hostname isolado e filesystem interno do rootfs.',
  },
  {
    id: 'hello-container',
    name: 'hello-container',
    preferredMode: 'container-linux',
    command: '/bin/sh',
    args: ['-c', 'echo hello from container'],
    objective: 'Validar execução e logs em runtime real ou fallback.',
  },
  {
    id: 'controlled-sleep',
    name: 'controlled-sleep',
    preferredMode: 'container-linux',
    command: '/bin/sh',
    args: ['-c', 'echo starting && sleep 10 && echo finished'],
    objective: 'Demonstrar Running -> Completed e stop manual.',
  },
  {
    id: 'runtime-diagnostics',
    name: 'runtime-diagnostics',
    preferredMode: 'container-linux',
    command: '/bin/sh',
    args: ['-c', 'hostname && uname -a && ps'],
    objective: 'Evidenciar diferenças do ambiente isolado.',
  },
  {
    id: 'rootfs-inspection',
    name: 'rootfs-inspection',
    preferredMode: 'container-linux',
    command: '/bin/sh',
    args: ['-c', 'ls -la / && ls -la /bin'],
    objective: 'Comprovar rootfs montado e estrutura interna.',
  },
  {
    id: 'fallback-demo',
    name: 'fallback-demo',
    preferredMode: 'container-linux',
    command: '/bin/sh',
    args: ['-c', 'echo fallback validation'],
    objective: 'Exibir fallback automático quando container-linux não estiver disponível.',
  },
  {
    id: 'postgres-demo',
    name: 'postgres-demo',
    preferredMode: 'container-linux',
    command: 'minidock-postgres-demo',
    args: [],
    objective: 'Demonstrar workload stateful com logs de inicialização e readiness.',
  },
]

export const NAV_ITEMS = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'demos', label: 'Demonstrações' },
  { key: 'workloads', label: 'Workloads' },
  { key: 'insights', label: 'Insights' },
  { key: 'activity', label: 'Atividade' },
] as const

export type NavigationKey = (typeof NAV_ITEMS)[number]['key']
