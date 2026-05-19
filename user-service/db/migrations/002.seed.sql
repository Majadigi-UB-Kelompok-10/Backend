-- Default superadmin account (password: Admin@12345)
-- Hash generated via: bcrypt.GenerateFromPassword([]byte("Admin@12345"), 12)
INSERT INTO users (
    first_name, last_name, phone, email, nik, address,
    password_hash, role, email_verified
) VALUES (
    'Super', 'Admin',
    '+6281234567890',
    'superadmin@majadigi.id',
    '1234567890123456',
    'Surabaya, Jawa Timur',
    '$2a$12$VPDgCGQvvoBdvX0krzrlU.W09r2qPiuUbORXASlWNtGFr/GJ9hcya',
    'superadmin',
    true
) ON CONFLICT DO NOTHING;
