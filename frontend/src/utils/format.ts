export function recordNumber(id: string | null | undefined, noun = "Запись") {
  if (!id) {
    return `${noun} не выбрана`;
  }

  const digits = id.replace(/\D/g, "").slice(-5);
  const suffix = digits || id.replace(/-/g, "").slice(-5).toUpperCase();

  return `${noun} №${suffix.padStart(4, "0")}`;
}

export function formatDate(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  }).format(date);
}

export function formatTime(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function formatDateTime(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function formatMoney(amount: number, currencyCode = "RUB") {
  return new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: currencyCode,
    maximumFractionDigits: Number.isInteger(amount) ? 0 : 2,
  }).format(amount);
}
