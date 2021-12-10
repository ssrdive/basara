package mysql

import (
	"database/sql"
	"fmt"

	"github.com/ssrdive/basara/pkg/models"
)

// ModelModel struct holds methods to query user table
type DropdownModel struct {
	DB *sql.DB
}

func (m *DropdownModel) Get(name string) ([]*models.Dropdown, error) {
	stmt := fmt.Sprintf(`SELECT id, name FROM %s ORDER BY name ASC`, name)

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []*models.Dropdown{}
	for rows.Next() {
		i := &models.Dropdown{}

		err = rows.Scan(&i.ID, &i.Name)
		if err != nil {
			return nil, err
		}

		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *DropdownModel) ConditionGet(name, where, value string) ([]*models.Dropdown, error) {
	stmt := fmt.Sprintf(`SELECT id, name FROM %s WHERE %s = %s ORDER BY name ASC`, name, where, value)

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []*models.Dropdown{}
	for rows.Next() {
		i := &models.Dropdown{}

		err = rows.Scan(&i.ID, &i.Name)
		if err != nil {
			return nil, err
		}

		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *DropdownModel) ConditionAccountsGet(name, where, value string) ([]*models.DropdownAccount, error) {
	stmt := fmt.Sprintf(`SELECT id, account_id, name FROM %s WHERE %s = %s ORDER BY name ASC`, name, where, value)

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []*models.DropdownAccount{}
	for rows.Next() {
		i := &models.DropdownAccount{}

		err = rows.Scan(&i.ID, &i.AccountID, &i.Name)
		if err != nil {
			return nil, err
		}

		items = append(items, i)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *DropdownModel) GetGrn() ([]*models.Dropdown, error) {
	stmt :=`SELECT GRN.id,  concat(GRN.id, ' - ', BP.name) as name FROM goods_received_note GRN LEFT JOIN business_partner BP ON BP.id = GRN.supplier_id WHERE landed_cost_id is null ORDER BY GRN.id ASC`;
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []*models.Dropdown{}
	
	for rows.Next() {
		i := &models.Dropdown{}

		err = rows.Scan(&i.ID, &i.Name)
		if err != nil {
			return nil, err
		}

		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
