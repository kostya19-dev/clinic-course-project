import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import { ScheduleView } from "../types";

export function useDoctorScheduleViewModel(doctorId: string, from: string, to: string) {
  const [schedule, setSchedule] = useState<ScheduleView | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!doctorId) {
      setSchedule(null);
      setIsLoading(false);
      setError(null);
      return;
    }

    let isMounted = true;
    setIsLoading(true);
    setError(null);

    apiClient
      .getDoctorSchedule(doctorId, from, to)
      .then((result) => {
        if (isMounted) {
          setSchedule(result);
        }
      })
      .catch((reason: Error) => {
        if (isMounted) {
          setError(reason.message);
        }
      })
      .finally(() => {
        if (isMounted) {
          setIsLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, [doctorId, from, to]);

  return { schedule, isLoading, error };
}
