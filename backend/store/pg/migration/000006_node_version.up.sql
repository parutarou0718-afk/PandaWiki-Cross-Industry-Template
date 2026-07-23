-- create node_releases
CREATE TABLE
    "public"."node_releases" (
    id text NOT NULL,
    kb_id text NOT NULL,
    node_id text NOT NULL,
    doc_id text NOT NULL,
    type smallint NULL,
    visibility smallint NULL,
    name text NULL,
    meta JSONB NULL,
    content text NULL,
    parent_id text null,
    position float null,
    created_at timestamptz NULL,
    PRIMARY KEY (id)
);

-- create index on node_releases table
CREATE INDEX "idx_node_releases_kb_id" ON "public"."node_releases" ("kb_id");
CREATE INDEX "idx_node_releases_node_id" ON "public"."node_releases" ("node_id");
CREATE INDEX "idx_node_releases_doc_id" ON "public"."node_releases" ("doc_id");

-- create kb_release
CREATE TABLE
    "public"."kb_releases" (
    id text NOT NULL,
    kb_id text NOT NULL,
    tag text NULL,
    message text NULL,
    created_at timestamptz NULL,
    PRIMARY KEY (id)
);

-- create index on kb_releases table
CREATE INDEX "idx_kb_releases_kb_id" ON "public"."kb_releases" ("kb_id");

-- create kb_release_node_releases
CREATE TABLE
    "public"."kb_release_node_releases" (
    id text NOT NULL,
    kb_id text NOT NULL,
    release_id text NOT NULL,
    node_id text NOT NULL,
    node_release_id text NOT NULL,
    created_at timestamptz NULL,
    PRIMARY KEY (id)
);

-- create index on kb_release_node_releases table
CREATE INDEX "idx_kb_release_node_releases_kb_id" ON "public"."kb_release_node_releases" ("kb_id");
CREATE INDEX "idx_kb_release_node_releases_release_id_node_release_id" ON "public"."kb_release_node_releases" ("release_id", "node_release_id");
CREATE INDEX "idx_kb_release_node_releases_node_id" ON "public"."kb_release_node_releases" ("node_id");

-- update nodes table
ALTER TABLE "public"."nodes" ADD COLUMN "status" smallint NOT NULL DEFAULT 1;
ALTER TABLE "public"."nodes" ADD COLUMN "visibility" smallint NOT NULL DEFAULT 1;

-- update nodes table
UPDATE "public"."nodes" SET "visibility" = 2;


-- create table migrations
CREATE TABLE "public"."migrations" (
    "id" serial PRIMARY KEY,
    "name" varchar(255) NOT NULL,
    "executed_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- create index on migrations table
CREATE UNIQUE INDEX "idx_migrations_name" ON "public"."migrations" ("name");
