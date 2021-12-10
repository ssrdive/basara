package queries

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

const PAYMENT_VOUCHERS = `
	SELECT PV.id, T.datetime, T.posting_date, A.name AS from_account, U.name AS user
	FROM payment_voucher PV
	LEFT JOIN transaction T ON T.id = PV.transaction_id
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id AND AT.type = 'CR'
	LEFT JOIN account A ON A.id = AT.account_id
	LEFT JOIN user U ON T.user_id = U.id
	ORDER BY T.datetime DESC
`

const ACCOUNT_LEDGER = `
	SELECT A.name, AT.transaction_id, DATE_FORMAT(T.posting_date, '%Y-%m-%d') as posting_date, AT.amount, AT.type, T.remark
	FROM account_transaction AT
	LEFT JOIN account A ON A.id = AT.account_id
	LEFT JOIN transaction T ON T.id = AT.transaction_id
	WHERE AT.account_id = ?
`

const CHART_OF_ACCOUNTS = `
	SELECT MA.account_id AS main_account_id, MA.name AS main_account, SA.account_id AS sub_account_id, SA.name AS sub_account, AC.account_id AS account_category_id, AC.name AS account_category, A.account_id, A.name AS account_name
	FROM account A
	RIGHT JOIN account_category AC ON AC.id = A.account_category_id
	RIGHT JOIN sub_account SA ON SA.id = AC.sub_account_id
	RIGHT JOIN main_account MA ON MA.id = SA.main_account_id
`

const PAYMENT_VOUCHER_CHECK_DETAILS = `
	SELECT PV.due_date, PV.check_number, PV.payee, T.remark, A.name AS account_name, T.datetime
	FROM payment_voucher PV
	LEFT JOIN transaction T ON T.id = PV.transaction_id
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id AND AT.type = 'CR'
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE PV.id = ?
`

const PAYMENT_VOUCHER_DETAILS = `
	SELECT A.account_id, A.name AS account_name, AT.amount, DATE(T.posting_date) as posting_date
	FROM payment_voucher PV
	LEFT JOIN transaction T ON T.id = PV.transaction_id
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id AND AT.type = 'DR'
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE PV.id = ?
`

const TRANSACTION = `
	SELECT AT.transaction_id, A.account_id, A.id AS account_id2, A.name AS account_name, AT.type, AT.amount
	FROM account_transaction AT
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE AT.transaction_id = ?
`

const TRIAL_BALANCE = `
	SELECT A.id, A.account_id, A.name, COALESCE(AT.debit, 0) AS debit, COALESCE(AT.credit, 0) AS credit, COALESCE(AT.debit-AT.credit, 0) AS balance
	FROM account A
	LEFT JOIN (
		SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit 
		FROM account_transaction AT
		GROUP BY AT.account_id
	) AT ON AT.account_id = A.id
	ORDER BY account_id ASC
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
	SELECT OI.id, I.name, OI.unit_price, OI.qty, OI.discount_type, OI.discount_amount, OI.price_before_discount, OI.total_price
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
	SELECT GRNI.id, I.name, GRNI.unit_price, GRNI.qty, GRNI.discount_type, GRNI.discount_amount, GRNI.price_before_discount, GRNI.total_price
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
	SELECT OI.id, OI.item_id, OI.unit_price, (OI.qty - (OI.total_reconciled + OI.total_cancelled)) as quantity, OI.discount_type, OI.discount_amount 
	FROM purchase_order_item OI
	WHERE OI.purchase_order_id = ?
`

const GRN_ITEM_DETAILS_WITH_ORDER_TOTAL = `
	SELECT GRNI.id, GRNI.item_id,  GRNI.total_price as total_cost_price,  GRNI.qty, GRN.price_before_discount as total_price, GRN.warehouse_id 
	FROM goods_received_note_item GRNI
	LEFT JOIN item I ON I.id = GRNI.item_id
	LEFT JOIN goods_received_note GRN ON GRN.id= GRNI.goods_received_note_id
	WHERE GRNI.goods_received_note_id = ?
`
