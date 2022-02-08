package mysql

import (
	"database/sql"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

type WarehouseStock struct {
	DB *sql.DB
}

func (m *WarehouseStock) GetWarehouseStock(wid int) ([]models.WarehouseStockItem, error) {
	var res []models.WarehouseStockItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.WAREHOUSE_STOCK, wid)
	if err != nil {
		return nil, err
	}

	return res, nil
}
