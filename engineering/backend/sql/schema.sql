CREATE TABLE IF NOT EXISTS employer (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_type (
    id          SERIAL PRIMARY KEY,
    type        VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS skills (
    id          SERIAL PRIMARY KEY,
    skill       VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ai_version (
    id          SERIAL PRIMARY KEY,
    version     VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS education_level (
    id          SERIAL PRIMARY KEY,
    level       VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS crawler_runs (
    id          BIGSERIAL PRIMARY KEY,
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    status      VARCHAR(50),
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS geo_data (
    id          BIGSERIAL PRIMARY KEY,
    longitude   DECIMAL(9,6),
    latitude    DECIMAL(9,6),
    province    VARCHAR(100),
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS source (
    id          SERIAL PRIMARY KEY,
    -- in the diagram this is renamed to 'name'
    source      VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS experience (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS formality (
    id              SERIAL PRIMARY KEY,
    formality_type  VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS gender (
    id          SERIAL PRIMARY KEY,
    gender_type VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS vocational_education (
    id          SERIAL PRIMARY KEY,
    level       VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS employment_sector (
    id          SERIAL PRIMARY KEY,
    sector      VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

-- 2. title comment is here but 1. is missing at the top of the file.
-- -----------------------------------------------------------------------------
-- 2. Occupation hierarchy (major -> sub-major -> minor -> unit -> occupation)
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS major_group (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sub_major_group (
    id              SERIAL PRIMARY KEY,
    major_group_id  INT NOT NULL REFERENCES major_group(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS minor_group (
    id                  SERIAL PRIMARY KEY,
    sub_major_group_id  INT NOT NULL REFERENCES sub_major_group(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50) NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS unit_group (
    id              SERIAL PRIMARY KEY,
    minor_group_id  INT NOT NULL REFERENCES minor_group(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS occupation_group (
    id              SERIAL PRIMARY KEY,
    unit_group_id   INT NOT NULL REFERENCES unit_group(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ
);

-- -----------------------------------------------------------------------------
-- 3. Industry hierarchy (sector -> division -> group -> class -> subclass)
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS industry_sector (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS industry_division (
    id                  SERIAL PRIMARY KEY,
    industry_sector_id  INT NOT NULL REFERENCES industry_sector(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50) NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS industry_group (
    id                      SERIAL PRIMARY KEY,
    industry_division_id    INT NOT NULL REFERENCES industry_division(id) ON DELETE CASCADE,
    name                    VARCHAR(255) NOT NULL,
    code                    VARCHAR(50) NOT NULL,
    created_at              TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at              TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS industry_class (
    id                  SERIAL PRIMARY KEY,
    industry_group_id   INT NOT NULL REFERENCES industry_group(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50) NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS industry_subclass (
    id                  SERIAL PRIMARY KEY,
    industry_class_id   INT NOT NULL REFERENCES industry_class(id) ON DELETE CASCADE,
    -- there is a type mismatch (in the diagram this is varchar(300))
    name                VARCHAR(500) NOT NULL,
    code                VARCHAR(50) NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMPTZ
);

-- -----------------------------------------------------------------------------
-- 4. Core job post tables
-- -----------------------------------------------------------------------------
-- here there is an issue. pg_type has all the types of schemas and if any schema has 'work_mode_enum' already this will be skipped.
-- so better option here is try to create the enum and catch it in the exception
-- DO $$
-- BEGIN
--     CREATE TYPE work_mode_enum AS ENUM ('remote', 'onsite', 'hybrid');
-- EXCEPTION
--     WHEN duplicate_object THEN NULL;
-- END
-- $$;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'work_mode_enum') THEN
        CREATE TYPE work_mode_enum AS ENUM ('remote', 'onsite', 'hybrid');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS job_post (
    id              BIGSERIAL PRIMARY KEY,
    -- employer_id and job_type_id are also nullable, so these should be employer |o--o{ job_post and job_type |o--o{ job_post in the diagram
    employer_id     BIGINT REFERENCES employer(id),
    job_type_id     INT REFERENCES job_type(id),
    job_role        VARCHAR(255) NOT NULL,
    work_mode       work_mode_enum NOT NULL DEFAULT 'onsite',
    job_description TEXT,
    location        VARCHAR(255),
    no_of_vacancies INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- why this tables doesnt have the created_at updated_at records?
CREATE TABLE IF NOT EXISTS job_post_skills (
    job_post_id BIGINT REFERENCES job_post(id) ON DELETE CASCADE,
    -- create an index with skill_id 
    skill_id    INT    REFERENCES skills(id)   ON DELETE CASCADE,
    PRIMARY KEY (job_post_id, skill_id)
);

-- What is the cardinality between job_post and meta_data? diagram says its one-to-may,
-- but in the meta_data schema says job_post_id is UNIQUE. That makes it one-to-one.
CREATE TABLE IF NOT EXISTS meta_data (
    id                      BIGSERIAL PRIMARY KEY,
    job_post_id             BIGINT UNIQUE REFERENCES job_post(id) ON DELETE CASCADE,

    -- if we use ON DELETE SET NULL for these mentioned attributes in the diagram the relation should be `|o` (zero or one), not `||` (exactly one) - ex: ai_version |o--o{ meta_data (correct this in the diagram)
    -- classification / enrichment lookups (all nullable, ON DELETE SET NULL)
    ai_version_id           INT    REFERENCES ai_version(id)           ON DELETE SET NULL,
    education_level_id      INT    REFERENCES education_level(id)      ON DELETE SET NULL,
    crawler_run_id          BIGINT REFERENCES crawler_runs(id)         ON DELETE SET NULL,
    geo_data_id             BIGINT REFERENCES geo_data(id)             ON DELETE SET NULL,
    source_id               INT    REFERENCES source(id)               ON DELETE SET NULL,
    experience_id           INT    REFERENCES experience(id)           ON DELETE SET NULL,
    occupation_group_id     INT    REFERENCES occupation_group(id)     ON DELETE SET NULL,
    industry_subclass_id    INT    REFERENCES industry_subclass(id)    ON DELETE SET NULL,
    formality_id            INT    REFERENCES formality(id)            ON DELETE SET NULL,
    gender_id               INT    REFERENCES gender(id)               ON DELETE SET NULL,
    vocational_education_id INT    REFERENCES vocational_education(id) ON DELETE SET NULL,
    employment_sector_id    INT    REFERENCES employment_sector(id)    ON DELETE SET NULL,

    posted_at               TIMESTAMPTZ,
    end_date                TIMESTAMPTZ,
    -- here there is a possibility for integer overflow. this holds only 32 bit signed values. 
    minhash_signature       INTEGER[],
    confidence_score        DECIMAL(5,4),

    created_at              TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 5. Deduplication (LSH) index table
-- -----------------------------------------------------------------------------

-- why this tables doesnt have the created_at updated_at records?
CREATE TABLE IF NOT EXISTS lsh_index (
    bucket_key  VARCHAR(64) NOT NULL,
    band_no     INT NOT NULL,
    -- there is type mismatch in job post id in the diagram (correct one is BIGINT)
    job_post_id BIGINT NOT NULL REFERENCES job_post(id) ON DELETE CASCADE,
    PRIMARY KEY (bucket_key, job_post_id)
);

-- -----------------------------------------------------------------------------
-- 6. Indexes
-- -----------------------------------------------------------------------------

-- postgres usually creates index on primary key right? if so idx_lsh_bucket_radar is redunant.
-- can we create an index on job_post_id in the lsh_index table
CREATE INDEX IF NOT EXISTS idx_lsh_bucket_radar
    ON lsh_index (bucket_key, job_post_id);

-- here the end_date is always null. so choosing it as the key index column is useless. can we use INCLUDE job_post_id here?
CREATE INDEX IF NOT EXISTS idx_metadata_snapshot_reconcile
    ON meta_data (crawler_run_id, end_date)
    WHERE end_date IS NULL;

CREATE INDEX CONCURRENTLY idx_metadata_posted_at ON meta_data (posted_at);
CREATE INDEX CONCURRENTLY idx_metadata_occgroup_posted_at ON meta_data (occupation_group_id, posted_at);
CREATE INDEX CONCURRENTLY idx_metadata_indsubclass_posted_at ON meta_data (industry_subclass_id, posted_at);

