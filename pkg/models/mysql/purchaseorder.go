package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ssrdive/basara/pkg/models"
	"github.com/ssrdive/basara/pkg/sql/queries"
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

func (m *PurchaseOrderModel) PurchaseOrderList() ([]models.PurchaseOrderEntry, error) {
	var res []models.PurchaseOrderEntry
	err := mysequel.QueryToStructs(&res, m.DB, queries.PURCHASE_ORDER_LIST)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *PurchaseOrderModel) PurchaseOrderDetails(oid int) (models.PurchaseOrderSummary, error) {
	var id, orderDate, supplier, warehouse, priceBeforeDiscount, discountType, discountAmount, totalPrice, remarks sql.NullString
	err := m.DB.QueryRow(queries.PURCHASE_ORDER_DETAILS, oid).Scan(&id, &orderDate, &supplier, &warehouse, &priceBeforeDiscount, &discountType, &discountAmount, &totalPrice, &remarks)
	
	if err != nil {
		fmt.Println(err)
		return models.PurchaseOrderSummary{}, err
	}

	var orderItems []models.OrderItemDetails
	err = mysequel.QueryToStructs(&orderItems, m.DB, queries.PURCHASE_ORDER_ITEM_DETAILS, oid)
	if err != nil {
		return models.PurchaseOrderSummary{}, err
	}

	return models.PurchaseOrderSummary{Order_ID: id, OrderDate: orderDate, Supplier: supplier, Warehouse: warehouse, PriceBeforeDiscount: priceBeforeDiscount, DiscountType: discountType, DiscountAmount:discountAmount, TotalPrice:totalPrice, Remarks:remarks, OrderItemDetails: orderItems}, nil
}


	