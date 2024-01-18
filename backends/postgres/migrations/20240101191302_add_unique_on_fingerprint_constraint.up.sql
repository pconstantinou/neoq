DROP INDEX IF EXISTS neoq_jobs_fingerprint_idx;
DROP INDEX IF EXISTS neoq_jobs_fingerprint_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS neoq_jobs_fingerprint_unique_idx ON neoq_jobs (queue, status, fingerprint, ran_at);
