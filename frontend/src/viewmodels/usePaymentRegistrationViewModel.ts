import { useState } from "react";
import { apiClient } from "../api/client";
import { RegisterPaymentRequest } from "../types";

export function usePaymentRegistrationViewModel() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [paymentId, setPaymentId] = useState<string | null>(null);

  async function submit(invoiceId: string, request: RegisterPaymentRequest) {
    setIsSubmitting(true);
    setError(null);

    try {
      const response = await apiClient.registerPayment(invoiceId, request);
      setPaymentId(response.id);
      return response.id;
    } catch (reason) {
      if (reason instanceof Error) {
        setError(reason.message);
      } else {
        setError("Не удалось зарегистрировать оплату");
      }
      return null;
    } finally {
      setIsSubmitting(false);
    }
  }

  return {
    isSubmitting,
    error,
    paymentId,
    submit,
  };
}