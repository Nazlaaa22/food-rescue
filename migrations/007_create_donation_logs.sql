CREATE TABLE IF NOT EXISTS donation_logs (
    id BIGSERIAL PRIMARY KEY,
    donation_id BIGINT NOT NULL REFERENCES donations(id),
    status VARCHAR(30) NOT NULL,
    changed_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
