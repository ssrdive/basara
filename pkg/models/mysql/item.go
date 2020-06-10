package mysql

import (
	"database/sql"
	"net/url"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

// ModelModel struct holds methods to query user table
type ItemModel struct {
	DB *sql.DB
}

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

func (m *ItemModel) All() ([]models.AllItemItem, error) {
	var res []models.AllItemItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.ALL_ITEMS)
	if err != nil {
		return nil, err
	}

	return res, nil
}
