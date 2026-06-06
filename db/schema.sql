CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE SCHEMA IF NOT EXISTS security;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS medical;
CREATE SCHEMA IF NOT EXISTS scheduling;
CREATE SCHEMA IF NOT EXISTS billing;

CREATE DOMAIN security.email_address AS VARCHAR(150)
    CHECK (VALUE IS NULL OR VALUE ~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$');

CREATE DOMAIN security.phone_number AS VARCHAR(20)
    CHECK (VALUE IS NULL OR VALUE ~ '^\+?[1-9][0-9]{10,14}$');

CREATE TYPE security.user_status AS ENUM ('active', 'inactive', 'archived');
CREATE TYPE medical.patient_status AS ENUM ('active', 'inactive', 'archived');
CREATE TYPE medical.doctor_status AS ENUM ('active', 'inactive', 'archived');
CREATE TYPE catalog.service_status AS ENUM ('active', 'inactive', 'archived');
CREATE TYPE scheduling.appointment_status AS ENUM ('booked', 'cancelled', 'completed');
CREATE TYPE billing.invoice_status AS ENUM ('issued', 'partially_paid', 'paid', 'cancelled');
CREATE TYPE billing.payment_status AS ENUM ('completed', 'cancelled');
CREATE TYPE billing.payment_method AS ENUM ('cash', 'card', 'transfer');

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clinic_owner') THEN
        CREATE ROLE clinic_owner NOLOGIN;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clinic_api') THEN
        CREATE ROLE clinic_api NOLOGIN;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clinic_reporting') THEN
        CREATE ROLE clinic_reporting NOLOGIN;
    END IF;
END
$$;

CREATE OR REPLACE FUNCTION security.set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS security.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS security.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    email security.email_address,
    phone security.phone_number,
    status security.user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS security.user_roles (
    user_id UUID NOT NULL REFERENCES security.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES security.roles(id) ON DELETE RESTRICT,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS catalog.specialties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medical.patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    birth_date DATE NOT NULL CHECK (birth_date <= CURRENT_DATE),
    phone security.phone_number NOT NULL,
    email security.email_address,
    insurance_number VARCHAR(50),
    status medical.patient_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medical.doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES security.users(id) ON DELETE RESTRICT,
    specialty_id UUID NOT NULL REFERENCES catalog.specialties(id) ON DELETE RESTRICT,
    cabinet_number VARCHAR(20),
    status medical.doctor_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS catalog.services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 480),
    status catalog.service_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS catalog.service_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES catalog.services(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    currency_code CHAR(3) NOT NULL DEFAULT 'RUB',
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period TSTZRANGE GENERATED ALWAYS AS (
        tstzrange(effective_from, COALESCE(effective_to, 'infinity'::timestamptz), '[)')
    ) STORED,
    CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE TABLE IF NOT EXISTS medical.doctor_services (
    doctor_id UUID NOT NULL REFERENCES medical.doctors(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES catalog.services(id) ON DELETE CASCADE,
    PRIMARY KEY (doctor_id, service_id)
);

CREATE TABLE IF NOT EXISTS scheduling.doctor_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES medical.doctors(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    slot_duration_minutes INTEGER NOT NULL CHECK (slot_duration_minutes > 0 AND slot_duration_minutes <= 480),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period TSTZRANGE GENERATED ALWAYS AS (tstzrange(starts_at, ends_at, '[)')) STORED,
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS scheduling.doctor_absences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES medical.doctors(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period TSTZRANGE GENERATED ALWAYS AS (tstzrange(starts_at, ends_at, '[)')) STORED,
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS scheduling.appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES medical.patients(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES medical.doctors(id) ON DELETE RESTRICT,
    booked_service_id UUID NOT NULL REFERENCES catalog.services(id) ON DELETE RESTRICT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status scheduling.appointment_status NOT NULL DEFAULT 'booked',
    note TEXT,
    cancellation_reason TEXT,
    created_by_user_id UUID NOT NULL REFERENCES security.users(id) ON DELETE RESTRICT,
    cancelled_by_user_id UUID REFERENCES security.users(id) ON DELETE RESTRICT,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period TSTZRANGE GENERATED ALWAYS AS (tstzrange(starts_at, ends_at, '[)')) STORED,
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS medical.visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL UNIQUE REFERENCES scheduling.appointments(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES medical.doctors(id) ON DELETE RESTRICT,
    patient_id UUID NOT NULL REFERENCES medical.patients(id) ON DELETE RESTRICT,
    diagnosis TEXT,
    summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medical.visit_services (
    visit_id UUID NOT NULL REFERENCES medical.visits(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES catalog.services(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    line_total NUMERIC(12,2) NOT NULL CHECK (line_total >= 0),
    PRIMARY KEY (visit_id, service_id)
);

CREATE TABLE IF NOT EXISTS billing.invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL UNIQUE REFERENCES medical.visits(id) ON DELETE RESTRICT,
    patient_id UUID NOT NULL REFERENCES medical.patients(id) ON DELETE RESTRICT,
    status billing.invoice_status NOT NULL DEFAULT 'issued',
    total_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    paid_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing.invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES billing.invoices(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES catalog.services(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    line_total NUMERIC(12,2) NOT NULL CHECK (line_total >= 0)
);

CREATE TABLE IF NOT EXISTS billing.payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES billing.invoices(id) ON DELETE RESTRICT,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    payment_method billing.payment_method NOT NULL,
    status billing.payment_status NOT NULL DEFAULT 'completed',
    external_reference VARCHAR(100),
    received_by_user_id UUID NOT NULL REFERENCES security.users(id) ON DELETE RESTRICT,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION billing.refresh_invoice_state(p_invoice_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    v_total NUMERIC(12,2);
    v_paid NUMERIC(12,2);
BEGIN
    SELECT COALESCE(SUM(line_total), 0)
      INTO v_total
      FROM billing.invoice_items
     WHERE invoice_id = p_invoice_id;

    SELECT COALESCE(SUM(amount), 0)
      INTO v_paid
      FROM billing.payments
     WHERE invoice_id = p_invoice_id
       AND status = 'completed';

    UPDATE billing.invoices
       SET total_amount = v_total,
           paid_amount = v_paid,
           status = CASE
                        WHEN v_paid = 0 THEN 'issued'::billing.invoice_status
                        WHEN v_paid < v_total THEN 'partially_paid'::billing.invoice_status
                        ELSE 'paid'::billing.invoice_status
                    END,
           updated_at = NOW()
     WHERE id = p_invoice_id;
END;
$$;

CREATE OR REPLACE FUNCTION billing.trg_refresh_invoice_from_items()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM billing.refresh_invoice_state(COALESCE(NEW.invoice_id, OLD.invoice_id));
    RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE OR REPLACE FUNCTION billing.trg_refresh_invoice_from_payments()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM billing.refresh_invoice_state(COALESCE(NEW.invoice_id, OLD.invoice_id));
    RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE OR REPLACE FUNCTION catalog.get_service_price(p_service_id UUID, p_at TIMESTAMPTZ)
RETURNS NUMERIC(12,2)
LANGUAGE plpgsql
AS $$
DECLARE
    v_price NUMERIC(12,2);
BEGIN
    SELECT sp.price
      INTO v_price
      FROM catalog.service_prices sp
     WHERE sp.service_id = p_service_id
       AND sp.effective_from <= p_at
       AND (sp.effective_to IS NULL OR sp.effective_to > p_at)
     ORDER BY sp.effective_from DESC
     LIMIT 1;

    IF v_price IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = format('Не найдена активная цена для услуги %s на момент %s', p_service_id, p_at);
    END IF;

    RETURN v_price;
END;
$$;

CREATE OR REPLACE FUNCTION security.has_role(p_user_id UUID, p_role_code VARCHAR)
RETURNS BOOLEAN
LANGUAGE sql
AS $$
    SELECT EXISTS (
        SELECT 1
          FROM security.user_roles ur
          JOIN security.roles r ON r.id = ur.role_id
         WHERE ur.user_id = p_user_id
           AND r.code = p_role_code
    );
$$;

CREATE OR REPLACE FUNCTION scheduling.book_appointment(
    p_patient_id UUID,
    p_doctor_id UUID,
    p_service_id UUID,
    p_starts_at TIMESTAMPTZ,
    p_created_by_user_id UUID,
    p_note TEXT DEFAULT NULL
)
RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    v_duration INTEGER;
    v_ends_at TIMESTAMPTZ;
    v_appointment_id UUID;
BEGIN
    IF p_starts_at <= NOW() THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Время начала приема должно быть в будущем';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM medical.patients WHERE id = p_patient_id AND status = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = format('Пациент %s неактивен', p_patient_id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM medical.doctors WHERE id = p_doctor_id AND status = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = format('Врач %s неактивен', p_doctor_id);
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM medical.doctor_services
         WHERE doctor_id = p_doctor_id
           AND service_id = p_service_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = format('Услуга %s не закреплена за врачом %s', p_service_id, p_doctor_id);
    END IF;

    SELECT duration_minutes
      INTO v_duration
      FROM catalog.services
     WHERE id = p_service_id
       AND status = 'active';

    IF v_duration IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = format('Услуга %s неактивна', p_service_id);
    END IF;

    v_ends_at := p_starts_at + make_interval(mins => v_duration);

    IF NOT EXISTS (
        SELECT 1
          FROM scheduling.doctor_availability da
         WHERE da.doctor_id = p_doctor_id
           AND da.is_active
           AND da.starts_at <= p_starts_at
           AND da.ends_at >= v_ends_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Выбранное время находится вне интервала доступности врача';
    END IF;

    IF EXISTS (
        SELECT 1
          FROM scheduling.doctor_absences abs
         WHERE abs.doctor_id = p_doctor_id
           AND abs.period && tstzrange(p_starts_at, v_ends_at, '[)')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Врач отсутствует в выбранный интервал времени';
    END IF;

    INSERT INTO scheduling.appointments (
        patient_id,
        doctor_id,
        booked_service_id,
        starts_at,
        ends_at,
        status,
        note,
        created_by_user_id
    )
    VALUES (
        p_patient_id,
        p_doctor_id,
        p_service_id,
        p_starts_at,
        v_ends_at,
        'booked',
        p_note,
        p_created_by_user_id
    )
    RETURNING id INTO v_appointment_id;

    RETURN v_appointment_id;
END;
$$;

CREATE OR REPLACE FUNCTION scheduling.cancel_appointment(
    p_appointment_id UUID,
    p_actor_user_id UUID,
    p_reason TEXT
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    v_starts_at TIMESTAMPTZ;
    v_status scheduling.appointment_status;
    v_can_override BOOLEAN;
BEGIN
    SELECT starts_at, status
      INTO v_starts_at, v_status
      FROM scheduling.appointments
     WHERE id = p_appointment_id
     FOR UPDATE;

    IF v_status IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002', MESSAGE = format('Запись %s не найдена', p_appointment_id);
    END IF;

    IF v_status <> 'booked' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Отменить можно только запись со статусом booked';
    END IF;

    v_can_override := security.has_role(p_actor_user_id, 'administrator');

    IF NOT v_can_override AND v_starts_at < NOW() + INTERVAL '24 hours' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Пациент может отменить запись не позднее чем за 24 часа до приема';
    END IF;

    UPDATE scheduling.appointments
       SET status = 'cancelled',
           cancellation_reason = p_reason,
           cancelled_by_user_id = p_actor_user_id,
           updated_at = NOW()
     WHERE id = p_appointment_id;
END;
$$;

CREATE OR REPLACE FUNCTION medical.complete_visit(
    p_appointment_id UUID,
    p_doctor_user_id UUID,
    p_summary TEXT,
    p_diagnosis TEXT,
    p_performed_services JSONB DEFAULT '[]'::JSONB
)
RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    v_visit_id UUID;
    v_invoice_id UUID;
    v_doctor_id UUID;
    v_patient_id UUID;
    v_booked_service_id UUID;
    v_started_at TIMESTAMPTZ;
    v_status scheduling.appointment_status;
    v_item JSONB;
    v_service_id UUID;
    v_quantity INTEGER;
    v_price NUMERIC(12,2);
BEGIN
    SELECT a.doctor_id, a.patient_id, a.booked_service_id, a.starts_at, a.status
      INTO v_doctor_id, v_patient_id, v_booked_service_id, v_started_at, v_status
      FROM scheduling.appointments a
      JOIN medical.doctors d ON d.id = a.doctor_id
     WHERE a.id = p_appointment_id
       AND d.user_id = p_doctor_user_id
     FOR UPDATE;

    IF v_status IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002', MESSAGE = format('Запись %s не найдена для врача-пользователя %s', p_appointment_id, p_doctor_user_id);
    END IF;

    IF v_status <> 'booked' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Завершить можно только запись со статусом booked';
    END IF;

    INSERT INTO medical.visits (
        appointment_id,
        doctor_id,
        patient_id,
        diagnosis,
        summary
    )
    VALUES (
        p_appointment_id,
        v_doctor_id,
        v_patient_id,
        p_diagnosis,
        p_summary
    )
    RETURNING id INTO v_visit_id;

    IF jsonb_array_length(COALESCE(p_performed_services, '[]'::jsonb)) = 0 THEN
        v_price := catalog.get_service_price(v_booked_service_id, v_started_at);

        INSERT INTO medical.visit_services (visit_id, service_id, quantity, unit_price, line_total)
        VALUES (v_visit_id, v_booked_service_id, 1, v_price, v_price);
    ELSE
        FOR v_item IN SELECT * FROM jsonb_array_elements(p_performed_services)
        LOOP
            v_service_id := (v_item ->> 'service_id')::UUID;
            v_quantity := COALESCE((v_item ->> 'quantity')::INTEGER, 1);

            IF NOT EXISTS (
                SELECT 1
                  FROM medical.doctor_services
                 WHERE doctor_id = v_doctor_id
                   AND service_id = v_service_id
            ) THEN
                RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = format('Услуга %s не закреплена за врачом %s', v_service_id, v_doctor_id);
            END IF;

            v_price := catalog.get_service_price(v_service_id, v_started_at);

            INSERT INTO medical.visit_services (visit_id, service_id, quantity, unit_price, line_total)
            VALUES (v_visit_id, v_service_id, v_quantity, v_price, v_price * v_quantity);
        END LOOP;
    END IF;

    INSERT INTO billing.invoices (visit_id, patient_id)
    VALUES (v_visit_id, v_patient_id)
    RETURNING id INTO v_invoice_id;

    INSERT INTO billing.invoice_items (invoice_id, service_id, quantity, unit_price, line_total)
    SELECT v_invoice_id, service_id, quantity, unit_price, line_total
      FROM medical.visit_services
     WHERE visit_id = v_visit_id;

    UPDATE scheduling.appointments
       SET status = 'completed',
           completed_at = NOW(),
           updated_at = NOW()
     WHERE id = p_appointment_id;

    PERFORM billing.refresh_invoice_state(v_invoice_id);

    RETURN v_visit_id;
END;
$$;

CREATE OR REPLACE FUNCTION billing.register_payment(
    p_invoice_id UUID,
    p_amount NUMERIC(12,2),
    p_payment_method billing.payment_method,
    p_received_by_user_id UUID,
    p_external_reference VARCHAR(100) DEFAULT NULL
)
RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    v_payment_id UUID;
BEGIN
    INSERT INTO billing.payments (
        invoice_id,
        amount,
        payment_method,
        status,
        external_reference,
        received_by_user_id
    )
    VALUES (
        p_invoice_id,
        p_amount,
        p_payment_method,
        'completed',
        p_external_reference,
        p_received_by_user_id
    )
    RETURNING id INTO v_payment_id;

    PERFORM billing.refresh_invoice_state(p_invoice_id);

    RETURN v_payment_id;
END;
$$;

CREATE OR REPLACE VIEW catalog.current_price_list AS
SELECT
    s.id AS service_id,
    s.code,
    s.name,
    s.duration_minutes,
    sp.price,
    sp.currency_code,
    sp.effective_from
FROM catalog.services s
JOIN LATERAL (
    SELECT price, currency_code, effective_from
      FROM catalog.service_prices sp
     WHERE sp.service_id = s.id
       AND sp.effective_from <= NOW()
       AND (sp.effective_to IS NULL OR sp.effective_to > NOW())
     ORDER BY sp.effective_from DESC
     LIMIT 1
) sp ON TRUE
WHERE s.status = 'active';

ALTER TABLE catalog.service_prices
    ADD CONSTRAINT service_prices_no_overlap
    EXCLUDE USING gist (service_id WITH =, period WITH &&);

ALTER TABLE scheduling.doctor_availability
    ADD CONSTRAINT doctor_availability_no_overlap
    EXCLUDE USING gist (doctor_id WITH =, period WITH &&)
    WHERE (is_active);

ALTER TABLE scheduling.appointments
    ADD CONSTRAINT appointments_doctor_no_overlap
    EXCLUDE USING gist (doctor_id WITH =, period WITH &&)
    WHERE (status IN ('booked', 'completed'));

CREATE INDEX IF NOT EXISTS idx_users_status ON security.users(status);
CREATE INDEX IF NOT EXISTS idx_patients_phone ON medical.patients(phone);
CREATE INDEX IF NOT EXISTS idx_patients_status ON medical.patients(status);
CREATE INDEX IF NOT EXISTS idx_doctors_specialty ON medical.doctors(specialty_id);
CREATE INDEX IF NOT EXISTS idx_doctors_status ON medical.doctors(status);
CREATE INDEX IF NOT EXISTS idx_services_status ON catalog.services(status);
CREATE INDEX IF NOT EXISTS idx_service_prices_lookup ON catalog.service_prices(service_id, effective_from DESC);
CREATE INDEX IF NOT EXISTS idx_schedule_availability_lookup ON scheduling.doctor_availability(doctor_id, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_schedule_absence_lookup ON scheduling.doctor_absences(doctor_id, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_appointments_doctor_time ON scheduling.appointments(doctor_id, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_appointments_patient_history ON scheduling.appointments(patient_id, starts_at DESC);
CREATE INDEX IF NOT EXISTS idx_visits_patient_created ON medical.visits(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_invoices_patient_status ON billing.invoices(patient_id, status, issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_invoice_paid_at ON billing.payments(invoice_id, paid_at DESC);

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON security.users
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_patients_updated_at
BEFORE UPDATE ON medical.patients
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_doctors_updated_at
BEFORE UPDATE ON medical.doctors
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_services_updated_at
BEFORE UPDATE ON catalog.services
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_doctor_availability_updated_at
BEFORE UPDATE ON scheduling.doctor_availability
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_doctor_absences_updated_at
BEFORE UPDATE ON scheduling.doctor_absences
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_appointments_updated_at
BEFORE UPDATE ON scheduling.appointments
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_visits_updated_at
BEFORE UPDATE ON medical.visits
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_invoices_updated_at
BEFORE UPDATE ON billing.invoices
FOR EACH ROW
EXECUTE FUNCTION security.set_updated_at();

CREATE TRIGGER trg_invoice_items_refresh_invoice
AFTER INSERT OR UPDATE OR DELETE ON billing.invoice_items
FOR EACH ROW
EXECUTE FUNCTION billing.trg_refresh_invoice_from_items();

CREATE TRIGGER trg_payments_refresh_invoice
AFTER INSERT OR UPDATE OR DELETE ON billing.payments
FOR EACH ROW
EXECUTE FUNCTION billing.trg_refresh_invoice_from_payments();

INSERT INTO security.roles (code, name)
VALUES
    ('administrator', 'Administrator'),
    ('patient', 'Patient'),
    ('doctor', 'Doctor')
ON CONFLICT (code) DO NOTHING;

GRANT USAGE ON SCHEMA security, catalog, medical, scheduling, billing TO clinic_api, clinic_reporting;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA security, catalog, medical, scheduling, billing TO clinic_api;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA security, catalog, medical, scheduling, billing TO clinic_api;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA security, catalog, medical, scheduling, billing TO clinic_api;
GRANT SELECT ON ALL TABLES IN SCHEMA security, catalog, medical, scheduling, billing TO clinic_reporting;

