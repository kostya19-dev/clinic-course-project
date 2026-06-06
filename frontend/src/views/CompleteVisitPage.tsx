import { useEffect, useMemo, useState } from "react";
import { EmptyState, InfoCard, PageTitle, Panel, StatusBadge } from "../components/ui";
import { demoClinicContext } from "../demoData";
import { formatMoney, recordNumber } from "../utils/format";
import { useCompleteVisitViewModel } from "../viewmodels/useCompleteVisitViewModel";

type Props = {
  appointmentId: string;
  doctorUserId: string;
  onVisitCompleted: (invoiceId: string) => void;
};

export function CompleteVisitPage({ appointmentId, doctorUserId, onVisitCompleted }: Props) {
  const [form, setForm] = useState({
    appointmentId,
    doctorUserId,
    complaints: "Периодические боли в области сердца после физической нагрузки.",
    diagnosis: "Артериальная гипертензия, требуется наблюдение кардиолога",
    recommendations: "Контроль давления утром и вечером, повторный прием через 14 дней.",
    extraServiceId: "",
    quantity: 1,
  });
  const { isSubmitting, error, result, submit } = useCompleteVisitViewModel();

  useEffect(() => {
    setForm((current) => ({ ...current, appointmentId }));
  }, [appointmentId]);

  useEffect(() => {
    setForm((current) => ({ ...current, doctorUserId }));
  }, [doctorUserId]);

  const availableExtraServices = useMemo(
    () => demoClinicContext.services.filter((service) => service.id !== demoClinicContext.services[0].id),
    [],
  );
  const hasAppointment = Boolean(form.appointmentId);

  return (
    <Panel>
      <PageTitle
        eyebrow="Кабинет врача"
        title="Завершение приема"
        description="Заполните жалобы, диагноз, рекомендации и дополнительные услуги для счета."
        action={<StatusBadge tone={hasAppointment ? "blue" : "slate"}>{hasAppointment ? recordNumber(form.appointmentId) : "Прием не выбран"}</StatusBadge>}
      />

      {!hasAppointment && (
        <EmptyState
          title="Нет выбранной записи"
          description="Создайте или выберите запись пациента, чтобы врач мог завершить прием."
        />
      )}

      <div className="visit-layout">
        <form
          className="form-card form-grid"
          aria-disabled={!hasAppointment}
          onSubmit={async (event) => {
            event.preventDefault();
            const response = await submit(form.appointmentId, {
              doctorUserId: form.doctorUserId,
              summary: `Жалобы: ${form.complaints}\nРекомендации: ${form.recommendations}`,
              diagnosis: form.diagnosis,
              services: form.extraServiceId
                ? [{ service_id: form.extraServiceId, quantity: form.quantity }]
                : [],
            });

            if (response) {
              onVisitCompleted(response.invoiceId);
            }
          }}
        >
          <label className="full-width">
            <span>Жалобы пациента</span>
            <textarea
              rows={4}
              value={form.complaints}
              onChange={(event) => setForm((current) => ({ ...current, complaints: event.target.value }))}
              disabled={!hasAppointment}
              required
            />
          </label>
          <label className="full-width">
            <span>Диагноз</span>
            <textarea
              rows={3}
              value={form.diagnosis}
              onChange={(event) => setForm((current) => ({ ...current, diagnosis: event.target.value }))}
              disabled={!hasAppointment}
              required
            />
          </label>
          <label className="full-width">
            <span>Рекомендации врача</span>
            <textarea
              rows={4}
              value={form.recommendations}
              onChange={(event) => setForm((current) => ({ ...current, recommendations: event.target.value }))}
              disabled={!hasAppointment}
              required
            />
          </label>
          <label>
            <span>Дополнительная услуга</span>
            <select
              value={form.extraServiceId}
              onChange={(event) => setForm((current) => ({ ...current, extraServiceId: event.target.value }))}
              disabled={!hasAppointment}
            >
              <option value="">Без дополнительных услуг</option>
              {availableExtraServices.map((service) => (
                <option key={service.id} value={service.id}>
                  {service.name}, {formatMoney(service.price)}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Количество</span>
            <input
              type="number"
              min={1}
              value={form.quantity}
              onChange={(event) => setForm((current) => ({ ...current, quantity: Number(event.target.value) || 1 }))}
              disabled={!hasAppointment}
            />
          </label>
          <div className="form-actions full-width">
            <button type="submit" disabled={isSubmitting || !hasAppointment}>
              {isSubmitting ? "Завершаем прием..." : "Завершить прием"}
            </button>
          </div>
        </form>

        <aside className="side-stack">
          <InfoCard
            title="История посещений"
            subtitle="Последние назначения будут отображаться в личном кабинете пациента."
            meta={<StatusBadge tone="green">Медкарта активна</StatusBadge>}
          >
            <p className="card-text">
              После завершения приема система сформирует счет, а пациент увидит рекомендации врача.
            </p>
          </InfoCard>
          <InfoCard
            title={result ? recordNumber(result.invoiceId, "Счет") : "Счет не сформирован"}
            subtitle={result ? "Передайте счет в кассу для оплаты." : "Счет появится после завершения приема."}
            meta={<StatusBadge tone={result ? "amber" : "slate"}>{result ? "Оплата ожидается" : "Ожидание"}</StatusBadge>}
          />
        </aside>
      </div>

      {error && <p className="error-box">{error}</p>}
      {result && (
        <p className="success-box">
          Прием завершен. {recordNumber(result.visitId, "Визит")} и {recordNumber(result.invoiceId, "счет")} созданы.
        </p>
      )}
    </Panel>
  );
}
