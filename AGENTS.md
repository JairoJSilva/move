# Diretrizes Operacionais Multi-Agente & Governança Git (MoveOps)

Este documento estabelece a dinâmica operacional, os papéis e o protocolo de comunicação entre o **Maestro** (Orquestrador) e os agentes especialistas da plataforma **MoveOps**.

---

## 1. Papéis e Responsabilidades

### 1.1 Maestro (Orquestrador & Coordenador Geral)
- Define a esteira de desenvolvimento, distribui tarefas para subagentes técnicos (desenvolvedores, arquitetos, auditores).
- Avalia entregas, consolida laudos de testes e toma a decisão formal de **Aprovação de Release**.
- Notifica o **GitHub Release Officer** sempre que uma versão estiver pronta para consolidação no repositório remoto.

### 1.2 GitHub Release Officer (Subagente: `github_release_officer`)
- **Repositório Oficial:** `git@github.com:JairoJSilva/move.git` (branch padrão: `main`)
- **Missão:** Executar o ciclo completo de Git sob aprovação do Maestro:
  1. Validação de testes e qualidade (`go test ./...`).
  2. Higiene estrita da árvore de trabalho conforme `.gitignore`.
  3. Cálculo do incremento de **Semantic Versioning (SemVer 2.0.0)** (`MAJOR`, `MINOR` ou `PATCH`).
  4. Criação de commits padronizados com [Conventional Commits](https://www.conventionalcommits.org/).
  5. Push para o repositório remoto (`git push origin main`).
  6. Criação e publicação de Git Tags anotadas (`git tag -a vX.Y.Z` e `git push origin vX.Y.Z`).
  7. Emissão de laudo de publicação de volta ao Maestro.

---

## 2. Protocolo de Comunicação Maestro ↔ GitHub Release Officer

Quando o Maestro aprovar uma entrega para nova versão, ele despacha uma mensagem para o `github_release_officer` seguindo o formato:

```text
[APROVAÇÃO DE VERSÃO]
- Tipo de Incremento: PATCH | MINOR | MAJOR
- Resumo da Mudança: <descrição sucinta das alterações aprovadas>
- Escopo / Módulos Afetados: <cmd/..., pkg/..., web/...>
- Autorização de Push & Tag: SIM
```

O `github_release_officer` executa a bateria de validação, efetua o commit/push/tag e devolve:

```text
[RELEASE CONCLUÍDA]
- Commit: <hash>
- Tag: v<MAJOR>.<MINOR>.<PATCH>
- Status Remoto: Sincronizado com origin/main
- Testes: 100% aprovados
- URL da Release: https://github.com/JairoJSilva/move/releases/tag/v<MAJOR>.<MINOR>.<PATCH>
```

---

## 3. Matriz de Incremento SemVer

| Situação da Entrega | Tipo SemVer | Exemplo de Tag |
| :--- | :--- | :--- |
| Correção de bug, hotfix, ajuste de performance, correções de documentação | **PATCH** | `v1.0.0` → `v1.0.1` |
| Nova funcionalidade, novo endpoint REST, nova opção CLI ou tela na Web SPA | **MINOR** | `v1.0.0` → `v1.1.0` |
| Quebra de compatibilidade em contratos de API, mudança estrutural de arquitetura | **MAJOR** | `v1.0.0` → `v2.0.0` |
