import { useEffect, useState } from "react";
import { EmptyState, InfoCard, PageTitle, Panel, StatusBadge } from "../components/ui";
import { demoClinicContext } from "../demoData";
import { PaymentMethod } from "../types";
import { formatMoney, recordNumber } from "../utils/format";
import { usePaymentRegistrationViewModel } from "../viewmodels/usePaymentRegistrationViewModel";

type Props = {
  invoiceId: string;
  receivedByUserId: string;
};

const paymentMethodLabels: Record<PaymentMethod, string> = {
  card: "Банковская карта",
  cash: "Наличные",
  transfer: "Безналичный перевод",
};

export function PaymentRegistrationPage({ invoiceId, receivedByUserId }: Props) {
  const [form, setForm] = useState({
    invoiceId,
    receivedByUserId,
    amount: demoClinicContext.services[0].price,
    paymentMethod: "card" as PaymentMethod,
    externalReference: "Оплата в кассе клиники",
  });
  const { isSubmitting, error, paymentId, submit } = usePaymentRegistrationViewModel();
  const hasInvoice = Boolean(form.invoiceId);

  useEffect(() => {
    setForm((current) => ({ ...current, invoiceId }));
  }, [invoiceId]);

  useEffect(() => {
    setForm((current) => ({ ...current, receivedByUserId }));
  }, [receivedByUserId]);

  return (
    <Panel>
      <PageTitle
        eyebrow="Касса"
        title="Регистрация оплаты"
        description="Примите оплату по счету, сформированному после приема врача."
        action={<StatusBadge tone={hasInvoice ? "amber" : "slate"}>{hasInvoice ? recordNumber(form.invoiceId, "Счет") : "Счет не выбран"}</StatusBadge>}
      />

      {!hasInvoice && (
        <EmptyState
          title="Нет счета для оплаты"
          description="Сначала завершите прием врача, после этого счет появится в кассе."
        />
      )}

      <div className="payment-layout">
        <form
          className="form-card form-grid"
          onSubmit={async (event) => {
            event.preventDefault();
            await submit(form.invoiceId, {
              receivedByUserId: form.receivedByUserId,
              amount: form.amount,
              paymentMethod: form.paymentMethod,
              externalReference: form.externalReference,
            });
          }}
        >
          <label>
            <span>Сумма к оплате</span>
            <input
              type="number"
              step="0.01"
              min={0.01}
              value={form.amount}
              onChange={(event) => setForm((current) => ({ ...current, amount: Number(event.target.value) || 0 }))}
              disabled={!hasInvoice}
              required
            />
          </label>
          <label>
            <span>Способ оплаты</span>
            <select
              value={form.paymentMethod}
              onChange={(event) =>
                setForm((current) => ({ ...current, paymentMethod: event.target.value as PaymentMethod }))
              }
              disabled={!hasInvoice}
            >
              {Object.entries(paymentMethodLabels).map(([method, label]) => (
                <option key={method} value={method}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <label className="full-width">
            <span>Комментарий к платежу</span>
            <input
              value={form.externalReference}
              onChange={(event) => setForm((current) => ({ ...current, externalReference: event.target.value }))}
              disabled={!hasInvoice}
            />
          </label>
          <div className="form-actions full-width">
            <button type="submit" disabled={isSubmitting || !hasInvoice}>
              {isSubmitting ? "Регистрируем оплату..." : "Зарегистрировать оплату"}
            </button>
          </div>
        </form>

        <aside className="side-stack">
          <InfoCard
            title={hasInvoice ? recordNumber(form.invoiceId, "Счет") : "Счет не выбран"}
            subtitle={hasInvoice ? `К оплате ${formatMoney(form.amount)}` : "Ожидается завершение приема"}
            meta={<StatusBadge tone={paymentId ? "green" : hasInvoice ? "amber" : "slate"}>{paymentId ? "Оплачено" : hasInvoice ? "Оплата ожидается" : "Нет счета"}</StatusBadge>}
          >
            <p className="card-text">
              После регистрации оплаты счет считается закрытым, а информация сохраняется в истории пациента.
            </p>
          </InfoCard>
        </aside>
      </div>

      {error && <p className="error-box">{error}</p>}
      {paymentId && <p className="success-box">Оплата зарегистрирована: {recordNumber(paymentId, "Платеж")}.</p>}
    </Panel>
  );
}
