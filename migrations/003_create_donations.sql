CREATE TABLE IF NOT EXISTS donations (
    id BIGSERIAL PRIMARY KEY,
    donor_id BIGINT NOT NULL REFERENCES users(id),
    category_id BIGINT NOT NULL REFERENCES food_categories(id),
    title VARCHAR(150) NOT NULL,
    description TEXT,
    portions INT NOT NULL CHECK (portions > 0),
    expired_at TIMESTAMP NOT NULL,
    pickup_address TEXT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'available'
        CHECK (status IN (
            'available',
            'allocated',
            'picked_up',
            'delivered',
            'expired',
            'cancelled'
        )),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
