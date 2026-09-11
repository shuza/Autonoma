# Autonoma — GitHub Copilot Instructions

## Mission

Autonoma is a production-oriented platform for autonomous AI business operations.

The project is intentionally built in stages:

1. Build one excellent autonomous workflow.
2. Prove reliability, safety, observability, and evaluation.
3. Add a second agent.
4. Demonstrate that the platform supports multiple agents without redesigning the core.

Do not start with a multi-agent system.

## Mandatory context

Before significant implementation work, read:

- `docs/PROJECT_CONTEXT.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `docs/SECURITY.md`
- `docs/EVALUATION.md`
- `docs/DEVELOPMENT.md`

## Current strategy

The first end-to-end workflow is a sales/research workflow:

Lead Created
→ Load Business Context
→ Research Company
→ Research Decision Makers
→ Qualify Lead
→ Generate Sales Strategy
→ Generate Outreach Proposal
→ Deterministic Policy Check
→ Autonomous Action OR Human Approval
→ Execute Through Tool Gateway
→ Update CRM
→ Audit
→ Complete

The first workflow must be demonstrably reliable before adding another agent.

## Architecture principles

AI is not the system. AI is one component inside a controlled execution platform.

Use:

- deterministic policy for authorization and business rules
- Tool Gateway for external actions
- persistent workflow state
- explicit state transitions
- structured LLM outputs
- validation of all LLM outputs
- idempotency for side effects
- retries only where safe
- audit logging
- metrics and tracing
- human approval for risky actions
- automated evaluation

Never allow the LLM to directly:

- execute arbitrary shell commands
- execute arbitrary SQL
- access arbitrary URLs
- bypass authorization
- decide whether a business policy permits an action

## Technology direction

Use Go for the platform:

- API
- persistence
- workflow infrastructure
- policy engine
- tool gateway
- audit
- observability

Use Python for the AI runtime when it provides a clear benefit:

- LangGraph
- LangChain components
- LLM integrations
- agent implementations
- AI-specific evaluation

Do not make the business domain depend directly on framework-specific LangChain/LangGraph types.

## LangGraph rule

Use LangGraph for stateful AI orchestration where it adds value.

Do not use LangGraph merely because it is available.

Do not put authorization, deterministic business rules, secrets, or irreversible side effects inside an LLM graph node without going through Autonoma's platform boundaries.

## Implementation rules

Before coding:

1. Inspect existing code.
2. Identify the current phase in `docs/IMPLEMENTATION_PLAN.md`.
3. Check existing architectural decisions.
4. Implement the smallest useful vertical slice.
5. Do not implement future phases without explicit instruction.

After coding:

1. Add or update tests.
2. Test failure paths.
3. Run formatting and linting.
4. Run relevant unit/integration tests.
5. Check observability.
6. Check security implications.
7. Update `docs/IMPLEMENTATION_PLAN.md`.
8. Record meaningful architectural decisions in `docs/DECISIONS.md`.

## Quality bar

A feature is not complete just because the happy path works.

Consider:

- malformed input
- malformed LLM output
- LLM timeout
- external API timeout
- API 5xx
- duplicate events
- duplicate side effects
- process restart
- workflow resume
- unauthorized actions
- prompt injection
- sensitive data leakage

Prefer simple, explicit, testable engineering over unnecessary framework complexity.