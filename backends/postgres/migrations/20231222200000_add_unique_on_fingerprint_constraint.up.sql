ALTER TABLE neoq_jobs ADD CONSTRAINT neoq_jobs_fingerprint_constraint_idx UNIQUE (fingerprint, status, ran_at);
