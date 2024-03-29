package mysql

import (
	"database/sql"
	"errors"
	"net/url"
	"strconv"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

// ItemModel struct holds methods to query item table
type ItemModel struct {
	DB *sql.DB
}

func (m *ItemModel) UpdateById(form url.Values) (int64, error) {
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

	var currentPrice float64
	err = tx.QueryRow("SELECT price FROM item WHERE id = ?", form.Get("item_id")).Scan(&currentPrice)

	if err != nil {
		tx.Rollback()
		return 0, err
	}
	updatedPrice, _ := strconv.ParseFloat(form.Get("item_price"), 64)

	if currentPrice > updatedPrice {
		return 0, errors.New("price cannot be lower than the current price")
	}

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
func (m *ItemModel) Create(rparams, oparams []string, form url.Values) (int64, error) {
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
		TableName: "item",
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
func (m *ItemModel) All() ([]models.AllItemItem, error) {
	var res []models.AllItemItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.AllItems)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DetailsById returns all items
func (m *ItemModel) DetailsById(id string) (models.ItemDetails, error) {
	var itemDetails models.ItemDetails
	err := m.DB.QueryRow(queries.ItemDetailsById, id).Scan(&itemDetails.ID, &itemDetails.ItemID, &itemDetails.ModelID, &itemDetails.ModelName, &itemDetails.ItemCategoryID, &itemDetails.ItemCategoryName, &itemDetails.PageNo, &itemDetails.ItemNo, &itemDetails.ForeignID, &itemDetails.ItemName, &itemDetails.Price)
	if err != nil {
		return models.ItemDetails{}, err
	}

	return itemDetails, nil
}

// Stock returns stock for given item
func (m *ItemModel) Stock(id string) ([]models.CurrentStock, error) {
	var res []models.CurrentStock
	err := mysequel.QueryToStructs(&res, m.DB, queries.ItemStock, id)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// All returns all items
func (m *ItemModel) Details(id string) (models.ItemDetails, error) {
	var itemDetails models.ItemDetails
	err := m.DB.QueryRow(queries.ItemDetailsByItemId, id).Scan(&itemDetails.ID, &itemDetails.ItemID, &itemDetails.ModelID, &itemDetails.ModelName, &itemDetails.ItemCategoryID, &itemDetails.ItemCategoryName, &itemDetails.PageNo, &itemDetails.ItemNo, &itemDetails.ForeignID, &itemDetails.ItemName, &itemDetails.Price)
	if err != nil {
		return models.ItemDetails{}, err
	}

	return itemDetails, nil
}

// Search returns search results
func (m *ItemModel) Search(search string) ([]models.AllItemItem, error) {
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
