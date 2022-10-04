package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/ssrdive/basara/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// user := app.extractUser(r)

	if app.runtimeEnv == "dev" {
		fmt.Fprintf(w, "It works! [dev]")
	} else {
		fmt.Fprintf(w, "It works!")
	}
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	u, err := app.user.Get(username, password)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["username"] = u.Username
	claims["name"] = u.Name
	claims["exp"] = time.Now().Add(time.Minute * 1440).Unix()

	ts, err := token.SignedString(app.secret)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := models.UserResponse{u.ID, u.Username, u.Name, u.Type, ts, u.WarehouseID, u.WarehouseName}
	js, err := json.Marshal(user)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (app *application) dropdownHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.dropdown.Get(name)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) dropdownConditionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	where := vars["where"]
	value := vars["value"]
	if name == "" || where == "" || value == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.dropdown.ConditionGet(name, where, value)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) dropdownMultiConditionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	where := vars["where"]
	value := vars["value"]
	operator := vars["operator"]

	if name == "" || where == "" || value == "" || operator == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.dropdown.MultiConditionGet(name, where, operator, value)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) dropdownItemsHandler(w http.ResponseWriter, r *http.Request) {

	items, err := app.dropdown.GetItems()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) dropdownGrnHandler(w http.ResponseWriter, r *http.Request) {

	items, err := app.dropdown.GetGrn()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) itemTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Item Test")
}

func (app *application) itemSearch(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	results, err := app.item.Search(search)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) itemDetailsById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.item.DetailsById(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) itemDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.item.Details(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) updateItemById(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"item_id", "name", "item_price"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.item.UpdateById(r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) createBusinessPartner(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "business_partner_type_id", "name", "address", "telephone", "email"}
	optionalParams := []string{}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.businessPartner.Create(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)

}

func (app *application) createItem(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "item_id", "model_id", "item_category_id", "page_no", "item_no", "foreign_id", "name", "price"}
	optionalParams := []string{}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.item.Create(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)

}

func (app *application) allItems(w http.ResponseWriter, r *http.Request) {
	results, err := app.item.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) dropdownConditionAccountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	where := vars["where"]
	value := vars["value"]
	if name == "" || where == "" || value == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.dropdown.ConditionAccountsGet(name, where, value)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) newAccountCategory(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"sub_account_id", "user_id", "account_id", "name"}
	optionalParams := []string{"datetime"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.account.CreateCategory(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) newAccount(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"account_category_id", "user_id", "account_id", "name"}
	optionalParams := []string{"datetime"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.account.CreateAccount(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) accountDeposit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "posting_date", "to_account_id", "amount", "entries", "remark"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	tid, err := app.account.Deposit(r.PostForm.Get("user_id"), r.PostForm.Get("posting_date"), r.PostForm.Get("to_account_id"), r.PostForm.Get("amount"), r.PostForm.Get("entries"), r.PostForm.Get("remark"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%v", tid)
}

func (app *application) accountJournalEntry(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "posting_date", "remark", "entries"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	tid, err := app.account.JournalEntry(r.PostForm.Get("user_id"), r.PostForm.Get("posting_date"), r.PostForm.Get("remark"), r.PostForm.Get("entries"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%v", tid)
}

func (app *application) accountPaymentVoucher(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "posting_date", "from_account_id", "amount", "entries", "remark"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	tid, err := app.account.PaymentVoucher(r.PostForm.Get("user_id"), r.PostForm.Get("posting_date"), r.PostForm.Get("from_account_id"), r.PostForm.Get("amount"), r.PostForm.Get("entries"), r.PostForm.Get("remark"), r.PostForm.Get("due_date"), r.PostForm.Get("check_number"), r.PostForm.Get("payee"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%v", tid)
}

func (app *application) accountLedger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["aid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ledger, err := app.account.Ledger(aid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledger)
}

func (app *application) accountChart(w http.ResponseWriter, r *http.Request) {
	accounts, err := app.account.ChartOfAccounts()
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func (app *application) paymentVouchers(w http.ResponseWriter, r *http.Request) {
	items, err := app.account.PaymentVouchers()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) paymentVoucherDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["pid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.account.PaymentVoucherDetails(pid)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) accountTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tid, err := strconv.Atoi(vars["tid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ledger, err := app.account.Transaction(tid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledger)
}

func (app *application) accountTrialBalance(w http.ResponseWriter, r *http.Request) {
	accounts, err := app.account.TrialBalance()
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func (app *application) journalEntryAudit(w http.ResponseWriter, r *http.Request) {
	entrydate := r.URL.Query().Get("entrydate")
	postingdate := r.URL.Query().Get("postingdate")

	results, err := app.account.JournalEntriesForAudit(entrydate, postingdate)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) pnlSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startdate")
	endDate := r.URL.Query().Get("enddate")

	results, err := app.account.AccountsForPNL(startDate, endDate)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) balanceSheetSummary(w http.ResponseWriter, r *http.Request) {
	postingdate := r.URL.Query().Get("postingdate")

	results, err := app.account.BalanceSheetSummary(postingdate)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) accountBalancesForReporting(w http.ResponseWriter, r *http.Request) {
	postingdate := r.URL.Query().Get("postingdate")

	results, err := app.account.AccountBalancesForReporting(postingdate)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) invoiceSearch(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startdate")
	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	endDate := r.URL.Query().Get("enddate")
	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	officer := r.URL.Query().Get("officer")

	results, err := app.reporting.InvoiceSearch(startDate, endDate, officer)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) createInvoice(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "from_warehouse", "customer_contact", "discount", "items"}
	optionalParams := []string{}

	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.transactions.CreateInvoice(requiredParams, optionalParams, app.fgAPIKey, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) inventoryTransferAction(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"inventory_transfer_id", "user_id", "resolution", "resolution_remarks"}
	optionalParams := []string{}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.transactions.InventoryTransferAction(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) createInventoryTransfer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "from_warehouse_id", "to_warehouse_id", "entries"}
	optionalParams := []string{"remark"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.transactions.CreateInventoryTransfer(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"supplier_id", "warehouse_id", "entries"}
	optionalParams := []string{"remark"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.purchaseOrder.CreatePurchaseOrder(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) getPendingInventoryTransfers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userType := vars["type"]
	wid, err := strconv.Atoi(vars["warehouse"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	pendingTransfers, err := app.transactions.GetPendingTransfers(wid, userType)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendingTransfers)
}

func (app *application) salesCommission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.Atoi(vars["uid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	inventoryTransferItems, err := app.transactions.GetSalesCommission(uid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventoryTransferItems)
}

func (app *application) cashInHand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.Atoi(vars["uid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	inventoryTransferItems, err := app.transactions.GetCashInHand(uid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventoryTransferItems)
}

func (app *application) inventoryTransferItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itid, err := strconv.Atoi(vars["itid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	inventoryTransferItems, err := app.transactions.GetInventoryTransferItems(itid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventoryTransferItems)
}

func (app *application) getWarehouseStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wid, err := strconv.Atoi(vars["wid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	goodsReceivedNote, err := app.transactions.GetWarehouseStock(wid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goodsReceivedNote)
}

func (app *application) purchaseOrderList(w http.ResponseWriter, r *http.Request) {
	orders, err := app.purchaseOrder.PurchaseOrderList()
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (app *application) purchaseOrderDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["pid"])

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	purchaseOrder, err := app.purchaseOrder.PurchaseOrderDetails(pid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(purchaseOrder)

}

func (app *application) createGoodsReceivedNote(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"supplier_id", "warehouse_id", "entries"}
	optionalParams := []string{"remark"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.goodsReceivedNote.CreateGoodsReceivedNote(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) goodsReceivedNoteList(w http.ResponseWriter, r *http.Request) {
	notes, err := app.goodsReceivedNote.GoodsReceivedNotesList()
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (app *application) goodsReceivedNoteDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	grnid, err := strconv.Atoi(vars["grnid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	goodsReceivedNote, err := app.goodsReceivedNote.GoodsReceivedNoteDetails(grnid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goodsReceivedNote)

}

func (app *application) purchaseOrderData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["pid"])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	purchaseOrder, err := app.purchaseOrder.PurchaseOrderData(pid)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(purchaseOrder)

}

func (app *application) createLandedCost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"grn_id", "entries", "user_id"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.landedCost.CreateLandedCost(requiredParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}
