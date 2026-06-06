package postgres

import (
    "context"
    "database/sql"

    "clinic.local/backend/internal/model"
)

type CatalogRepository struct {
    db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository {
    return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListCurrentPriceList(ctx context.Context) ([]model.PriceListItem, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT service_id, code, name, duration_minutes, price, currency_code, effective_from
        FROM catalog.current_price_list
        ORDER BY name
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    items := make([]model.PriceListItem, 0)
    for rows.Next() {
        var item model.PriceListItem
        if err := rows.Scan(
            &item.ServiceID,
            &item.Code,
            &item.Name,
            &item.DurationMinutes,
            &item.Price,
            &item.CurrencyCode,
            &item.EffectiveFrom,
        ); err != nil {
            return nil, err
        }

        items = append(items, item)
    }

    return items, rows.Err()
}
