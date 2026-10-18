INSERT INTO users (id, email, password_hash, role)
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        'dummy-user@room-booking.local',
        'dummy-login-only-no-password',
        'user'
    ),
    (
        '00000000-0000-0000-0000-000000000002',
        'dummy-admin@room-booking.local',
        'dummy-login-only-no-password',
        'admin'
    )
ON CONFLICT DO NOTHING;
