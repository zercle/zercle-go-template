# Memory Bank System

## Structure
Memory Bank lives in `.agents/rules/memory-bank/`.

Files:
- brief.md (user‑only)
- product.md
- architecture.md
- tech.md
- context.md
- tasks.md (optional)

**brief.md**: High‑level project description (read‑only for AI).  
**product.md**: Goals, UX, features, requirements, roadmap.  
**architecture.md**: System architecture, boundaries, flows, module rules.  
**tech.md**: Coding standards, patterns, performance rules, testing rules.  
**context.md**: Evolving decisions, key terms, domain rules, file summaries.  
**tasks.md**: Repetitive workflows (test, refactor, review pipelines).

## Startup Behavior
At the beginning of each task:
1. Load all Memory Bank files.  
2. Print `Memory Bank: Active` if Memory Bank files loaded, otherwise `Memory Bank: Missing`.  
3. Summarize project in 3–6 bullets, using Memory Bank only.  
4. Use this reconstructed context for all reasoning.  
5. Do not scan project files unless user approves.

## When the AI Must Use Memory Bank
Always reference Memory Bank before:
- Designing or modifying architecture  
- Naming decisions  
- Implementing features  
- Applying coding standards  
- Testing  
- Refactoring  
- Reviewing for compliance  
- Preventing regressions  

Memory Bank is the authoritative source of project rules.

## When the AI Must NOT Read the Codebase
Do not scan files unless:
- User says `update memory bank`  
- User says `initialize memory bank`  
- AI asks: `I need to inspect files—OK to update Memory Bank?`

This protects performance and context window.

## Updating the Memory Bank
Triggered by explicit user commands or new essential info.

Update rules:
1. Scan project directory (scoped or full).  
2. Extract only long‑term knowledge (not code).  
3. Update context.md first.  
4. Update architecture.md, tech.md, product.md if needed.  
5. Ask before overwriting major sections.  
6. Never store raw code, each file limit 300 lines.

## Files the AI May Modify
Allowed: context.md, architecture.md, tech.md, product.md, tasks.md.  
Forbidden (need user approval): brief.md.

## Thread & Context Window Rules
If conversation drifts, AI should suggest:  
`Update Memory Bank and start a fresh thread?`

A new thread must always reload Memory Bank.

## Consistency Enforcement
AI must enforce consistency with:
- architecture.md  
- tech.md  
- product.md  
- context.md  

If user requests a conflicting change:
- AI asks for confirmation  
- AI issues a consistency warning  

## Goals of the Memory Bank System
- Reduce repetitive context loading  
- Prevent architectural drift  
- Maintain long‑term understanding  
- Improve coding quality  
- Minimize token overhead  
- Enable stable agentic workflows  

# Generic Programming Guidelines
These rules apply across languages, frameworks, and systems.

## Performance Patterns
- Use object pooling to reduce allocations.  
- Preallocate collections to avoid resizing.  
- Optimize data layout for cache locality.  
- Avoid unnecessary type conversions.  
- Use zero‑copy techniques when possible.  
- Minimize heap usage; prefer stack allocation.  
- Ensure escape analysis keeps values local.  

## Concurrency Guidelines
- Limit concurrency with controlled pools.  
- Use atomic operations over locks for simple counters.  
- Apply lazy initialization for expensive work.  
- Share immutable data safely.  
- Use cancellation tokens for timeouts.  

## I/O Optimization
- Use buffered I/O.  
- Batch reads/writes.  
- Use proper pooling for connections.  

## Generic Programming Principles
- Use generics/templates for reusable, type‑safe code.  
- Apply type constraints to enforce correctness.  
- Use generic interfaces for polymorphism.  
- Prefer generic collections.  
- Reuse generic algorithms.  

## SOLID Principles
- SRP: One responsibility per component.  
- OCP: Extend without modifying.  
- LSP: Subtypes must behave as base types.  
- ISP: Favor small, focused interfaces.  
- DIP: Depend on abstractions.  

## Idiomatic Practices
- Keep functions small and focused.  
- Use descriptive names.  
- Avoid global state.  
- Favor composition over inheritance.  
- Keep interfaces minimal.  
- Maintain consistent abstraction levels.  

## Error Handling Best Practices
- Always handle errors explicitly.  
- Don’t use exceptions for normal flow.  
- Use typed errors/enums.  
- Wrap errors with context.  
- Implement robust recovery paths.  

## Testing Strategies
- Write unit tests for critical logic.  
- Use table‑driven tests.  
- Benchmark performance‑critical code.  
- Use mocks/stubs for external dependencies.  
- Use property‑based tests for generic logic.  

## Anti‑Patterns
- Excessive concurrency  
- Large “god” interfaces  
- Ignoring errors  
- Overuse of reflection/dynamic typing  
- Tight coupling  
- Premature optimization  

## Generic Design Patterns
- Factory  
- Strategy  
- Observer  
- Decorator  
- Command  
- Template Method  

## Memory Management
- Understand target platform memory model.  
- Reduce object creation in hot paths.  
- Use value types where appropriate.  
- Dispose/close resources properly.  
- Profile memory usage.  

## Additional Best Practices
- Design for composition and extensibility.  
- Use strong typing intentionally.  
- Balance generic vs. specific implementations.  
- Clearly document interfaces.  
- Maintain consistent coding conventions.  

# Combined Agent Behavior Summary
The AI Coder Agent must:
- Load Memory Bank on every task.  
- Use Memory Bank as primary context.  
- Follow all programming best practices above.  
- Ensure architectural, naming, and coding consistency.  
- Avoid reading source unless asked.  
- Update Memory Bank only when user authorizes.  
- Warn about contradictions.  
- Never store raw code in Memory Bank.