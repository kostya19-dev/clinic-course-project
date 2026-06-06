import { useState } from "react";
import { apiClient } from "../api/client";
import { BookAppointmentRequest, CancelAppointmentRequest } from "../types";

export function useAppointmentWorkflowViewModel() {
  const [isBooking, setIsBooking] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [bookError, setBookError] = useState<string | null>(null);
  const [cancelError, setCancelError] = useState<string | null>(null);
  const [lastAppointmentId, setLastAppointmentId] = useState<string | null>(null);
  const [cancelSuccessMessage, setCancelSuccessMessage] = useState<string | null>(null);

  async function book(request: BookAppointmentRequest) {
    setIsBooking(true);
    setBookError(null);
    setCancelSuccessMessage(null);

    try {
      const response = await apiClient.bookAppointment(request);
      setLastAppointmentId(response.id);
      return response.id;
    } catch (reason) {
      if (reason instanceof Error) {
        setBookError(reason.message);
      } else {
        setBookError("Не удалось выполнить запись на прием");
      }
      return null;
    } finally {
      setIsBooking(false);
    }
  }

  async function cancel(appointmentId: string, request: CancelAppointmentRequest) {
    setIsCancelling(true);
    setCancelError(null);
    setCancelSuccessMessage(null);

    try {
      await apiClient.cancelAppointment(appointmentId, request);
      setCancelSuccessMessage("Запись успешно отменена");
      return true;
    } catch (reason) {
      if (reason instanceof Error) {
        setCancelError(reason.message);
      } else {
        setCancelError("Не удалось отменить запись");
      }
      return false;
    } finally {
      setIsCancelling(false);
    }
  }

  return {
    isBooking,
    isCancelling,
    bookError,
    cancelError,
    lastAppointmentId,
    cancelSuccessMessage,
    book,
    cancel,
  };
}