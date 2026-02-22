---
globs: []
alwaysApply: true
description: AI Coder Unified Ruleset - Memory bank system + coding standards for backend development.
---

# Quick Reference

## Priority Levels
| Priority | When | Examples |
|----------|------|----------|
| **P1** | Non-negotiable | Memory Bank startup, error handling, security |
| **P2** | Context-dependent | Concurrency patterns, testing strategies |
| **P3** | Recommended | Code style, documentation, naming |

## Thresholds
- Max 50 lines/function | Max 200 lines/Memory Bank file
- 80%+ test coverage for critical paths | 5-10 items/checklist | 3-6 bullets/summary

---

# Memory Bank System

## Files
```
.agents/rules/memory-bank/
├── brief.md        # Project overview (READ-ONLY)
├── product.md      # Goals, features, roadmap
├── architecture.md # Components, data flows
├── tech.md         # Standards, patterns, testing
├── context.md      # Decisions, rules, dependencies
├── tasks.md        # Workflows (TDD, reviews)
└── decisions_log.md # Reversed decisions
```

## Startup (P1)
1. Load: `brief` → `product` → `architecture` → `tech` → `context` → `tasks`
2. Print: `Memory Bank: Active (6 files)` OR `Memory Bank: Missing: [list]`
3. Summarize: 3-6 bullets (purpose, architecture, tech, features, state, constraints)

## Reference Rules
**P1 - Always reference before:** architecture changes, naming, implementation, coding standards, testing, refactoring, security, performance

**P2 - Scan only when:** user commands `update memory bank`, `initialize memory bank`, `audit codebase`, or confirms `OK to update Memory Bank?`

## Update Process
1. Extract knowledge (summarize, don't copy code)
2. Update `context.md` first
3. Cascade to relevant files
4. Confirm before overwriting >50 lines or critical decisions
5. Add timestamp: `YYYY-MM-DD HH:MM - rationale`
6. Log reversed decisions in `decisions_log.md`

Max 200 lines/file

## Permissions
| File | Edit | Delete |
|------|------|--------|
| `brief.md` | No | No |
| Others | Yes | No |

Structural changes: always require approval

## Deviation & Conflict
**Deviation:** Identify type (minor/major/breaking), assess impact, propose options, ask: `Update Memory Bank and start fresh thread?`

**Conflict:** Identify which files, check timestamps (most recent wins), prioritize: brief > product > architecture > tech > context, report to user

---

# Programming Guidelines

## Performance & Concurrency
- GC pause: < 2ms | Worker pool: `2 * runtime.NumCPU()` CPU-bound
- Batch size: 100-1000 (tune via profiling) | Memory: < 10MB/hour sustained

### P1 Rules
- Preallocate collections when the final size is known in advance
- Use object pooling to reuse frequently allocated objects in hot paths
- Avoid per-iteration allocations in tight loops - reuse buffers or pool objects

### P2 Rules
- Lock contention: profile with appropriate tools, aim < 5% in mutex
- Worker lifecycle: always pair with context cancellation for clean shutdown
- Atomics for simple counters/flags; mutex for complex state

### Decisions
- **Hot path + allocation-heavy** → pooling + preallocation
- **Concurrent writes** → read-write locks or lock-free if <10 items
- **Worker pool** → errgroup pattern with context cancellation

## I/O & Memory

- **Buffered I/O:** Wrap file readers with buffered readers to reduce system calls
- **Batch DB operations:** Group multiple inserts/updates into single batch operations using parameterized queries
- **Connection pooling:** Configure max open connections, idle connections, and connection lifetime (e.g., 25 max, 5 idle, 5-minute lifetime)

**Thresholds:** DB: 25 max, 5 idle, 5min | HTTP: 30s default, 5s internal | Circuit: 5 failures → open, 30s half-open

## SOLID Examples

- **SRP (Single Responsibility Principle):** Separate concerns into distinct types - each type should have one reason to change. For example, a user creator handles only database operations, while an email notifier handles only messaging.

- **DIP (Dependency Inversion Principle):** Depend on abstractions rather than concrete implementations. Define interfaces for repositories and services that describe what operations are needed without specifying how they're implemented.

## Functions
**Max 50 lines** - break larger functions into focused helpers:

- Decompose complex operations into a main function that orchestrates multiple smaller functions
- Each helper function should perform one clear task with a descriptive name
- The main function chains the helpers together, handling the control flow

## Error Handling (P1)

- Create domain-specific error types that wrap underlying errors with context
- Include relevant information in errors (resource type, identifier, operation attempted)
- Use error wrapping to preserve the error chain while adding actionable details

## Testing
| Component | Coverage |
|-----------|----------|
| Business logic | 80% |
| Critical paths | 90% |
| Error handling | 70% |
| Public APIs | 85% |

**Unit (P1):** Write table-driven tests with clear setup, execution, and assertion phases. Mock external dependencies to isolate the code under test.

**Integration (P2):** Use real databases or containerized test instances. Test HTTP handlers end-to-end with actual requests.

**Benchmarks (P3):** Profile before optimizing. Run benchmarks to measure performance improvements.

## Patterns
| Pattern | Use When | Example |
|---------|----------|---------|
| Factory | Complex creation | Service constructors with optional configuration |
| Strategy | Interchangeable algos | Storage strategy interfaces |
| Repository | Data access | Repository interfaces for data access |
| Builder | Complex config | Fluent configuration builders |
| Decorator | Add behavior | HTTP middleware, logging wrappers |

## Anti-Patterns
**P1 - Never:** Silent errors, panic in production, global state, hardcoded secrets
**P2 - Avoid:** God interfaces (>5 methods), deep nesting (>3 levels), magic numbers, premature optimization

---

# Agent Behavior

## Loading (P1)
Every task: Load 6 Memory Bank files → Print status + 3-6 bullet summary

## Tools
| Tool | Use For | Avoid |
|------|---------|-------|
| `codebase_search` | Unknown code | Known files |
| `read_file` | Specific paths | Discovery |
| `execute_command` | Tests, builds | File creation |
| `write_to_file` | New files | Existing files |
| `edit_file` | Modifications | Rewrites |

**Selection:** Unknown → `codebase_search` | Modify → `read_file`+`edit_file` | Create → `write_to_file` | Run → `execute_command`

## Error Handling
Tools fail: Parse error → Check common causes (exists? syntax? permission?) → Retry once → Ask user

Rules conflict: Flag → Apply priority (P1>P2>P3) → Propose options → Wait

## Thread Drift
**Triggers:** Feature not in Memory Bank, off-topic discussion, new tech proposed
**Response:** Suggest `Update Memory Bank?` + document gaps in `context.md`

## Ask User When
- Missing critical info (API keys, credentials)
- Unclear requirements
- Multiple valid approaches
- Risk of breaking changes

**Proceed Without Asking:** Routine implementation, established patterns, small refactors, test additions

---

# Metrics
| Metric | Value |
|--------|-------|
| Max function lines | 50 |
| Max Memory Bank file | 200 |
| Test coverage | 80% |
| DB max connections | 25 |
| DB idle | 5 |
| HTTP timeout | 30s/5s |
| Circuit failures | 5 |
| Circuit reset | 30s |
| Worker pool | 2*NumCPU |
