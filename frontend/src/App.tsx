import { FormEvent, useMemo, useState } from "react";
import { apiClient } from "./api/client";
import {
  AppLayout,
  BrandMark,
  EmptyState,
  InfoCard,
  PageTitle,
  Panel,
  StatCard,
  StatusBadge,
} from "./components/ui";
import { demoClinicContext } from "./demoData";
import { AuthenticatedUser, LoginRequest, RegisterPatientAccountRequest, UserRole } from "./types";
import { AppointmentWorkflowPage } from "./views/AppointmentWorkflowPage";
import { CompleteVisitPage } from "./views/CompleteVisitPage";
import { DoctorSchedulePage } from "./views/DoctorSchedulePage";
import { PatientRegistrationPage } from "./views/PatientRegistrationPage";
import { PaymentRegistrationPage } from "./views/PaymentRegistrationPage";
import { PriceListPage } from "./views/PriceListPage";
import { recordNumber } from "./utils/format";

type WorkspaceTab = "overview" | "patient" | "appointment" | "visit" | "payment" | "schedule" | "prices";

type TabDefinition = {
  key: WorkspaceTab;
  label: string;
  caption: string;
};

const sessionStorageKey = "clinic.session";

const roleLabels: Record<UserRole, string> = {
  administrator: "Администратор",
  patient: "Пациент",
  doctor: "Врач",
};

const workspaceTitles: Record<UserRole, string> = {
  administrator: "Панель администратора",
  doctor: "Кабинет врача",
  patient: "Личный кабинет пациента",
};

const roleDescriptions: Record<UserRole, string> = {
  administrator: "Регистратура, расписание, записи и касса",
  patient: "Запись на прием, история посещений и счета",
  doctor: "Расписание, приемы и медицинские рекомендации",
};

const displayNamesByLogin: Record<string, string> = {
  admin: "Мария Андреевна Орлова",
  doctor_ivanov: "Иван Сергеевич Иванов",
  patient_demo: "Анна Сергеевна Петрова",
};

const roleTabs: Record<UserRole, TabDefinition[]> = {
  administrator: [
    { key: "overview", label: "Главная", caption: "сводка клиники" },
    { key: "patient", label: "Новый пациент", caption: "карта пациента" },
    { key: "appointment", label: "Оформить запись", caption: "прием и отмена" },
    { key: "payment", label: "Оплата", caption: "счет и касса" },
    { key: "schedule", label: "Расписание", caption: "окна врача" },
    { key: "prices", label: "Услуги", caption: "прайс-лист" },
  ],
  doctor: [
    { key: "overview", label: "Сегодня", caption: "пациенты и прием" },
    { key: "visit", label: "Прием", caption: "диагноз и рекомендации" },
    { key: "schedule", label: "Расписание", caption: "занятые окна" },
    { key: "prices", label: "Услуги", caption: "назначения" },
  ],
  patient: [
    { key: "overview", label: "Мой кабинет", caption: "записи и счета" },
    { key: "appointment", label: "Записаться", caption: "выбор услуги" },
    { key: "prices", label: "Услуги", caption: "специализации" },
  ],
};

function defaultTabForRole(role: UserRole): WorkspaceTab {
  return roleTabs[role][0].key;
}

function readStoredSession() {
  try {
    const raw = window.sessionStorage.getItem(sessionStorageKey);
    return raw ? (JSON.parse(raw) as AuthenticatedUser) : null;
  } catch {
    return null;
  }
}

function getInitials(fullName: string) {
  const parts = fullName
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "З";
}

function getDisplayName(user: AuthenticatedUser) {
  return displayNamesByLogin[user.login] ?? user.fullName;
}

export default function App() {
  const [session, setSession] = useState<AuthenticatedUser | null>(() => readStoredSession());
  const [selectedPatientId, setSelectedPatientId] = useState(demoClinicContext.patients[0].id);
  const [selectedAppointmentId, setSelectedAppointmentId] = useState("");
  const [selectedInvoiceId, setSelectedInvoiceId] = useState("");
  const [activeTab, setActiveTab] = useState<WorkspaceTab>(() =>
    defaultTabForRole(readStoredSession()?.role ?? "patient"),
  );

  function handleLogin(user: AuthenticatedUser) {
    window.sessionStorage.setItem(sessionStorageKey, JSON.stringify(user));
    setSession(user);
    setActiveTab(defaultTabForRole(user.role));
  }

  function handleLogout() {
    window.sessionStorage.removeItem(sessionStorageKey);
    setSession(null);
    setSelectedAppointmentId("");
    setSelectedInvoiceId("");
  }

  if (!session) {
    return <LoginScreen onLogin={handleLogin} />;
  }

  const tabs = roleTabs[session.role];
  const currentPatientId =
    session.role === "patient" ? (session.patientId ?? demoClinicContext.patients[0].id) : selectedPatientId;
  const currentPatient = demoClinicContext.patients.find((patient) => patient.id === currentPatientId);
  const currentDoctorId = session.doctorId ?? demoClinicContext.doctorId;
  const activeTabDefinition = tabs.find((tab) => tab.key === activeTab) ?? tabs[0];
  const displayName = getDisplayName(session);

  return (
    <AppLayout
      sidebar={
        <Sidebar
          activeTab={activeTabDefinition.key}
          displayName={displayName}
          role={session.role}
          tabs={tabs}
          onNavigate={setActiveTab}
        />
      }
      header={
        <WorkspaceHeader
          activeTab={activeTabDefinition}
          currentPatientId={currentPatientId}
          currentPatientName={currentPatient?.fullName}
          displayName={displayName}
          role={session.role}
          selectedAppointmentId={selectedAppointmentId}
          selectedInvoiceId={selectedInvoiceId}
          onLogout={handleLogout}
          onPatientChange={(patientId) => {
            setSelectedPatientId(patientId);
            setSelectedAppointmentId("");
            setSelectedInvoiceId("");
          }}
        />
      }
    >
      <WorkspaceView
        activeTab={activeTabDefinition.key}
        currentDoctorId={currentDoctorId}
        currentPatientId={currentPatientId}
        currentPatientName={currentPatient?.fullName ?? "Пациент клиники"}
        selectedAppointmentId={selectedAppointmentId}
        selectedInvoiceId={selectedInvoiceId}
        session={session}
        onAppointmentBooked={(appointmentId) => {
          setSelectedAppointmentId(appointmentId);
          setSelectedInvoiceId("");
        }}
        onAppointmentCancelled={(appointmentId) => {
          if (appointmentId === selectedAppointmentId) {
            setSelectedAppointmentId("");
            setSelectedInvoiceId("");
          }
        }}
        onNavigate={setActiveTab}
        onPatientCreated={(patientId) => {
          setSelectedPatientId(patientId);
          setSelectedAppointmentId("");
          setSelectedInvoiceId("");
        }}
        onVisitCompleted={(invoiceId) => setSelectedInvoiceId(invoiceId)}
      />
    </AppLayout>
  );
}

function Sidebar({
  activeTab,
  displayName,
  role,
  tabs,
  onNavigate,
}: {
  activeTab: WorkspaceTab;
  displayName: string;
  role: UserRole;
  tabs: TabDefinition[];
  onNavigate: (tab: WorkspaceTab) => void;
}) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <BrandMark />
        <span>
          <strong>МЦ «Здоровье»</strong>
          <small>медицинская информационная система</small>
        </span>
      </div>
      <div className="sidebar-user">
        <span className="user-avatar">{getInitials(displayName)}</span>
        <span>
          <strong>{displayName}</strong>
          <small>{roleLabels[role]}</small>
        </span>
      </div>
      <nav className="nav-list" aria-label="Разделы кабинета">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            className={tab.key === activeTab ? "nav-tab active" : "nav-tab"}
            type="button"
            onClick={() => onNavigate(tab.key)}
          >
            <span>{tab.label}</span>
            <small>{tab.caption}</small>
          </button>
        ))}
      </nav>
      <div className="sidebar-note">
        <span className="status-dot" />
        Клиника открыта сегодня с 08:00 до 20:00
      </div>
    </aside>
  );
}

function WorkspaceHeader({
  activeTab,
  currentPatientId,
  currentPatientName,
  displayName,
  role,
  selectedAppointmentId,
  selectedInvoiceId,
  onLogout,
  onPatientChange,
}: {
  activeTab: TabDefinition;
  currentPatientId: string;
  currentPatientName?: string;
  displayName: string;
  role: UserRole;
  selectedAppointmentId: string;
  selectedInvoiceId: string;
  onLogout: () => void;
  onPatientChange: (patientId: string) => void;
}) {
  return (
    <header className="workspace-header">
      <div className="header-topline">
        <div>
          <p className="eyebrow">Медицинский центр «Здоровье»</p>
          <h1>{workspaceTitles[role]}</h1>
          <p className="header-subtitle">{roleDescriptions[role]}</p>
        </div>
        <div className="header-profile">
          <span className="user-avatar">{getInitials(displayName)}</span>
          <span>
            <strong>{displayName}</strong>
            <small>{roleLabels[role]}</small>
          </span>
          <button className="ghost-button" type="button" onClick={onLogout}>
            Выйти
          </button>
        </div>
      </div>

      <div className="header-context">
        <div>
          <span>Раздел</span>
          <strong>{activeTab.label}</strong>
        </div>
        <div>
          <span>Пациент</span>
          {role === "patient" ? (
            <strong>{currentPatientName ?? "Пациент клиники"}</strong>
          ) : (
            <select
              aria-label="Выбранный пациент"
              value={currentPatientId}
              onChange={(event) => onPatientChange(event.target.value)}
            >
              {demoClinicContext.patients.map((patient) => (
                <option key={patient.id} value={patient.id}>
                  {patient.fullName}
                </option>
              ))}
            </select>
          )}
        </div>
        <div>
          <span>Врач</span>
          <strong>{demoClinicContext.doctorName}, каб. {demoClinicContext.cabinetNumber}</strong>
        </div>
        <div>
          <span>Текущий прием</span>
          <strong>{selectedAppointmentId ? recordNumber(selectedAppointmentId) : "Не выбран"}</strong>
        </div>
        <div>
          <span>Счет</span>
          <strong>{selectedInvoiceId ? recordNumber(selectedInvoiceId, "Счет") : "Не выставлен"}</strong>
        </div>
      </div>
    </header>
  );
}

type LoginScreenProps = {
  onLogin: (user: AuthenticatedUser) => void;
};

function LoginScreen({ onLogin }: LoginScreenProps) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [form, setForm] = useState<LoginRequest>({
    login: "",
    password: "",
  });
  const [registrationForm, setRegistrationForm] = useState<RegisterPatientAccountRequest>({
    lastName: "",
    firstName: "",
    middleName: "",
    birthDate: "",
    phone: "",
    email: "",
    password: "",
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function signIn(credentials: LoginRequest) {
    setIsSubmitting(true);
    setError(null);

    try {
      const user = await apiClient.login(credentials);
      onLogin(user);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось выполнить вход");
    } finally {
      setIsSubmitting(false);
    }
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void signIn(form);
  }

  async function registerPatient(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const user = await apiClient.registerPatientAccount(registrationForm);
      onLogin(user);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось создать кабинет пациента");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-shell">
        <aside className="login-panel">
          <div className="login-brand">
            <BrandMark size="large" />
            <span>
              <strong>Медицинский центр «Здоровье»</strong>
              <small>единый вход для пациентов и сотрудников</small>
            </span>
          </div>
          <div className="login-panel-copy">
            <p className="eyebrow">Медицинская система</p>
            <h2>Вход в защищенную рабочую область</h2>
          </div>
        </aside>

        <div className="login-card">
          <p className="eyebrow">{mode === "login" ? "Вход в систему" : "Новый пациент"}</p>
          <h1>{mode === "login" ? "Авторизация" : "Создание личного кабинета"}</h1>
          <p className="login-copy">
            {mode === "login"
              ? "Введите учетные данные, выданные администратором клиники."
              : "Заполните данные пациента, чтобы создать личный кабинет и карточку в регистратуре."}
          </p>

          {mode === "login" && (
            <form className="login-form" onSubmit={submit}>
              <label>
                <span>Логин, телефон или email</span>
                <input
                  autoComplete="username"
                  placeholder="Введите логин"
                  value={form.login}
                  onChange={(event) => setForm((current) => ({ ...current, login: event.target.value }))}
                  required
                />
              </label>
              <label>
                <span>Пароль</span>
                <input
                  autoComplete="current-password"
                  type="password"
                  value={form.password}
                  onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
                  required
                />
              </label>
              <div className="auth-actions">
                <button className="primary-button" type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "Проверяем данные..." : "Войти"}
                </button>
                <button className="ghost-button" type="button" onClick={() => setMode("register")}>
                  Зарегистрироваться
                </button>
              </div>
            </form>
          )}

          {mode === "register" && (
            <form className="login-form registration-form" onSubmit={registerPatient}>
              <div className="form-grid">
                <label>
                  <span>Фамилия</span>
                  <input
                    value={registrationForm.lastName}
                    onChange={(event) =>
                      setRegistrationForm((current) => ({ ...current, lastName: event.target.value }))
                    }
                    required
                  />
                </label>
                <label>
                  <span>Имя</span>
                  <input
                    value={registrationForm.firstName}
                    onChange={(event) =>
                      setRegistrationForm((current) => ({ ...current, firstName: event.target.value }))
                    }
                    required
                  />
                </label>
                <label>
                  <span>Отчество</span>
                  <input
                    value={registrationForm.middleName}
                    onChange={(event) =>
                      setRegistrationForm((current) => ({ ...current, middleName: event.target.value }))
                    }
                  />
                </label>
                <label>
                  <span>Дата рождения</span>
                  <input
                    type="date"
                    value={registrationForm.birthDate}
                    onChange={(event) =>
                      setRegistrationForm((current) => ({ ...current, birthDate: event.target.value }))
                    }
                    required
                  />
                </label>
              </div>
              <label>
                <span>Телефон</span>
                <input
                  autoComplete="tel"
                  placeholder="+7 999 555-44-33"
                  value={registrationForm.phone}
                  onChange={(event) => setRegistrationForm((current) => ({ ...current, phone: event.target.value }))}
                  required
                />
              </label>
              <label>
                <span>Email</span>
                <input
                  autoComplete="email"
                  type="email"
                  value={registrationForm.email}
                  onChange={(event) => setRegistrationForm((current) => ({ ...current, email: event.target.value }))}
                  required
                />
              </label>
              <label>
                <span>Пароль</span>
                <input
                  autoComplete="new-password"
                  minLength={8}
                  type="password"
                  value={registrationForm.password}
                  onChange={(event) =>
                    setRegistrationForm((current) => ({ ...current, password: event.target.value }))
                  }
                  required
                />
              </label>
              <div className="auth-actions">
                <button className="primary-button" type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "Создаем кабинет..." : "Создать кабинет"}
                </button>
                <button className="ghost-button" type="button" onClick={() => setMode("login")}>
                  Вернуться ко входу
                </button>
              </div>
            </form>
          )}

          {error && <p className="error-box">{error}</p>}
        </div>
      </section>
    </main>
  );
}

type WorkspaceViewProps = {
  activeTab: WorkspaceTab;
  currentDoctorId: string;
  currentPatientId: string;
  currentPatientName: string;
  selectedAppointmentId: string;
  selectedInvoiceId: string;
  session: AuthenticatedUser;
  onAppointmentBooked: (appointmentId: string) => void;
  onAppointmentCancelled: (appointmentId: string) => void;
  onNavigate: (tab: WorkspaceTab) => void;
  onPatientCreated: (patientId: string) => void;
  onVisitCompleted: (invoiceId: string) => void;
};

function WorkspaceView({
  activeTab,
  currentDoctorId,
  currentPatientId,
  currentPatientName,
  selectedAppointmentId,
  selectedInvoiceId,
  session,
  onAppointmentBooked,
  onAppointmentCancelled,
  onNavigate,
  onPatientCreated,
  onVisitCompleted,
}: WorkspaceViewProps) {
  switch (activeTab) {
    case "patient":
      return <PatientRegistrationPage onPatientCreated={onPatientCreated} />;
    case "appointment":
      return (
        <AppointmentWorkflowPage
          actorUserId={session.id}
          patientId={currentPatientId}
          patientName={currentPatientName}
          selectedAppointmentId={selectedAppointmentId}
          onAppointmentBooked={onAppointmentBooked}
          onAppointmentCancelled={onAppointmentCancelled}
        />
      );
    case "visit":
      return (
        <CompleteVisitPage
          appointmentId={selectedAppointmentId}
          doctorUserId={session.id}
          onVisitCompleted={onVisitCompleted}
        />
      );
    case "payment":
      return <PaymentRegistrationPage invoiceId={selectedInvoiceId} receivedByUserId={session.id} />;
    case "schedule":
      return <DoctorSchedulePage defaultDoctorId={currentDoctorId} />;
    case "prices":
      return <PriceListPage />;
    case "overview":
    default:
      return (
        <OverviewPanel
          currentPatientName={currentPatientName}
          selectedAppointmentId={selectedAppointmentId}
          selectedInvoiceId={selectedInvoiceId}
          session={session}
          onNavigate={onNavigate}
        />
      );
  }
}

type OverviewPanelProps = {
  currentPatientName: string;
  selectedAppointmentId: string;
  selectedInvoiceId: string;
  session: AuthenticatedUser;
  onNavigate: (tab: WorkspaceTab) => void;
};

function OverviewPanel({
  currentPatientName,
  selectedAppointmentId,
  selectedInvoiceId,
  session,
  onNavigate,
}: OverviewPanelProps) {
  const stats = useMemo(() => {
    if (session.role === "administrator") {
      return [
        { label: "Пациенты", value: demoClinicContext.patients.length, detail: "активные карты", tone: "blue" as const },
        { label: "Врачи", value: 1, detail: "принимают сегодня", tone: "green" as const },
        { label: "Записи сегодня", value: selectedAppointmentId ? 1 : 0, detail: "по текущей смене", tone: "amber" as const },
        { label: "Ожидают оплаты", value: selectedInvoiceId ? 1 : 0, detail: "счета к закрытию", tone: "red" as const },
      ];
    }

    if (session.role === "doctor") {
      return [
        { label: "Пациенты сегодня", value: selectedAppointmentId ? 1 : 0, detail: "в расписании", tone: "blue" as const },
        { label: "Свободные окна", value: 6, detail: "можно записать пациента", tone: "green" as const },
        { label: "Приемы завершены", value: selectedInvoiceId ? 1 : 0, detail: "за смену", tone: "amber" as const },
        { label: "Рекомендации", value: selectedInvoiceId ? 1 : 0, detail: "выданы пациентам", tone: "slate" as const },
      ];
    }

    return [
      { label: "Ближайшие записи", value: selectedAppointmentId ? 1 : 0, detail: "активные визиты", tone: "blue" as const },
      { label: "Врачи", value: 1, detail: "доступны для записи", tone: "green" as const },
      { label: "История посещений", value: selectedInvoiceId ? 1 : 0, detail: "завершенные приемы", tone: "amber" as const },
      { label: "Счета", value: selectedInvoiceId ? 1 : 0, detail: "к оплате или оплачены", tone: "red" as const },
    ];
  }, [selectedAppointmentId, selectedInvoiceId, session.role]);

  return (
    <Panel className="overview-panel">
      <PageTitle
        eyebrow={workspaceTitles[session.role]}
        title={session.role === "doctor" ? "Расписание и текущий прием" : "Сводка на сегодня"}
        description="Главные события и быстрые действия без служебных данных."
      />

      <div className="stat-grid">
        {stats.map((tile) => (
          <StatCard key={tile.label} {...tile} />
        ))}
      </div>

      <div className="dashboard-grid">
        <InfoCard
          title={selectedAppointmentId ? recordNumber(selectedAppointmentId) : "Ближайшая запись не выбрана"}
          subtitle={
            selectedAppointmentId
              ? `${currentPatientName}, ${demoClinicContext.doctorName}, кабинет ${demoClinicContext.cabinetNumber}`
              : "Оформите запись пациента, чтобы она появилась в сводке."
          }
          meta={<StatusBadge tone={selectedAppointmentId ? "blue" : "slate"}>{selectedAppointmentId ? "Запланировано" : "Ожидание"}</StatusBadge>}
        >
          <div className="card-actions">
            <button type="button" onClick={() => onNavigate("appointment")}>
              Оформить запись
            </button>
            {session.role === "doctor" && (
              <button className="ghost-button" type="button" onClick={() => onNavigate("visit")}>
                Открыть прием
              </button>
            )}
          </div>
        </InfoCard>

        <InfoCard
          title={selectedInvoiceId ? recordNumber(selectedInvoiceId, "Счет") : "Оплата пока не требуется"}
          subtitle={
            selectedInvoiceId
              ? "Счет сформирован после завершения приема и готов к оплате."
              : "Счет появится после завершения приема врачом."
          }
          meta={<StatusBadge tone={selectedInvoiceId ? "amber" : "green"}>{selectedInvoiceId ? "Оплата ожидается" : "Без задолженности"}</StatusBadge>}
        >
          <div className="card-actions">
            {session.role === "administrator" && (
              <button type="button" onClick={() => onNavigate("payment")}>
                Перейти к оплате
              </button>
            )}
            <button className="ghost-button" type="button" onClick={() => onNavigate("prices")}>
              Смотреть услуги
            </button>
          </div>
        </InfoCard>

        <InfoCard
          title="Карточка пациента"
          subtitle={`${currentPatientName}. История посещений и рекомендации доступны после приема.`}
          meta={<StatusBadge tone="green">Наблюдение активно</StatusBadge>}
        >
          <p className="card-text">
            В интерфейсе отображаются только понятные номера записей и счетов. Служебные идентификаторы остаются внутри
            запросов к API.
          </p>
        </InfoCard>
      </div>

      {session.role === "patient" && !selectedAppointmentId && (
        <EmptyState
          title="У вас пока нет ближайшей записи"
          description="Выберите услугу и удобное время приема в разделе записи."
          action={
            <button type="button" onClick={() => onNavigate("appointment")}>
              Записаться на прием
            </button>
          }
        />
      )}
    </Panel>
  );
}
