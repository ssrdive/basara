package main

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	r := mux.NewRouter()
	r.Handle("/", http.HandlerFunc(app.home)).Methods("GET")
	r.HandleFunc("/authenticate", http.HandlerFunc(app.authenticate)).Methods("POST")
	r.Handle("/dropdown/{name}", app.validateToken(http.HandlerFunc(app.dropdownHandler))).Methods("GET")
	r.Handle("/dropdown/condition/{name}/{where}/{value}", app.validateToken(http.HandlerFunc(app.dropdownConditionHandler))).Methods("GET")
	r.Handle("/dropdown/condition/accounts/{name}/{where}/{value}", app.validateToken(http.HandlerFunc(app.dropdownConditionAccountsHandler))).Methods("GET")
	r.Handle("/dropdown/custom/grn", app.validateToken(http.HandlerFunc(app.dropdownGrnHandler))).Methods("GET")
	r.Handle("/dropdown/custom/items", app.validateToken(http.HandlerFunc(app.dropdownItemsHandler))).Methods("GET")
	r.Handle("/dropdown/multicondition/{name}/{where}/{value}/{operator}", app.validateToken(http.HandlerFunc(app.dropdownMultiConditionHandler))).Methods("GET")

	r.Handle("/item/create", app.validateToken(http.HandlerFunc(app.createItem))).Methods("POST")
	r.Handle("/item/all", app.validateToken(http.HandlerFunc(app.allItems))).Methods("GET")
	r.Handle("/item/search", app.validateToken(http.HandlerFunc(app.itemSearch))).Methods("GET")
	r.Handle("/item/{id}", app.validateToken(http.HandlerFunc(app.itemDetails))).Methods("GET")
	r.Handle("/item/details/byid/{id}", app.validateToken(http.HandlerFunc(app.itemDetailsById))).Methods("GET")
	r.Handle("/item/update/byid", app.validateToken(http.HandlerFunc(app.updateItemById))).Methods("POST")
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	r.Handle("/businesspartner/create", app.validateToken(http.HandlerFunc(app.createBusinessPartner))).Methods("POST")
	r.Handle("/businesspartner/balances", app.validateToken(http.HandlerFunc(app.businessPartnerBalances))).Methods("GET")
	r.Handle("/businesspartner/payment", app.validateToken(http.HandlerFunc(app.businessPartnerPayment))).Methods("POST")
	r.Handle("/businesspartner/balance/{bpid}", app.validateToken(http.HandlerFunc(app.bpBalanceDetail))).Methods("GET")

	r.Handle("/account/category/new", app.validateToken(http.HandlerFunc(app.newAccountCategory))).Methods("POST")
	r.Handle("/account/new", app.validateToken(http.HandlerFunc(app.newAccount))).Methods("POST")
	r.Handle("/account/deposit", app.validateToken(http.HandlerFunc(app.accountDeposit))).Methods("POST")
	r.Handle("/account/journalentry", app.validateToken(http.HandlerFunc(app.accountJournalEntry))).Methods("POST")
	r.Handle("/account/paymentvoucher", app.validateToken(http.HandlerFunc(app.accountPaymentVoucher))).Methods("POST")
	r.Handle("/account/ledger/{aid}", app.validateToken(http.HandlerFunc(app.accountLedger))).Methods("GET")
	r.Handle("/account/chart", app.validateToken(http.HandlerFunc(app.accountChart))).Methods("GET")
	r.Handle("/paymentvouchers", app.validateToken(http.HandlerFunc(app.paymentVouchers))).Methods("GET")
	r.Handle("/paymentvoucher/{pid}", app.validateToken(http.HandlerFunc(app.paymentVoucherDetails))).Methods("GET")
	r.Handle("/transaction/{tid}", app.validateToken(http.HandlerFunc(app.accountTransaction))).Methods("GET")
	r.Handle("/account/trialbalance", app.validateToken(http.HandlerFunc(app.accountTrialBalance))).Methods("GET")

	r.Handle("/account/journalentryaudit", app.validateToken(http.HandlerFunc(app.journalEntryAudit))).Methods("GET")
	r.Handle("/account/balancesforreporting", app.validateToken(http.HandlerFunc(app.accountBalancesForReporting))).Methods("GET")
	r.Handle("/account/balancesheetsummary", app.validateToken(http.HandlerFunc(app.balanceSheetSummary))).Methods("GET")
	r.Handle("/account/pnlsummary", app.validateToken(http.HandlerFunc(app.pnlSummary))).Methods("GET")

	r.Handle("/account/cashinhand/{uid}", app.validateToken(http.HandlerFunc(app.cashInHand))).Methods("GET")
	r.Handle("/account/commission/{uid}", app.validateToken(http.HandlerFunc(app.salesCommission))).Methods("GET")

	r.Handle("/transaction/purchaseorder/new", app.validateToken(http.HandlerFunc(app.createOrder))).Methods("POST")
	r.Handle("/transaction/purchaseorder/list", app.validateToken(http.HandlerFunc(app.purchaseOrderList))).Methods("GET")
	r.Handle("/transaction/purchaseorder/{pid}", app.validateToken(http.HandlerFunc(app.purchaseOrderDetails))).Methods("GET")

	r.Handle("/transaction/goodsreceivednote/new", app.validateToken(http.HandlerFunc(app.createGoodsReceivedNote))).Methods("POST")
	r.Handle("/transaction/goodsreceivednote/list", app.validateToken(http.HandlerFunc(app.goodsReceivedNoteList))).Methods("GET")
	r.Handle("/transaction/goodsreceivednote/{grnid}", app.validateToken(http.HandlerFunc(app.goodsReceivedNoteDetails))).Methods("GET")
	r.Handle("/transaction/copypurchaseorder/{pid}", app.validateToken(http.HandlerFunc(app.purchaseOrderData))).Methods("GET")

	r.Handle("/transaction/landedcost/new", app.validateToken(http.HandlerFunc(app.createLandedCost))).Methods("POST")

	r.Handle("/transaction/warehousestock/{wid}", app.validateToken(http.HandlerFunc(app.getWarehouseStock))).Methods("GET")

	r.Handle("/transaction/inventorytransfer/new", app.validateToken(http.HandlerFunc(app.createInventoryTransfer))).Methods("POST")
	r.Handle("/transaction/inventorytransfer/list", app.validateToken(http.HandlerFunc(app.inventoryTransferList))).Methods("GET")
	r.Handle("/transaction/inventorytransfer/{itid}", app.validateToken(http.HandlerFunc(app.inventoryTransferDetails))).Methods("GET")
	r.Handle("/transaction/inventorytransfer/{type}/{warehouse}", app.validateToken(http.HandlerFunc(app.getPendingInventoryTransfers))).Methods("GET")
	r.Handle("/transaction/inventorytransferitems/{itid}", app.validateToken(http.HandlerFunc(app.inventoryTransferItems))).Methods("GET")
	r.Handle("/transaction/inventorytransferaction", app.validateToken(http.HandlerFunc(app.inventoryTransferAction))).Methods("POST")

	r.Handle("/transaction/invoice", app.validateToken(http.HandlerFunc(app.createInvoice))).Methods("POST")
	r.Handle("/transaction/invoice/{iid}", app.validateToken(http.HandlerFunc(app.invoiceDetails))).Methods("GET")

	r.Handle("/reporting/invoicesearch", app.validateToken(http.HandlerFunc(app.invoiceSearch))).Methods("GET")

	r.Handle("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(r))
}
