package models

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type UserResponse struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Token       string `json:"token"`
	WarehouseID int    `json:"warehouse_id"`
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
	ID          int
	Username    string
	Password    string
	Name        string
	Type        string
	WarehouseID int
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

type JournalEntry struct {
	Account string
	Debit   string
	Credit  string
}

type DropdownAccount struct {
	ID        string `json:"id"`
	AccountID int    `json:"account_id"`
	Name      string `json:"name"`
}

type PaymentVoucherList struct {
	ID          int       `json:"id"`
	Datetime    time.Time `json:"date_time"`
	PostingDate string    `json:"posting_date"`
	FromAccount string    `json:"from_account"`
	User        string    `json:"user"`
}

type PaymentVoucherEntry struct {
	Account string
	Amount  string
}

type LedgerEntry struct {
	Name          string  `json:"account_name"`
	TransactionID int     `json:"transaction_id"`
	PostingDate   string  `json:"posting_date"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	Remark        string  `json:"remark"`
}

type ChartOfAccount struct {
	MainAccountID     int            `json:"main_account_id"`
	MainAccount       string         `json:"main_account"`
	SubAccountID      int            `json:"sub_account_id"`
	SubAccount        string         `json:"sub_account"`
	AccountCategoryID sql.NullInt32  `json:"account_category_id"`
	AccountCategory   sql.NullString `json:"account_category"`
	AccountID         sql.NullInt32  `json:"account_id"`
	AccountName       sql.NullString `json:"account_name"`
}

type PaymentVoucherSummary struct {
	DueDate               sql.NullString          `json:"due_date"`
	CheckNumber           sql.NullString          `json:"check_number"`
	Payee                 sql.NullString          `json:"payee"`
	Remark                sql.NullString          `json:"remark"`
	Account               sql.NullString          `json:"account"`
	Datetime              sql.NullString          `json:"datetime"`
	PaymentVoucherDetails []PaymentVoucherDetails `json:"payment_voucher_details"`
}

type PaymentVoucherDetails struct {
	AccountID   int     `json:"account_id"`
	AccountName string  `json:"account_name"`
	Amount      float64 `json:"amount"`
	PostingDate string  `json:"posting_date"`
}

type Transaction struct {
	TransactionID int     `json:"transaction_id"`
	AccountID     int     `json:"account_id"`
	AccountID2    int     `json:"account_id2"`
	AccountName   string  `json:"account_name"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
}

type TrialEntry struct {
	ID          int     `json:"id"`
	AccountID   int     `json:"account_id"`
	AccountName string  `json:"account_name"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
	Balance     float64 `json:"balance"`
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

type GoodReceivedNoteEntry struct {
	GRNID      int     `json:"grn_id"`
	Supplier   string  `json:"supplier"`
	Warehouse  string  `json:"warehouse"`
	TotalPrice float64 `json:"total_price"`
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
