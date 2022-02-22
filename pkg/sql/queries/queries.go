package queries

import "fmt"

const ALL_ITEMS = `
	SELECT id, item_id, model_id, item_category_id, page_no, item_no, foreign_id, name, price FROM item
`

const ITEM_DETAILS_BY_ITEM_ID = `
	SELECT I.id, I.item_id, I.model_id, M.name AS model_name, I.item_category_id, IC.name AS item_category_name, I.page_no, I.item_no, I.foreign_id, I.name, I.price
	FROM item I 
	LEFT JOIN model M ON M.id = I.model_id
	LEFT JOIN item_category IC ON IC.id = I.item_category_id
	WHERE I.item_id = ?
`

const ITEM_DETAILS_BY_ID = `
	SELECT I.id, I.item_id, I.model_id, M.name AS model_name, I.item_category_id, IC.name AS item_category_name, I.page_no, I.item_no, I.foreign_id, I.name, I.price
	FROM item I 
	LEFT JOIN model M ON M.id = I.model_id
	LEFT JOIN item_category IC ON IC.id = I.item_category_id
	WHERE I.id = ?
`

const SEARCH_ITEMS = `
	SELECT id, item_id, model_id, item_category_id, page_no, item_no, foreign_id, name, price 
	FROM item
	WHERE (? IS NULL OR CONCAT(item_id, foreign_id, name) LIKE ?)
`

const PURCHASE_ORDER_LIST = `
	SELECT PO.id, BP.name, BP2.name, PO.total_price
	FROM purchase_order PO
	LEFT JOIN business_partner BP ON BP.id = PO.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = PO.warehouse_id
	ORDER BY PO.id ASC
`

const PURCHASE_ORDER_DETAILS = `
	SELECT PO.id, PO.created,  BP.name as supplier, BP2.name as warehouse, PO.price_before_discount, PO.discount_type, PO.discount_amount, PO.total_price, PO.remarks
	FROM purchase_order PO
	LEFT JOIN business_partner BP ON BP.id = PO.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = PO.warehouse_id
	WHERE PO.id = ?
	ORDER BY PO.id ASC
`

const PURCHASE_ORDER_ITEM_DETAILS = `
	SELECT OI.id, I.name, OI.unit_price, OI.qty, OI.total_price
	FROM purchase_order_item OI
	LEFT JOIN item I ON I.id = OI.item_id
	WHERE OI.purchase_order_id = ?
`

const PURCHASE_ORDER_ITEM_COUNT = `
	SELECT OI.id, OI.qty, OI.total_reconciled, OI.total_cancelled
	FROM purchase_order_item OI
	LEFT JOIN item I ON I.id = OI.item_id
	WHERE OI.item_id = ? AND OI.purchase_order_id = ?
`

const GOODS_RECEIVED_NOTE_LIST = `
	SELECT GRN.id, BP.name, BP2.name, GRN.total_price
	FROM goods_received_note GRN
	LEFT JOIN business_partner BP ON BP.id = GRN.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = GRN.warehouse_id
	ORDER BY GRN.id ASC
`

const GOODS_RECEIVED_NOTE_DETAILS = `
	SELECT GRN.id, GRN.created,  BP.name as supplier, BP2.name as warehouse, GRN.price_before_discount, GRN.discount_type, GRN.discount_amount, GRN.total_price, GRN.remarks
	FROM goods_received_note GRN
	LEFT JOIN business_partner BP ON BP.id = GRN.supplier_id
	LEFT JOIN business_partner BP2 ON BP2.id = GRN.warehouse_id
	WHERE GRN.id = ?
	ORDER BY GRN.id ASC
`

const GRN_ITEM_DETAILS = `
	SELECT GRNI.id, I.name, GRNI.unit_price, GRNI.qty, GRNI.total_price
	FROM goods_received_note_item GRNI
	LEFT JOIN item I ON I.id = GRNI.item_id
	WHERE GRNI.goods_received_note_id = ?
`

const PURCHASE_ORDER_DATA = `
	SELECT PO.id,  PO.supplier_id , PO.warehouse_id, PO.discount_type, PO.discount_amount
	FROM purchase_order PO
	WHERE PO.id = ?
	ORDER BY PO.id ASC
`

const PURCHASE_ORDER_ITEM_DATA = `
	SELECT OI.id, OI.item_id, OI.unit_price, (OI.qty - (OI.total_reconciled + OI.total_cancelled)) as quantity
	FROM purchase_order_item OI
	WHERE OI.purchase_order_id = ?
	AND (OI.qty - (OI.total_reconciled + OI.total_cancelled)) != 0
`

const GRN_ITEM_DETAILS_WITH_ORDER_TOTAL = `
	SELECT GRNI.id, GRNI.item_id,  GRNI.total_price as total_cost_price,  GRNI.qty, GRN.price_before_discount as total_price, GRN.warehouse_id 
	FROM goods_received_note_item GRNI
	LEFT JOIN item I ON I.id = GRNI.item_id
	LEFT JOIN goods_received_note GRN ON GRN.id= GRNI.goods_received_note_id
	WHERE GRNI.goods_received_note_id = ?
`

const WAREHOUSE_STOCK = `
	SELECT BP.name AS warehouse_name, I.name AS item_name, SUM(CS.qty) AS quantity, I.price
	FROM current_stock CS
	LEFT JOIN item I ON I.id = CS.item_id
	LEFT JOIN business_partner BP ON BP.id = CS.warehouse_id
	WHERE CS.warehouse_id = ?
	GROUP BY warehouse_name, item_name, I.price
	HAVING SUM(CS.qty) > 0
`

func WAREHOSUE_ITEM_QTY(warehouseID, itemIDs interface{}) string {
	return fmt.Sprintf(`
		SELECT CS.item_id, SUM(CS.qty) AS quantity
		FROM current_stock CS
		WHERE CS.warehouse_id = %v AND CS.item_id IN (%v)
		GROUP BY CS.item_id`,
		warehouseID, itemIDs)
}

const WAREHOUSE_ITEM_STOCK_WITH_DOCUMENT_IDS = `
	SELECT CS.warehouse_id, CS.item_id, CS.goods_received_note_id, CS.inventory_transfer_id, CS.qty
	FROM current_stock CS
	LEFT JOIN goods_received_note GRN ON GRN.id = CS.goods_received_note_id
	WHERE CS.warehouse_id = ? AND CS.item_id = ?
	ORDER BY GRN.created ASC
`

const WAREHOUSE_ITEM_STOCK_WITH_DOCUMENT_IDS_AND_PRICES = `
	SELECT CS.warehouse_id, CS.item_id, CS.goods_received_note_id, CS.inventory_transfer_id, CS.qty, CS.price AS cost_price, I.price
	FROM current_stock CS
	LEFT JOIN goods_received_note GRN ON GRN.id = CS.goods_received_note_id
	LEFT JOIN item I ON I.id = CS.item_id
	WHERE CS.warehouse_id = ? AND CS.item_id = ?
	ORDER BY GRN.created ASC
`

const GET_PENDING_TRANSFERS = `
	SELECT IT.id, IT.created, FBP.name AS from_warehouse, TBP.name AS to_warehouse
	FROM inventory_transfer IT
	LEFT JOIN business_partner FBP ON FBP.id = IT.from_warehouse_id
	LEFT JOIN business_partner TBP ON TBP.id = IT.to_warehouse_id
	WHERE resolution IS NULL
`

const GET_PENDING_TRANSFERS_BY_WAREHOUSE = `
	SELECT IT.id, IT.created, FBP.name AS from_warehouse, TBP.name AS to_warehouse
	FROM inventory_transfer IT
	LEFT JOIN business_partner FBP ON FBP.id = IT.from_warehouse_id
	LEFT JOIN business_partner TBP ON TBP.id = IT.to_warehouse_id
	WHERE resolution IS NULL AND IT.to_warehouse_id = ?
`

const INVENTORY_TRANSFER_ITEMS = `
	SELECT I.name AS item_name, I.item_id, SUM(ITI.qty) AS qty
	FROM inventory_transfer_item ITI
	LEFT JOIN item I ON ITI.item_id = I.id
	WHERE ITI.inventory_transfer_id = ?
	GROUP BY I.name, I.item_id
`

const INVENTORY_TRANSFER_ITEMS_FOR_ACTION = `
	SELECT IT.from_warehouse_id, IT.to_warehouse_id, ITI.prev_inventory_transfer_id, ITI.goods_received_note_id, item_id, qty
	FROM inventory_transfer_item ITI
	LEFT JOIN inventory_transfer IT ON IT.id = ITI.inventory_transfer_id
	WHERE ITI.inventory_transfer_id = ?
`
