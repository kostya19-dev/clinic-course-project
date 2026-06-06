export type PatientStatus = "active" | "inactive" | "archived";
export type AppointmentStatus = "booked" | "cancelled" | "completed";
export type PaymentMethod = "cash" | "card" | "transfer";
export type UserRole = "administrator" | "patient" | "doctor";

export type PriceListItem = {
  serviceId: string;
  code: string;
  name: string;
  durationMinutes: number;
  price: number;
  currencyCode: string;
  effectiveFrom: string;
};

export type ScheduleSlot = {
  id: string;
  doctorId: string;
  startsAt: string;
  endsAt: string;
  slotDurationMinutes: number;
};

export type AppointmentSummary = {
  id: string;
  patientId: string;
  doctorId: string;
  bookedServiceId: string;
  startsAt: string;
  endsAt: string;
  status: AppointmentStatus;
};

export type ScheduleView = {
  doctorId: string;
  from: string;
  to: string;
  availability: ScheduleSlot[];
  appointments: AppointmentSummary[];
};

export type CreatePatientRequest = {
  lastName: string;
  firstName: string;
  middleName: string;
  birthDate: string;
  phone: string;
  email: string;
  status: PatientStatus;
};

export type BookAppointmentRequest = {
  patientId: string;
  doctorId: string;
  serviceId: string;
  startsAt: string;
  createdByUserId: string;
  note: string;
};

export type CancelAppointmentRequest = {
  actorUserId: string;
  reason: string;
};

export type PerformedServiceRequest = {
  service_id: string;
  quantity: number;
};

export type CompleteVisitRequest = {
  doctorUserId: string;
  summary: string;
  diagnosis: string;
  services: PerformedServiceRequest[];
};

export type RegisterPaymentRequest = {
  receivedByUserId: string;
  amount: number;
  paymentMethod: PaymentMethod;
  externalReference: string;
};

export type LoginRequest = {
  login: string;
  password: string;
};

export type RegisterPatientAccountRequest = {
  lastName: string;
  firstName: string;
  middleName: string;
  birthDate: string;
  phone: string;
  email: string;
  password: string;
};

export type AuthenticatedUser = {
  id: string;
  login: string;
  fullName: string;
  email: string;
  role: UserRole;
  doctorId?: string;
  patientId?: string;
};

export type IDResponse = {
  id: string;
};

export type CompleteVisitResponse = {
  visitId: string;
  invoiceId: string;
};
