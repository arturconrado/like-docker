# MiniDock Executive AI

**Subtítulo:** Inteligência de Runtime para Liderança Técnica

MiniDock Executive AI é um MVP local-first para demonstrar engenharia full stack com UX executiva.
O produto oferece execução de workloads, observabilidade visual e narrativa de decisão com heurísticas locais.

## 1. Visão geral
- Backend em Go com estado em memória, eventos e SSE.
- Frontend em React + Vite + TypeScript + Tailwind.
- Camada de AI UX local: resumo, risco, insights e próxima ação.
- Runtime com três modos:
  - `demo`
  - `processo-local`
  - `container-linux` (avançado, Linux only)

## 2. Funcionalidades
- Criar workloads com preset de comando.
- Executar comandos reais locais.
- Visualizar logs, status, duração e exit code.
- Parar e remover workloads.
- Ler resumo executivo global + insights por workload.
- Ver classificação de risco (`Safe`, `Review`, `Risky`).
- Detectar capabilities do host via API.
- Biblioteca com exemplos clicáveis para demo guiada.
- Área **Demonstrações** com jornada guiada em etapas.
- PostgreSQL Demo com metadados operacionais (porta, data dir, readiness, modo efetivo).

## 3. Arquitetura
```text
/like-docker
  /apps
    /web     # React + Vite + TypeScript + Tailwind
    /api     # Go HTTP API + runtime engines + heurísticas
  /examples  # rootfs/bundles de demonstração
  /docs      # arquitetura e roteiro de demo
  /scripts   # dev + preparação de rootfs
```

Detalhes: [`docs/architecture.md`](docs/architecture.md)
Comparativo com artigo InfoQ: [`docs/report-infoq-comparison.md`](docs/report-infoq-comparison.md)

## 4. Como rodar localmente

### Pré-requisitos
- Go 1.22+
- Node.js 20+
- npm 10+

### Setup
```bash
make setup
```

### Rodar API + Web
```bash
make dev
```

- Web: `http://localhost:5173`
- API: `http://localhost:8080`

### Rodar separadamente
```bash
make api
make web
```

### Testes e build
```bash
make test
make build
```

## 5. Modos de execução

### `demo`
Execução simulada para storytelling rápido e ambiente sem dependências.

### `processo-local`
Execução real com `exec.CommandContext`, compatível com qualquer ambiente local.

### `container-linux`
Execução avançada em Linux com:
- namespaces `UTS`, `PID` e `mount`
- rootfs dedicado com `pivot_root`
- hostname isolado
- processo PID 1 no namespace isolado
- cgroups com estratégia `v2 -> v1` para limites de recursos

### Cadeia de fallback
Quando `container-linux` não está disponível:
1. fallback para `processo-local`
2. se necessário, fallback para `demo`

## 6. Execução em Container Linux Real (modo avançado)

### O que esse modo faz
- Isola a workload em namespaces Linux.
- Executa dentro de rootfs dedicado via `pivot_root`.
- Mantém logs, stop/remove e insights no mesmo fluxo da UI.
- Expõe evidências técnicas (`pivotRootApplied`, `cgroupPath`, `cgroupVersion`).

### O que esse modo não é
- Não é Docker Engine completo.
- Não é runtime de produção.
- Não promete hardening/segurança de produção.

### Requisitos mínimos
- Host Linux.
- Privilégios root para API (`sudo`).
- Rootfs local com `/bin/sh` disponível.

### Preparar rootfs de demo
```bash
make prepare-rootfs
```

Isso cria (por padrão) `examples/rootfs/demo` copiando binários essenciais do host.

Opcional:
```bash
./scripts/prepare-rootfs.sh /caminho/customizado/rootfs
```

Defina o rootfs usado pela API:
```bash
export MINIDOCK_CONTAINER_ROOTFS=./examples/rootfs/demo
```

### Rodar API com privilégios (Linux)
```bash
sudo MINIDOCK_RUNTIME_MODE=container-linux make api
```

### Como validar que funcionou
- Endpoint `GET /api/capabilities` deve indicar:
  - `supportsContainers: true`
  - `supportsPivotRoot: true`
  - `supportsCgroups: true` (quando disponível)
  - `rootfsAvailable: true`
- Na UI, executar `container-identity-check`.
- No drawer, validar:
  - modo efetivo `container-linux`
  - `pivotRootApplied=true`
  - `cgroupPath` preenchido
  - `rootfs` preenchido
  - `containerHostname` preenchido
  - `mainPid` preenchido

### Como o fallback acontece
- Se Linux/root/rootfs não estiverem disponíveis, o backend registra fallback automático.
- A UI mostra modo efetivo e motivo (`fallbackReason`).
- A demo continua funcional sem quebrar fluxo ponta a ponta.

## 7. Exemplos recomendados para demonstração
- `container-identity-check`: `hostname && pwd && ls /`
- `hello-container`: `echo hello from container`
- `controlled-sleep`: `echo starting && sleep 10 && echo finished`
- `runtime-diagnostics`: `hostname && uname -a && ps`
- `rootfs-inspection`: `ls -la / && ls -la /bin`
- `fallback-demo`: `echo fallback validation`

Todos estão disponíveis no dashboard com botão **Executar exemplo**.

## 8. Demonstrações Guiadas

A navegação principal agora inclui a área **Demonstrações**, com:
- catálogo visual de demos prontas;
- execução com um clique;
- jornada guiada por etapas (contexto, capabilities, preparação, execução, validação e encerramento);
- painel de evidências técnicas e operacionais;
- resumo executivo da demonstração.

Demos incluídas:
1. `hello-container`
2. `filesystem-inspection`
3. `runtime-diagnostics`
4. `controlled-sleep`
5. `postgres-demo`

`postgres-demo` é tratado como demonstração principal de maturidade e aparece com destaque na interface.

### PostgreSQL Demo
- Nome: `postgres-demo`
- Tipo: `Database`
- Preferência: `container-linux`
- Evidências exibidas na UI:
  - logs de inicialização;
  - readiness;
  - porta;
  - data directory;
  - pid principal;
  - hostname quando aplicável.

Comportamento por modo:
- `demo`: simulação plausível com sinais de readiness.
- `processo-local`: tenta execução simplificada com binários locais (`postgres`/`initdb`); se indisponível, fallback para `demo`.
- `container-linux`: tenta execução isolada com rootfs; se faltar requisito, fallback controlado para `processo-local` ou `demo`.

## 9. Fluxo recomendado para apresentação
1. Abrir **Demonstrações**.
2. Executar **PostgreSQL Demo**.
3. Acompanhar stepper de preparação e execução.
4. Mostrar logs e readiness no card de status PostgreSQL.
5. Abrir detalhes completos da workload.
6. Destacar modo efetivo, fallback (se houver), insights e resumo executivo.

## 10. Endpoints principais
- `GET /health`
- `GET /api/workloads`
- `POST /api/workloads`
- `GET /api/workloads/:id`
- `POST /api/workloads/:id/stop`
- `DELETE /api/workloads/:id`
- `GET /api/workloads/:id/logs`
- `GET /api/events`
- `GET /api/stream`
- `POST /api/demo/seed`
- `GET /api/demos`
- `GET /api/demos/:id`
- `POST /api/demos/:id/run`
- `GET /api/demos/:id/validate`
- `GET /api/summary`
- `GET /api/capabilities`

## 11. Variáveis de ambiente
Arquivo de exemplo: [`.env.example`](.env.example)

- `API_PORT=8080`
- `MINIDOCK_RUNTIME_MODE=processo-local|demo|container-linux`
- `MINIDOCK_SEED_DEMO=true|false`
- `MINIDOCK_CONTAINER_ROOTFS=./examples/rootfs/demo`
- `MINIDOCK_CGROUP_PIDS_MAX=256`
- `MINIDOCK_CGROUP_MEMORY_MAX=1073741824`
- `MINIDOCK_CGROUP_CPU_MAX=200000 100000`
- `VITE_API_BASE_URL=http://localhost:8080`

## 12. Exemplo Didático (estilo artigo)

Para estudo técnico, o repositório inclui um binário standalone:

- `apps/api/cmd/container100` (Linux only)

Uso:

```bash
cd apps/api
go run ./cmd/container100 run ./../../examples/rootfs/demo /bin/sh -c "hostname && pwd && ls /"
```

Este comando é educacional e não deve ser usado em produção.

## 13. Limitações
- Estado em memória (sem persistência).
- Sem autenticação.
- Sem orquestração Kubernetes.
- Sem runtime container completo.
- Sem garantias de segurança para produção.
