import { useState } from "react";
import { InfoCard, PageTitle, Panel, PatientStatusBadge } from "../components/ui";
import { PatientStatus } from "../types";
import { recordNumber } from "../utils/format";
import { usePatientRegistrationViewModel } from "../viewmodels/usePatientRegistrationViewModel";

type Props = {
  onPatientCreated: (patientId: string) => void;
};

const statusLabels: Record<PatientStatus, string> = {
  active: "Активный пациент",
  inactive: "Наблюдение приостановлено",
  archived: "Архивная карта",
};

export function PatientRegistrationPage({ onPatientCreated }: Props) {
  const [form, setForm] = useState({
    lastName: "Смирнова",
    firstName: "Елена",
    middleName: "Игоревна",
    birthDate: "1994-08-21",
    phone: "+7 999 555-44-33",
    email: "elena.smirnova@example.ru",
    status: "active" as PatientStatus,
  });
  const { isSubmitting, error, createdPatientId, submit } = usePatientRegistrationViewModel(onPatientCreated);

  return (
    <Panel>
      <PageTitle
        eyebrow="Регистратура"
        title="Новый пациент"
        description="Создайте карточку пациента для записи на прием и дальнейшего наблюдения."
        action={<PatientStatusBadge status={form.status} />}
      />

      <div className="patient-layout">
        <form
          className="form-card form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            void submit(form);
          }}
        >
          <label>
            <span>Фамилия</span>
            <input
              value={form.lastName}
              onChange={(event) => setForm((current) => ({ ...current, lastName: event.target.value }))}
              required
            />
          </label>
          <label>
            <span>Имя</span>
            <input
              value={form.firstName}
              onChange={(event) => setForm((current) => ({ ...current, firstName: event.target.value }))}
              required
            />
          </label>
          <label>
            <span>Отчество</span>
            <input
              value={form.middleName}
              onChange={(event) => setForm((current) => ({ ...current, middleName: event.target.value }))}
            />
          </label>
          <label>
            <span>Дата рождения</span>
            <input
              type="date"
              value={form.birthDate}
              onChange={(event) => setForm((current) => ({ ...current, birthDate: event.target.value }))}
              required
            />
          </label>
          <label>
            <span>Телефон</span>
            <input
              value={form.phone}
              onChange={(event) => setForm((current) => ({ ...current, phone: event.target.value }))}
              required
            />
          </label>
          <label>
            <span>Email</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
            />
          </label>
          <label>
            <span>Статус карты</span>
            <select
              value={form.status}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  status: event.target.value as PatientStatus,
                }))
              }
            >
              {Object.entries(statusLabels).map(([status, label]) => (
                <option key={status} value={status}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <div className="form-actions">
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Сохраняем карту..." : "Зарегистрировать пациента"}
            </button>
          </div>
        </form>

        <aside className="side-stack">
          <InfoCard
            title="После регистрации"
            subtitle="Пациента можно сразу выбрать для записи на прием."
            meta={<PatientStatusBadge status={form.status} />}
          >
            <p className="card-text">
              Контактные данные будут использоваться для подтверждения записи и уведомлений о приеме.
            </p>
          </InfoCard>
          {createdPatientId && (
            <InfoCard
              title={recordNumber(createdPatientId, "Карта пациента")}
              subtitle="Карточка создана и выбрана в текущем рабочем месте."
              meta={<PatientStatusBadge status="active" />}
            />
          )}
        </aside>
      </div>

      {error && <p className="error-box">{error}</p>}
      {createdPatientId && <p className="success-box">Карточка пациента создана. Можно оформить запись на прием.</p>}
    </Panel>
  );
}
