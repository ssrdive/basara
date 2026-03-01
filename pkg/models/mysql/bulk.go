package mysql

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
)

// BulkModel struct holds methods to query bulk related operations
type BulkModel struct {
	DB *sql.DB
}

// BulkStockError represents a validation error on a specific row
type BulkStockError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

// VerifyBulkStock validates the uploaded CSV file for bulk stock updates
func (m *BulkModel) VerifyBulkStock(file multipart.File) ([]BulkStockError, error) {
	var validationErrors []BulkStockError

	// Load valid warehouse names
	warehouseRows, err := m.DB.Query(queries.GetWarehouseNames)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch warehouse names: %v", err)
	}
	defer warehouseRows.Close()

	validWarehouses := make(map[string]bool)
	for warehouseRows.Next() {
		var name string
		if err := warehouseRows.Scan(&name); err != nil {
			return nil, err
		}
		validWarehouses[name] = true
	}

	// Load valid item IDs
	itemRows, err := m.DB.Query(queries.GetItemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item IDs: %v", err)
	}
	defer itemRows.Close()

	validItemIDs := make(map[string]bool)
	for itemRows.Next() {
		var itemID string
		if err := itemRows.Scan(&itemID); err != nil {
			return nil, err
		}
		validItemIDs[itemID] = true
	}

	csvReader := csv.NewReader(file)
	// We do not trim any column as requested
	csvReader.TrimLeadingSpace = false

	// Read and Validate Headers
	headers, err := csvReader.Read()
	if err != nil {
		validationErrors = append(validationErrors, BulkStockError{Row: 1, Error: "Failed to read headers"})
		return validationErrors, nil
	}

	expectedHeaders := []string{"warehouse_name", "item_id", "quantity"}
	for i, header := range expectedHeaders {
		if i >= len(headers) || headers[i] != header {
			validationErrors = append(validationErrors, BulkStockError{
				Row:   1,
				Error: fmt.Sprintf("Invalid header at column %d. Expected '%s', got '%s'", i+1, header, headers[i]),
			})
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors, nil
	}

	rowNum := 1
	var expectedWarehouse string
	seenItemIDs := make(map[string]bool)

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNum++

		if err != nil {
			validationErrors = append(validationErrors, BulkStockError{Row: rowNum, Error: "Error parsing row"})
			continue
		}

		if len(row) < 3 {
			validationErrors = append(validationErrors, BulkStockError{Row: rowNum, Error: "Row has insufficient columns"})
			continue
		}

		warehouseName := row[0]
		itemID := row[1]
		quantityStr := strings.TrimSpace(row[2])

		// Multiple warehouse check
		if expectedWarehouse == "" {
			expectedWarehouse = warehouseName
		} else if warehouseName != expectedWarehouse {
			validationErrors = append(validationErrors, BulkStockError{
				Row:   rowNum,
				Error: fmt.Sprintf("Multiple warehouses found in CSV. Expected '%s', got '%s'", expectedWarehouse, warehouseName),
			})
		}

		// Validate warehouse exists
		if !validWarehouses[warehouseName] {
			validationErrors = append(validationErrors, BulkStockError{
				Row:   rowNum,
				Error: fmt.Sprintf("Invalid warehouse_name '%s'", warehouseName),
			})
		}

		// Validate item ID exists
		if !validItemIDs[itemID] {
			validationErrors = append(validationErrors, BulkStockError{
				Row:   rowNum,
				Error: fmt.Sprintf("Invalid item_id '%s'", itemID),
			})
		}

		// Duplicate item ID check
		if seenItemIDs[itemID] {
			validationErrors = append(validationErrors, BulkStockError{
				Row:   rowNum,
				Error: fmt.Sprintf("Duplicate item_id '%s' found", itemID),
			})
		}
		seenItemIDs[itemID] = true

		// Validate quantity
		if qty, err := strconv.Atoi(quantityStr); err != nil || qty < 0 {
			validationErrors = append(validationErrors, BulkStockError{
				Row:   rowNum,
				Error: fmt.Sprintf("Invalid quantity. Expected integer >= 0, got '%s'", quantityStr),
			})
		}
	}

	return validationErrors, nil
}

// ComparisonResult holds the difference between CSV and DB stock
type ComparisonResult struct {
	WarehouseName   string  `json:"warehouse_name"`
	ItemID          string  `json:"item_id"`
	ItemName        string  `json:"item_name"`
	CostPrice       float64 `json:"cost_price"`
	CurrentQuantity int     `json:"current_quantity"`
	NewQuantity     int     `json:"new_quantity"`
	Difference      int     `json:"difference"`
	Price           float64 `json:"price"`
}

// CompareBulkStock compares the validated uploaded CSV file against current DB stock
func (m *BulkModel) CompareBulkStock(file multipart.File) ([]ComparisonResult, []BulkStockError, error) {
	// 1. Run validation first to be safe
	validationErrors, err := m.VerifyBulkStock(file)
	if err != nil {
		return nil, nil, err
	}
	if len(validationErrors) > 0 {
		return nil, validationErrors, nil
	}

	// Rewind file to read again for comparison
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, fmt.Errorf("could not rewind file: %v", err)
	}

	csvReader := csv.NewReader(file)
	csvReader.TrimLeadingSpace = false

	// Skip headers
	headers, err := csvReader.Read()
	if err != nil || len(headers) < 3 {
		return nil, nil, fmt.Errorf("failed to read headers on comparison pass: %v", err)
	}

	// 2. Load Item Prices and Names into memory
	itemRows, err := m.DB.Query(queries.GetAllItemsWithPrices)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch items for comparison: %v", err)
	}
	defer itemRows.Close()

	type itemData struct {
		Name  string
		Price float64
	}
	itemMap := make(map[string]itemData)
	for itemRows.Next() {
		var id, name string
		var price float64
		if err := itemRows.Scan(&id, &name, &price); err != nil {
			return nil, nil, err
		}
		itemMap[id] = itemData{Name: name, Price: price}
	}

	var comparisonResults []ComparisonResult
	var expectedWarehouse string
	var expectedWarehouseID int

	// 3. Process each row
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Should have been caught by VerifyBulkStock
		}
		if len(row) < 3 {
			continue
		}

		warehouseName := row[0]
		itemID := row[1]
		quantityStr := strings.TrimSpace(row[2])

		newQty, _ := strconv.Atoi(quantityStr) // Safe due to VerifyBulkStock

		// Retrieve Warehouse ID once
		if expectedWarehouse == "" {
			expectedWarehouse = warehouseName
			err := m.DB.QueryRow(queries.GetWarehouseIDByName, expectedWarehouse).Scan(&expectedWarehouseID)
			if err != nil {
				return nil, nil, fmt.Errorf("could not find warehouse id for '%s': %v", expectedWarehouse, err)
			}
		}

		// Retrieve current stock for this specific item in this warehouse
		var currentQty int
		var dbItemName string
		var dbPrice float64
		// Using the WarehouseStock query to get current quantity. It groups by item and warehouse.
		// Since we need it for a single item, we construct a quick direct query or reuse WarehouseItemQty.
		// Actually, let's just query current_stock directly for exact values to aggregate.

		err = m.DB.QueryRow(`SELECT SUM(qty) FROM current_stock WHERE warehouse_id = ? AND item_id = (SELECT id FROM item WHERE item_id = ?)`, expectedWarehouseID, itemID).Scan(&currentQty)
		if err != nil && err != sql.ErrNoRows {
			// If completely missing, quantity is 0, which is fine.
			// But if it's a real DB error, panic.
			// We'll ignore ErrNoRows.
		}

		// The sum query above returns NULL if no rows found. Scan(&currentQty) on NULL will panic if not a sql.NullInt64.
		// Let's use NullInt64 to be safe.
		var nullCurrentQty sql.NullInt64
		err = m.DB.QueryRow(`
            SELECT SUM(CS.qty) 
            FROM current_stock CS 
            JOIN item I ON I.id = CS.item_id 
            WHERE CS.warehouse_id = ? AND I.item_id = ?`,
			expectedWarehouseID, itemID).Scan(&nullCurrentQty)

		currentQty = 0
		if err == nil && nullCurrentQty.Valid {
			currentQty = int(nullCurrentQty.Int64)
		}

		// Get item details from our preloaded map
		iData := itemMap[itemID]
		dbItemName = iData.Name
		dbPrice = iData.Price
		costPrice := dbPrice * 0.70

		difference := newQty - currentQty

		comparisonResults = append(comparisonResults, ComparisonResult{
			WarehouseName:   expectedWarehouse,
			ItemID:          itemID,
			ItemName:        dbItemName,
			CostPrice:       costPrice,
			CurrentQuantity: currentQty,
			NewQuantity:     newQty,
			Difference:      difference,
			Price:           dbPrice,
		})
	}

	return comparisonResults, nil, nil
}

// SubmitBulkStock saves the validated CSV stock data as a pending report
func (m *BulkModel) SubmitBulkStock(userID int, file multipart.File) (int64, []BulkStockError, error) {
	// 1. Run validation first
	validationErrors, err := m.VerifyBulkStock(file)
	if err != nil {
		return 0, nil, err
	}
	if len(validationErrors) > 0 {
		return 0, validationErrors, nil
	}

	// Rewind file
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return 0, nil, fmt.Errorf("could not rewind file: %v", err)
	}

	csvReader := csv.NewReader(file)
	csvReader.TrimLeadingSpace = false

	// Skip headers
	headers, err := csvReader.Read()
	if err != nil || len(headers) < 3 {
		return 0, nil, fmt.Errorf("failed to read headers on submit pass: %v", err)
	}

	var expectedWarehouseID int
	var expectedWarehouse string

	// Read first data row to get warehouse name
	firstRow, err := csvReader.Read()
	if err == io.EOF {
		return 0, nil, fmt.Errorf("CSV has no data rows")
	}
	if err != nil {
		return 0, nil, err
	}
	expectedWarehouse = firstRow[0]

	err = m.DB.QueryRow(queries.GetWarehouseIDByName, expectedWarehouse).Scan(&expectedWarehouseID)
	if err != nil {
		return 0, nil, fmt.Errorf("could not find warehouse id for '%s': %v", expectedWarehouse, err)
	}

	// Start a transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback() // Safe to defer, will be ignored if tx is already committed

	// Insert main report record
	result, err := tx.Exec(`
		INSERT INTO bulk_stock_report (user_id, warehouse_id, status, created) 
		VALUES (?, ?, 'PENDING', NOW())`,
		userID, expectedWarehouseID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to insert bulk stock report: %v", err)
	}

	reportID, err := result.LastInsertId()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	// Process first row and write to items table
	itemID := firstRow[1]
	quantityStr := strings.TrimSpace(firstRow[2])
	newQty, _ := strconv.Atoi(quantityStr)

	_, err = tx.Exec(`
		INSERT INTO bulk_stock_report_item (report_id, item_id, new_quantity) 
		VALUES (?, ?, ?)`,
		reportID, itemID, newQty)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to insert first item: %v", err)
	}

	// Process remaining rows
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, nil, err
		}
		if len(row) < 3 {
			continue
		}

		itemID := row[1]
		quantityStr := strings.TrimSpace(row[2])
		newQty, _ := strconv.Atoi(quantityStr)

		_, err = tx.Exec(`
			INSERT INTO bulk_stock_report_item (report_id, item_id, new_quantity) 
			VALUES (?, ?, ?)`,
			reportID, itemID, newQty)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to insert item %s: %v", itemID, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return 0, nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return reportID, nil, nil
}

// BulkStockReport represents a summary of a submitted report
type BulkStockReport struct {
	ID             int            `json:"id"`
	Status         string         `json:"status"`
	Created        string         `json:"created"`
	ActionOn       sql.NullString `json:"action_on"`
	UserName       string         `json:"user_name"`
	WarehouseName  string         `json:"warehouse_name"`
	ActionUserName string         `json:"action_user_name"`
}

// GetPendingBulkStocks returns a list of all stock reports
func (m *BulkModel) GetPendingBulkStocks() ([]BulkStockReport, error) {
	rows, err := m.DB.Query(queries.GetPendingBulkStockReports)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []BulkStockReport
	for rows.Next() {
		var report BulkStockReport
		if err := rows.Scan(&report.ID, &report.Status, &report.Created, &report.ActionOn, &report.UserName, &report.WarehouseName, &report.ActionUserName); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reports, nil
}

// GetBulkStockDetails returns the comparison items for a specific report
func (m *BulkModel) GetBulkStockDetails(reportID int) ([]ComparisonResult, error) {
	// 1. Fetch report base info to get warehouse ID/Name
	var warehouseID int
	var warehouseName string
	err := m.DB.QueryRow(`
		SELECT BSR.warehouse_id, BP.name 
		FROM bulk_stock_report BSR 
		JOIN business_partner BP ON BSR.warehouse_id = BP.id 
		WHERE BSR.id = ?`, reportID).Scan(&warehouseID, &warehouseName)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch report base info: %v", err)
	}

	// 2. Fetch the items for this report
	rows, err := m.DB.Query(queries.GetBulkStockReportDetails, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ComparisonResult
	for rows.Next() {
		var res ComparisonResult
		res.WarehouseName = warehouseName

		if err := rows.Scan(&res.ItemID, &res.ItemName, &res.Price, &res.NewQuantity); err != nil {
			return nil, err
		}

		res.CostPrice = res.Price * 0.70

		// Retrieve current stock dynamically
		var nullCurrentQty sql.NullInt64
		err = m.DB.QueryRow(`
            SELECT SUM(CS.qty) 
            FROM current_stock CS 
            JOIN item I ON I.id = CS.item_id 
            WHERE CS.warehouse_id = ? AND I.item_id = ?`,
			warehouseID, res.ItemID).Scan(&nullCurrentQty)

		res.CurrentQuantity = 0
		if err == nil && nullCurrentQty.Valid {
			res.CurrentQuantity = int(nullCurrentQty.Int64)
		}

		res.Difference = res.NewQuantity - res.CurrentQuantity

		results = append(results, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// ProcessBulkStockReport approves or rejects a pending stock report
func (m *BulkModel) ProcessBulkStockReport(reportID int, status string, actionUserID int) error {
	if status == "REJECTED" {
		_, err := m.DB.Exec(`
			UPDATE bulk_stock_report 
			SET status = 'REJECTED', action_by = ?, action_on = NOW() 
			WHERE id = ?`, actionUserID, reportID)
		return err
	}

	if status != "APPROVED" {
		return fmt.Errorf("invalid status: %s", status)
	}

	// 1. Fetch Report Data
	var creatorID, warehouseID int
	err := m.DB.QueryRow(`
		SELECT user_id, warehouse_id 
		FROM bulk_stock_report 
		WHERE id = ? AND status = 'PENDING'`, reportID).Scan(&creatorID, &warehouseID)
	if err != nil {
		return fmt.Errorf("failed to fetch pending report or report already processed: %v", err)
	}

	if creatorID == actionUserID {
		return fmt.Errorf("you cannot approve your own stock report")
	}

	// 2. Fetch Supplier ID for Stock Correction
	var supplierID int
	err = m.DB.QueryRow(`
		SELECT id 
		FROM business_partner 
		WHERE name = 'Stock Re-Correction Supplier' AND business_partner_type_id = 3`).Scan(&supplierID)
	if err != nil {
		return fmt.Errorf("stock Re-Correction Supplier not found. Please create it first: %v", err)
	}

	// 3. Fetch Items
	rows, err := m.DB.Query(`
		SELECT I.id, I.item_id, I.price, BSRI.new_quantity 
		FROM bulk_stock_report_item BSRI
		JOIN item I ON BSRI.item_id = I.item_id
		WHERE BSRI.report_id = ?`, reportID)
	if err != nil {
		return fmt.Errorf("failed to fetch report items: %v", err)
	}
	defer rows.Close()

	var grnItems []models.GRNItemEntry
	var totalGRNPrice float64
	var itemIDs []string

	for rows.Next() {
		var itemPk int
		var itemID string
		var price float64
		var newQty int
		if err := rows.Scan(&itemPk, &itemID, &price, &newQty); err != nil {
			return err
		}

		// Calculate Cost Price (Selling Price * 0.70)
		costPrice := price * 0.70
		totalPrice := costPrice * float64(newQty)
		totalGRNPrice += totalPrice

		// Build GRN Entry
		grnItems = append(grnItems, models.GRNItemEntry{
			ItemID:    strconv.Itoa(itemPk),
			UnitPrice: fmt.Sprintf("%f", costPrice),
			Quantity:  strconv.Itoa(newQty),
		})

		itemIDs = append(itemIDs, itemID)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	if len(grnItems) == 0 {
		return fmt.Errorf("no items found in report")
	}

	// 4. Delete existing stock for these items in the specific warehouse
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Using a query with IN clause to delete multiple items
	placeholders := make([]string, len(itemIDs))
	args := make([]interface{}, len(itemIDs)+1)
	args[0] = warehouseID
	for i, id := range itemIDs {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		DELETE CS FROM current_stock CS
		JOIN item I ON CS.item_id = I.id
		WHERE CS.warehouse_id = ? AND I.item_id IN (%s)`, strings.Join(placeholders, ","))

	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to clear existing stock: %v", err)
	}

	// Fetch approver and creator names for the remark
	var approverName, creatorName string
	m.DB.QueryRow("SELECT name FROM user WHERE id = ?", actionUserID).Scan(&approverName)
	m.DB.QueryRow("SELECT name FROM user WHERE id = ?", creatorID).Scan(&creatorName)

	// Commit the stock deletion so GRN model can read true state if needed (though GRN only inserts)
	// It's safer to commit here, then call the independent models.
	// The independent models create their own transactions.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit stock clearance: %v", err)
	}

	// 5. Create GRN
	grnModel := &GoodsReceivedNoteModel{DB: m.DB}
	grnEntriesJSON, _ := json.Marshal(grnItems)

	formValues := url.Values{}
	formValues.Set("user_id", strconv.Itoa(actionUserID))
	formValues.Set("supplier_id", strconv.Itoa(supplierID))
	formValues.Set("warehouse_id", strconv.Itoa(warehouseID))
	formValues.Set("effective_date", time.Now().Format("2006-01-02 15:04:05"))
	formValues.Set("discount_type", "per")
	formValues.Set("discount_amount", "0")
	formValues.Set("total_price", fmt.Sprintf("%f", totalGRNPrice))
	formValues.Set("remark", fmt.Sprintf("BULK STOCK UPDATE SUBMITTED BY %s, APPROVED BY %s", creatorName, approverName))
	formValues.Set("entries", string(grnEntriesJSON))

	grnID, err := grnModel.CreateGoodsReceivedNote(nil, nil, formValues)
	if err != nil {
		return fmt.Errorf("failed to create GRN: %v", err)
	}

	// 6. Fetch "Import Tax" Landed Cost Type ID
	var landedCostTypeID int
	err = m.DB.QueryRow(`SELECT id FROM landed_cost_type WHERE name = 'Import Tax'`).Scan(&landedCostTypeID)
	if err != nil {
		return fmt.Errorf("landed cost type 'Import Tax' not found: %v", err)
	}

	// 7. Create Landed Cost
	landedCostModel := &LandedCostModel{DB: m.DB}

	lcItems := []models.LandedCostItemEntry{
		{
			CostTypeID: strconv.Itoa(landedCostTypeID),
			Amount:     "0",
		},
	}
	lcEntriesJSON, _ := json.Marshal(lcItems)

	lcFormValues := url.Values{}
	lcFormValues.Set("user_id", strconv.Itoa(actionUserID))
	lcFormValues.Set("grn_id", strconv.FormatInt(grnID, 10))
	lcFormValues.Set("entries", string(lcEntriesJSON))

	_, err = landedCostModel.CreateLandedCost(nil, lcFormValues)
	if err != nil {
		return fmt.Errorf("failed to process landed cost: %v", err)
	}

	// 8. Update Report Status
	_, err = m.DB.Exec(`
		UPDATE bulk_stock_report 
		SET status = 'APPROVED', action_by = ?, action_on = NOW() 
		WHERE id = ?`, actionUserID, reportID)
	if err != nil {
		return fmt.Errorf("failed to update report status: %v", err)
	}

	return nil
}
