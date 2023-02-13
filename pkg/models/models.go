package models

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type UserResponse struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Token         string `json:"token"`
	WarehouseID   int    `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
}

type User struct {
	ID        int
	GroupID   int
	Username  string
	Password  string
	Name      string
	CreatedAt time.Time
}

type JWTUser struct {
	ID            int
	Username      string
	Password      string
	Name          string
	Type          string
	WarehouseID   int
	WarehouseName string
}

type Dropdown struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AllItemItem struct {
	ID             int     `json:"id"`
	ItemID         string  `json:"item_id"`
	ModelID        string  `json:"model_id"`
	ItemCategoryID string  `json:"item_category_id"`
	PageNo         string  `json:"page_no"`
	ItemNo         string  `json:"item_no"`
	ForeignID      string  `json:"foreign_id"`
	ItemName       string  `json:"name"`
	Price          float64 `json:"price"`
}

type BusinessPartnerBalance struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	BalanceToday float64 `json:"balance_today"`
	Balance      float64 `json:"balance"`
}

type ItemDetails struct {
	ID               int     `json:"id"`
	ItemID           string  `json:"item_id"`
	ModelID          string  `json:"model_id"`
	ModelName        string  `json:"model_name"`
	ItemCategoryID   string  `json:"item_category_id"`
	ItemCategoryName string  `json:"item_category_name"`
	PageNo           string  `json:"page_no"`
	ItemNo           string  `json:"item_no"`
	ForeignID        string  `json:"foreign_id"`
	ItemName         string  `json:"name"`
	Price            float64 `json:"price"`
}

type DropdownAccount struct {
	ID        string `json:"id"`
	AccountID int    `json:"account_id"`
	Name      string `json:"name"`
}

type OrderItemEntry struct {
	ItemID         string `json:"item_id"`
	Quantity       string `json:"qty"`
	UnitPrice      string `json:"unit_price"`
	DiscountType   string `json:"discount_type"`
	DiscountAmount string `json:"discount_amount"`
}

type TransferItem struct {
	ItemID   string `json:"item_id"`
	Quantity string `json:"qty"`
}

type PurchaseOrderEntry struct {
	OrderID    int     `json:"order_id"`
	Supplier   string  `json:"supplier"`
	Warehouse  string  `json:"warehouse"`
	TotalPrice float64 `json:"total_price"`
}

type PurchaseOrderSummary struct {
	OrderID             sql.NullString     `json:"order_id"`
	OrderDate           sql.NullString     `json:"order_date"`
	Supplier            sql.NullString     `json:"supplier"`
	Warehouse           sql.NullString     `json:"warehouse"`
	PriceBeforeDiscount sql.NullString     `json:"price_before_discount"`
	DiscountType        sql.NullString     `json:"discount_type"`
	DiscountAmount      sql.NullString     `json:"discount_amount"`
	TotalPrice          sql.NullString     `json:"total_price"`
	Remarks             sql.NullString     `json:"remarks"`
	OrderItemDetails    []OrderItemDetails `json:"order_item_details"`
}

type OrderItemDetails struct {
	OrderID    sql.NullString `json:"order_id"`
	ItemID     sql.NullString `json:"item_id"`
	ItemName   sql.NullString `json:"item_name"`
	UnitPrice  sql.NullString `json:"unit_price"`
	Quantity   sql.NullString `json:"quantity"`
	TotalPrice sql.NullString `json:"total_price"`
}

type GRNItemEntry struct {
	ItemID         string `json:"item_id"`
	Quantity       string `json:"qty"`
	UnitPrice      string `json:"unit_price"`
	DiscountType   string `json:"discount_type"`
	DiscountAmount string `json:"discount_amount"`
}

type CashInHand struct {
	Amount float64 `json:"amount"`
}

type GoodReceivedNoteEntry struct {
	GRNID      int     `json:"grn_id"`
	Supplier   string  `json:"supplier"`
	Warehouse  string  `json:"warehouse"`
	TotalPrice float64 `json:"total_price"`
}

type InventoryTransferEntry struct {
	ITID              int            `json:"inventory_transfer_id"`
	Created           string         `json:"created"`
	Issuer            string         `json:"issuer"`
	FromWarehouse     string         `json:"from_warehouse"`
	ToWarehouse       string         `json:"to_warehouse"`
	Resolution        sql.NullString `json:"resolution"`
	ResolvedBy        sql.NullString `json:"resolved_by"`
	ResolvedOn        sql.NullString `json:"resolved_on"`
	ResolutionRemarks sql.NullString `json:"resolution_remarks"`
}

type GoodReceivedNoteSummary struct {
	GRNID               sql.NullString   `json:"grn_id"`
	OrderDate           sql.NullString   `json:"order_date"`
	Supplier            sql.NullString   `json:"supplier"`
	Warehouse           sql.NullString   `json:"warehouse"`
	PriceBeforeDiscount sql.NullString   `json:"price_before_discount"`
	DiscountType        sql.NullString   `json:"discount_type"`
	DiscountAmount      sql.NullString   `json:"discount_amount"`
	TotalPrice          sql.NullString   `json:"total_price"`
	Remarks             sql.NullString   `json:"remarks"`
	GRNItemDetails      []GRNItemDetails `json:"grn_item_details"`
}

type GRNItemDetails struct {
	GRNID      sql.NullString `json:"grn_id"`
	ItemID     sql.NullString `json:"item_id"`
	ItemName   sql.NullString `json:"item_name"`
	UnitPrice  sql.NullString `json:"unit_price"`
	Quantity   sql.NullString `json:"quantity"`
	TotalPrice sql.NullString `json:"total_price"`
}

type PurchaseOrderData struct {
	OrderID             sql.NullString  `json:"order_id"`
	SupplierID          sql.NullString  `json:"supplier_id"`
	WarehouseID         sql.NullString  `json:"warehouse_id"`
	DiscountType        sql.NullString  `json:"discount_type"`
	DiscountAmount      sql.NullString  `json:"discount_amount"`
	PriceBeforeDiscount float64         `json:"price_before_discount"`
	TotalPrice          float64         `json:"total_price"`
	OrderItemData       []OrderItemData `json:"order_item_details"`
}

type OrderItemData struct {
	OrderID   sql.NullString `json:"order_id"`
	ItemID    sql.NullString `json:"item_id"`
	UnitPrice sql.NullString `json:"unit_price"`
	Quantity  sql.NullString `json:"quantity"`
}

type LandedCostItemEntry struct {
	CostTypeID string `json:"landed_cost_type_id"`
	Amount     string `json:"amount"`
}

type GRNItemDetailsWithTotal struct {
	GRNID          sql.NullString `json:"grn_id"`
	ItemID         int            `json:"item_id"`
	ToatlCostPrice float64        `json:"total_cost_price"`
	Quantity       float64        `json:"quantity"`
	TotalPrice     float64        `json:"total_price"`
	WarehouseId    int            `json:"warehouse_id"`
}

type WarehouseStockItem struct {
	WarehouseName string  `json:"warehouse_name"`
	ID            int     `json:"id"`
	ItemID        string  `json:"item_id"`
	ForeignID     string  `json:"foreign_id"`
	ItemName      string  `json:"item_name"`
	Quantity      int     `json:"quantity"`
	Price         float64 `json:"price"`
}

type WarehouseStockItemQty struct {
	ItemID   string `json:"item_id"`
	Quantity string `json:"quantity"`
}

type WarehouseItemStockWithDocumentIDs struct {
	WarehouseID         int
	ItemID              int
	GoodsReceivedNoteID int
	InventoryTransferID sql.NullInt32
	Qty                 int
}

type WarehouseItemStockWithDocumentIDsAndPrices struct {
	WarehouseID                 int
	ItemID                      int
	GoodsReceivedNoteID         int
	InventoryTransferID         sql.NullInt32
	Qty                         int
	CostPriceWithoutLandedCosts float64
	CostPrice                   float64
	Price                       float64
}

type PendingInventoryTransfer struct {
	Id      int    `json:"id"`
	Created string `json:"created"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type PendingInventoryTransferItem struct {
	ItemName string `json:"item_name"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"qty"`
}

type InventoryTransferItemForAction struct {
	FromWarehouseID         int
	ToWarehouseID           int
	PrevInventoryTransferID sql.NullInt32
	GoodsReceivedNoteID     int
	ItemID                  int
	Quantity                int
}

type InvoiceSearchItem struct {
	ID                  int     `json:"id"`
	Created             string  `json:"created"`
	Issuer              string  `json:"issuer"`
	IssuingLocation     string  `json:"issuing_location"`
	CostPrice           float64 `json:"cost_price"`
	PriceBeforeDiscount float64 `json:"price_before_discount"`
	Discount            float64 `json:"discount"`
	PriceAfterDiscount  float64 `json:"price_after_discount"`
	CustomerName        string  `json:"customer_name"`
	CustomerContact     string  `json:"customer_contact"`
}

type BPPaymentEntry struct {
	BP     string
	Amount string
}

type BusinessPartnerBalanceDetail struct {
	BPName        string  `json:"bp_name"`
	TransactionID int     `json:"transaction_id"`
	PostingDate   string  `json:"posting_date"`
	EffectiveDate string  `json:"effective_date"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Remark        string  `json:"remark"`
}
