\set doctor_id 'PUT_DOCTOR_UUID_HERE'
\set patient_id 'PUT_PATIENT_UUID_HERE'

EXPLAIN ANALYZE
SELECT id, patient_id, doctor_id, booked_service_id, starts_at, ends_at, status
FROM scheduling.appointments
WHERE doctor_id = :'doctor_id'::uuid
  AND starts_at >= TIMESTAMPTZ '2026-04-01 00:00:00+03'
  AND starts_at < TIMESTAMPTZ '2026-04-08 00:00:00+03'
ORDER BY starts_at;

EXPLAIN ANALYZE
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
WHERE a.patient_id = :'patient_id'::uuid
ORDER BY a.starts_at DESC
LIMIT 50;

EXPLAIN ANALYZE
SELECT i.id,
       i.status,
       i.total_amount,
       i.paid_amount,
       i.issued_at
FROM billing.invoices i
WHERE i.patient_id = :'patient_id'::uuid
ORDER BY i.issued_at DESC;

EXPLAIN ANALYZE
SELECT a.doctor_id,
       COUNT(*) AS visit_count,
       SUM(i.total_amount) AS revenue
FROM scheduling.appointments a
JOIN medical.visits v ON v.appointment_id = a.id
JOIN billing.invoices i ON i.visit_id = v.id
WHERE a.starts_at >= TIMESTAMPTZ '2026-04-01 00:00:00+03'
  AND a.starts_at < TIMESTAMPTZ '2026-05-01 00:00:00+03'
  AND a.status = 'completed'
GROUP BY a.doctor_id
ORDER BY revenue DESC;
