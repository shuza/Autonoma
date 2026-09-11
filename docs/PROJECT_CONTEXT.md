# Autonoma — Project Context

## What is Autonoma?

Autonoma is an autonomous AI operations platform designed to move real business work from manual execution to reliable AI-driven workflows.

It is not a chatbot.

The platform combines:

- business context
- real data
- AI reasoning
- external tools and APIs
- deterministic business rules
- multi-step workflows
- controlled autonomy
- human approval
- reliability
- observability
- evaluation

## Portfolio objective

The project demonstrates production-level AI engineering capability.

The portfolio story is:

> I designed and implemented an autonomous business workflow, added deterministic controls around AI decisions, made execution resumable and observable, evaluated its behavior, and then extended the platform with a second agent without redesigning the core.

## Build strategy

Do not build a multi-agent system first.

Build one excellent autonomous workflow.

Then prove:

- correctness
- reliability
- security
- observability
- cost awareness
- evaluation
- failure recovery
- human escalation

Only then add a second agent.

## First workflow

The first workflow is an autonomous sales/research workflow.

Lead Created
→ Load Context
→ Research Company
→ Research Decision Makers
→ Qualify Lead
→ Generate Sales Strategy
→ Generate Outreach Proposal
→ Policy Evaluation
→ Allow / Require Approval / Deny
→ Execute Action
→ Update CRM
→ Audit
→ Complete

## Important distinction

The AI proposes reasoning and actions.

The platform controls execution.

Example:

AI proposes:

> Send a personalized introductory email to this prospect.

Policy Engine:

> ALLOW / REQUIRE_APPROVAL / DENY

Tool Gateway:

> Executes the approved email operation.

Audit:

> Records what happened, why, who/what approved it, and the result.

## Autonomy model

### Level 0 — Recommendation

AI suggests an action.

No execution.

### Level 1 — Human approval

AI proposes an action.

A human approves it.

### Level 2 — Policy-controlled autonomy

AI proposes an action.

Deterministic policy allows it.

The platform executes it.

### Level 3 — Autonomous workflow

Multiple low-risk actions execute automatically.

High-risk actions still require approval.

The first workflow should demonstrate Levels 1–3.

## Platform vs AI

### Platform

Responsible for:

- API
- persistence
- workflow state
- policy
- authorization
- tool execution
- audit
- observability
- reliability
- idempotency

### AI runtime

Responsible for:

- reasoning
- planning
- structured generation
- research
- classification
- recommendations
- agent-specific behavior

The AI runtime must not become the security boundary.

## Technology direction

### Go

Use Go for the platform.

### Python

Use Python for AI-specific runtime components.

Potential libraries:

- LangGraph
- LangChain components
- provider SDKs
- evaluation tooling

Frameworks are implementation details, not the architecture.

## Second agent

After the first workflow is reliable, add a second agent.

Recommended options:

- Research/Knowledge Agent
- Analytics Agent

The second agent must reuse:

- tools
- policy
- workflow infrastructure
- audit
- observability
- evaluation

## Non-goals for first MVP

Do not initially build:

- dozens of agents
- complex multi-agent conversations
- autonomous financial transactions
- arbitrary code execution
- unrestricted browser automation
- Kubernetes before local functionality works
- unnecessary event infrastructure
- huge RAG infrastructure
- complex frontend before the backend workflow works

## Definition of success

A reviewer should be able to see:

1. A business event starts a workflow.
2. AI performs useful reasoning.
3. AI uses tools.
4. Deterministic policies control execution.
5. The system acts autonomously or requests approval.
6. Actions are executed safely.
7. Important steps are auditable.
8. Failures are recoverable.
9. Metrics and traces show what happened.
10. Automated evaluations demonstrate quality.