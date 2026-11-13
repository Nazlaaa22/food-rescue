CREATE TABLE IF NOT EXISTS pickups (
    id BIGSERIAL PRIMARY KEY,
    distribution_id BIGINT NOT NULL REFERENCES distributions(id),
    volunteer_id BIGINT NOT NULL REFERENCES users(id),
    status VARCHAR(30) NOT NULL DEFAULT 'assigned'
        CHECK (status IN (
            'assigned',
            'picked_up',
            'delivered',
            'cancelled'
        )),
    picked_at TIMESTAMP,
    delivered_at TIMESTAMP,
    notes TEXT
);
