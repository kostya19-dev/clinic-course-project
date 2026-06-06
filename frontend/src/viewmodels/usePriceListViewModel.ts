import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import { PriceListItem } from "../types";

export function usePriceListViewModel() {
  const [items, setItems] = useState<PriceListItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    apiClient
      .getPriceList()
      .then((result) => {
        if (isMounted) {
          setItems(result);
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
  }, []);

  return { items, isLoading, error };
}
