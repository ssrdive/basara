package mysql

import (
	"database/sql"
	"net/url"

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
			Columns:   []string{"item_name", "price"},
			Vals:      []interface{}{form.Get("item_name"), form.Get("item_price")},
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
	err := mysequel.QueryToStructs(&res, m.DB, queries.ALL_ITEMS)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DetailsById returns all items
func (m *BusinessPartnerModel) DetailsById(id string) (models.ItemDetails, error) {
	var itemDetails models.ItemDetails
	err := m.DB.QueryRow(queries.ITEM_DETAILS_BY_ID, id).Scan(&itemDetails.ID, &itemDetails.ItemID, &itemDetails.ModelID, &itemDetails.ModelName, &itemDetails.ItemCategoryID, &itemDetails.ItemCategoryName, &itemDetails.PageNo, &itemDetails.ItemNo, &itemDetails.ForeignID, &itemDetails.ItemName, &itemDetails.Price)
	if err != nil {
		return models.ItemDetails{}, err
	}

	return itemDetails, nil
}

// All returns all items
func (m *BusinessPartnerModel) Details(id string) (models.ItemDetails, error) {
	var itemDetails models.ItemDetails
	err := m.DB.QueryRow(queries.ITEM_DETAILS_BY_ITEM_ID, id).Scan(&itemDetails.ID, &itemDetails.ItemID, &itemDetails.ModelID, &itemDetails.ModelName, &itemDetails.ItemCategoryID, &itemDetails.ItemCategoryName, &itemDetails.PageNo, &itemDetails.ItemNo, &itemDetails.ForeignID, &itemDetails.ItemName, &itemDetails.Price)
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
	err := mysequel.QueryToStructs(&res, m.DB, queries.SEARCH_ITEMS, k, k)
	if err != nil {
		return nil, err
	}

	return res, nil
}
