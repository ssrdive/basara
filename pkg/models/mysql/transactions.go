package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/ssrdive/scribe"
	smodels "github.com/ssrdive/scribe/models"
	"log"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

type Transactions struct {
	DB                 *sql.DB
	TransactionsLogger *log.Logger
}

const (
	SparePartsSalesAccountID       = 200
	SparePartsCostOfSalesAccountID = 202
)

func (m *Transactions) GetSalesCommission(uid int) (models.CashInHand, error) {
	var r models.CashInHand
	err := m.DB.QueryRow(queries.GetSalesCommission, uid).Scan(&r.Amount)
	if err != nil {
		return models.CashInHand{}, nil
	}
	return r, nil
}

func (m *Transactions) GetCashInHand(uid int) (models.CashInHand, error) {
	var r models.CashInHand
	err := m.DB.QueryRow(queries.GetCashInHand, uid, uid).Scan(&r.Amount)
	if err != nil {
		return models.CashInHand{}, nil
	}
	return r, nil
}

func (m *Transactions) GetInventoryTransferItems(itid int) ([]models.PendingInventoryTransferItem, error) {
	var res []models.PendingInventoryTransferItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.InventoryTransferItems, itid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) InventoryTransferList() ([]models.InventoryTransferEntry, error) {
	var res []models.InventoryTransferEntry
	err := mysequel.QueryToStructs(&res, m.DB, queries.InventoryTransferList)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) GetWarehouseStock(wid int) ([]models.WarehouseStockItem, error) {
	var res []models.WarehouseStockItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.WarehouseStock, wid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) GetPendingTransfers(warehouse int, userType string) ([]models.PendingInventoryTransfer, error) {
	var res []models.PendingInventoryTransfer
	var err error
	if userType == "Admin" {
		err = mysequel.QueryToStructs(&res, m.DB, queries.GetPendingTransfers)
	} else {
		err = mysequel.QueryToStructs(&res, m.DB, queries.GetPendingTransfersByWarehouse, warehouse)
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) CreateInventoryTransfer(rparams, oparams []string, form url.Values) (int64, error) {
	m.TransactionsLogger.Println("------------------- START: CreateInventoryTransfer -------------------")

	var err error

	tx, err := m.DB.Begin()
	if err != nil {
		m.TransactionsLogger.Printf("Error beginning transaction: %v", err)
		return 0, err
	}
	defer func() {
		if err != nil {
			m.TransactionsLogger.Printf("Error occurred, rolling back transaction: %v", err)
			tx.Rollback()
			m.TransactionsLogger.Println("------------------- END: CreateInventoryTransfer -------------------")
		} else {
			m.TransactionsLogger.Println("Committing transaction")
			_ = tx.Commit()
			m.TransactionsLogger.Println("------------------- END: CreateInventoryTransfer -------------------")
		}
	}()

	// Converting JSON transfer entries to structs
	entries := form.Get("entries")
	m.TransactionsLogger.Printf("Transfer entries JSON: %s", entries)
	var transferItems []models.TransferItem
	err = json.Unmarshal([]byte(entries), &transferItems)
	if err != nil {
		m.TransactionsLogger.Printf("Error unmarshalling transfer items: %v", err)
		return 0, err
	}
	m.TransactionsLogger.Printf("Parsed transfer items:")
	PrintStructOrSliceAsTable(transferItems, m.TransactionsLogger)

	var transferItemIDs = make([]interface{}, len(transferItems))
	for i, item := range transferItems {
		transferQty, _ := strconv.Atoi(item.Quantity)
		if transferQty < 1 {
			err = errors.New("invalid transfer quantity")
			return 0, err
		}
		transferItemIDs[i] = item.ItemID
	}

	// Locking all transfer items from the source warehouse to avoid race conditions
	//_, err = tx.Exec(fmt.Sprintf("SELECT * FROM current_stock WHERE item_id IN (%v) AND warehouse_id = %v FOR UPDATE", ConvertArrayToString(transferItemIDs), form.Get("from_warehouse_id")))
	//if err != nil {
	//	m.TransactionsLogger.Printf("CreateInventoryTransfer: SELECT FOR UPDATE Failed: %v", err)
	//	return 0, err
	//}

	m.TransactionsLogger.Printf("Transfer item IDs: %v", transferItemIDs)

	// Load warehouse stock with transfer items to validate
	// if the transferring items are present in the source warehouse
	var warehouseStock []models.WarehouseStockItemQty
	err = mysequel.QueryToStructs(&warehouseStock, m.DB, queries.WarehouseItemQty(form.Get("from_warehouse_id"), ConvertArrayToString(transferItemIDs)))
	if err != nil {
		m.TransactionsLogger.Printf("Error loading warehouse stock: %v", err)
		return 0, err
	}
	m.TransactionsLogger.Printf("Loaded warehouse stock:")
	PrintStructOrSliceAsTable(warehouseStock, m.TransactionsLogger)

	if len(warehouseStock) != len(transferItemIDs) {
		err = errors.New("selected items do not exist in the source warehouse")
		m.TransactionsLogger.Println(err.Error())
		return 0, err
	}

	// Sorting transfer items / stocks for comparing
	// quantity validation through single loop
	sort.Slice(transferItems, func(i, j int) bool {
		lValue, _ := strconv.Atoi(transferItems[i].ItemID)
		rValue, _ := strconv.Atoi(transferItems[j].ItemID)
		return lValue < rValue
	})
	sort.Slice(warehouseStock, func(i, j int) bool {
		lValue, _ := strconv.Atoi(warehouseStock[i].ItemID)
		rValue, _ := strconv.Atoi(warehouseStock[j].ItemID)
		return lValue < rValue
	})

	m.TransactionsLogger.Printf("Sorted transfer items:")
	PrintStructOrSliceAsTable(transferItems, m.TransactionsLogger)
	m.TransactionsLogger.Printf("Sorted warehouse stock:")
	PrintStructOrSliceAsTable(warehouseStock, m.TransactionsLogger)

	// Validate if the transferring quantities are
	// present in the source warehouse
	for i, transferItem := range transferItems {
		transferQty, _ := strconv.Atoi(transferItem.Quantity)
		presentQty, _ := strconv.Atoi(warehouseStock[i].Quantity)

		m.TransactionsLogger.Printf("Validating item ID %s: transferQty = %d, presentQty = %d", transferItem.ItemID, transferQty, presentQty)

		if transferQty > presentQty {
			err = errors.New("transfer quantity is higher than the present quantity")
			m.TransactionsLogger.Println(err.Error())
			return 0, err
		}
	}

	var transfers []models.WarehouseItemStockWithDocumentIDs

	// Select items to be transferred from the source warehouse
	// based on their goods received note IDs. Priority is given to
	// move the items from old GRNs first.
	for _, transferItem := range transferItems {
		itemQty, _ := strconv.Atoi(transferItem.Quantity)
		m.TransactionsLogger.Printf("Preparing transfer for item ID %s with quantity %d", transferItem.ItemID, itemQty)

		var warehouseItemWithDocumentIDs []models.WarehouseItemStockWithDocumentIDs
		err = mysequel.QueryToStructs(&warehouseItemWithDocumentIDs, m.DB, queries.WarehouseItemStockWithDocumentIds, form.Get("from_warehouse_id"), transferItem.ItemID)
		if err != nil {
			m.TransactionsLogger.Printf("Error loading stock with document IDs: %v", err)
			return 0, err
		}
		m.TransactionsLogger.Printf("Loaded warehouse items with document IDs:")
		PrintStructOrSliceAsTable(warehouseItemWithDocumentIDs, m.TransactionsLogger)

		m.TransactionsLogger.Printf("Starting processing warehouse items for transfers loop")
		for i, stockItem := range warehouseItemWithDocumentIDs {
			m.TransactionsLogger.Printf("------------ Loop Index: %d", i)
			m.TransactionsLogger.Printf("Processing Item:")
			PrintStructOrSliceAsTable(stockItem, m.TransactionsLogger)
			fromWarehouseID, _ := strconv.Atoi(form.Get("from_warehouse_id"))
			subtractQty := 0
			if stockItem.Qty > itemQty {
				subtractQty = itemQty
				itemQty = 0
			} else {
				subtractQty = stockItem.Qty
				itemQty = itemQty - stockItem.Qty
			}
			m.TransactionsLogger.Printf("Subtracting %d from item ID %s (GoodsReceivedNoteID %s)", subtractQty, stockItem.ItemID, stockItem.GoodsReceivedNoteID)
			transfers = append(transfers, models.WarehouseItemStockWithDocumentIDs{
				EntrySpecifier:      stockItem.EntrySpecifier,
				WarehouseID:         fromWarehouseID,
				ItemID:              stockItem.ItemID,
				GoodsReceivedNoteID: stockItem.GoodsReceivedNoteID,
				InventoryTransferID: stockItem.InventoryTransferID,
				Qty:                 subtractQty,
			})
			m.TransactionsLogger.Printf("Transfers Status:")
			PrintStructOrSliceAsTable(transfers, m.TransactionsLogger)
			if itemQty == 0 {
				break
			}
		}
		m.TransactionsLogger.Printf("End processing warehouse items for transfers loop")
		m.TransactionsLogger.Printf("Transfers prepared for item ID %s: %+v", transferItem.ItemID, transfers)
	}

	var itid int64
	itid, err = mysequel.Insert(mysequel.Table{
		TableName: "inventory_transfer",
		Columns:   []string{"user_id", "from_warehouse_id", "to_warehouse_id"},
		Vals:      []interface{}{form.Get("user_id"), form.Get("from_warehouse_id"), form.Get("to_warehouse_id")},
		Tx:        tx,
	})
	if err != nil {
		m.TransactionsLogger.Printf("Error inserting into inventory_transfer: %v", err)
		return 0, err
	}
	m.TransactionsLogger.Printf("Inserted into inventory_transfer with ID %d", itid)

	var res sql.Result
	var rowsAffected int64
	// Moving the items from the current stock table to
	// float field for the transferring items based on the
	// GRN selected. inventory_transfer_item is also populated
	for _, transfer := range transfers {
		m.TransactionsLogger.Printf("Processing transfer for item ID %s, GoodsReceivedNoteID %s, Quantity %d", transfer.ItemID, transfer.GoodsReceivedNoteID, transfer.Qty)
		if transfer.InventoryTransferID.Valid {
			_, err = mysequel.Insert(mysequel.Table{
				TableName: "inventory_transfer_item",
				Columns:   []string{"entry_specifier", "inventory_transfer_id", "prev_inventory_transfer_id", "goods_received_note_id", "item_id", "qty"},
				Vals:      []interface{}{transfer.EntrySpecifier, itid, transfer.InventoryTransferID.Int32, transfer.GoodsReceivedNoteID, transfer.ItemID, transfer.Qty},
				Tx:        tx,
			})
			if err != nil {
				m.TransactionsLogger.Printf("Error inserting into inventory_transfer_item: %v", err)
				return 0, err
			}

			res, err = tx.Exec("UPDATE current_stock SET qty = qty - ?, float_qty = float_qty + ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ? AND entry_specifier = ?", transfer.Qty, transfer.Qty, form.Get("from_warehouse_id"), transfer.ItemID, transfer.GoodsReceivedNoteID, transfer.InventoryTransferID.Int32, transfer.EntrySpecifier)
			if err != nil {
				m.TransactionsLogger.Printf("Error updating current_stock: %v", err)
				return 0, err
			}

			rowsAffected, err = res.RowsAffected()
			if err != nil {
				m.TransactionsLogger.Printf("Failed to get rows affected: %v", err)
				return 0, err
			}

			if rowsAffected != 1 {
				err = errors.New("ERROR: More than 1 row affected!")
				m.TransactionsLogger.Printf("ERROR: %v", err)
				return 0, err
			}
		} else {
			_, err = mysequel.Insert(mysequel.Table{
				TableName: "inventory_transfer_item",
				Columns:   []string{"entry_specifier", "inventory_transfer_id", "goods_received_note_id", "item_id", "qty"},
				Vals:      []interface{}{transfer.EntrySpecifier, itid, transfer.GoodsReceivedNoteID, transfer.ItemID, transfer.Qty},
				Tx:        tx,
			})
			if err != nil {
				m.TransactionsLogger.Printf("Error inserting into inventory_transfer_item (no prev ID): %v", err)
				return 0, err
			}

			res, err = tx.Exec("UPDATE current_stock SET qty = qty - ?, float_qty = float_qty + ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id IS NULL AND entry_specifier = ?", transfer.Qty, transfer.Qty, form.Get("from_warehouse_id"), transfer.ItemID, transfer.GoodsReceivedNoteID, transfer.EntrySpecifier)
			if err != nil {
				m.TransactionsLogger.Printf("Error updating current_stock (no prev ID): %v", err)
				return 0, err
			}

			rowsAffected, err = res.RowsAffected()
			if err != nil {
				m.TransactionsLogger.Printf("Failed to get rows affected: %v", err)
				return 0, err
			}

			if rowsAffected != 1 {
				err = errors.New("ERROR: More than 1 row affected!")
				m.TransactionsLogger.Printf("ERROR: %v", err)
				return 0, err
			}
		}
	}

	m.TransactionsLogger.Printf("Completed: CreateInventoryTransfer ID: %d", itid)
	return itid, nil
}

func (m *Transactions) CreateInvoice(rparams, oparams []string, apiKey string, form url.Values) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}

	if requestExists(tx, form.Get("request_id")) {
		tx.Rollback()
		return 0, nil
	}

	if form.Get("request_id") != "" {
		if form.Get("execution_type") == "plan" {
			tx2, err2 := m.DB.Begin()
			if err2 != nil {
				tx.Rollback()
				return 0, err
			}
			_, err := mysequel.Insert(mysequel.Table{
				TableName: "unique_requests",
				Columns:   []string{"request_id"},
				Vals:      []interface{}{form.Get("request_id")},
				Tx:        tx2,
			})
			if err != nil {
				tx.Rollback()
				tx2.Rollback()
				return 0, err
			}
			tx2.Commit()
		} else {
			_, err := mysequel.Insert(mysequel.Table{
				TableName: "unique_requests",
				Columns:   []string{"request_id"},
				Vals:      []interface{}{form.Get("request_id")},
				Tx:        tx,
			})
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	items := form.Get("items")
	var invoiceItems []models.TransferItem
	err = json.Unmarshal([]byte(items), &invoiceItems)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	var invoiceItemIDs = make([]interface{}, len(invoiceItems))

	for i, item := range invoiceItems {
		invoiceItemIDs[i] = item.ItemID
	}

	// Locking all transfer items from the source warehouse to avoid race conditions
	//_, err = tx.Exec(fmt.Sprintf("SELECT * FROM current_stock WHERE item_id IN (%v) AND warehouse_id = %v FOR UPDATE", ConvertArrayToString(invoiceItemIDs), form.Get("from_warehouse")))
	//if err != nil {
	//	m.TransactionsLogger.Printf("CreateInvoice: SELECT FOR UPDATE Failed: %v", err)
	//	return 0, err
	//}

	// Load warehouse stock with transfer items to validate
	// if the transferring items are present in the source warehouse

	var warehouseStock []models.WarehouseStockItemQty
	err = mysequel.QueryToStructs(&warehouseStock, m.DB, queries.WarehouseItemQty(form.Get("from_warehouse"), ConvertArrayToString(invoiceItemIDs)))
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(warehouseStock) != len(invoiceItemIDs) {
		return 0, errors.New("selected items does not exist in the source warehouse")
	}

	sort.Slice(invoiceItems, func(i, j int) bool {
		lValue, _ := strconv.Atoi(invoiceItems[i].ItemID)
		rValue, _ := strconv.Atoi(invoiceItems[j].ItemID)
		return lValue < rValue
	})

	sort.Slice(warehouseStock, func(i, j int) bool {
		lValue, _ := strconv.Atoi(warehouseStock[i].ItemID)
		rValue, _ := strconv.Atoi(warehouseStock[j].ItemID)
		return lValue < rValue
	})

	// Validate if the transferring quantities are
	// present in the source warehouse

	for i, invoiceItem := range invoiceItems {
		transferQty, _ := strconv.Atoi(invoiceItem.Quantity)
		presentQty, _ := strconv.Atoi(warehouseStock[i].Quantity)

		if transferQty > presentQty {
			return 0, errors.New("invoice quantity is higher than the present quantity")
		}
	}

	var invoice []models.WarehouseItemStockWithDocumentIDsAndPrices

	// Select items to be transferred from the source warehouse
	// based on their goods received note ids. Priority is given to
	// move the items from old GRNs first.

	for _, invoiceItem := range invoiceItems {
		itemQty, _ := strconv.Atoi(invoiceItem.Quantity)

		var warehouseItemWithDocumentIDs []models.WarehouseItemStockWithDocumentIDsAndPrices
		err = mysequel.QueryToStructs(&warehouseItemWithDocumentIDs, m.DB, queries.WarehouseItemStockWithDocumentIdsAndPrices, form.Get("from_warehouse"), invoiceItem.ItemID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		for _, stockItem := range warehouseItemWithDocumentIDs {
			fromWarehouseID, _ := strconv.Atoi(form.Get("from_warehouse"))
			subtractQty := 0
			if stockItem.Qty > itemQty {
				subtractQty = itemQty
				itemQty = 0
			} else {
				subtractQty = stockItem.Qty
				itemQty = itemQty - stockItem.Qty
			}
			invoice = append(invoice, models.WarehouseItemStockWithDocumentIDsAndPrices{
				EntrySpecifier:              stockItem.EntrySpecifier,
				WarehouseID:                 fromWarehouseID,
				ItemID:                      stockItem.ItemID,
				GoodsReceivedNoteID:         stockItem.GoodsReceivedNoteID,
				InventoryTransferID:         stockItem.InventoryTransferID,
				Qty:                         subtractQty,
				CostPriceWithoutLandedCosts: stockItem.CostPriceWithoutLandedCosts,
				CostPrice:                   stockItem.CostPrice,
				Price:                       stockItem.Price,
			})
			if itemQty == 0 {
				break
			}
		}
	}

	var cashAccountID sql.NullInt32
	err = tx.QueryRow(queries.OfficerAccNo, form.Get("user_id")).Scan(&cashAccountID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if !cashAccountID.Valid {
		tx.Rollback()
		err = errors.New("cash in hand account not specififed")
		return 0, err
	}

	if form.Get("execution_type") == "plan" {
		tx.Rollback()
		return 0, nil
	}

	iid, err := mysequel.Insert(mysequel.Table{
		TableName: "invoice",
		Columns:   []string{"user_id", "warehouse_id", "cost_price", "price_before_discount", "discount", "price_after_discount", "customer_name", "customer_contact"},
		Vals:      []interface{}{form.Get("user_id"), form.Get("from_warehouse"), 0, 0, form.Get("discount"), 0, form.Get("customer_name"), form.Get("customer_contact")},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	var costPrice float64
	costPrice = 0
	var price float64
	price = 0
	var costPriceWithoutLCs float64
	costPriceWithoutLCs = 0

	for _, item := range invoice {
		costPrice = costPrice + (item.CostPrice * float64(item.Qty))
		price = price + (item.Price * float64(item.Qty))
		costPriceWithoutLCs = costPriceWithoutLCs + (item.CostPriceWithoutLandedCosts * float64(item.Qty))
		if item.InventoryTransferID.Valid {
			_, err = tx.Exec("UPDATE current_stock SET qty = qty - ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ? AND entry_specifier = ?", item.Qty, item.WarehouseID, item.ItemID, item.GoodsReceivedNoteID, item.InventoryTransferID.Int32, item.EntrySpecifier)
			if err != nil {
				tx.Rollback()
				return 0, err
			}

			_, err = mysequel.Insert(mysequel.Table{
				TableName: "invoice_item",
				Columns:   []string{"entry_specifier", "invoice_id", "item_id", "goods_received_note_id", "inventory_transfer_id", "qty", "cost_price", "price"},
				Vals:      []interface{}{item.EntrySpecifier, iid, item.ItemID, item.GoodsReceivedNoteID, item.InventoryTransferID.Int32, item.Qty, item.CostPrice, item.Price},
				Tx:        tx,
			})
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		} else {
			_, err = tx.Exec("UPDATE current_stock SET qty = qty - ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id IS NULL AND entry_specifier = ?", item.Qty, item.WarehouseID, item.ItemID, item.GoodsReceivedNoteID, item.EntrySpecifier)
			if err != nil {
				tx.Rollback()
				return 0, err
			}

			_, err = mysequel.Insert(mysequel.Table{
				TableName: "invoice_item",
				Columns:   []string{"entry_specifier", "invoice_id", "item_id", "goods_received_note_id", "qty", "cost_price", "price"},
				Vals:      []interface{}{item.EntrySpecifier, iid, item.ItemID, item.GoodsReceivedNoteID, item.Qty, item.CostPrice, item.Price},
				Tx:        tx,
			})
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	discount, _ := strconv.ParseFloat(form.Get("discount"), 32)
	priceAfterDiscount := math.Round((price*(float64(100)-discount)/100)*100) / 100

	_, err = mysequel.Update(mysequel.UpdateTable{
		Table: mysequel.Table{
			TableName: "invoice",
			Columns:   []string{"cost_price", "price_before_discount", "price_after_discount"},
			Vals:      []interface{}{costPrice, price, priceAfterDiscount},
			Tx:        tx,
		},
		WColumns: []string{"id"},
		WVals:    []string{strconv.FormatInt(iid, 10)},
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "remark"},
		Vals:      []interface{}{form.Get("user_id"), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02"), fmt.Sprintf("INVOICE %d", iid)},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	journalEntries := []smodels.JournalEntry{
		{Account: fmt.Sprintf("%d", cashAccountID.Int32), Debit: fmt.Sprintf("%f", priceAfterDiscount), Credit: ""},
		{Account: fmt.Sprintf("%d", SparePartsSalesAccountID), Debit: "", Credit: fmt.Sprintf("%f", priceAfterDiscount)},
		{Account: fmt.Sprintf("%d", SparePartsCostOfSalesAccountID), Debit: fmt.Sprintf("%f", costPriceWithoutLCs), Credit: ""},
		{Account: fmt.Sprintf("%d", StockAccountID), Debit: "", Credit: fmt.Sprintf("%f", costPriceWithoutLCs)},
	}
	err = scribe.IssueJournalEntries(tx, tid, journalEntries)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	message := fmt.Sprintf("Dear Customer, Thank you for your purchase of LKR %s. We look forward to serving you again", humanize.Comma(int64(priceAfterDiscount)))

	telephone := fmt.Sprintf("%s,768237192,703524279,775607777,703524274", form.Get("customer_contact"))

	requestURL := fmt.Sprintf("https://richcommunication.dialog.lk/api/sms/inline/send.php?destination=%s&q=%s&message=%s", telephone, apiKey, url.QueryEscape(message))

	if form.Get("execution_type") == "plan" {
		tx.Rollback()
		return 0, nil
	} else {
		resp, _ := http.Get(requestURL)
		defer resp.Body.Close()

		tx.Commit()
		return iid, nil
	}
}

func (m *Transactions) InventoryTransferDetails(itid int) (models.InventoryTransferSummary, error) {
	var inventoryTransferSummary models.InventoryTransferSummary
	err := m.DB.QueryRow(queries.InventoryTransferDetails, itid).Scan(&inventoryTransferSummary.InventoryTransferID,
		&inventoryTransferSummary.Created, &inventoryTransferSummary.IssuedBy, &inventoryTransferSummary.FromWarehouse,
		&inventoryTransferSummary.ToWarehouse, &inventoryTransferSummary.Resolution,
		&inventoryTransferSummary.ResolvedBy, &inventoryTransferSummary.ResolvedOn,
		&inventoryTransferSummary.ResolutionRemarks)

	if err != nil {
		return models.InventoryTransferSummary{}, err
	}

	var inventoryTransferItems []models.InventoryTransferItemDetails
	err = mysequel.QueryToStructs(&inventoryTransferItems, m.DB, queries.InventoryTransferItemsDetails, itid)
	if err != nil {
		return models.InventoryTransferSummary{}, err
	}

	inventoryTransferSummary.TransferItems = inventoryTransferItems

	return inventoryTransferSummary, nil
}

func (m *Transactions) InvoiceDetails(iid int) (models.InvoiceSummary, error) {
	var invoiceSummary models.InvoiceSummary
	err := m.DB.QueryRow(queries.InvoiceDetails, iid).Scan(&invoiceSummary.InvoiceID, &invoiceSummary.IssuedBy,
		&invoiceSummary.Warehouse, &invoiceSummary.PriceBeforeDiscount, &invoiceSummary.Discount,
		&invoiceSummary.PriceAfterDiscount, &invoiceSummary.CustomerName, &invoiceSummary.CustomerContact)

	if err != nil {
		return models.InvoiceSummary{}, err
	}

	var invoiceItems []models.InvoiceItemDetails
	err = mysequel.QueryToStructs(&invoiceItems, m.DB, queries.InvoiceItemDetails, iid)
	if err != nil {
		return models.InvoiceSummary{}, err
	}

	invoiceSummary.ItemDetails = invoiceItems

	return invoiceSummary, nil
}

func requestExists(tx *sql.Tx, requestId string) bool {
	var requestDBId int
	err := tx.QueryRow(queries.RequestPresentCheck, requestId).Scan(&requestDBId)
	if err != nil {
		return false
	}
	return true
}

func (m *Transactions) InventoryTransferAction(rparams, oparams []string, form url.Values) (int64, error) {
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

	if requestExists(tx, form.Get("request_id")) {
		m.TransactionsLogger.Printf("Request dropped: %s", form.Get("request_id"))
		return 0, nil
	}

	if form.Get("request_id") != "" {
		tx2, err2 := m.DB.Begin()
		if err2 != nil {
			return 0, err
		}
		_, err := mysequel.Insert(mysequel.Table{
			TableName: "unique_requests",
			Columns:   []string{"request_id"},
			Vals:      []interface{}{form.Get("request_id")},
			Tx:        tx2,
		})
		if err != nil {
			_ = tx2.Rollback()
			return 0, err
		}
		_ = tx2.Commit()
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	m.TransactionsLogger.Println("Log Time: " + time.Now().Format("2006-01-02 15:04:05.000") + " Start Time: " + timestamp + " Request Received: " + form.Get("request_id"))

	itid := form.Get("inventory_transfer_id")
	userID := form.Get("user_id")
	resolution := form.Get("resolution")
	resolutionRemarks := form.Get("resolution_remarks")

	var transferItemsForAction []models.InventoryTransferItemForAction
	err = mysequel.QueryToStructs(&transferItemsForAction, m.DB, queries.InventoryTransferItemsForAction, itid)
	if err != nil {
		m.TransactionsLogger.Println(err)
		return 0, err
	}
	m.TransactionsLogger.Println("Log Time: " + time.Now().Format("2006-01-02 15:04:05.000") + " Start Time: " + timestamp + " Lock Aquired: " + form.Get("request_id"))

	var transferActionsItemIDs = make([]interface{}, len(transferItemsForAction))
	for i, actionItem := range transferItemsForAction {
		transferActionsItemIDs[i] = actionItem.ItemID
	}

	// Locking all transfer items from the source warehouse to avoid race conditions
	//_, err = tx.Exec(fmt.Sprintf("SELECT * FROM current_stock WHERE item_id IN (%v) AND warehouse_id = %v FOR UPDATE", ConvertArrayToString(transferActionsItemIDs), transferItemsForAction[0].FromWarehouseID))
	//if err != nil {
	//	m.TransactionsLogger.Printf("InventoryTransferAction: SELECT FOR UPDATE Failed: %v", err)
	//	return 0, err
	//}
	//time.Sleep(2 * time.Second)
	if requestExists(tx, form.Get("request_id")) {
		m.TransactionsLogger.Println("Request dropped: " + form.Get("request_id"))
		return 0, nil
	}

	var resolvedBy sql.NullInt32
	err = tx.QueryRow("SELECT resolved_by FROM inventory_transfer WHERE id = ?", itid).Scan(&resolvedBy)
	if resolvedBy.Valid {
		return 0, nil
	}

	for _, actionItem := range transferItemsForAction {
		if resolution == "Approved" || resolution == "Provisional" {
			var costPrice float64
			var landedCost float64
			var price float64
			if actionItem.PrevInventoryTransferID.Valid {
				_, err = tx.Exec("UPDATE current_stock SET float_qty = float_qty - ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ? AND entry_specifier = ?", actionItem.Quantity, actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.PrevInventoryTransferID.Int32, actionItem.EntrySpecifier)
				if err != nil {
					m.TransactionsLogger.Println(err)
					return 0, err
				}

				err = tx.QueryRow("SELECT cost_price, landed_costs, price FROM current_stock WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ? AND entry_specifier = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.PrevInventoryTransferID.Int32, actionItem.EntrySpecifier).Scan(&costPrice, &landedCost, &price)
				if err != nil {
					m.TransactionsLogger.Println(err)
					return 0, err
				}
			} else {
				_, err = tx.Exec("UPDATE current_stock SET float_qty = float_qty - ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id IS NULL AND entry_specifier = ?", actionItem.Quantity, actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.EntrySpecifier)
				if err != nil {
					m.TransactionsLogger.Println(err)
					return 0, err
				}

				err = tx.QueryRow("SELECT cost_price, landed_costs, price FROM current_stock WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ?  AND inventory_transfer_id IS NULL AND entry_specifier = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.EntrySpecifier).Scan(&costPrice, &landedCost, &price)
				if err != nil {
					m.TransactionsLogger.Println(err)
					return 0, err
				}
			}

			_, err = mysequel.Insert(mysequel.Table{
				TableName: "current_stock",
				Columns:   []string{"entry_specifier", "warehouse_id", "item_id", "goods_received_note_id", "inventory_transfer_id", "cost_price", "landed_costs", "qty", "float_qty", "price"},
				Vals:      []interface{}{uuid.New(), actionItem.ToWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, itid, costPrice, landedCost, actionItem.Quantity, 0, price},
				Tx:        tx,
			})
			if err != nil {
				m.TransactionsLogger.Println(err)
				return 0, err
			}
		} else if resolution == "Rejected" {
			if actionItem.PrevInventoryTransferID.Valid {
				_, err = tx.Exec("UPDATE current_stock SET qty = qty + ?, float_qty = float_qty - ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ?", actionItem.Quantity, actionItem.Quantity, actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.PrevInventoryTransferID.Int32)
				if err != nil {
					m.TransactionsLogger.Println(err)
					return 0, err
				}
			} else {
				_, err = tx.Exec("UPDATE current_stock SET qty = qty + ?, float_qty = float_qty - ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ?  AND inventory_transfer_id IS NULL", actionItem.Quantity, actionItem.Quantity, actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID)
				if err != nil {
					m.TransactionsLogger.Println(err)
					return 0, err
				}
			}
		}
	}

	_, err = mysequel.Update(mysequel.UpdateTable{
		Table: mysequel.Table{
			TableName: "inventory_transfer",
			Columns:   []string{"resolved_by", "resolved_on", "resolution", "resolution_remarks"},
			Vals:      []interface{}{userID, time.Now().Format("2006-01-02 15:04:05"), resolution, resolutionRemarks},
			Tx:        tx,
		},
		WColumns: []string{"id"},
		WVals:    []string{itid},
	})
	if err != nil {
		m.TransactionsLogger.Println(err)
		return 0, err
	}

	m.TransactionsLogger.Println("Log Time: " + time.Now().Format("2006-01-02 15:04:05.000") + " Start Time: " + timestamp + " Transfer Complete: " + form.Get("request_id"))
	return 0, nil
}

func ConvertArrayToString(arr []interface{}) string {
	str := ""
	for i, elem := range arr {
		if i != len(arr)-1 {
			str = str + fmt.Sprintf("%v", elem) + ","
		} else {
			str = str + fmt.Sprintf("%v", elem)
		}
	}
	return str
}

// PrintStructOrSliceAsTable prints a single struct or a slice of structs as a formatted table.
func PrintStructOrSliceAsTable(data interface{}, logger *log.Logger) {
	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			logger.Println("Slice is empty")
			return
		}
		elemType := v.Type().Elem()
		if elemType.Kind() != reflect.Struct {
			logger.Println("Slice elements are not structs")
			return
		}
		printTable(v, logger)
	case reflect.Struct:
		// Wrap single struct into a slice for table printing
		slice := reflect.MakeSlice(reflect.SliceOf(v.Type()), 1, 1)
		slice.Index(0).Set(v)
		printTable(slice, logger)
	default:
		logger.Println("Input is neither a struct nor a slice of structs")
	}
}

func printTable(slice reflect.Value, logger *log.Logger) {
	elemType := slice.Type().Elem()
	// Extract field names for the table header
	fieldNames := make([]string, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		fieldNames[i] = elemType.Field(i).Name
	}

	// Prepare rows of the table
	rows := make([][]string, slice.Len()+1)
	rows[0] = fieldNames // Header row

	for i := 0; i < slice.Len(); i++ {
		row := make([]string, elemType.NumField())
		structValue := slice.Index(i)
		for j := 0; j < structValue.NumField(); j++ {
			value := structValue.Field(j).Interface()
			row[j] = formatValue(value)
		}
		rows[i+1] = row
	}

	// Calculate column widths
	colWidths := make([]int, len(fieldNames))
	for _, row := range rows {
		for colIndex, col := range row {
			if len(col) > colWidths[colIndex] {
				colWidths[colIndex] = len(col)
			}
		}
	}

	// Print table with proper alignment and separators
	printSeparator(colWidths, logger)
	printRow(rows[0], colWidths, logger) // Print header
	printSeparator(colWidths, logger)
	for _, row := range rows[1:] {
		printRow(row, colWidths, logger)
	}
	printSeparator(colWidths, logger)
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
	case uint, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(value).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func printSeparator(colWidths []int, logger *log.Logger) {
	separator := "+"
	for _, width := range colWidths {
		separator += strings.Repeat("-", width+2) + "+"
	}
	logger.Println(separator)
}

func printRow(row []string, colWidths []int, logger *log.Logger) {
	line := "|"
	for colIndex, col := range row {
		line += " " + padRight(col, colWidths[colIndex]) + " |"
	}
	logger.Println(line)
}

func padRight(str string, length int) string {
	if len(str) < length {
		return str + strings.Repeat(" ", length-len(str))
	}
	return str
}
