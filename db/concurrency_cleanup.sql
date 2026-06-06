-- Очистка данных после сценария конкурентной записи.

DELETE FROM billing.payments
 WHERE invoice_id IN (
     SELECT i.id
       FROM billing.invoices i
       JOIN medical.visits v ON v.id = i.visit_id
       JOIN scheduling.appointments a ON a.id = v.appointment_id
      WHERE a.note IN ('SCENARIO_CONCURRENT_SESSION_1', 'SCENARIO_CONCURRENT_SESSION_2')
 );

DELETE FROM billing.invoice_items
 WHERE invoice_id IN (
     SELECT i.id
       FROM billing.invoices i
       JOIN medical.visits v ON v.id = i.visit_id
       JOIN scheduling.appointments a ON a.id = v.appointment_id
      WHERE a.note IN ('SCENARIO_CONCURRENT_SESSION_1', 'SCENARIO_CONCURRENT_SESSION_2')
 );

DELETE FROM billing.invoices
 WHERE visit_id IN (
     SELECT v.id
       FROM medical.visits v
       JOIN scheduling.appointments a ON a.id = v.appointment_id
      WHERE a.note IN ('SCENARIO_CONCURRENT_SESSION_1', 'SCENARIO_CONCURRENT_SESSION_2')
 );

DELETE FROM medical.visit_services
 WHERE visit_id IN (
     SELECT v.id
       FROM medical.visits v
       JOIN scheduling.appointments a ON a.id = v.appointment_id
      WHERE a.note IN ('SCENARIO_CONCURRENT_SESSION_1', 'SCENARIO_CONCURRENT_SESSION_2')
 );

DELETE FROM medical.visits
 WHERE appointment_id IN (
     SELECT id
       FROM scheduling.appointments
      WHERE note IN ('SCENARIO_CONCURRENT_SESSION_1', 'SCENARIO_CONCURRENT_SESSION_2')
 );

DELETE FROM scheduling.appointments
 WHERE note IN ('SCENARIO_CONCURRENT_SESSION_1', 'SCENARIO_CONCURRENT_SESSION_2');
