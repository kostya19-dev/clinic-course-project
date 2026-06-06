import { useState } from "react";
import { apiClient } from "../api/client";
import { CreatePatientRequest } from "../types";

export function usePatientRegistrationViewModel(onPatientCreated?: (patientId: string) => void) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdPatientId, setCreatedPatientId] = useState<string | null>(null);

  async function submit(request: CreatePatientRequest) {
    setIsSubmitting(true);
    setError(null);

    try {
      const response = await apiClient.createPatient(request);
      setCreatedPatientId(response.id);
      onPatientCreated?.(response.id);
    } catch (reason) {
      if (reason instanceof Error) {
        setError(reason.message);
      } else {
        setError("Не удалось зарегистрировать пациента");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return {
    isSubmitting,
    error,
    createdPatientId,
    submit,
  };
}