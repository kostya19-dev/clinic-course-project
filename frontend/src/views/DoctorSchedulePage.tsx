import { useEffect, useMemo, useState } from "react";
import { AppointmentStatusBadge, EmptyState, InfoCard, LoadingState, PageTitle, Panel, StatusBadge } from "../components/ui";
import { demoClinicContext } from "../demoData";
import { formatDate, formatTime, recordNumber } from "../utils/format";
import { useDoctorScheduleViewModel } from "../viewmodels/useDoctorScheduleViewModel";

function toRangeStart(value: string) {
  return `${value}T00:00:00+03:00`;
}

function toRangeEnd(value: string) {
  return `${value}T23:59:59+03:00`;
}

type Props = {
  defaultDoctorId?: string;
};

export function DoctorSchedulePage({ defaultDoctorId = demoClinicContext.doctorId }: Props) {
  const [doctorId, setDoctorId] = useState(defaultDoctorId);
  const [fromDate, setFromDate] = useState(demoClinicContext.defaultScheduleFromDate);
  const [toDate, setToDate] = useState(demoClinicContext.defaultScheduleToDate);
  const from = toRangeStart(fromDate);
  const to = toRangeEnd(toDate);
  const { schedule, isLoading, error } = useDoctorScheduleViewModel(doctorId, from, to);

  useEffect(() => {
    setDoctorId(defaultDoctorId);
  }, [defaultDoctorId]);

  const doctorSummary = useMemo(
    () => `${demoClinicContext.doctorName}, кабинет ${demoClinicContext.cabinetNumber}`,
    [],
  );

  return (
    <Panel>
      <PageTitle
        eyebrow="Расписание врача"
        title="Рабочие окна и записи"
        description="Проверьте свободные интервалы и занятые слоты врача за выбранный период."
        action={<StatusBadge tone="green">{doctorSummary}</StatusBadge>}
      />

      <form className="filter-bar">
        <label>
          <span>Начало периода</span>
          <input type="date" value={fromDate} onChange={(event) => setFromDate(event.target.value)} />
        </label>
        <label>
          <span>Окончание периода</span>
          <input type="date" value={toDate} onChange={(event) => setToDate(event.target.value)} />
        </label>
      </form>

      {isLoading && <LoadingState text="Загружаем расписание врача..." />}
      {error && <p className="error-box">Не удалось загрузить расписание врача. {error}</p>}
      {!isLoading && !error && schedule && (
        <div className="schedule-grid">
          <section>
            <div className="subheading-row">
              <h3>Свободные интервалы</h3>
              <StatusBadge tone="blue">{schedule.availability.length}</StatusBadge>
            </div>
            <div className="card-list">
              {schedule.availability.length === 0 && (
                <EmptyState
                  title="Свободных окон нет"
                  description="Попробуйте выбрать другой день или проверьте записи врача."
                />
              )}
              {schedule.availability.map((slot) => (
                <InfoCard
                  key={slot.id}
                  title={`${formatTime(slot.startsAt)} — ${formatTime(slot.endsAt)}`}
                  subtitle={`${formatDate(slot.startsAt)}, ${slot.slotDurationMinutes} минут`}
                  meta={<StatusBadge tone="green">Доступно</StatusBadge>}
                />
              ))}
            </div>
          </section>

          <section>
            <div className="subheading-row">
              <h3>Ближайшие записи</h3>
              <StatusBadge tone="amber">{schedule.appointments.length}</StatusBadge>
            </div>
            <div className="card-list">
              {schedule.appointments.length === 0 && (
                <EmptyState
                  title="Записей в периоде нет"
                  description="Когда пациент будет записан, прием появится в этом списке."
                />
              )}
              {schedule.appointments.map((appointment) => {
                const patientName =
                  demoClinicContext.patients.find((patient) => patient.id === appointment.patientId)?.fullName ??
                  "Пациент клиники";
                const serviceName =
                  demoClinicContext.services.find((service) => service.id === appointment.bookedServiceId)?.name ??
                  "Медицинская услуга";

                return (
                  <InfoCard
                    key={appointment.id}
                    title={recordNumber(appointment.id)}
                    subtitle={`${patientName}, ${formatTime(appointment.startsAt)} — ${formatTime(appointment.endsAt)}`}
                    meta={<AppointmentStatusBadge status={appointment.status} />}
                  >
                    <p className="card-text">{serviceName}</p>
                  </InfoCard>
                );
              })}
            </div>
          </section>
        </div>
      )}
    </Panel>
  );
}
