import { useEffect, useMemo, useState } from "react";
import { EmptyState, InfoCard, PageTitle, Panel, StatusBadge } from "../components/ui";
import { demoClinicContext } from "../demoData";
import { formatDateTime, formatMoney, recordNumber } from "../utils/format";
import { useAppointmentWorkflowViewModel } from "../viewmodels/useAppointmentWorkflowViewModel";

type Props = {
  patientId: string;
  patientName: string;
  actorUserId: string;
  selectedAppointmentId: string;
  onAppointmentBooked: (appointmentId: string) => void;
  onAppointmentCancelled: (appointmentId: string) => void;
};

function toMoscowDateTime(value: string) {
  return `${value}:00+03:00`;
}

export function AppointmentWorkflowPage({
  patientId,
  patientName,
  actorUserId,
  selectedAppointmentId,
  onAppointmentBooked,
  onAppointmentCancelled,
}: Props) {
  const [bookingForm, setBookingForm] = useState({
    patientId,
    doctorId: demoClinicContext.doctorId,
    serviceId: demoClinicContext.services[0].id,
    startsAt: demoClinicContext.defaultAppointmentAt,
    createdByUserId: actorUserId,
    note: "Пациент просит подтвердить запись по телефону за день до приема.",
  });
  const [cancelReason, setCancelReason] = useState("Пациент выбрал другое время приема");
  const {
    isBooking,
    isCancelling,
    bookError,
    cancelError,
    lastAppointmentId,
    cancelSuccessMessage,
    book,
    cancel,
  } = useAppointmentWorkflowViewModel();

  useEffect(() => {
    setBookingForm((current) => ({ ...current, patientId }));
  }, [patientId]);

  useEffect(() => {
    setBookingForm((current) => ({ ...current, createdByUserId: actorUserId }));
  }, [actorUserId]);

  const selectedService = useMemo(
    () => demoClinicContext.services.find((service) => service.id === bookingForm.serviceId),
    [bookingForm.serviceId],
  );
  const appointmentForCancel = selectedAppointmentId || lastAppointmentId || "";

  return (
    <Panel>
      <PageTitle
        eyebrow="Запись на прием"
        title="Оформить запись"
        description="Выберите услугу, время и при необходимости добавьте комментарий для регистратуры."
        action={<StatusBadge tone="blue">{patientName}</StatusBadge>}
      />

      <div className="appointment-layout">
        <form
          className="form-card form-grid"
          onSubmit={async (event) => {
            event.preventDefault();
            const appointmentId = await book({
              ...bookingForm,
              startsAt: toMoscowDateTime(bookingForm.startsAt),
            });

            if (appointmentId) {
              onAppointmentBooked(appointmentId);
            }
          }}
        >
          <label className="full-width">
            <span>Пациент</span>
            <input value={patientName} readOnly />
          </label>
          <label>
            <span>Врач</span>
            <select
              value={bookingForm.doctorId}
              onChange={(event) => setBookingForm((current) => ({ ...current, doctorId: event.target.value }))}
            >
              <option value={demoClinicContext.doctorId}>
                {demoClinicContext.doctorName}, кабинет {demoClinicContext.cabinetNumber}
              </option>
            </select>
          </label>
          <label>
            <span>Услуга</span>
            <select
              value={bookingForm.serviceId}
              onChange={(event) => setBookingForm((current) => ({ ...current, serviceId: event.target.value }))}
            >
              {demoClinicContext.services.map((service) => (
                <option key={service.id} value={service.id}>
                  {service.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Дата и время приема</span>
            <input
              type="datetime-local"
              value={bookingForm.startsAt}
              onChange={(event) => setBookingForm((current) => ({ ...current, startsAt: event.target.value }))}
              required
            />
          </label>
          <label>
            <span>Стоимость</span>
            <input value={selectedService ? formatMoney(selectedService.price) : "Будет рассчитана"} readOnly />
          </label>
          <label className="full-width">
            <span>Комментарий для клиники</span>
            <textarea
              rows={3}
              value={bookingForm.note}
              onChange={(event) => setBookingForm((current) => ({ ...current, note: event.target.value }))}
            />
          </label>
          <div className="form-actions full-width">
            <button type="submit" disabled={isBooking}>
              {isBooking ? "Оформляем запись..." : "Записать на прием"}
            </button>
          </div>
        </form>

        <aside className="side-stack">
          <InfoCard
            title={lastAppointmentId ? recordNumber(lastAppointmentId) : "Новая запись"}
            subtitle={`${demoClinicContext.doctorName}, ${formatDateTime(toMoscowDateTime(bookingForm.startsAt))}`}
            meta={<StatusBadge tone={lastAppointmentId ? "green" : "blue"}>{lastAppointmentId ? "Создана" : "Черновик"}</StatusBadge>}
          >
            <p className="card-text">
              После оформления запись станет доступна врачу для приема, а администратору — для дальнейшего
              сопровождения.
            </p>
          </InfoCard>

          <InfoCard
            title={appointmentForCancel ? "Отмена записи" : "Нет записи для отмены"}
            subtitle={
              appointmentForCancel
                ? `${recordNumber(appointmentForCancel)} можно отменить с указанием причины.`
                : "Сначала оформите запись или выберите существующий прием."
            }
            meta={<StatusBadge tone={appointmentForCancel ? "amber" : "slate"}>{appointmentForCancel ? "Доступно" : "Ожидание"}</StatusBadge>}
          >
            {appointmentForCancel ? (
              <form
                className="inline-form"
                onSubmit={async (event) => {
                  event.preventDefault();
                  const isSuccess = await cancel(appointmentForCancel, {
                    actorUserId,
                    reason: cancelReason,
                  });

                  if (isSuccess) {
                    onAppointmentCancelled(appointmentForCancel);
                  }
                }}
              >
                <label>
                  <span>Причина отмены</span>
                  <textarea
                    rows={3}
                    value={cancelReason}
                    onChange={(event) => setCancelReason(event.target.value)}
                    required
                  />
                </label>
                <button className="danger-button" type="submit" disabled={isCancelling}>
                  {isCancelling ? "Отменяем запись..." : "Отменить запись"}
                </button>
              </form>
            ) : (
              <EmptyState
                title="Запись еще не выбрана"
                description="После создания приема здесь появится действие для отмены."
              />
            )}
          </InfoCard>
        </aside>
      </div>

      {bookError && <p className="error-box">{bookError}</p>}
      {lastAppointmentId && <p className="success-box">Запись создана: {recordNumber(lastAppointmentId)}.</p>}
      {cancelError && <p className="error-box">{cancelError}</p>}
      {cancelSuccessMessage && <p className="success-box">Запись отменена. Пациенту можно подобрать другое время.</p>}
    </Panel>
  );
}
