-- Массовое наполнение базы данных детерминированными данными
-- для тестирования производительности и выполнения требований по объему данных.
-- Выполнять после db/schema.sql.

CREATE OR REPLACE FUNCTION pg_temp.make_uuid(p_prefix TEXT, p_seq INTEGER)
RETURNS UUID
LANGUAGE sql
IMMUTABLE
AS $$
    SELECT (
        substr(md5(p_prefix || ':' || p_seq::text), 1, 8) || '-' ||
        substr(md5(p_prefix || ':' || p_seq::text), 9, 4) || '-' ||
        substr(md5(p_prefix || ':' || p_seq::text), 13, 4) || '-' ||
        substr(md5(p_prefix || ':' || p_seq::text), 17, 4) || '-' ||
        substr(md5(p_prefix || ':' || p_seq::text), 21, 12)
    )::uuid;
$$;

INSERT INTO security.roles (id, code, name)
SELECT pg_temp.make_uuid('bulk-role', g),
       format('bulk_role_%s', lpad(g::text, 4, '0')),
       format('Дополнительная роль %s', lpad(g::text, 4, '0'))
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO security.users (id, login, password_hash, full_name, email, phone, status)
SELECT pg_temp.make_uuid('bulk-user', g),
       format('bulk_user_%s', lpad(g::text, 4, '0')),
       repeat('x', 60),
       format('Пользователь %s', lpad(g::text, 4, '0')),
       format('bulk_user_%s@clinic.local', lpad(g::text, 4, '0')),
       '+7900' || lpad(g::text, 7, '0'),
       'active'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

WITH doctor_role AS (
    SELECT id FROM security.roles WHERE code = 'doctor' LIMIT 1
)
INSERT INTO security.user_roles (user_id, role_id)
SELECT pg_temp.make_uuid('bulk-user', g), (SELECT id FROM doctor_role)
FROM generate_series(1, 1000) AS g
ON CONFLICT DO NOTHING;

INSERT INTO security.user_roles (user_id, role_id)
SELECT pg_temp.make_uuid('bulk-user', g), pg_temp.make_uuid('bulk-role', g)
FROM generate_series(1, 1000) AS g
ON CONFLICT DO NOTHING;

INSERT INTO catalog.specialties (id, code, name)
SELECT pg_temp.make_uuid('bulk-specialty', g),
       format('bulk_specialty_%s', lpad(g::text, 4, '0')),
       format('Специальность %s', lpad(g::text, 4, '0'))
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.patients (id, last_name, first_name, middle_name, birth_date, phone, email, insurance_number, status)
SELECT pg_temp.make_uuid('bulk-patient', g),
       format('Пациент%s', lpad(g::text, 4, '0')),
       format('Имя%s', lpad(g::text, 4, '0')),
       format('Отчество%s', lpad(g::text, 4, '0')),
       DATE '1980-01-01' + ((g - 1) % 12000),
       '+7910' || lpad(g::text, 7, '0'),
       format('bulk_patient_%s@clinic.local', lpad(g::text, 4, '0')),
       format('INS-%s', lpad(g::text, 8, '0')),
       'active'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO catalog.services (id, code, name, description, duration_minutes, status)
SELECT pg_temp.make_uuid('bulk-service', g),
       format('bulk_service_%s', lpad(g::text, 4, '0')),
       format('Услуга %s', lpad(g::text, 4, '0')),
       format('Описание услуги %s', lpad(g::text, 4, '0')),
       30,
       'active'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO catalog.service_prices (id, service_id, price, currency_code, effective_from, effective_to)
SELECT pg_temp.make_uuid('bulk-price', g),
       pg_temp.make_uuid('bulk-service', g),
       (1000 + (g % 500))::numeric(12,2),
       'RUB',
       TIMESTAMPTZ '2026-01-01 00:00:00+03',
       NULL
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.doctors (id, user_id, specialty_id, cabinet_number, status)
SELECT pg_temp.make_uuid('bulk-doctor', g),
       pg_temp.make_uuid('bulk-user', g),
       pg_temp.make_uuid('bulk-specialty', g),
       format('C-%s', lpad(((g - 1) % 200 + 1)::text, 3, '0')),
       'active'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.doctor_services (doctor_id, service_id)
SELECT pg_temp.make_uuid('bulk-doctor', g),
       pg_temp.make_uuid('bulk-service', g)
FROM generate_series(1, 1000) AS g
ON CONFLICT DO NOTHING;

INSERT INTO scheduling.doctor_availability (id, doctor_id, starts_at, ends_at, slot_duration_minutes, is_active)
SELECT pg_temp.make_uuid('bulk-availability', g),
       pg_temp.make_uuid('bulk-doctor', g),
       TIMESTAMPTZ '2026-04-15 09:00:00+03',
       TIMESTAMPTZ '2026-04-15 18:00:00+03',
       30,
       TRUE
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO scheduling.doctor_absences (id, doctor_id, starts_at, ends_at, reason)
SELECT pg_temp.make_uuid('bulk-absence', g),
       pg_temp.make_uuid('bulk-doctor', g),
       TIMESTAMPTZ '2026-04-16 16:00:00+03',
       TIMESTAMPTZ '2026-04-16 17:00:00+03',
       format('Плановое отсутствие %s', lpad(g::text, 4, '0'))
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO scheduling.appointments (
    id,
    patient_id,
    doctor_id,
    booked_service_id,
    starts_at,
    ends_at,
    status,
    note,
    created_by_user_id,
    cancelled_by_user_id,
    completed_at,
    created_at,
    updated_at
)
SELECT pg_temp.make_uuid('bulk-appointment', g),
       pg_temp.make_uuid('bulk-patient', g),
       pg_temp.make_uuid('bulk-doctor', g),
       pg_temp.make_uuid('bulk-service', g),
       TIMESTAMPTZ '2026-04-15 09:00:00+03' + ((g - 1) % 16) * INTERVAL '30 minutes',
       TIMESTAMPTZ '2026-04-15 09:30:00+03' + ((g - 1) % 16) * INTERVAL '30 minutes',
       'completed',
       format('BULK_APPOINTMENT_%s', lpad(g::text, 4, '0')),
       pg_temp.make_uuid('bulk-user', 1),
       NULL,
       TIMESTAMPTZ '2026-04-15 10:00:00+03' + ((g - 1) % 16) * INTERVAL '30 minutes',
       TIMESTAMPTZ '2026-04-14 09:00:00+03',
       TIMESTAMPTZ '2026-04-15 10:00:00+03'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.visits (id, appointment_id, doctor_id, patient_id, diagnosis, summary, created_at, updated_at)
SELECT pg_temp.make_uuid('bulk-visit', g),
       pg_temp.make_uuid('bulk-appointment', g),
       pg_temp.make_uuid('bulk-doctor', g),
       pg_temp.make_uuid('bulk-patient', g),
       format('Диагноз %s', lpad(g::text, 4, '0')),
       format('Краткое описание приема %s', lpad(g::text, 4, '0')),
       TIMESTAMPTZ '2026-04-15 10:05:00+03',
       TIMESTAMPTZ '2026-04-15 10:05:00+03'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO medical.visit_services (visit_id, service_id, quantity, unit_price, line_total)
SELECT pg_temp.make_uuid('bulk-visit', g),
       pg_temp.make_uuid('bulk-service', g),
       1,
       (1000 + (g % 500))::numeric(12,2),
       (1000 + (g % 500))::numeric(12,2)
FROM generate_series(1, 1000) AS g
ON CONFLICT DO NOTHING;

INSERT INTO billing.invoices (id, visit_id, patient_id, status, total_amount, paid_amount, issued_at, updated_at)
SELECT pg_temp.make_uuid('bulk-invoice', g),
       pg_temp.make_uuid('bulk-visit', g),
       pg_temp.make_uuid('bulk-patient', g),
       'issued',
       0,
       0,
       TIMESTAMPTZ '2026-04-15 10:10:00+03',
       TIMESTAMPTZ '2026-04-15 10:10:00+03'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO billing.invoice_items (id, invoice_id, service_id, quantity, unit_price, line_total)
SELECT pg_temp.make_uuid('bulk-invoice-item', g),
       pg_temp.make_uuid('bulk-invoice', g),
       pg_temp.make_uuid('bulk-service', g),
       1,
       (1000 + (g % 500))::numeric(12,2),
       (1000 + (g % 500))::numeric(12,2)
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

INSERT INTO billing.payments (id, invoice_id, amount, payment_method, status, external_reference, received_by_user_id, paid_at, created_at)
SELECT pg_temp.make_uuid('bulk-payment', g),
       pg_temp.make_uuid('bulk-invoice', g),
       (1000 + (g % 500))::numeric(12,2),
       CASE g % 3
           WHEN 0 THEN 'cash'::billing.payment_method
           WHEN 1 THEN 'card'::billing.payment_method
           ELSE 'transfer'::billing.payment_method
       END,
       'completed',
       format('BULK_PAYMENT_%s', lpad(g::text, 4, '0')),
       pg_temp.make_uuid('bulk-user', 2),
       TIMESTAMPTZ '2026-04-15 10:20:00+03',
       TIMESTAMPTZ '2026-04-15 10:20:00+03'
FROM generate_series(1, 1000) AS g
ON CONFLICT (id) DO NOTHING;

SELECT 'security.roles' AS table_name, COUNT(*) AS row_count FROM security.roles
UNION ALL
SELECT 'security.users', COUNT(*) FROM security.users
UNION ALL
SELECT 'security.user_roles', COUNT(*) FROM security.user_roles
UNION ALL
SELECT 'catalog.specialties', COUNT(*) FROM catalog.specialties
UNION ALL
SELECT 'medical.patients', COUNT(*) FROM medical.patients
UNION ALL
SELECT 'medical.doctors', COUNT(*) FROM medical.doctors
UNION ALL
SELECT 'catalog.services', COUNT(*) FROM catalog.services
UNION ALL
SELECT 'catalog.service_prices', COUNT(*) FROM catalog.service_prices
UNION ALL
SELECT 'medical.doctor_services', COUNT(*) FROM medical.doctor_services
UNION ALL
SELECT 'scheduling.doctor_availability', COUNT(*) FROM scheduling.doctor_availability
UNION ALL
SELECT 'scheduling.doctor_absences', COUNT(*) FROM scheduling.doctor_absences
UNION ALL
SELECT 'scheduling.appointments', COUNT(*) FROM scheduling.appointments
UNION ALL
SELECT 'medical.visits', COUNT(*) FROM medical.visits
UNION ALL
SELECT 'medical.visit_services', COUNT(*) FROM medical.visit_services
UNION ALL
SELECT 'billing.invoices', COUNT(*) FROM billing.invoices
UNION ALL
SELECT 'billing.invoice_items', COUNT(*) FROM billing.invoice_items
UNION ALL
SELECT 'billing.payments', COUNT(*) FROM billing.payments
ORDER BY table_name;