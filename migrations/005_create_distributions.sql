CREATE TABLE IF NOT EXISTS distributions (
    id BIGSERIAL PRIMARY KEY,
    donation_id BIGINT NOT NULL REFERENCES donations(id),
    partner_id BIGINT NOT NULL REFERENCES partners(id),
    admin_id BIGINT NOT NULL REFERENCES users(id),
    portions_allocated INT NOT NULL CHECK (portions_allocated > 0),
    status VARCHAR(30) NOT NULL DEFAULT 'allocated'
        CHECK (status IN (
            'allocated',
            'picked_up',
            'delivered',
            'cancelled'
        )),
    allocated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
