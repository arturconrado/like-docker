# Arquitetura - MiniDock Executive AI

## Visão geral
- `apps/api` (Go): API HTTP, estado em memória, arquitetura de runtime por engine, heurísticas de AI UX, eventos e SSE.
- `apps/web` (React + Vite + TypeScript): UI executiva, dashboard, lista de workloads, drawer de detalhes, feed de atividade e fallback de polling.

## Backend
- Estado em memória com `map` de workloads + buffer circular de eventos.
- Arquitetura de runtime:
  - `DemoEngine`
  - `LocalProcessEngine`
  - `LinuxContainerEngine` (Linux + namespaces + pivot_root + cgroups)
- Execução real via `exec.CommandContext` no modo `processo-local`.
- Modo `container-linux` com isolamento real (UTS/PID/MNT namespace + rootfs dedicado via `pivot_root` + cgroup v2/v1).
- Fallback automático: `container-linux` -> `processo-local` -> `demo`.
- Modo `demo` para execução simulada e seed de dados plausíveis.
- Endpoints centrais:
  - `GET /health`
  - `GET/POST /api/workloads`
  - `GET/POST/DELETE /api/workloads/:id`
  - `GET /api/workloads/:id/logs`
  - `GET /api/demos`
  - `GET /api/demos/:id`
  - `POST /api/demos/:id/run`
  - `GET /api/demos/:id/validate`
  - `GET /api/events`
  - `GET /api/stream` (SSE)
  - `POST /api/demo/seed`
  - `GET /api/summary`
  - `GET /api/capabilities`

## AI UX (heurísticas locais)
- Nome inteligente por comando.
- Resumo executivo por padrão de comando.
- Classificação de risco (`Safe`, `Review`, `Risky`).
- Insights por workload (2 a 4 linhas).
- Próxima ação sugerida.
- Resumo executivo global para dashboard.

## Frontend
- React Query para dados e mutações.
- SSE para atualização em tempo real com fallback para polling quando necessário.
- Estados visuais: loading, vazio, erro parcial, backend indisponível.
- Interface dark mode premium com tipografia local e microanimações via Framer Motion.
- Área `Demonstrações` com catálogo e runner guiado (stepper, evidence panel, capability panel e resumo executivo).
- Card especializado para estado do PostgreSQL Demo.
- Exibição de capabilities do host e disponibilidade de `container-linux`.
