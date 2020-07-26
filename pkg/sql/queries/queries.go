package queries

const ALL_ITEMS = `
	SELECT id, item_id, model_id, item_category_id, page_no, item_no, foreign_id, item_name, price FROM item
`

const ITEM_DETAILS_BY_ITEM_ID = `
	SELECT I.id, I.item_id, I.model_id, M.name AS model_name, I.item_category_id, IC.name AS item_category_name, I.page_no, I.item_no, I.foreign_id, I.item_name, I.price
	FROM item I 
	LEFT JOIN model M ON M.id = I.model_id
	LEFT JOIN item_category IC ON IC.id = I.item_category_id
	WHERE I.item_id = ?
`

const ITEM_DETAILS_BY_ID = `
	SELECT I.id, I.item_id, I.model_id, M.name AS model_name, I.item_category_id, IC.name AS item_category_name, I.page_no, I.item_no, I.foreign_id, I.item_name, I.price
	FROM item I 
	LEFT JOIN model M ON M.id = I.model_id
	LEFT JOIN item_category IC ON IC.id = I.item_category_id
	WHERE I.id = ?
`

const SEARCH_ITEMS = `
	SELECT id, item_id, model_id, item_category_id, page_no, item_no, foreign_id, item_name, price 
	FROM item
	WHERE (? IS NULL OR CONCAT(item_id, foreign_id, item_name) LIKE ?)
`
