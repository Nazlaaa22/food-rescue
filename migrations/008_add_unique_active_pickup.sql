-- Satu distribusi hanya boleh punya satu pickup yang masih aktif.
-- Pickup yang dibatalkan relawan tidak dihitung, jadi tugasnya bisa diambil ulang
-- oleh relawan lain (sesuai alur di pickup_repository.go).
CREATE UNIQUE INDEX IF NOT EXISTS uq_pickups_active_distribution
    ON pickups (distribution_id)
    WHERE status <> 'cancelled';
