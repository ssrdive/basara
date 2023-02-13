package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

// BusinessPartnerModel struct holds methods to query item table
type BusinessPartnerModel struct {
	DB *sql.DB
}

func (m *BusinessPartnerModel) UpdateById(form url.Values) (int64, error) {
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

	id, err := mysequel.Update(mysequel.UpdateTable{
		Table: mysequel.Table{
			TableName: "item",
			Columns:   []string{"name", "price"},
			Vals:      []interface{}{form.Get("name"), form.Get("item_price")},
			Tx:        tx,
		},
		WColumns: []string{"id"},
		WVals:    []string{form.Get("item_id")},
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Create creates an item
func (m *BusinessPartnerModel) Create(rparams, oparams []string, form url.Values) (int64, error) {
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

	id, err := mysequel.Insert(mysequel.FormTable{
		TableName: "business_partner",
		RCols:     rparams,
		OCols:     oparams,
		Form:      form,
		Tx:        tx,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// All returns all items
func (m *BusinessPartnerModel) All() ([]models.AllItemItem, error) {
	var res []models.AllItemItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.AllItems)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Return all business partner balances
func (m *BusinessPartnerModel) Balances() ([]models.BusinessPartnerBalance, error) {
	var res []models.BusinessPartnerBalance
	err := mysequel.QueryToStructs(&res, m.DB, queries.BusinessPartnerBalances)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DetailsById returns all items
func (m *BusinessPartnerModel) DetailsById(id string) (models.ItemDetails, error) {
	var itemDetails models.ItemDetails
	err := m.DB.QueryRow(queries.ItemDetailsById, id).Scan(&itemDetails.ID, &itemDetails.ItemID, &itemDetails.ModelID, &itemDetails.ModelName, &itemDetails.ItemCategoryID, &itemDetails.ItemCategoryName, &itemDetails.PageNo, &itemDetails.ItemNo, &itemDetails.ForeignID, &itemDetails.ItemName, &itemDetails.Price)
	if err != nil {
		return models.ItemDetails{}, err
	}

	return itemDetails, nil
}

// All returns all items
func (m *BusinessPartnerModel) Details(id string) (models.ItemDetails, error) {
	var itemDetails models.ItemDetails
	err := m.DB.QueryRow(queries.ItemDetailsByItemId, id).Scan(&itemDetails.ID, &itemDetails.ItemID, &itemDetails.ModelID, &itemDetails.ModelName, &itemDetails.ItemCategoryID, &itemDetails.ItemCategoryName, &itemDetails.PageNo, &itemDetails.ItemNo, &itemDetails.ForeignID, &itemDetails.ItemName, &itemDetails.Price)
	if err != nil {
		return models.ItemDetails{}, err
	}

	return itemDetails, nil
}

// Search returns search results
func (m *BusinessPartnerModel) Search(search string) ([]models.AllItemItem, error) {
	var k sql.NullString
	if search == "" {
		k = sql.NullString{}
	} else {
		k = sql.NullString{
			Valid:  true,
			String: "%" + search + "%",
		}
	}

	var res []models.AllItemItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.SearchItems, k, k)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *BusinessPartnerModel) BPBalanceDetail(bpID int) ([]models.BusinessPartnerBalanceDetail, error) {
	var res []models.BusinessPartnerBalanceDetail
	err := mysequel.QueryToStructs(&res, m.DB, queries.BusinessPartnerBalanceDetail, bpID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *BusinessPartnerModel) Payment(userID, postingDate, fromAccountID, amount, entries, remark, effectiveDate, checkNumber string) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	err = validatePostingDate(postingDate)
	if err != nil {
		return 0, err
	}

	var bpPayments []models.BPPaymentEntry
	_ = json.Unmarshal([]byte(entries), &bpPayments)

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "remark"},
		Vals:      []interface{}{userID, time.Now().Format("2006-01-02 15:04:05"), postingDate, remark},
		Tx:        tx,
	})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	_, err = mysequel.Insert(mysequel.Table{
		TableName: "account_transaction",
		Columns:   []string{"transaction_id", "account_id", "type", "amount"},
		Vals:      []interface{}{tid, fromAccountID, "CR", amount},
		Tx:        tx,
	})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	_, err = mysequel.Insert(mysequel.Table{
		TableName: "account_transaction",
		Columns:   []string{"transaction_id", "account_id", "type", "amount"},
		Vals:      []interface{}{tid, PayableAccountID, "DR", amount},
		Tx:        tx,
	})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	for _, bpPayment := range bpPayments {
		_, err = mysequel.Insert(mysequel.Table{
			TableName: "business_partner_financial",
			Columns:   []string{"effective_date", "business_partner_id", "type", "amount", "transaction_id"},
			Vals:      []interface{}{effectiveDate, bpPayment.BP, "DR", bpPayment.Amount, tid},
			Tx:        tx,
		})
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	return tid, nil
}

func validatePostingDate(postingDate string) error {
	now := time.Now()

	var m time.Month
	year, m, _ := now.Date()
	month := int(m)

	var oldestDate time.Time
	if month < 4 {
		oldestDate = time.Date(year-1, 4, 1, 0, 0, 0, 0, time.UTC)
	} else {
		oldestDate = time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC)
	}

	parsedDate, err := time.Parse("2006-01-02", postingDate)
	if err != nil {
		return errors.New("invalid posting date")
	}

	if parsedDate.Before(oldestDate) {
		return errors.New("posting date does not fall within the financial year")
	} else {
		return nil
	}
}
