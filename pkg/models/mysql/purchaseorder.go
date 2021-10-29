package mysql

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/mysequel"
)

// PurchaseOrderModel struct holds database instance
type PurchaseOrderModel struct {
	DB *sql.DB
}

// CreatePurchaseOrder creats an purchase order
func (m *PurchaseOrderModel) CreatePurchaseOrder(rparams, oparams []string, form url.Values) (int64, error) {
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
	var orderItem []models.OrderItemEntry
	json.Unmarshal([]byte(entities), &orderItem)
	
	oid, err := mysequel.Insert(mysequel.Table{
		TableName: "purchase_order",
		Columns:   []string{"user_id", "supplier_id", "warehouse_id", "discount_type", "discount_amount", "price_before_discount", "total_price", "remarks"},
		Vals:      []interface{}{form.Get("user_id"), form.Get("supplier_id"), form.Get("warehouse_id"), form.Get("discount_type"), form.Get("discount_amount") , form.Get("price_before_discount"), form.Get("total_price"), form.Get("remark")},
		Tx:        tx,
	})

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, entry := range orderItem {	

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

		discountAmount, err := strconv.ParseFloat(entry.DiscountAmount, 32)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		var totalPrice float64 
		if entry.DiscountType == "per" {
			totalPrice = unitPrice * (100 - discountAmount) * quantity / 100 ;
		} else {
			totalPrice = (unitPrice - discountAmount) * quantity ;
		}

		_, err = mysequel.Insert(mysequel.Table{
			TableName: "purchase_order_item",
			Columns:   []string{"purchase_order_id", "item_id", "unit_price", "qty", "discount_type", "discount_amount", "price_before_discount", "total_price"},
			Vals:      []interface{}{oid, entry.ItemID, unitPrice, quantity, entry.DiscountType, discountAmount,  unitPrice * quantity, totalPrice},
			Tx:        tx,
		})
		
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	return oid, nil
}
