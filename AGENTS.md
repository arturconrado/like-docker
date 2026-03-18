# AGENTS.md

## Objetivo do projeto

Este repositório contém o projeto **MiniDock Executive AI**, um MVP técnico demonstrável, inspirado visualmente em consoles de runtime/container, com backend em Go, frontend em React e uma camada de UX “inteligente” baseada em heurísticas locais.

O objetivo principal é entregar um produto:
- funcional
- simples
- bonito
- rodável localmente
- publicável no GitHub
- útil como showcase técnico e demo

Este projeto **não é**:
- um clone do Docker
- uma plataforma de produção
- uma solução real de orquestração
- um runtime seguro para uso produtivo

## Princípios obrigatórios

Ao trabalhar neste repositório, siga estes princípios:

1. **Executável localmente acima de tudo**
   - qualquer decisão deve favorecer facilidade de execução local
   - se algo dificultar rodar localmente, simplifique

2. **Polimento visual importa**
   - a interface deve parecer premium
   - evitar estética de CRUD genérico
   - evitar aparência de hackathon improvisado

3. **Simplicidade vence complexidade**
   - não adicionar infraestrutura desnecessária
   - não usar serviços externos sem necessidade real
   - não criar abstrações excessivas

4. **Produto demonstrável > perfeição técnica**
   - a experiência ponta a ponta deve funcionar
   - uma demo completa vale mais do que um runtime low-level quebradiço

5. **IA local e crível**
   - não usar API externa de IA
   - a “inteligência” do produto deve vir de heurísticas, regras e boa escrita UX
   - evitar gimmicks exagerados

## Resultado esperado

O projeto final deve permitir:
- criar workloads
- executar comandos locais simples
- visualizar logs
- parar workloads
- remover workloads
- mostrar status e duração
- exibir resumo executivo
- exibir classificação de risco
- exibir insights e próxima ação sugerida
- apresentar dashboard com métricas resumidas

## Arquitetura esperada

Estrutura preferencial:

- `apps/web` → frontend React
- `apps/api` → backend Go
- `docs/` → documentação complementar
- `README.md` → instruções principais
- `Makefile` → atalhos de desenvolvimento
- `.env.example` → variáveis de ambiente de exemplo

## Regras para o frontend

O frontend deve:
- usar React + Vite + TypeScript
- usar Tailwind CSS
- ter dark mode por padrão
- ter layout limpo, premium e executivo
- usar animações discretas
- possuir estados de loading, vazio, erro e sucesso
- ser resiliente a indisponibilidade temporária do backend

Áreas principais:
- sidebar
- dashboard
- lista de workloads
- detalhes em drawer/painel lateral
- feed de atividade
- cards de KPI
- insights executivos

## Regras para o backend

O backend deve:
- usar Go
- manter estado em memória
- expor API HTTP simples
- controlar ciclo de vida das workloads
- armazenar logs
- suportar parada e remoção
- preferir fallback confiável se o runtime low-level for frágil

Endpoints esperados:
- `GET /health`
- `GET /api/workloads`
- `POST /api/workloads`
- `GET /api/workloads/:id`
- `POST /api/workloads/:id/stop`
- `DELETE /api/workloads/:id`
- `GET /api/workloads/:id/logs`
- `GET /api/events`
- opcional: `GET /api/stream`

## Estratégia de runtime

Se possível, pode existir uma implementação opcional inspirada em conceitos de container como namespaces Linux e `pivot_root`.

Mas isso **não pode** comprometer a executabilidade local do MVP.

Se houver risco de fragilidade:
- usar fallback com execução local de processos
- manter a mesma UX
- preservar a narrativa do produto

## AI UX obrigatória

A camada “AI UX” deve ser local e baseada em heurísticas.

Implementar:
- nome inteligente da workload
- resumo executivo do comando
- classificação de risco
- insights curtos
- ação sugerida
- resumo executivo global no dashboard

Essa camada deve soar:
- madura
- elegante
- útil
- não caricata

## Critérios de qualidade

Antes de considerar qualquer tarefa concluída, verificar:

- isso ficou mais fácil de rodar localmente?
- isso deixou a demo mais forte?
- isso melhorou a clareza do produto?
- isso aumentou a credibilidade visual?
- isso evitou complexidade desnecessária?

Se a resposta for “não”, simplifique.

## Estratégia de implementação

Seguir esta ordem sempre que possível:

1. estrutura do projeto
2. shell visual
3. dashboard
4. lista de workloads
5. API backend
6. engine de execução
7. heurísticas de AI UX
8. logs e eventos
9. README
10. polimento final

## Restrições

Não fazer:
- integração com Kubernetes
- autenticação complexa
- banco de dados para o MVP
- image registry
- engine real de containers completa
- promessas de segurança de produção
- dependências externas de IA pagas

## Tom do produto

O produto deve comunicar:
- confiança
- controle
- clareza
- maturidade técnica
- qualidade de produto

Pense como:
“um conceito premium de plataforma técnica que uma consultoria ou startup enterprise teria orgulho de demonstrar.”

## Definição de pronto

O projeto está pronto quando:
- outro dev consegue clonar e rodar
- a UI parece premium
- a API funciona
- a criação de workload funciona
- o fluxo ponta a ponta funciona
- os detalhes mostram logs e insights
- a UX de risco e resumo está visível
- o README está claro
- o projeto está publicável

## Diretriz adicional: experiência guiada de demonstração

Este projeto deve incluir uma experiência guiada de demonstração visível na interface.

Objetivos dessa área:
- evidenciar a completude da ferramenta
- facilitar apresentação e demo
- mostrar o valor do produto sem depender de explicação externa
- destacar workloads mais sofisticadas, especialmente PostgreSQL

A demonstração principal deve ser o **PostgreSQL Demo**.

Essa demonstração precisa:
- existir como item visível do catálogo
- ter uma jornada guiada clara
- exibir evidências de execução
- destacar status, logs, modo, readiness e recomendações
- manter fallback confiável se o host não suportar o modo avançado