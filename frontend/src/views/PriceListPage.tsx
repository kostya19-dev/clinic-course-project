import { EmptyState, LoadingState, PageTitle, Panel, StatusBadge } from "../components/ui";
import { PriceListItem } from "../types";
import { formatMoney } from "../utils/format";
import { usePriceListViewModel } from "../viewmodels/usePriceListViewModel";

const serviceNames: Record<string, string> = {
  CONSULT_CARDIO: "Консультация кардиолога",
  ECG_STANDARD: "Стандартная ЭКГ",
  research_service: "Исследовательская услуга",
  "Cardiologist consultation": "Консультация кардиолога",
  "Standard ECG": "Стандартная ЭКГ",
  "Research Service": "Исследовательская услуга",
};

function getServiceName(item: PriceListItem) {
  if (serviceNames[item.code] || serviceNames[item.name]) {
    return serviceNames[item.code] ?? serviceNames[item.name];
  }

  const hasCyrillic = /[А-Яа-яЁё]/.test(item.name);

  return hasCyrillic ? item.name : "Медицинская услуга";
}

export function PriceListPage() {
  const { items, isLoading, error } = usePriceListViewModel();

  return (
    <Panel>
      <PageTitle
        eyebrow="Медицинские услуги"
        title="Прайс-лист клиники"
        description="Актуальные услуги, длительность приема и стоимость для пациента."
        action={<StatusBadge tone="blue">{items.length} услуг</StatusBadge>}
      />

      {isLoading && <LoadingState text="Загружаем список услуг..." />}
      {error && <p className="error-box">Не удалось загрузить прайс-лист. {error}</p>}
      {!isLoading && !error && items.length === 0 && (
        <EmptyState
          title="Прайс-лист пока пуст"
          description="После добавления услуг они появятся в этом разделе."
        />
      )}
      {!isLoading && !error && items.length > 0 && (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Услуга</th>
                <th>Длительность</th>
                <th>Стоимость</th>
                <th>Действует с</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.serviceId}>
                  <td>
                    <strong>{getServiceName(item)}</strong>
                  </td>
                  <td>{item.durationMinutes} мин</td>
                  <td>{formatMoney(item.price, item.currencyCode)}</td>
                  <td>{new Intl.DateTimeFormat("ru-RU").format(new Date(item.effectiveFrom))}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  );
}
