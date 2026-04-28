# Relatório Comparativo: InfoQ vs MiniDock Executive AI

## 1) Paridade Técnica com o Artigo

Referência: [Build Your Own Container Using Less than 100 Lines of Go (InfoQ)](https://www.infoq.com/articles/build-a-container-golang/)

MiniDock implementa os blocos centrais do walkthrough:

- Skeleton `run/child` com reexecução do binário atual.
- Namespaces Linux no processo filho:
  - `CLONE_NEWUTS`
  - `CLONE_NEWPID`
  - `CLONE_NEWNS`
- Root filesystem isolado com `pivot_root`:
  - bind mount do rootfs
  - criação de `oldroot`
  - `pivot_root`
  - `chdir("/")`
  - `unmount` do oldroot com `MNT_DETACH`
- Inicialização do ambiente dentro do container:
  - configuração de hostname
  - montagem de `/proc`
  - execução do comando final via `syscall.Exec`
- Cgroups reais com fallback de versão:
  - v2 prioritário (`pids.max`, `memory.max`, `cpu.max`, `cgroup.procs`)
  - v1 fallback (`pids`, `memory`, `cpu`)

## 2) O Que o MiniDock Tem a Mais (e Melhor)

Além do protótipo didático do artigo, o MiniDock entrega capacidades de produto:

- Fallback resiliente de runtime:
  - `container-linux` -> `processo-local` -> `demo`
  - fallback também durante falhas de startup em `container-linux`.
- API HTTP completa para ciclo de vida:
  - create/list/get/stop/delete workloads
  - logs, eventos, stream, summary, capabilities.
- Observabilidade operacional no payload:
  - `pivotRootApplied`
  - `cgroupPath`
  - `cgroupVersion`
  - `mainPid`, `rootfs`, `containerHostname`, `readiness`.
- UX executiva local-first:
  - classificação de risco
  - resumo executivo
  - insights e próxima ação sugerida.
- Demonstração guiada de PostgreSQL com evidências operacionais.
- Frontend premium com estados de erro/loading/vazio/sucesso e atualização em tempo real.
- Testes automatizados backend + frontend.

## 3) Limites Atuais e Diferenças para Runtimes de Produção

MiniDock permanece um MVP demonstrável, não um engine de produção:

- Sem namespace de rede dedicado (`CLONE_NEWNET`) com veth/routing.
- Sem user namespace e sem mapeamento UID/GID.
- Sem hardening avançado (drop de capabilities, seccomp, AppArmor/SELinux, etc.).
- Sem layered filesystem/OCI image engine completo.
- Estado em memória, sem persistência transacional.

Esses limites são intencionais para manter:

- executabilidade local;
- simplicidade do setup;
- demonstração ponta a ponta com boa UX e alta clareza técnica.
