package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

type Transactions struct {
	DB *sql.DB
}

func (m *Transactions) GetInventoryTransferItems(itid int) ([]models.PendingInventoryTransferItem, error) {
	var res []models.PendingInventoryTransferItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.INVENTORY_TRANSFER_ITEMS, itid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) GetWarehouseStock(wid int) ([]models.WarehouseStockItem, error) {
	var res []models.WarehouseStockItem
	err := mysequel.QueryToStructs(&res, m.DB, queries.WAREHOUSE_STOCK, wid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) GetPendingTransfers(warehouse int, userType string) ([]models.PendingInventoryTransfer, error) {
	var res []models.PendingInventoryTransfer
	var err error
	if userType == "Admin" {
		err = mysequel.QueryToStructs(&res, m.DB, queries.GET_PENDING_TRANSFERS)
	} else {
		err = mysequel.QueryToStructs(&res, m.DB, queries.GET_PENDING_TRANSFERS_BY_WAREHOUSE, warehouse)
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Transactions) CreateInventoryTransfer(rparams, oparams []string, form url.Values) (int64, error) {
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

	// Converting JSON transfer entries to structs

	entries := form.Get("entries")
	var transferItems []models.TransferItem
	json.Unmarshal([]byte(entries), &transferItems)

	var transferItemIDs = make([]interface{}, len(transferItems))

	for i, item := range transferItems {
		transferItemIDs[i] = item.ItemID
	}

	// Load warehouse stock with transfer items to validate
	// if the transferring items are present in the source warehouse

	var warehouseStock []models.WarehouseStockItemQty
	err = mysequel.QueryToStructs(&warehouseStock, m.DB, queries.WAREHOSUE_ITEM_QTY(form.Get("from_warehouse_id"), ConvertArrayToString(transferItemIDs)))
	if err != nil {
		return 0, err
	}

	if len(warehouseStock) != len(transferItemIDs) {
		return 0, errors.New("selected items does not exist in the source warehouse")
	}

	// Sorting transfer items / stocks for comparing
	// quantity validation through single loop

	sort.Slice(transferItems, func(i, j int) bool {
		lValue, _ := strconv.Atoi(transferItems[i].ItemID)
		rValue, _ := strconv.Atoi(transferItems[j].ItemID)
		return lValue < rValue
	})

	sort.Slice(warehouseStock, func(i, j int) bool {
		lValue, _ := strconv.Atoi(transferItems[i].ItemID)
		rValue, _ := strconv.Atoi(transferItems[j].ItemID)
		return lValue < rValue
	})

	// Validate if the transferring quantities are
	// present in the source warehouse

	for i, transferItem := range transferItems {
		transferQty, _ := strconv.Atoi(transferItem.Quantity)
		presentQty, _ := strconv.Atoi(warehouseStock[i].Quantity)

		if transferQty > presentQty {
			return 0, errors.New("transfer quantity is higher than the present quantity")
		}
	}

	var transfers []models.WarehouseItemStockWithDocumentIDs

	// Select items to be transferred from the source warehouse
	// based on their goods received note ids. Priority is given to
	// move the items from old GRNs first.

	for _, transferItem := range transferItems {
		itemQty, _ := strconv.Atoi(transferItem.Quantity)

		var warehouseItemWithDocumentIDs []models.WarehouseItemStockWithDocumentIDs
		err = mysequel.QueryToStructs(&warehouseItemWithDocumentIDs, m.DB, queries.WAREHOUSE_ITEM_STOCK_WITH_DOCUMENT_IDS, form.Get("from_warehouse_id"), transferItem.ItemID)

		for _, stockItem := range warehouseItemWithDocumentIDs {
			fromWarehouseID, _ := strconv.Atoi(form.Get("from_warehouse_id"))
			subtractQty := 0
			if stockItem.Qty > itemQty {
				subtractQty = itemQty
				itemQty = 0
			} else {
				subtractQty = stockItem.Qty
				itemQty = itemQty - stockItem.Qty
			}
			transfers = append(transfers, models.WarehouseItemStockWithDocumentIDs{
				WarehouseID:         fromWarehouseID,
				ItemID:              stockItem.ItemID,
				GoodsReceivedNoteID: stockItem.GoodsReceivedNoteID,
				InventoryTransferID: stockItem.InventoryTransferID,
				Qty:                 subtractQty,
			})
			if itemQty == 0 {
				break
			}
		}
	}

	itid, err := mysequel.Insert(mysequel.Table{
		TableName: "inventory_transfer",
		Columns:   []string{"user_id", "from_warehouse_id", "to_warehouse_id"},
		Vals:      []interface{}{form.Get("user_id"), form.Get("from_warehouse_id"), form.Get("to_warehouse_id")},
		Tx:        tx,
	})
	if err != nil {
		return 0, err
	}

	// Moving the items from the current stock table to
	// float field for the transferring items based on the
	// GRN selected. inventory_transfer_item is also populated

	for _, transfer := range transfers {
		if transfer.InventoryTransferID.Valid {
			_, err = mysequel.Insert(mysequel.Table{
				TableName: "inventory_transfer_item",
				Columns:   []string{"inventory_transfer_id", "prev_inventory_transfer_id", "goods_received_note_id", "item_id", "qty"},
				Vals:      []interface{}{itid, transfer.InventoryTransferID.Int32, transfer.GoodsReceivedNoteID, transfer.ItemID, transfer.Qty},
				Tx:        tx,
			})
			if err != nil {
				return 0, err
			}

			_, err = tx.Exec("UPDATE current_stock SET qty = qty - ?, float_qty = ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ?", transfer.Qty, transfer.Qty, form.Get("from_warehouse_id"), transfer.ItemID, transfer.GoodsReceivedNoteID, transfer.InventoryTransferID.Int32)
			if err != nil {
				return 0, err
			}
		} else {
			_, err = mysequel.Insert(mysequel.Table{
				TableName: "inventory_transfer_item",
				Columns:   []string{"inventory_transfer_id", "goods_received_note_id", "item_id", "qty"},
				Vals:      []interface{}{itid, transfer.GoodsReceivedNoteID, transfer.ItemID, transfer.Qty},
				Tx:        tx,
			})
			if err != nil {
				return 0, err
			}

			_, err = tx.Exec("UPDATE current_stock SET qty = qty - ?, float_qty = ? WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ?", transfer.Qty, transfer.Qty, form.Get("from_warehouse_id"), transfer.ItemID, transfer.GoodsReceivedNoteID)
			if err != nil {
				return 0, err
			}
		}
	}

	return itid, nil
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

	itid := form.Get("inventory_transfer_id")
	userID := form.Get("user_id")
	resolution := form.Get("resolution")
	resolutionRemarks := form.Get("resolution_remarks")

	var resolvedBy sql.NullInt32
	err = tx.QueryRow("SELECT resolved_by FROM inventory_transfer WHERE id = ?", itid).Scan(&resolvedBy)
	if resolution.Valid {
		return 0, nil
	}

	var transferItemsForAction []models.InventoryTransferItemForAction
	err = mysequel.QueryToStructs(&transferItemsForAction, m.DB, queries.INVENTORY_TRANSFER_ITEMS_FOR_ACTION, itid)
	if err != nil {
		return 0, err
	}

	for _, actionItem := range transferItemsForAction {
		if resolution == "Approved" || resolution == "Provisional" {
			var costPrice float64
			var landedCost float64
			var price float64
			if actionItem.PrevInventoryTransferID.Valid {
				_, err = tx.Exec("UPDATE current_stock SET float_qty = 0 WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.PrevInventoryTransferID.Int32)

				err = tx.QueryRow("SELECT cost_price, landed_costs, price FROM current_stock WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.PrevInventoryTransferID.Int32).Scan(&costPrice, &landedCost, &price)
				if err != nil {
					return 0, err
				}
			} else {
				_, err = tx.Exec("UPDATE current_stock SET float_qty = 0 WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID)

				err = tx.QueryRow("SELECT cost_price, landed_costs, price FROM current_stock WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID).Scan(&costPrice, &landedCost, &price)
				if err != nil {
					return 0, err
				}
			}

			_, err = mysequel.Insert(mysequel.Table{
				TableName: "current_stock",
				Columns:   []string{"warehouse_id", "item_id", "goods_received_note_id", "inventory_transfer_id", "cost_price", "landed_costs", "qty", "float_qty", "price"},
				Vals:      []interface{}{actionItem.ToWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, itid, costPrice, landedCost, actionItem.Quantity, 0, price},
				Tx:        tx,
			})
			if err != nil {
				return 0, err
			}
		} else if resolution == "Rejected" {
			if actionItem.PrevInventoryTransferID.Valid {
				_, err = tx.Exec("UPDATE current_stock SET qty = qty + float_qty WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ? AND inventory_transfer_id = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID, actionItem.PrevInventoryTransferID.Int32)
				if err != nil {
					return 0, err
				}
			} else {
				_, err = tx.Exec("UPDATE current_stock SET fqty = qty + float_qty WHERE warehouse_id = ? AND item_id = ? AND goods_received_note_id = ?", actionItem.FromWarehouseID, actionItem.ItemID, actionItem.GoodsReceivedNoteID)
				if err != nil {
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
		return 0, err
	}

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
