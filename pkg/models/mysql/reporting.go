package mysql

import (
	"database/sql"
	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

// ReportingModel struct holds database instance
type ReportingModel struct {
	DB *sql.DB
}

// ReceiptSearch returns receipt search
func (m *ReportingModel) InvoiceSearch(startDate, endDate, officer string) ([]models.InvoiceSearchItem, error) {
	o := mysequel.NewNullString(officer)

	var res []models.InvoiceSearchItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.INVOICE_SEARCH, o, o, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return res, nil
}
