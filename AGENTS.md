---
globs: []
alwaysApply: true
description: AI Coder Unified Ruleset - Concise memory bank system + coding standards.
---

# Memory Bank System

## Files (`.agents/rules/memory-bank/`)
- `brief.md` - project overview (read-only)
- `product.md` - goals, UX, features, roadmap, acceptance criteria
- `architecture.md` - architecture, components, data flows, integrations, scalability
- `tech.md` - standards, patterns, conventions, testing, security, deployment
- `context.md` - decisions, rules, logic summaries, dependencies
- `tasks.md` - workflows (TDD, refactoring, reviews, debugging)

## Startup
1. Load all files: brief → product → architecture → tech → context → tasks
2. Print `Memory Bank: Active` (count, timestamps) or `Memory Bank: Missing` (list files)
3. Summarize project (3–6 bullets): purpose, architecture, tech, features, state, constraints
4. Use as single source of truth; never scan codebase unless user commands
5. Validate consistency; report conflicts immediately

## Rules
**Reference before:** architecture changes, naming, implementation, coding standards, testing, refactoring, reviews, security, performance

**Scan only when:** user commands `update memory bank`, `initialize memory bank`, `audit codebase`, or confirms `OK to update Memory Bank?`

**Updates:**
- Extract long-term knowledge only (no raw code, no ephemeral details)
- Update `context.md` first, cascade to others
- Confirm before overwriting >50 lines or critical decisions
- Max 300 lines/file; use concise summaries, cross-references
- Timestamp + rationale for significant updates
- Log reversed/changed decisions in `decisions_log.md`

**Permissions:**
- Editable: `context.md`, `architecture.md`, `tech.md`, `product.md`, `tasks.md`
- Read-only: `brief.md` (requires approval)
- Structural changes need explicit approval

**Thread drift:** Detect deviation; suggest `Update Memory Bank and start fresh thread?`; document gaps

**Consistency:** Align proposals with all MD files; flag conflicts before implementation; provide resolution options; warn on breaking changes

# Programming Guidelines

## Performance & Concurrency
- Object pooling, preallocate collections, optimize cache locality
- Minimize heap allocations; prefer stack for short-lived objects
- Zero-copy patterns, controlled thread/worker pools, atomic ops
- Lazy init, share immutable objects, proper cancellation
- Avoid lock contention; prefer lock-free or partitioned locks
- Profile before optimizing; measure improvements

## I/O & Memory
- Buffered I/O, batch DB operations, connection pooling
- Respect platform memory limits, reuse hot-path objects
- Use value types for critical paths, proper disposal patterns
- Stream large data, compress at rest/transit, circuit breakers

## Core Principles
- **SOLID:** SRP, OCP, LSP, ISP, DIP
- **Generics:** type-safe reuse, constraints, cross-component interfaces
- **Idiomatic:** small focused functions (<50 lines), descriptive names, DI over global state, composition > inheritance
- **Error Handling:** typed errors with context, domain wrapping, recovery paths, detailed logging, explicit error surface

## Testing & Patterns
- **Tests:** unit (logic), integration (interactions), table-driven (scenarios), benchmarks (performance), mocks/stubs, property-based; 80%+ coverage for critical paths
- **Patterns:** Factory, Strategy, Observer, Decorator, Command, Template Method, Builder, Repository, Singleton (use DI), Adapter

## Anti‑Patterns
- Excessive complexity, god interfaces, silent errors, overuse of reflection
- Tight coupling, premature optimization, magic numbers, duplication
- Feature bloat, circular dependencies, deep nesting (early returns preferred)

## Additional
- Design for composition/extensibility, strong typing, clear interface docs
- Consistent naming, pure functions where appropriate, proper logging
- Security best practices (validation, encoding, least privilege)
- Readable/maintainable > clever code, meaningful commits, document non-obvious decisions

# Agent Behavior
- Load Memory Bank at task start; use as primary context
- Follow guidelines; verify alignment before implementation
- Never read source files unless user requests
- Update Memory Bank only with explicit authorization
- Warn contradictions; never store raw code; summarize only
- Detect thread drift; suggest updates when new knowledge emerges
- Provide specific, actionable guidance
- Balance thoroughness with pragmatism; ask when uncertain
