# Roteiro de Demo

## 1. Abrir o produto
- Executar `make dev`.
- Mostrar dashboard inicial já populado (seed demo).
- Abrir painel de `Capabilities do Host` e destacar o modo recomendado.

## 2. Ler narrativa executiva
- Destacar KPIs, saúde do runtime e resumo executivo global.
- Abrir painel de Insights para mostrar fila de revisão.

## 3. Executar demonstração guiada principal
- Abrir área `Demonstrações`.
- Rodar `PostgreSQL Demo`.
- Mostrar stepper da jornada (contexto -> ambiente -> preparação -> execução -> validação).
- Evidenciar `status`, `modo`, `porta`, `data dir`, `readiness`, `logs`.

## 4. Operar ciclo de vida
- Rodar `controlled-sleep` no catálogo de demos.
- Mostrar status `Running` -> `Completed` ou interromper com `Stop`.
- Abrir drawer para exibir modo solicitado vs modo efetivo, rootfs/hostname/PID e logs.

## 5. Diagnóstico e fallback
- Rodar `runtime-diagnostics`.
- Destacar classificação de risco, insights e próxima ação sugerida.
- Mostrar feed de atividade e sinal de fallback quando aplicável.

## 6. Encerrar com valor de produto
- Mostrar que não há cloud, banco ou IA externa.
- Reforçar que o MVP é local-first, visualmente premium, com runtime avançado opcional e fallback confiável.
