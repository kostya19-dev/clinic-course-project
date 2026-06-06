\set ON_ERROR_STOP on
\pset format unaligned
\pset fieldsep '|'
\pset tuples_only on

SET client_min_messages = warning;
SET jit = off;
SET max_parallel_workers_per_gather = 0;

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

CREATE OR REPLACE FUNCTION pg_temp.explain_ms(p_sql TEXT)
RETURNS NUMERIC
LANGUAGE plpgsql
AS $$
DECLARE
    v_plan TEXT;
BEGIN
    EXECUTE 'EXPLAIN (ANALYZE, TIMING OFF, FORMAT JSON) ' || p_sql INTO v_plan;
    RETURN ((v_plan::json -> 0) ->> 'Execution Time')::NUMERIC;
END;
$$;

CREATE OR REPLACE FUNCTION pg_temp.ensure_research_data(p_rows INTEGER)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    v_admin_role_id UUID := pg_temp.make_uuid('research-role', 1);
    v_patient_role_id UUID := pg_temp.make_uuid('research-role', 2);
    v_doctor_role_id UUID := pg_temp.make_uuid('research-role', 3);
    v_admin_user_id UUID := pg_temp.make_uuid('research-user', 1);
    v_doctor_user_id UUID := pg_temp.make_uuid('research-user', 2);
    v_specialty_id UUID := pg_temp.make_uuid('research-specialty', 1);
    v_doctor_id UUID := pg_temp.make_uuid('research-doctor', 1);
    v_service_id UUID := pg_temp.make_uuid('research-service', 1);
BEGIN
    INSERT INTO security.roles (id, code, name)
    VALUES
        (v_admin_role_id, 'research_administrator', 'Research administrator'),
        (v_patient_role_id, 'research_patient', 'Research patient'),
        (v_doctor_role_id, 'research_doctor', 'Research doctor')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO security.users (id, login, password_hash, full_name, email, phone, status)
    VALUES
        (v_admin_user_id, 'research_admin', repeat('x', 60), 'Research Admin', 'research_admin@clinic.local', '+79000000001', 'active'),
        (v_doctor_user_id, 'research_doctor', repeat('x', 60), 'Research Doctor', 'research_doctor@clinic.local', '+79000000002', 'active')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO security.user_roles (user_id, role_id)
    VALUES
        (v_admin_user_id, v_admin_role_id),
        (v_doctor_user_id, v_doctor_role_id)
    ON CONFLICT DO NOTHING;

    INSERT INTO catalog.specialties (id, code, name)
    VALUES (v_specialty_id, 'research_specialty', 'Research Specialty')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO medical.patients (id, last_name, first_name, birth_date, phone, email, status)
    SELECT pg_temp.make_uuid('research-patient', g),
           format('Patient%s', g),
           format('Name%s', g),
           DATE '1980-01-01' + ((g - 1) % 12000),
           '+7920' || lpad(g::text, 7, '0'),
           format('research_patient_%s@clinic.local', g),
           'active'
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO catalog.services (id, code, name, description, duration_minutes, status)
    VALUES (v_service_id, 'research_service', 'Research Service', 'Research service for benchmark', 30, 'active')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO catalog.service_prices (id, service_id, price, currency_code, effective_from, effective_to)
    VALUES (pg_temp.make_uuid('research-price', 1), v_service_id, 1500.00, 'RUB', TIMESTAMPTZ '2026-01-01 00:00:00+03', NULL)
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO medical.doctors (id, user_id, specialty_id, cabinet_number, status)
    VALUES (v_doctor_id, v_doctor_user_id, v_specialty_id, 'R-101', 'active')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO medical.doctor_services (doctor_id, service_id)
    VALUES (v_doctor_id, v_service_id)
    ON CONFLICT DO NOTHING;

    INSERT INTO scheduling.doctor_availability (id, doctor_id, starts_at, ends_at, slot_duration_minutes, is_active)
    VALUES (
        pg_temp.make_uuid('research-availability', 1),
        v_doctor_id,
        TIMESTAMPTZ '2027-01-01 08:00:00+03',
        TIMESTAMPTZ '2032-01-01 00:00:00+03',
        30,
        TRUE
    )
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
        completed_at,
        created_at,
        updated_at
    )
    SELECT pg_temp.make_uuid('research-appointment', g),
           pg_temp.make_uuid('research-patient', ((g - 1) % LEAST(p_rows, 100)) + 1),
           v_doctor_id,
           v_service_id,
           TIMESTAMPTZ '2027-01-01 08:00:00+03' + ((g - 1) * INTERVAL '30 minutes'),
           TIMESTAMPTZ '2027-01-01 08:30:00+03' + ((g - 1) * INTERVAL '30 minutes'),
           'completed',
           format('research appointment %s', g),
           v_admin_user_id,
           TIMESTAMPTZ '2027-01-01 08:30:00+03' + ((g - 1) * INTERVAL '30 minutes'),
           TIMESTAMPTZ '2026-12-31 08:00:00+03',
           TIMESTAMPTZ '2027-01-01 08:30:00+03' + ((g - 1) * INTERVAL '30 minutes')
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO medical.visits (id, appointment_id, doctor_id, patient_id, diagnosis, summary, created_at, updated_at)
    SELECT pg_temp.make_uuid('research-visit', g),
           pg_temp.make_uuid('research-appointment', g),
           v_doctor_id,
           pg_temp.make_uuid('research-patient', ((g - 1) % LEAST(p_rows, 100)) + 1),
           'Research diagnosis',
           'Research summary',
           TIMESTAMPTZ '2027-01-01 08:35:00+03' + ((g - 1) * INTERVAL '30 minutes'),
           TIMESTAMPTZ '2027-01-01 08:35:00+03' + ((g - 1) * INTERVAL '30 minutes')
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO medical.visit_services (visit_id, service_id, quantity, unit_price, line_total)
    SELECT pg_temp.make_uuid('research-visit', g),
           v_service_id,
           1,
           1500.00,
           1500.00
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT DO NOTHING;

    INSERT INTO billing.invoices (id, visit_id, patient_id, status, total_amount, paid_amount, issued_at, updated_at)
    SELECT pg_temp.make_uuid('research-invoice', g),
           pg_temp.make_uuid('research-visit', g),
           pg_temp.make_uuid('research-patient', ((g - 1) % LEAST(p_rows, 100)) + 1),
           'issued',
           0,
           0,
           TIMESTAMPTZ '2027-01-01 08:40:00+03' + ((g - 1) * INTERVAL '30 minutes'),
           TIMESTAMPTZ '2027-01-01 08:40:00+03' + ((g - 1) * INTERVAL '30 minutes')
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO billing.invoice_items (id, invoice_id, service_id, quantity, unit_price, line_total)
    SELECT pg_temp.make_uuid('research-invoice-item', g),
           pg_temp.make_uuid('research-invoice', g),
           v_service_id,
           1,
           1500.00,
           1500.00
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO billing.payments (id, invoice_id, amount, payment_method, status, external_reference, received_by_user_id, paid_at, created_at)
    SELECT pg_temp.make_uuid('research-payment', g),
           pg_temp.make_uuid('research-invoice', g),
           1500.00,
           'card',
           'completed',
           format('research-payment-%s', g),
           v_admin_user_id,
           TIMESTAMPTZ '2027-01-01 08:45:00+03' + ((g - 1) * INTERVAL '30 minutes'),
           TIMESTAMPTZ '2027-01-01 08:45:00+03' + ((g - 1) * INTERVAL '30 minutes')
    FROM generate_series(1, p_rows) AS g
    ON CONFLICT (id) DO NOTHING;
END;
$$;

CREATE OR REPLACE FUNCTION pg_temp.cleanup_research_data()
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    v_doctor_id UUID := pg_temp.make_uuid('research-doctor', 1);
    v_service_id UUID := pg_temp.make_uuid('research-service', 1);
BEGIN
    DELETE FROM billing.payments
     WHERE external_reference LIKE 'research-payment-%'
        OR invoice_id IN (
            SELECT id
              FROM billing.invoices
             WHERE patient_id IN (
                SELECT id
                  FROM medical.patients
                 WHERE email LIKE 'research_patient_%@clinic.local'
             )
        );

    DELETE FROM billing.invoice_items
     WHERE invoice_id IN (
        SELECT id
          FROM billing.invoices
         WHERE patient_id IN (
            SELECT id
              FROM medical.patients
             WHERE email LIKE 'research_patient_%@clinic.local'
         )
     );

    DELETE FROM billing.invoices
     WHERE patient_id IN (
        SELECT id
          FROM medical.patients
         WHERE email LIKE 'research_patient_%@clinic.local'
     );

    DELETE FROM medical.visit_services
     WHERE visit_id IN (
        SELECT id
          FROM medical.visits
         WHERE diagnosis = 'Research diagnosis'
     );

    DELETE FROM medical.visits
     WHERE diagnosis = 'Research diagnosis';

    DELETE FROM scheduling.appointments
     WHERE doctor_id = v_doctor_id
        OR note LIKE 'research appointment %'
        OR note LIKE 'research measured insert%';

    DELETE FROM scheduling.doctor_availability
     WHERE doctor_id = v_doctor_id;

    DELETE FROM medical.doctor_services
     WHERE doctor_id = v_doctor_id
        OR service_id = v_service_id;

    DELETE FROM catalog.service_prices
     WHERE service_id = v_service_id;

    DELETE FROM catalog.services
     WHERE id = v_service_id
        OR code = 'research_service';

    DELETE FROM medical.doctors
     WHERE id = v_doctor_id;

    DELETE FROM medical.patients
     WHERE email LIKE 'research_patient_%@clinic.local';

    DELETE FROM security.user_roles
     WHERE user_id IN (
        SELECT id
          FROM security.users
         WHERE login IN ('research_admin', 'research_doctor')
     )
        OR role_id IN (
        SELECT id
          FROM security.roles
         WHERE code IN ('research_administrator', 'research_patient', 'research_doctor')
     );

    DELETE FROM security.users
     WHERE login IN ('research_admin', 'research_doctor');

    DELETE FROM catalog.specialties
     WHERE code = 'research_specialty';

    DELETE FROM security.roles
     WHERE code IN ('research_administrator', 'research_patient', 'research_doctor');
END;
$$;

DO $$
BEGIN
    PERFORM pg_temp.cleanup_research_data();
END;
$$;

CREATE TEMP TABLE research_results (
    scale_rows INTEGER NOT NULL,
    operation TEXT NOT NULL,
    run_no INTEGER NOT NULL,
    execution_ms NUMERIC NOT NULL
);

DO $$
DECLARE
    v_scale INTEGER;
    v_run INTEGER;
    v_doctor_id UUID := pg_temp.make_uuid('research-doctor', 1);
    v_service_id UUID := pg_temp.make_uuid('research-service', 1);
    v_admin_user_id UUID := pg_temp.make_uuid('research-user', 1);
    v_patient_id UUID := pg_temp.make_uuid('research-patient', 1);
    v_start_at TIMESTAMPTZ;
    v_sql TEXT;
BEGIN
    FOREACH v_scale IN ARRAY ARRAY[5000, 20000, 50000]
    LOOP
        DELETE FROM scheduling.appointments
         WHERE doctor_id = v_doctor_id
           AND note LIKE 'research measured insert%';

        PERFORM pg_temp.ensure_research_data(v_scale);
        ANALYZE;

        v_sql := format($sql$
            SELECT id, patient_id, doctor_id, booked_service_id, starts_at, ends_at, status
              FROM scheduling.appointments
             WHERE doctor_id = %L::uuid
               AND starts_at >= TIMESTAMPTZ '2027-01-01 08:00:00+03'
               AND starts_at < TIMESTAMPTZ '2032-01-01 00:00:00+03'
             ORDER BY starts_at
        $sql$, v_doctor_id);
        PERFORM pg_temp.explain_ms(v_sql);

        FOR v_run IN 1..7 LOOP
            INSERT INTO research_results
            VALUES (v_scale, 'schedule_indexed', v_run, pg_temp.explain_ms(v_sql));
        END LOOP;

        PERFORM set_config('enable_indexscan', 'off', TRUE);
        PERFORM set_config('enable_bitmapscan', 'off', TRUE);
        FOR v_run IN 1..7 LOOP
            INSERT INTO research_results
            VALUES (v_scale, 'schedule_without_index_scan', v_run, pg_temp.explain_ms(v_sql));
        END LOOP;
        PERFORM set_config('enable_indexscan', 'on', TRUE);
        PERFORM set_config('enable_bitmapscan', 'on', TRUE);

        v_sql := format($sql$
            SELECT a.id,
                   a.starts_at,
                   a.ends_at,
                   a.status,
                   s.name AS service_name,
                   u.full_name AS doctor_name
              FROM scheduling.appointments a
              JOIN catalog.services s ON s.id = a.booked_service_id
              JOIN medical.doctors d ON d.id = a.doctor_id
              JOIN security.users u ON u.id = d.user_id
             WHERE a.patient_id = %L::uuid
             ORDER BY a.starts_at DESC
             LIMIT 50
        $sql$, v_patient_id);
        PERFORM pg_temp.explain_ms(v_sql);

        FOR v_run IN 1..7 LOOP
            INSERT INTO research_results
            VALUES (v_scale, 'patient_history', v_run, pg_temp.explain_ms(v_sql));
        END LOOP;

        v_sql := format($sql$
            SELECT i.id,
                   i.status,
                   i.total_amount,
                   i.paid_amount,
                   i.issued_at
              FROM billing.invoices i
             WHERE i.patient_id = %L::uuid
             ORDER BY i.issued_at DESC
             LIMIT 50
        $sql$, v_patient_id);
        PERFORM pg_temp.explain_ms(v_sql);

        FOR v_run IN 1..7 LOOP
            INSERT INTO research_results
            VALUES (v_scale, 'patient_invoices', v_run, pg_temp.explain_ms(v_sql));
        END LOOP;

        v_sql := $sql$
            SELECT a.doctor_id,
                   COUNT(*) AS visit_count,
                   SUM(i.total_amount) AS revenue
              FROM scheduling.appointments a
              JOIN medical.visits v ON v.appointment_id = a.id
              JOIN billing.invoices i ON i.visit_id = v.id
             WHERE a.starts_at >= TIMESTAMPTZ '2027-01-01 08:00:00+03'
               AND a.starts_at < TIMESTAMPTZ '2032-01-01 00:00:00+03'
               AND a.status = 'completed'
             GROUP BY a.doctor_id
        $sql$;
        PERFORM pg_temp.explain_ms(v_sql);

        FOR v_run IN 1..7 LOOP
            INSERT INTO research_results
            VALUES (v_scale, 'revenue_report', v_run, pg_temp.explain_ms(v_sql));
        END LOOP;

        v_start_at := TIMESTAMPTZ '2027-01-01 08:00:00+03' + (60000 * INTERVAL '30 minutes');
        v_sql := format(
            $sql$
                SELECT scheduling.book_appointment(
                    %L::uuid,
                    %L::uuid,
                    %L::uuid,
                    %L::timestamptz,
                    %L::uuid,
                    'research measured insert warmup'
                )
            $sql$,
            v_patient_id,
            v_doctor_id,
            v_service_id,
            v_start_at,
            v_admin_user_id
        );
        PERFORM pg_temp.explain_ms(v_sql);

        FOR v_run IN 1..7 LOOP
            v_start_at := TIMESTAMPTZ '2027-01-01 08:00:00+03' + ((60100 + v_run) * INTERVAL '30 minutes');
            v_sql := format(
                $sql$
                    SELECT scheduling.book_appointment(
                        %L::uuid,
                        %L::uuid,
                        %L::uuid,
                        %L::timestamptz,
                        %L::uuid,
                        'research measured insert scale %s run %s'
                    )
                $sql$,
                v_patient_id,
                v_doctor_id,
                v_service_id,
                v_start_at,
                v_admin_user_id,
                v_scale,
                v_run
            );

            INSERT INTO research_results
            VALUES (v_scale, 'book_appointment', v_run, pg_temp.explain_ms(v_sql));
        END LOOP;
    END LOOP;
END;
$$;

SELECT 'RESULT',
       scale_rows,
       operation,
       ROUND(AVG(execution_ms), 3) AS avg_ms,
       ROUND(MIN(execution_ms), 3) AS min_ms,
       ROUND(MAX(execution_ms), 3) AS max_ms
  FROM research_results
 GROUP BY scale_rows, operation
 ORDER BY scale_rows, operation;

SELECT 'COUNT',
       'medical.patients',
       COUNT(*)
  FROM medical.patients
UNION ALL
SELECT 'COUNT', 'scheduling.appointments', COUNT(*) FROM scheduling.appointments
UNION ALL
SELECT 'COUNT', 'medical.visits', COUNT(*) FROM medical.visits
UNION ALL
SELECT 'COUNT', 'billing.invoices', COUNT(*) FROM billing.invoices
UNION ALL
SELECT 'COUNT', 'billing.invoice_items', COUNT(*) FROM billing.invoice_items
UNION ALL
SELECT 'COUNT', 'billing.payments', COUNT(*) FROM billing.payments
ORDER BY 2;
