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

CREATE TABLE workflows
(
    id         UUID PRIMARY KEY,
    lead_id    UUID NOT NULL REFERENCES leads (id) ON DELETE CASCADE,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);