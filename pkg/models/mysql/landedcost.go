package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
	"github.com/ssrdive/scribe"
	smodels "github.com/ssrdive/scribe/models"
)

const (
	StockAccountID = 183
)

// LandedCostModel struct holds database instance
type LandedCostModel struct {
	DB *sql.DB
}

// CreateLandedCost creates a Landed Cost
func (m *LandedCostModel) CreateLandedCost(rparams []string, form url.Values) (int64, error) {
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

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "remark"},
		Vals:      []interface{}{form.Get("user_id"), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02"), fmt.Sprintf("GOODS RECEIVED NOTE %s", form.Get("grn_id"))},
		Tx:        tx,
	})
	if err != nil {
		return 0, err
	}

	var journalEntries []smodels.JournalEntry
	for _, entry := range landedCostTypes {
		costAmount, err := strconv.ParseFloat(entry.Amount, 32)
		if err != nil {
			return 0, err
		}
		totalLandedCost = totalLandedCost + costAmount

		if costAmount == 0 {
			continue
		}

		_, err = mysequel.Insert(mysequel.Table{
			TableName: "landed_cost_item",
			Columns:   []string{"landed_cost_id", "landed_cost_type_id", "amount"},
			Vals:      []interface{}{lcid, entry.CostTypeID, entry.Amount},
			Tx:        tx,
		})
		if err != nil {
			return 0, err
		}

		var expenseAccountID sql.NullInt32
		var payableAccountID sql.NullInt32
		err = tx.QueryRow("SELECT expense_account_id, payable_account_id FROM landed_cost_type WHERE id = ?", entry.CostTypeID).Scan(&expenseAccountID, &payableAccountID)
		if err != nil {
			return 0, err
		}

		if !expenseAccountID.Valid || !payableAccountID.Valid {
			err = errors.New("expense account or payable account for landed cost item is not configured")
			return 0, err
		}

		journalEntries = append(journalEntries,
			smodels.JournalEntry{Account: fmt.Sprintf("%d", expenseAccountID.Int32), Debit: entry.Amount, Credit: ""},
			smodels.JournalEntry{Account: fmt.Sprintf("%d", payableAccountID.Int32), Debit: "", Credit: entry.Amount},
		)
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

	grnCostPrice := grnItems[0].TotalPrice
	var supplierAccountID sql.NullInt32
	err = tx.QueryRow("SELECT account_id FROM business_partner WHERE id = (SELECT supplier_id FROM goods_received_note WHERE id = ?)", form.Get("grn_id")).Scan(&supplierAccountID)
	if err != nil {
		return 0, err
	}

	if !supplierAccountID.Valid {
		err = errors.New("account id not specified for supplier")
		return 0, err
	}

	journalEntries = append(journalEntries,
		smodels.JournalEntry{Account: fmt.Sprintf("%d", StockAccountID), Debit: fmt.Sprintf("%f", grnCostPrice), Credit: ""},
		smodels.JournalEntry{Account: fmt.Sprintf("%d", supplierAccountID.Int32), Debit: "", Credit: fmt.Sprintf("%f", grnCostPrice)},
	)
	err = scribe.IssueJournalEntries(tx, tid, journalEntries)
	if err != nil {
		return 0, err
	}

	return lcid, nil
}
