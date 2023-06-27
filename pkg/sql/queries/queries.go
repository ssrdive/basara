package queries

import "fmt"

const AllItems = `
	SELECT id, item_id, model_id, item_category_id, page_no, item_no, foreign_id, name, price FROM item
`

const ItemDetailsByItemId = `
	SELECT I.id, I.item_id, I.model_id, M.name AS model_name, I.item_category_id, IC.name AS item_category_name, I.page_no, I.item_no, I.foreign_id, I.name, I.price
	FROM item I 
	LEFT JOIN model M ON M.id = I.model_id
	LEFT JOIN item_category IC ON IC.id = I.item_category_id
	WHERE I.item_id = ?
`

const ItemDetailsById = `
	SELECT I.id, I.item_id, I.model_id, M.name AS model_name, I.item_category_id, IC.name AS item_category_name, I.page_no, I.item_no, I.foreign_id, I.name, I.price
	FROM item I 
	LEFT JOIN model M ON M.id = I.model_id
	LEFT JOIN item_category IC ON IC.id = I.item_category_id
	WHERE I.id = ?
`

const SearchItems = `
	SELECT id, item_id, model_id, item_category_id, page_no, item_no, foreign_id, name, price 
	FROM item
	WHERE (? IS NULL OR CONCAT(item_id, foreign_id, name) LIKE ?)
`

const PurchaseOrderList = `
	SELECT PO.id, BP.name, BP2.name, PO.total_price
	FROM purchase_order PO
	LEFT JOIN business_partner BP ON BP.id = PO.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = PO.warehouse_id
	ORDER BY PO.id ASC
`

const PurchaseOrderDetails = `
	SELECT PO.id, PO.created,  BP.name as supplier, BP2.name as warehouse, PO.price_before_discount, PO.discount_type, PO.discount_amount, PO.total_price, PO.remarks
	FROM purchase_order PO
	LEFT JOIN business_partner BP ON BP.id = PO.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = PO.warehouse_id
	WHERE PO.id = ?
	ORDER BY PO.id ASC
`

const PurchaseOrderItemDetails = `
	SELECT OI.id, I.item_id AS item_id, I.name, OI.unit_price, OI.qty, OI.total_price
	FROM purchase_order_item OI
	LEFT JOIN item I ON I.id = OI.item_id
	WHERE OI.purchase_order_id = ?
`

const PurchaseOrderItemCount = `
	SELECT OI.id, OI.qty, OI.total_reconciled, OI.total_cancelled
	FROM purchase_order_item OI
	LEFT JOIN item I ON I.id = OI.item_id
	WHERE OI.item_id = ? AND OI.purchase_order_id = ?
`

const GoodsReceivedNoteList = `
	SELECT GRN.id, BP.name, BP2.name, GRN.total_price
	FROM goods_received_note GRN
	LEFT JOIN business_partner BP ON BP.id = GRN.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = GRN.warehouse_id
	ORDER BY GRN.id ASC
`

const InventoryTransferList = `
	SELECT IT.id, IT.created, U.name as issuer, BP.name AS from_warehouse, BP2.name AS to_warehouse, IT.resolution, U2.name as resolved_by, IT.resolved_on, IT.resolution_remarks
	FROM inventory_transfer IT
	LEFT JOIN user U ON U.id = IT.user_id
	LEFT JOIN business_partner BP ON IT.from_warehouse_id = BP.id
	LEFT JOIN business_partner BP2 ON IT.to_warehouse_id = BP2.id
	LEFT JOIN user U2 ON U2.id = IT.resolved_by
`

const GoodsReceivedNoteDetails = `
	SELECT GRN.id, GRN.created,  BP.name as supplier, BP2.name as warehouse, GRN.price_before_discount, GRN.discount_type, GRN.discount_amount, GRN.total_price, GRN.remarks
	FROM goods_received_note GRN
	LEFT JOIN business_partner BP ON BP.id = GRN.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = GRN.warehouse_id
	WHERE GRN.id = ?
	ORDER BY GRN.id ASC
`

const GrnItemDetails = `
	SELECT GRNI.id, I.item_id as item_id, I.name, GRNI.unit_price, GRNI.qty, GRNI.total_price
	FROM goods_received_note_item GRNI
	LEFT JOIN item I ON I.id = GRNI.item_id
	WHERE GRNI.goods_received_note_id = ?
`

const InvoiceDetails = `
	SELECT invoice.id, U.name AS issued_by, BP.name AS warehouse, price_before_discount, discount, price_after_discount, customer_name, customer_contact
	FROM invoice
	LEFT JOIN user U ON U.id = invoice.user_id
	LEFT JOIN business_partner BP ON BP.id = invoice.warehouse_id
	WHERE invoice.id = ?
`

const InvoiceItemDetails = `
	SELECT II.item_id, I.item_id, I.name, SUM(II.qty), II.price
	FROM invoice_item II
	LEFT JOIN item I ON II.item_id = I.id
	WHERE invoice_id = ?
	GROUP BY II.item_id, II.price
`

const PurchaseOrderData = `
	SELECT PO.id,  PO.supplier_id , PO.warehouse_id, PO.discount_type, PO.discount_amount
	FROM purchase_order PO
	WHERE PO.id = ?
	ORDER BY PO.id ASC
`

const PurchaseOrderItemData = `
	SELECT OI.id, OI.item_id, OI.unit_price, (OI.qty - (OI.total_reconciled + OI.total_cancelled)) as quantity
	FROM purchase_order_item OI
	WHERE OI.purchase_order_id = ?
	AND (OI.qty - (OI.total_reconciled + OI.total_cancelled)) != 0
`

const GrnItemDetailsWithOrderTotal = `
	SELECT GRNI.id, GRNI.item_id,  GRNI.total_price as total_cost_price,  GRNI.qty, GRN.price_before_discount as total_price, GRN.warehouse_id 
	FROM goods_received_note_item GRNI
	LEFT JOIN item I ON I.id = GRNI.item_id
	LEFT JOIN goods_received_note GRN ON GRN.id= GRNI.goods_received_note_id
	WHERE GRNI.goods_received_note_id = ?
`

const WarehouseStock = `
	SELECT BP.name AS warehouse_name, I.id, I.item_id, I.foreign_id, I.name AS item_name, SUM(CS.qty) AS quantity, I.price
	FROM current_stock CS
	LEFT JOIN item I ON I.id = CS.item_id
	LEFT JOIN business_partner BP ON BP.id = CS.warehouse_id
	WHERE CS.warehouse_id = ?
	GROUP BY warehouse_name, I.id, item_id, foreign_id, item_name, I.price
	HAVING SUM(CS.qty) > 0
`

func WarehosueItemQty(warehouseID, itemIDs interface{}) string {
	return fmt.Sprintf(`
		SELECT CS.item_id, SUM(CS.qty) AS quantity
		FROM current_stock CS
		WHERE CS.warehouse_id = %v AND CS.item_id IN (%v)
		GROUP BY CS.item_id`,
		warehouseID, itemIDs)
}

const WarehouseItemStockWithDocumentIds = `
	SELECT CS.warehouse_id, CS.item_id, CS.goods_received_note_id, CS.inventory_transfer_id, CS.qty
	FROM current_stock CS
	LEFT JOIN goods_received_note GRN ON GRN.id = CS.goods_received_note_id
	WHERE CS.warehouse_id = ? AND CS.item_id = ?
	ORDER BY GRN.created ASC
`

const WarehouseItemStockWithDocumentIdsAndPrices = `
	SELECT CS.warehouse_id, CS.item_id, CS.goods_received_note_id, CS.inventory_transfer_id, CS.qty, 
	CS.cost_price AS cost_price_without_landed_costs, CS.price AS cost_price, I.price
	FROM current_stock CS
	LEFT JOIN goods_received_note GRN ON GRN.id = CS.goods_received_note_id
	LEFT JOIN item I ON I.id = CS.item_id
	WHERE CS.warehouse_id = ? AND CS.item_id = ?
	ORDER BY GRN.created ASC
`

const GetPendingTransfers = `
	SELECT IT.id, IT.created, FBP.name AS from_warehouse, TBP.name AS to_warehouse
	FROM inventory_transfer IT
	LEFT JOIN business_partner FBP ON FBP.id = IT.from_warehouse_id
	LEFT JOIN business_partner TBP ON TBP.id = IT.to_warehouse_id
	WHERE resolution IS NULL
`

const GetPendingTransfersByWarehouse = `
	SELECT IT.id, IT.created, FBP.name AS from_warehouse, TBP.name AS to_warehouse
	FROM inventory_transfer IT
	LEFT JOIN business_partner FBP ON FBP.id = IT.from_warehouse_id
	LEFT JOIN business_partner TBP ON TBP.id = IT.to_warehouse_id
	WHERE resolution IS NULL AND IT.to_warehouse_id = ?
`

const InventoryTransferDetails = `
	SELECT IT.id, IT.created, ISS_USR.name AS issued_by, FR_WH.name AS from_warehouse,
       TO_WH.name AS to_warehouse, IT.resolution, RSV_USR.name AS resolved_by,
        resolved_on, resolution_remarks
	FROM inventory_transfer IT
	LEFT JOIN user ISS_USR ON IT.user_id = ISS_USR.id
	LEFT JOIN business_partner FR_WH ON IT.from_warehouse_id = FR_WH.id
	LEFT JOIN business_partner TO_WH ON IT.to_warehouse_id = TO_WH.id
	LEFT JOIN user RSV_USR ON IT.resolved_by = RSV_USR.id
	WHERE IT.id = ?
`

const InventoryTransferItemsDetails = `
	SELECT I.id, I.item_id, I.name AS item_name, SUM(ITI.qty) AS qty
	FROM inventory_transfer_item ITI
	LEFT JOIN item I ON ITI.item_id = I.id
	WHERE ITI.inventory_transfer_id = ?
	GROUP BY I.name, I.item_id

`

const InventoryTransferItems = `
	SELECT I.name AS item_name, I.item_id, SUM(ITI.qty) AS qty
	FROM inventory_transfer_item ITI
	LEFT JOIN item I ON ITI.item_id = I.id
	WHERE ITI.inventory_transfer_id = ?
	GROUP BY I.name, I.item_id
`

const InventoryTransferItemsForAction = `
	SELECT IT.from_warehouse_id, IT.to_warehouse_id, ITI.prev_inventory_transfer_id, ITI.goods_received_note_id, item_id, qty
	FROM inventory_transfer_item ITI
	LEFT JOIN inventory_transfer IT ON IT.id = ITI.inventory_transfer_id
	WHERE ITI.inventory_transfer_id = ?
`

const OfficerAccNo = `
	SELECT account_id FROM user WHERE id = ?
`

const GetCashInHand = `
	SELECT COALESCE(AT.debit-AT.credit, 0) AS balance
	FROM account A
	LEFT JOIN (
		SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit 
		FROM account_transaction AT
        WHERE AT.account_id = (SELECT account_id FROM user WHERE id = ?)
		GROUP BY AT.account_id
	) AT ON AT.account_id = A.id
	WHERE AT.account_id = (SELECT account_id FROM user WHERE id = ?)
`

const GetSalesCommission = `
	SELECT ROUND(COALESCE(SUM(price_after_discount-cost_price)*0.025, 0), 2)
	FROM invoice WHERE YEAR(created) = YEAR(NOW()) AND MONTH(created) = MONTH(NOW()) AND invoice_type_id = 1 AND user_id = ?
`

const InvoiceSearch = `
	SELECT I.id, I.created, U.name AS issuer, BP.name AS issuing_location, cost_price, price_before_discount, discount, price_after_discount, customer_name, customer_contact
	FROM invoice I
	LEFT JOIN user U ON U.id = I.user_id
	LEFT JOIN business_partner BP ON BP.id = I.warehouse_id
	WHERE (? IS NULL OR I.user_id = ?) AND DATE(I.created) BETWEEN ? AND ?
`

const BusinessPartnerBalances = `
	SELECT BP.id, BP.name AS business_partner, (SUM(CASE WHEN BPF.type = "DR" AND effective_date <= DATE(NOW()) THEN BPF.amount ELSE 0 END) - SUM(CASE WHEN BPF.type = "CR" AND effective_date <= DATE(NOW()) THEN BPF.amount ELSE 0 END)) AS balance_today, (SUM(CASE WHEN BPF.type = "DR" THEN BPF.amount ELSE 0 END) - SUM(CASE WHEN BPF.type = "CR" THEN BPF.amount ELSE 0 END)) AS balance
	FROM business_partner_financial BPF
	LEFT JOIN business_partner BP on BPF.business_partner_id = BP.id
	GROUP BY BPF.business_partner_id
`

const BusinessPartnerBalanceDetail = `
	SELECT BP.name AS business_partner_name, T.id AS transaction_id, DATE_FORMAT(T.posting_date, '%Y-%m-%d') AS posting_date, DATE_FORMAT(BPF.effective_date, '%Y-%m-%d') AS effective_date, BPF.type, BPF.amount, T.remark
	FROM business_partner_financial BPF
	LEFT JOIN transaction T on BPF.transaction_id = T.id
	LEFT JOIN business_partner BP on BPF.business_partner_id = BP.id
	WHERE BPF.business_partner_id = ?
	ORDER BY BPF.effective_date
`

const ItemStock = `
	SELECT BP.name AS warehouse, I.item_id, I.name, SUM(CS.qty) AS qty, SUM(CS.float_qty) AS float_qty
	FROM current_stock CS
	LEFT JOIN item I ON I.id = CS.item_id
	LEFT JOIN business_partner BP ON BP.id = CS.warehouse_id
	WHERE CS.item_id = ?
	GROUP BY CS.warehouse_id, I.item_id, I.name
	HAVING SUM(CS.qty) > 0 OR SUM(CS.float_qty) > 0
`
