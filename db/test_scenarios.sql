-- �������������� SQL-�������� ��� �������� ���������� ��������� �������.
-- ��������� ����� `db/schema.sql` � `db/seed.sql`.

DO $$
DECLARE
    v_patient_1 UUID := '00000000-0000-0000-0000-000000000601';
    v_patient_2 UUID := '00000000-0000-0000-0000-000000000602';
    v_doctor UUID := '00000000-0000-0000-0000-000000000301';
    v_doctor_user UUID := '00000000-0000-0000-0000-000000000103';
    v_patient_user UUID := '00000000-0000-0000-0000-000000000102';
    v_admin_user UUID := '00000000-0000-0000-0000-000000000101';
    v_service_consult UUID := '00000000-0000-0000-0000-000000000401';
    v_service_ecg UUID := '00000000-0000-0000-0000-000000000402';

    v_appointment_cancel UUID;
    v_appointment_complete UUID;
    v_appointment_extra UUID;
    v_visit_id UUID;
    v_invoice_id UUID;
    v_payment_id UUID;
    v_total NUMERIC(12,2);
    v_paid NUMERIC(12,2);
    v_status billing.invoice_status;
    v_count INTEGER;
    v_rejected BOOLEAN;
BEGIN
    DELETE FROM billing.payments
     WHERE invoice_id IN (
         SELECT i.id
           FROM billing.invoices i
           JOIN medical.visits v ON v.id = i.visit_id
           JOIN scheduling.appointments a ON a.id = v.appointment_id
          WHERE a.note LIKE 'SCENARIO_%'
     );

    DELETE FROM billing.invoice_items
     WHERE invoice_id IN (
         SELECT i.id
           FROM billing.invoices i
           JOIN medical.visits v ON v.id = i.visit_id
           JOIN scheduling.appointments a ON a.id = v.appointment_id
          WHERE a.note LIKE 'SCENARIO_%'
     );

    DELETE FROM billing.invoices
     WHERE visit_id IN (
         SELECT v.id
           FROM medical.visits v
           JOIN scheduling.appointments a ON a.id = v.appointment_id
          WHERE a.note LIKE 'SCENARIO_%'
     );

    DELETE FROM medical.visit_services
     WHERE visit_id IN (
         SELECT v.id
           FROM medical.visits v
           JOIN scheduling.appointments a ON a.id = v.appointment_id
          WHERE a.note LIKE 'SCENARIO_%'
     );

    DELETE FROM medical.visits
     WHERE appointment_id IN (
         SELECT id
           FROM scheduling.appointments
          WHERE note LIKE 'SCENARIO_%'
     );

    DELETE FROM scheduling.appointments
     WHERE note LIKE 'SCENARIO_%';

    BEGIN
        INSERT INTO medical.patients (last_name, first_name, birth_date, phone, email, status)
        VALUES ('Invalid', 'Email', DATE '1990-01-01', '+79991110000', 'not-an-email', 'active');
        v_rejected := FALSE;
    EXCEPTION WHEN OTHERS THEN
        v_rejected := TRUE;
    END;

    IF NOT v_rejected THEN
        RAISE EXCEPTION 'Invalid email was not rejected';
    END IF;

    BEGIN
        INSERT INTO medical.patients (last_name, first_name, birth_date, phone, email, status)
        VALUES ('Invalid', 'Phone', DATE '1990-01-01', '12345', 'valid@clinic.local', 'active');
        v_rejected := FALSE;
    EXCEPTION WHEN OTHERS THEN
        v_rejected := TRUE;
    END;

    IF NOT v_rejected THEN
        RAISE EXCEPTION 'Invalid phone was not rejected';
    END IF;

    v_appointment_cancel := scheduling.book_appointment(
        v_patient_1,
        v_doctor,
        v_service_consult,
        TIMESTAMPTZ '2030-01-15 10:00:00+03',
        v_patient_user,
        'SCENARIO_BOOK_CANCEL'
    );

    IF v_appointment_cancel IS NULL THEN
        RAISE EXCEPTION 'Expected successful booking for cancellation scenario';
    END IF;

    BEGIN
        PERFORM scheduling.book_appointment(
            v_patient_2,
            v_doctor,
            v_service_consult,
            TIMESTAMPTZ '2030-01-15 10:15:00+03',
            v_patient_user,
            'SCENARIO_OVERLAP_REJECT'
        );
        v_rejected := FALSE;
    EXCEPTION WHEN OTHERS THEN
        v_rejected := TRUE;
    END;

    IF NOT v_rejected THEN
        RAISE EXCEPTION 'Overlapping booking was not rejected';
    END IF;

    BEGIN
        PERFORM scheduling.book_appointment(
            v_patient_2,
            v_doctor,
            v_service_consult,
            TIMESTAMPTZ '2030-01-17 08:00:00+03',
            v_patient_user,
            'SCENARIO_OUTSIDE_AVAILABILITY'
        );
        v_rejected := FALSE;
    EXCEPTION WHEN OTHERS THEN
        v_rejected := TRUE;
    END;

    IF NOT v_rejected THEN
        RAISE EXCEPTION 'Booking outside availability was not rejected';
    END IF;

    PERFORM scheduling.cancel_appointment(v_appointment_cancel, v_patient_user, 'SCENARIO_CANCEL_OK');

    IF (SELECT status FROM scheduling.appointments WHERE id = v_appointment_cancel) <> 'cancelled' THEN
        RAISE EXCEPTION 'Appointment cancellation did not change status to cancelled';
    END IF;

    v_appointment_complete := scheduling.book_appointment(
        v_patient_1,
        v_doctor,
        v_service_consult,
        TIMESTAMPTZ '2030-01-15 11:00:00+03',
        v_patient_user,
        'SCENARIO_COMPLETE_DEFAULT'
    );

    v_visit_id := medical.complete_visit(
        v_appointment_complete,
        v_doctor_user,
        'SCENARIO_DEFAULT_SUMMARY',
        'SCENARIO_DEFAULT_DIAGNOSIS',
        '[]'::jsonb
    );

    SELECT id, total_amount, paid_amount, status
      INTO v_invoice_id, v_total, v_paid, v_status
      FROM billing.invoices
     WHERE visit_id = v_visit_id;

    IF v_invoice_id IS NULL OR v_total <= 0 OR v_status <> 'issued' THEN
        RAISE EXCEPTION 'Default completion did not create expected invoice';
    END IF;

    v_payment_id := billing.register_payment(v_invoice_id, 500.00, 'cash', v_admin_user, 'SCENARIO_PARTIAL_PAYMENT');

    IF v_payment_id IS NULL THEN
        RAISE EXCEPTION 'Partial payment did not return payment id';
    END IF;

    SELECT total_amount, paid_amount, status
      INTO v_total, v_paid, v_status
      FROM billing.invoices
     WHERE id = v_invoice_id;

    IF v_status <> 'partially_paid' OR v_paid <> 500.00 THEN
        RAISE EXCEPTION 'Partial payment did not update invoice status correctly';
    END IF;

    PERFORM billing.register_payment(v_invoice_id, v_total - v_paid, 'card', v_admin_user, 'SCENARIO_FINAL_PAYMENT');

    SELECT paid_amount, status
      INTO v_paid, v_status
      FROM billing.invoices
     WHERE id = v_invoice_id;

    IF v_status <> 'paid' OR v_paid < v_total THEN
        RAISE EXCEPTION 'Final payment did not close the invoice';
    END IF;

    v_appointment_extra := scheduling.book_appointment(
        v_patient_2,
        v_doctor,
        v_service_consult,
        TIMESTAMPTZ '2030-01-15 12:00:00+03',
        v_patient_user,
        'SCENARIO_COMPLETE_EXTRA'
    );

    v_visit_id := medical.complete_visit(
        v_appointment_extra,
        v_doctor_user,
        'SCENARIO_EXTRA_SUMMARY',
        'SCENARIO_EXTRA_DIAGNOSIS',
        jsonb_build_array(
            jsonb_build_object('service_id', v_service_consult, 'quantity', 1),
            jsonb_build_object('service_id', v_service_ecg, 'quantity', 1)
        )
    );

    SELECT COUNT(*)
      INTO v_count
      FROM medical.visit_services
     WHERE visit_id = v_visit_id;

    IF v_count <> 2 THEN
        RAISE EXCEPTION 'Expected two performed services in additional service scenario';
    END IF;

    RAISE NOTICE 'All SQL coursework scenarios completed successfully';
END;
$$;
