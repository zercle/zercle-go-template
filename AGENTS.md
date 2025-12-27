---
globs: []
alwaysApply: true
description: AI Coder Unified Ruleset - Memory Bank persistence system replacing full codebase scans.
---

# Memory Bank System

## Location & Files
`.agents/rules/memory-bank/`
- `brief.md` - user-only project overview (read-only)
- `product.md` - goals, UX, features, roadmap
- `architecture.md` - architecture, boundaries, flows, modules
- `tech.md` - standards, patterns, testing rules
- `context.md` - decisions, domain rules, file summaries
- `tasks.md` - workflows (test, refactor, review)

## Startup Behavior
1. Load all Memory Bank files
2. Print `Memory Bank: Active` or `Memory Bank: Missing`
3. Summarize project (3–6 bullets) from Memory Bank only
4. Use this context for all reasoning
5. Never scan codebase unless user approves

## Memory Bank Rules
**Always reference before:** architecture changes, naming, implementation, coding standards, testing, refactoring, reviews.

**Never scan files unless:** `update memory bank`, `initialize memory bank`, or AI asks `OK to update Memory Bank?`

**Updating rules:**
- Extract long-term knowledge only (no raw code)
- Update `context.md` first, then others as needed
- Ask before overwriting major sections
- Max 300 lines per file

**Permissions:**
- Editable: `context.md`, `architecture.md`, `tech.md`, `product.md`, `tasks.md`
- Read-only: `brief.md` (needs user approval)

**Thread drift:** Suggest `Update Memory Bank and start fresh thread?`; new thread must reload Memory Bank.

**Consistency:** Enforce with architecture.md, tech.md, product.md, context.md. Confirm conflicts with user + warn.

# Generic Programming Guidelines

## Performance & Concurrency
- Pool objects, preallocate collections, optimize cache locality
- Minimize heap, prefer stack, zero-copy when possible
- Controlled pools, atomic ops, lazy init, immutable sharing, cancellation tokens

## I/O & Memory
- Buffered I/O, batch ops, connection pooling
- Know platform model, reduce hot-path allocations, use value types, dispose properly

## Core Principles
- **SOLID**: SRP, OCP, LSP, ISP, DIP
- **Generics**: type-safe reuse, constraints, interfaces, collections
- **Idiomatic**: small focused functions, descriptive names, no globals, composition over inheritance, minimal interfaces
- **Error handling**: explicit handling, typed errors, wrap context, robust recovery

## Testing & Patterns
- Unit tests critical logic, table-driven tests, benchmarks, mocks/stubs, property-based tests
- Patterns: Factory, Strategy, Observer, Decorator, Command, Template Method

## Anti‑Patterns
- Excessive concurrency, god interfaces, ignored errors, overuse of reflection/dynamic, tight coupling, premature optimization

## Additional
- Design for composition/extensibility, intentional strong typing, clear interface docs, consistent conventions

# Agent Behavior Summary
- Load Memory Bank every task, use as primary context
- Follow all guidelines above, ensure consistency
- Never read source unless asked, update Memory Bank only with authorization
- Warn contradictions, never store raw code