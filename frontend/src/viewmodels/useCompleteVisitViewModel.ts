import { useState } from "react";
import { apiClient } from "../api/client";
import { CompleteVisitRequest, CompleteVisitResponse } from "../types";

export function useCompleteVisitViewModel() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<CompleteVisitResponse | null>(null);

  async function submit(appointmentId: string, request: CompleteVisitRequest) {
    setIsSubmitting(true);
    setError(null);

    try {
      const response = await apiClient.completeVisit(appointmentId, request);
      setResult(response);
      return response;
    } catch (reason) {
      if (reason instanceof Error) {
        setError(reason.message);
      } else {
        setError("Не удалось завершить прием");
      }
      return null;
    } finally {
      setIsSubmitting(false);
    }
  }

  return {
    isSubmitting,
    error,
    result,
    submit,
  };
}