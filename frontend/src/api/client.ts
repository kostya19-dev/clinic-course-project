import {
  BookAppointmentRequest,
  CancelAppointmentRequest,
  CompleteVisitRequest,
  CompleteVisitResponse,
  AuthenticatedUser,
  CreatePatientRequest,
  IDResponse,
  LoginRequest,
  PriceListItem,
  RegisterPatientAccountRequest,
  RegisterPaymentRequest,
  ScheduleView,
} from "../types";

async function readApiError(response: Response, fallbackMessage: string) {
  try {
    const payload = (await response.json()) as { error?: string };
    if (payload.error) {
      return payload.error;
    }
  } catch {
    return fallbackMessage;
  }

  return fallbackMessage;
}

export class ApiClient {
  constructor(private readonly baseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api/v1") {}

  async login(request: LoginRequest): Promise<AuthenticatedUser> {
    const response = await fetch(`${this.baseUrl}/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось выполнить вход"));
    }

    return response.json();
  }

  async registerPatientAccount(request: RegisterPatientAccountRequest): Promise<AuthenticatedUser> {
    const response = await fetch(`${this.baseUrl}/auth/patient/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось создать кабинет пациента"));
    }

    return response.json();
  }

  async getPriceList(): Promise<PriceListItem[]> {
    const response = await fetch(`${this.baseUrl}/services/pricelist`);
    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось загрузить прайс-лист"));
    }

    return response.json();
  }

  async getDoctorSchedule(doctorId: string, from: string, to: string): Promise<ScheduleView> {
    const params = new URLSearchParams({ from, to });
    const response = await fetch(`${this.baseUrl}/doctors/${doctorId}/schedule?${params.toString()}`);
    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось загрузить расписание врача"));
    }

    return response.json();
  }

  async createPatient(request: CreatePatientRequest): Promise<IDResponse> {
    const response = await fetch(`${this.baseUrl}/patients`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось зарегистрировать пациента"));
    }

    return response.json();
  }

  async bookAppointment(request: BookAppointmentRequest): Promise<IDResponse> {
    const response = await fetch(`${this.baseUrl}/appointments`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось выполнить запись на прием"));
    }

    return response.json();
  }

  async cancelAppointment(appointmentId: string, request: CancelAppointmentRequest): Promise<void> {
    const response = await fetch(`${this.baseUrl}/appointments/${appointmentId}/cancel`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось отменить запись"));
    }
  }

  async completeVisit(appointmentId: string, request: CompleteVisitRequest): Promise<CompleteVisitResponse> {
    const response = await fetch(`${this.baseUrl}/appointments/${appointmentId}/complete`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось завершить прием"));
    }

    return response.json();
  }

  async registerPayment(invoiceId: string, request: RegisterPaymentRequest): Promise<IDResponse> {
    const response = await fetch(`${this.baseUrl}/invoices/${invoiceId}/payments`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(await readApiError(response, "Не удалось зарегистрировать оплату"));
    }

    return response.json();
  }
}

export const apiClient = new ApiClient();
