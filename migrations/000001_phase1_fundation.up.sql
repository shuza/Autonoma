CREATE TABLE companies
(
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    website    TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE leads
(
    id         UUID PRIMARY KEY,
    company_id UUID NOT NULL REFERENCES companies (id) ON DELETE RESTRICT,
    source     TEXT NOT NULL,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE contacts
(
    id         UUID PRIMARY KEY,
    company_id UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    email      TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE workflows
(
    id         UUID PRIMARY KEY,
    lead_id    UUID NOT NULL REFERENCES leads (id) ON DELETE CASCADE,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE workflow_steps
(
    id          UUID PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE tool_executions
(
    id               UUID PRIMARY KEY,
    workflow_step_id UUID NOT NULL REFERENCES workflow_steps (id) ON DELETE CASCADE,
    tool_name        TEXT NOT NULL,
    status           TEXT NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE approvals
(
    id                UUID PRIMARY KEY,
    tool_execution_id UUID NOT NULL REFERENCES tool_executions (id) ON DELETE CASCADE,
    status            TEXT NOT NULL,
    requested_by      TEXT NOT NULL,
    created_at        TIMESTAMPTZ DEFAULT now(),
    updated_at        TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE audit_evets
(
    id          UUID PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    event_type  TEXT NOT NULL,
    actor       TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);