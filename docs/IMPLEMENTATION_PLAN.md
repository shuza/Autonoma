# Autonoma — Implementation Plan

## Guiding rule

Build one excellent autonomous workflow first.

Do not build a multi-agent system before proving the first workflow.

---

## Phase 0 — Foundation

- [x] Repository initialized
- [x] Go module initialized
- [x] Python AI runtime skeleton
- [x] Makefile
- [x] Docker Compose
- [x] PostgreSQL
- [x] Redis
- [x] API health endpoint
- [x] CI
- [x] Formatting
- [x] Linting
- [x] Tests
- [x] `.env.example`

Definition of done:

- application starts locally
- dependencies start locally
- tests pass
- CI passes

---

## Phase 1 — Domain and persistence

Implement:

- [x] Company
- [x] Contact
- [x] Lead
- [x] Workflow
- [x] WorkflowStep
- [x] ToolExecution
- [x] Approval
- [x] AuditEvent

Add:

- [x] migrations
- [x] repositories
- [x] integration tests

Definition of done:

All workflow state and business context survives a process restart.

---

## Phase 2 — Workflow engine

Implement:

- [x] workflow state machine
- [x] step execution
- [ ] persistent state
- [x] resume
- [ ] cancellation
- [ ] failure handling
- [ ] idempotency
- [ ] retry policy

States:

- PENDING
- RUNNING
- WAITING_APPROVAL
- COMPLETED
- FAILED
- CANCELLED

Definition of done:

A workflow can stop, the process can restart, and the workflow resumes correctly.

---

## Phase 3 — AI runtime

Create the Python AI runtime.

Implement:

- [ ] LLM provider abstraction
- [ ] structured output
- [ ] validation
- [ ] token/cost tracking
- [ ] mock provider
- [ ] provider integration
- [ ] LangGraph where stateful orchestration is useful

Do not tightly couple domain models to LangChain/LangGraph types.

---

## Phase 4 — Research capability

Implement company research.

- [ ] research input schema
- [ ] research output schema
- [ ] web research tool
- [ ] source validation
- [ ] confidence information
- [ ] research agent
- [ ] evaluation cases
- [ ] tests

Definition of done:

Given a company, Autonoma produces structured research with evidence.

---

## Phase 5 — Tool Gateway

Implement a secure tool boundary.

Initial tools:

- [ ] web search
- [ ] webpage reader
- [ ] company lookup
- [ ] mock CRM
- [ ] mock email

Every tool defines:

- name
- description
- input schema
- output schema
- risk level
- authorization requirements
- timeout
- retry policy

Definition of done:

AI cannot directly execute external side effects.

All actions go through the Tool Gateway.

---

## Phase 6 — First autonomous workflow

Implement:

Lead Created
→ Load Context
→ Research
→ Qualification
→ Strategy
→ Outreach Proposal
→ Policy
→ Action/Approval
→ CRM
→ Audit
→ Complete

- [ ] lead API
- [ ] trigger
- [ ] research
- [ ] qualification
- [ ] strategy
- [ ] outreach
- [ ] policy
- [ ] approval
- [ ] mock email
- [ ] mock CRM
- [ ] audit
- [ ] end-to-end test

Definition of done:

A complete business workflow runs from lead creation to completion.

---

## Phase 7 — Deterministic policy engine

Implement:

- [ ] action model
- [ ] risk levels
- [ ] allow
- [ ] deny
- [ ] require approval
- [ ] configurable rules
- [ ] policy tests

The LLM must never be the final authority for authorization.

---

## Phase 8 — Human approval

Implement:

- [ ] approval persistence
- [ ] approval API
- [ ] pause workflow
- [ ] approve
- [ ] reject
- [ ] resume
- [ ] audit approval decision

Definition of done:

A workflow can pause and resume without losing state.

---

## Phase 9 — Reliability

Test:

- [ ] LLM timeout
- [ ] malformed LLM output
- [ ] external API timeout
- [ ] external API 5xx
- [ ] duplicate event
- [ ] duplicate tool request
- [ ] process restart
- [ ] workflow resume
- [ ] retry exhaustion
- [ ] non-idempotent operation

Implement:

- [ ] timeouts
- [ ] retries
- [ ] backoff
- [ ] idempotency
- [ ] failure classification
- [ ] recovery

---

## Phase 10 — Observability

Implement:

- [ ] structured logs
- [ ] metrics
- [ ] distributed tracing
- [ ] workflow correlation ID
- [ ] agent spans
- [ ] tool spans
- [ ] LLM spans
- [ ] token metrics
- [ ] cost metrics

Dashboard:

- active workflows
- completed workflows
- failures
- approval requests
- tool calls
- LLM latency
- LLM cost

---

## Phase 11 — Evaluation

Create a fixed evaluation dataset.

Measure:

- qualification accuracy
- research accuracy
- tool selection
- policy compliance
- action correctness
- hallucination rate
- workflow completion
- cost
- latency

Run evaluations before and after significant prompt/model changes.

---

## Phase 12 — Production-style deployment

Only after local functionality and evaluation are solid.

Implement:

- [ ] Docker
- [ ] Helm
- [ ] Terraform
- [ ] cloud deployment
- [ ] secrets
- [ ] IAM
- [ ] CI/CD
- [ ] monitoring

Do not introduce Kubernetes just to make the portfolio look complex.

---

## Phase 13 — Second agent

Only start after the first workflow meets the reliability definition.

Second agent:

- Research/Knowledge Agent OR
- Analytics Agent

The second agent must reuse:

- Tool Gateway
- Policy Engine
- Memory
- Workflow infrastructure
- Audit
- Observability
- Evaluation framework

Definition of done:

Two different agents run on the same Autonoma platform without duplicating platform infrastructure.

---

## Phase 14 — Portfolio demonstration

Create a reproducible demo:

1. Create lead.
2. Observe workflow start.
3. Inspect research.
4. Inspect AI output.
5. Inspect tool calls.
6. Inspect policy decision.
7. Approve if required.
8. Observe execution.
9. Inspect audit trail.
10. Show metrics/traces.
11. Run evaluation.
12. Demonstrate failure recovery.
13. Run the second agent using the same platform.

---

## Current sprint

Current phase:

**Phase 2 — Workflow engine**

Current task:

**Implement the next smallest useful Phase 2 vertical slice.**

Do not implement future phases until the current phase is complete.

## Definition of done

A feature is not complete until appropriate tests, error handling, observability, and documentation are present.