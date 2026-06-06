-- ����������������� �������� ������ ��� ��������� ���������� � ������������ ��������� �������.

INSERT INTO security.users (id, login, password_hash, full_name, email, phone, status)
VALUES
    ('00000000-0000-0000-0000-000000000101', 'admin', '$2a$12$demo_admin_hash', 'Clinic Administrator', 'admin@clinic.local', '+79990000001', 'active'),
    ('00000000-0000-0000-0000-000000000102', 'patient_demo', '$2a$12$demo_patient_hash', 'Anna Petrova', 'patient@clinic.local', '+79990000002', 'active'),
    ('00000000-0000-0000-0000-000000000103', 'doctor_ivanov', '$2a$12$demo_doctor_hash', 'Dr. Ivan Ivanov', 'doctor@clinic.local', '+79990000003', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO security.user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000101'::uuid, id FROM security.roles WHERE code = 'administrator'
ON CONFLICT DO NOTHING;

INSERT INTO security.user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000102'::uuid, id FROM security.roles WHERE code = 'patient'
ON CONFLICT DO NOTHING;

INSERT INTO security.user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000103'::uuid, id FROM security.roles WHERE code = 'doctor'
ON CONFLICT DO NOTHING;


INSERT INTO catalog.specialties (id, code, name)
VALUES
    ('00000000-0000-0000-0000-000000000201', 'cardiology', 'Cardiology')
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.doctors (id, user_id, specialty_id, cabinet_number, status)
VALUES
    ('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000103', '00000000-0000-0000-0000-000000000201', '101', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO catalog.services (id, code, name, description, duration_minutes, status)
VALUES
    ('00000000-0000-0000-0000-000000000401', 'CONSULT_CARDIO', 'Cardiologist consultation', 'Primary cardiology consultation', 30, 'active'),
    ('00000000-0000-0000-0000-000000000402', 'ECG_STANDARD', 'Standard ECG', 'Electrocardiogram', 20, 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO catalog.service_prices (id, service_id, price, currency_code, effective_from, effective_to)
VALUES
    ('00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000401', 1500.00, 'RUB', TIMESTAMPTZ '2026-01-01 00:00:00+03', NULL),
    ('00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000402', 900.00, 'RUB', TIMESTAMPTZ '2026-01-01 00:00:00+03', NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.doctor_services (doctor_id, service_id)
VALUES
    ('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000401'),
    ('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000402')
ON CONFLICT DO NOTHING;

INSERT INTO medical.patients (id, last_name, first_name, middle_name, birth_date, phone, email, insurance_number, status)
VALUES
    ('00000000-0000-0000-0000-000000000601', 'Petrova', 'Anna', 'Sergeevna', DATE '1991-04-12', '+79991112233', 'anna.petrova@clinic.local', 'INS-0001', 'active'),
    ('00000000-0000-0000-0000-000000000602', 'Sidorov', 'Pavel', 'Olegovich', DATE '1987-11-03', '+79994445566', 'pavel.sidorov@clinic.local', 'INS-0002', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO scheduling.doctor_availability (id, doctor_id, starts_at, ends_at, slot_duration_minutes, is_active)
VALUES
    ('00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000301', TIMESTAMPTZ '2030-01-15 09:00:00+03', TIMESTAMPTZ '2030-01-15 18:00:00+03', 30, TRUE),
    ('00000000-0000-0000-0000-000000000702', '00000000-0000-0000-0000-000000000301', TIMESTAMPTZ '2030-01-16 09:00:00+03', TIMESTAMPTZ '2030-01-16 18:00:00+03', 30, TRUE)
ON CONFLICT (id) DO NOTHING;
