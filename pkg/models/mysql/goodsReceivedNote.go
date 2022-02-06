package mysql

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
	"github.com/ssrdive/mysequel"
)

// PurchaseOrderModel struct holds database instance
type GoodsReceivedNoteModel struct {
	DB *sql.DB
}

// CreateGoodsReceivedNote creats an Goods Received Note
func (m *GoodsReceivedNoteModel) CreateGoodsReceivedNote(rparams, oparams []string, form url.Values) (int64, error) {
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

	entities := form.Get("entries")
	var gRNItem []models.GRNItemEntry
	json.Unmarshal([]byte(entities), &gRNItem)

	grnid, err := mysequel.Insert(mysequel.Table{
		TableName: "goods_received_note",
		Columns:   []string{"user_id", "purcahse_order_id", "supplier_id", "warehouse_id", "discount_type", "discount_amount", "price_before_discount", "total_price", "remarks"},
		Vals:      []interface{}{form.Get("user_id"), form.Get("order_id"), form.Get("supplier_id"), form.Get("warehouse_id"), form.Get("discount_type"), form.Get("discount_amount"), 0, form.Get("total_price"), form.Get("remark")},
		Tx:        tx,
	})

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	var totalPriceBeforeDiscount = 0.0

	for _, entry := range gRNItem {

		unitPrice, err := strconv.ParseFloat(entry.UnitPrice, 32)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		quantity, err := strconv.ParseFloat(entry.Quantity, 32)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		totalPrice := unitPrice * quantity

		_, err = mysequel.Insert(mysequel.Table{
			TableName: "goods_received_note_item",
			Columns:   []string{"goods_received_note_id", "item_id", "unit_price", "qty", "total_price"},
			Vals:      []interface{}{grnid, entry.ItemID, unitPrice, quantity, totalPrice},
			Tx:        tx,
		})

		if err != nil {
			tx.Rollback()
			return 0, err
		}

		if form.Get("order_id") != "" {
			var id, orderItemQty, totalReconciled, totalCancelled float64

			ids := []interface{}{entry.ItemID, form.Get("order_id")}

			err := m.DB.QueryRow(queries.PURCHASE_ORDER_ITEM_COUNT, ids...).Scan(&id, &orderItemQty, &totalReconciled, &totalCancelled)

			if err == nil {
				leftToreconciled := orderItemQty - (totalReconciled + totalCancelled)
				var totalReconciledToBe float64
				if leftToreconciled >= quantity {
					totalReconciledToBe = totalReconciled + quantity
				} else {
					totalReconciledToBe = leftToreconciled
				}

				_, err := mysequel.Update(mysequel.UpdateTable{
					Table: mysequel.Table{
						TableName: "purchase_order_item",
						Columns:   []string{"total_reconciled"},
						Vals:      []interface{}{totalReconciledToBe},
						Tx:        tx,
					},
					WColumns: []string{"item_id", "purchase_order_id"},
					WVals:    []string{entry.ItemID, form.Get("order_id")},
				})

				if err != nil {
					tx.Rollback()
					return 0, err
				}

				_, err = mysequel.Insert(mysequel.Table{
					TableName: "purchase_order_item_reconciliation",
					Columns:   []string{"purcahse_order_id", "goods_received_note_id", "item_id", "qty"},
					Vals:      []interface{}{form.Get("order_id"), grnid, entry.ItemID, quantity},
					Tx:        tx,
				})

				if err != nil {
					tx.Rollback()
					return 0, err
				}
			}
		}

		totalPriceBeforeDiscount = totalPriceBeforeDiscount + totalPrice
	}

	_, err = mysequel.Update(mysequel.UpdateTable{
		Table: mysequel.Table{
			TableName: "goods_received_note",
			Columns:   []string{"price_before_discount"},
			Vals:      []interface{}{totalPriceBeforeDiscount},
			Tx:        tx,
		},
		WColumns: []string{"id"},
		WVals:    []string{strconv.FormatInt(grnid, 10)},
	})

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return grnid, nil
}

func (m *GoodsReceivedNoteModel) GoodsReceivedNotesList() ([]models.GoodReceivedNoteEntry, error) {
	var res []models.GoodReceivedNoteEntry
	err := mysequel.QueryToStructs(&res, m.DB, queries.GOODS_RECEIVED_NOTE_LIST)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *GoodsReceivedNoteModel) GoodsReceivedNoteDetails(grnid int) (models.GoodReceivedNoteSummary, error) {
	var id, orderDate, supplier, warehouse, priceBeforeDiscount, discountType, discountAmount, totalPrice, remarks sql.NullString
	err := m.DB.QueryRow(queries.GOODS_RECEIVED_NOTE_DETAILS, grnid).Scan(&id, &orderDate, &supplier, &warehouse, &priceBeforeDiscount, &discountType, &discountAmount, &totalPrice, &remarks)

	if err != nil {
		return models.GoodReceivedNoteSummary{}, err
	}

	var grnItems []models.GRNItemDetails
	err = mysequel.QueryToStructs(&grnItems, m.DB, queries.GRN_ITEM_DETAILS, grnid)
	if err != nil {
		return models.GoodReceivedNoteSummary{}, err
	}

	return models.GoodReceivedNoteSummary{GRNID: id, OrderDate: orderDate, Supplier: supplier, Warehouse: warehouse, PriceBeforeDiscount: priceBeforeDiscount, DiscountType: discountType, DiscountAmount: discountAmount, TotalPrice: totalPrice, Remarks: remarks, GRNItemDetails: grnItems}, nil
}
