import { ReactNode } from "react";
import { AppointmentStatus, PatientStatus } from "../types";

type Tone = "blue" | "green" | "amber" | "red" | "slate";

function classes(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(" ");
}

export function AppLayout({
  sidebar,
  header,
  children,
}: {
  sidebar: ReactNode;
  header: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="app-layout">
      {sidebar}
      <div className="app-main">
        {header}
        <main className="page-stack">{children}</main>
      </div>
    </div>
  );
}

export function BrandMark({ size = "normal" }: { size?: "normal" | "large" }) {
  return (
    <span className={classes("brand-mark", size === "large" && "brand-mark-large")} aria-hidden="true">
      З
    </span>
  );
}

export function PageTitle({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow?: string;
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="section-heading">
      <div>
        {eyebrow && <p className="eyebrow">{eyebrow}</p>}
        <h2>{title}</h2>
        {description && <p className="section-description">{description}</p>}
      </div>
      {action && <div className="section-action">{action}</div>}
    </div>
  );
}

export function Panel({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return <section className={classes("panel", className)}>{children}</section>;
}

export function StatCard({
  label,
  value,
  detail,
  tone = "blue",
}: {
  label: string;
  value: string | number;
  detail?: string;
  tone?: Tone;
}) {
  return (
    <article className={classes("stat-card", `tone-${tone}`)}>
      <span>{label}</span>
      <strong>{value}</strong>
      {detail && <small>{detail}</small>}
    </article>
  );
}

export function InfoCard({
  title,
  subtitle,
  meta,
  children,
}: {
  title: string;
  subtitle?: string;
  meta?: ReactNode;
  children?: ReactNode;
}) {
  return (
    <article className="info-card">
      <div className="info-card-head">
        <div>
          <h3>{title}</h3>
          {subtitle && <p>{subtitle}</p>}
        </div>
        {meta}
      </div>
      {children}
    </article>
  );
}

export function StatusBadge({
  children,
  tone = "slate",
}: {
  children: ReactNode;
  tone?: Tone;
}) {
  return <span className={classes("status-badge", `tone-${tone}`)}>{children}</span>;
}

export function AppointmentStatusBadge({ status }: { status: AppointmentStatus }) {
  const labels: Record<AppointmentStatus, string> = {
    booked: "Запланировано",
    cancelled: "Отменено",
    completed: "Прием завершен",
  };
  const tones: Record<AppointmentStatus, Tone> = {
    booked: "blue",
    cancelled: "red",
    completed: "green",
  };

  return <StatusBadge tone={tones[status]}>{labels[status]}</StatusBadge>;
}

export function PatientStatusBadge({ status }: { status: PatientStatus }) {
  const labels: Record<PatientStatus, string> = {
    active: "Активный пациент",
    inactive: "Наблюдение приостановлено",
    archived: "Архивная карта",
  };
  const tones: Record<PatientStatus, Tone> = {
    active: "green",
    inactive: "amber",
    archived: "slate",
  };

  return <StatusBadge tone={tones[status]}>{labels[status]}</StatusBadge>;
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="empty-state">
      <div className="empty-state-mark" aria-hidden="true">
        +
      </div>
      <h3>{title}</h3>
      {description && <p>{description}</p>}
      {action && <div>{action}</div>}
    </div>
  );
}

export function LoadingState({ text = "Загружаем данные..." }: { text?: string }) {
  return (
    <div className="loading-state" role="status">
      <span className="loader" aria-hidden="true" />
      {text}
    </div>
  );
}
