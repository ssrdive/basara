package mysql

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

// LandedCostModel struct holds database instance
type LandedCostModel struct {
	DB *sql.DB
}

// CreatelandedCost creats a Landed Cost
func (m *LandedCostModel) CreatelandedCost(rparams []string, form url.Values) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	lcid, err := mysequel.Insert(mysequel.Table{
		TableName: "landed_cost",
		Columns:   []string{"user_id", "goods_received_note_id"},
		Vals:      []interface{}{form.Get("user_id"), form.Get("grn_id")},
		Tx:        tx,
	})

	if err != nil {
		return 0, err
	}

	_, err = mysequel.Update(mysequel.UpdateTable{
		Table: mysequel.Table{
			TableName: "goods_received_note",
			Columns:   []string{"landed_cost_id"},
			Vals:      []interface{}{lcid},
			Tx:        tx,
		},
		WColumns: []string{"id"},
		WVals:    []string{form.Get("grn_id")},
	})

	if err != nil {
		return 0, err
	}

	entries := form.Get("entries")
	var landedCostTypes []models.LandedCostItemEntry
	json.Unmarshal([]byte(entries), &landedCostTypes)

	var totalLandedCost = 0.0

	for _, entry := range landedCostTypes {
		_, err = mysequel.Insert(mysequel.Table{
			TableName: "landed_cost_item",
			Columns:   []string{"landed_cost_id", "landed_cost_type_id", "amount"},
			Vals:      []interface{}{lcid, entry.CostTypeID, entry.Amount},
			Tx:        tx,
		})

		if err != nil {
			return 0, err
		}

		costAmount, err := strconv.ParseFloat(entry.Amount, 32)
		if err != nil {
			return 0, err
		}

		totalLandedCost = totalLandedCost + costAmount

		if err != nil {
			return 0, err
		}
	}

	var grnItems []models.GRNItemDetailsWithTotal
	err = mysequel.QueryToStructs(&grnItems, m.DB, queries.GRN_ITEM_DETAILS_WITH_ORDER_TOTAL, form.Get("grn_id"))
	if err != nil {
		return 0, err
	}

	for _, entry := range grnItems {

		landedCost := (entry.ToatlCostPrice * totalLandedCost) / (entry.TotalPrice * entry.Quantity)
		unitCost := (entry.ToatlCostPrice / entry.Quantity)

		_, err = mysequel.Insert(mysequel.Table{
			TableName: "current_stock",
			Columns:   []string{"warehouse_id", "item_id", "goods_received_note_id", "cost_price", "landed_costs", "qty", "float_qty", "price"},
			Vals:      []interface{}{entry.WarehouseId, entry.ItemID, form.Get("grn_id"), unitCost, landedCost, entry.Quantity, 0, (unitCost + landedCost)},
			Tx:        tx,
		})

		if err != nil {
			return 0, err
		}
	}

	return lcid, nil
}
