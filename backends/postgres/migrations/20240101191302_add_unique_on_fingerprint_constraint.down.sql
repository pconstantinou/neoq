DROP CONSTRAINT neoq_jobs_fingerprint_constraint_idx;
CREATE INDEX IF NOT EXISTS neoq_jobs_fingerprint_idx ON neoq_jobs (fingerprint, status);
CREATE UNIQUE INDEX IF NOT EXISTS neoq_jobs_fingerprint_unique_idx ON neoq_jobs (queue, fingerprint, status) WHERE NOT (status = 'processed');