package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	plaidinternal "gooava/internal/plaid"
	"net/http"
	"strconv"

	plaid "github.com/plaid/plaid-go/v44/plaid"
)

/************************************************************************
* Creates a Plaid Link token for the authenticated user. The browser uses
* this short-lived token to open Plaid Link.
*
* status: ✅
************************************************************************/
func PlaidCreateLinkToken(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: strconv.FormatUint(uint64(authUser.Id), 10),
	}

	request := plaid.NewLinkTokenCreateRequest(
		"gooava",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
	)
	request.SetUser(user)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	resp, _, err := plaidinternal.Client().PlaidApi.LinkTokenCreate(r.Context()).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		httpResponse.Error = fmt.Sprintf("could not create link token: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"link_token": resp.GetLinkToken(),
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Exchanges a public_token from Plaid Link for an access_token, stores the
* item and its accounts, then runs the initial transactions sync.
*
* status: ✅
************************************************************************/
func PlaidExchangePublicToken(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var requestBody struct {
		PublicToken     string `json:"public_token"`
		InstitutionName string `json:"institution_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.PublicToken == "" {
		httpResponse.Error = "public_token is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	exchangeRequest := plaid.NewItemPublicTokenExchangeRequest(requestBody.PublicToken)
	exchangeResp, _, err := plaidinternal.Client().PlaidApi.ItemPublicTokenExchange(r.Context()).ItemPublicTokenExchangeRequest(*exchangeRequest).Execute()
	if err != nil {
		httpResponse.Error = fmt.Sprintf("could not exchange public token: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	itemDBID, err := models.SavePlaidItem(authUser.Id, exchangeResp.GetItemId(), exchangeResp.GetAccessToken(), requestBody.InstitutionName)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	accountsRequest := plaid.NewAccountsGetRequest(exchangeResp.GetAccessToken())
	accountsResp, _, err := plaidinternal.Client().PlaidApi.AccountsGet(r.Context()).AccountsGetRequest(*accountsRequest).Execute()
	if err != nil {
		httpResponse.Error = fmt.Sprintf("could not fetch accounts: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	accountRows := make([]models.PlaidAccountRow, 0, len(accountsResp.GetAccounts()))
	for _, account := range accountsResp.GetAccounts() {
		accountRows = append(accountRows, mapAccount(account))
	}

	if err := models.SavePlaidAccounts(authUser.Id, itemDBID, accountRows); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// Initial transactions sync for this item so the list is populated right after linking.
	item, err := models.GetPlaidItemById(authUser.Id, itemDBID)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("bank connected, but could not load it: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	_, _, _, syncErr := runTransactionsSync(r.Context(), item)
	if syncErr != nil {
		httpResponse.Error = fmt.Sprintf("bank connected, but the initial sync failed: %v", syncErr)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]int{
		"accounts": len(accountRows),
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Runs a transactions/sync for the authenticated user and returns the
* number of added/modified/removed transactions.
*
* status: ✅
************************************************************************/
func PlaidSyncTransactions(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	items, err := models.ListPlaidItems(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if len(items) == 0 {
		httpResponse.Error = "no bank connected"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// Sync every connected source; one failing bank should not block the others.
	added, modified, removed, failed := 0, 0, 0, 0
	var firstErr error
	for i := range items {
		a, m, r, syncErr := runTransactionsSync(r.Context(), &items[i])
		if syncErr != nil {
			failed++
			if firstErr == nil {
				firstErr = syncErr
			}
			continue
		}
		added += a
		modified += m
		removed += r
	}

	if failed == len(items) {
		httpResponse.Error = fmt.Sprintf("%v", firstErr)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]int{
		"added":    added,
		"modified": modified,
		"removed":  removed,
		"failed":   failed,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Lists the authenticated user's synced transactions, newest first.
* Optional ?limit= query param (default 50, max 200).
*
* status: ✅
************************************************************************/
func PlaidListTransactions(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	limit := 50
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 200 {
		limit = 200
	}

	transactions, err := models.ListPlaidTransactions(authUser.Id, limit)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = transactions
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Lists the authenticated user's connected bank accounts with their last
* known balances.
*
* status: ✅
************************************************************************/
func PlaidListAccounts(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	accounts, err := models.ListPlaidAccounts(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = accounts
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Lists the authenticated user's connected Plaid items (bank sources).
* Access tokens are never exposed (json:"-" on the struct field).
*
* status: ✅
************************************************************************/
func PlaidListItems(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	items, err := models.ListPlaidItems(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = items
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Reports whether the authenticated user has a bank connected.
*
* status: ✅
************************************************************************/
func PlaidStatus(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	item, err := models.GetPlaidItemByUser(authUser.Id)
	if err != nil && err != sql.ErrNoRows {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	connected := err == nil
	institutionName := ""
	if connected {
		institutionName = item.InstitutionName
	}

	httpResponse.Data = map[string]interface{}{
		"connected":        connected,
		"institution_name": institutionName,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/**************************************************************************************
* runTransactionsSync pages through /transactions/sync for one Plaid item with its
* stored cursor (empty cursor = full history), applies added/modified/removed to the
* database, refreshes account balances, and persists the final cursor.
**************************************************************************************/
func runTransactionsSync(ctx context.Context, item *models.PlaidItem) (addedN, modifiedN, removedN int, err error) {
	userID := item.UserID
	cursor := item.Cursor
	added := make([]models.PlaidTransactionRow, 0)
	modified := make([]models.PlaidTransactionRow, 0)
	removedIDs := make([]string, 0)
	var syncedAccounts []plaid.AccountBase

	hasMore := true
	for hasMore {
		request := plaid.NewTransactionsSyncRequest(item.AccessToken)
		if cursor != "" {
			request.SetCursor(cursor)
		}

		resp, _, apiErr := plaidinternal.Client().PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*request).Execute()
		if apiErr != nil {
			return 0, 0, 0, fmt.Errorf("plaid sync failed: %v", apiErr)
		}

		for _, txn := range resp.GetAdded() {
			added = append(added, mapTransaction(txn))
		}
		for _, txn := range resp.GetModified() {
			modified = append(modified, mapTransaction(txn))
		}
		for _, removed := range resp.GetRemoved() {
			removedIDs = append(removedIDs, removed.GetTransactionId())
		}
		// every sync page carries the accounts with their latest balances
		syncedAccounts = resp.GetAccounts()

		hasMore = resp.GetHasMore()
		cursor = resp.GetNextCursor()
	}

	if len(syncedAccounts) > 0 {
		accountRows := make([]models.PlaidAccountRow, 0, len(syncedAccounts))
		for _, account := range syncedAccounts {
			accountRows = append(accountRows, mapAccount(account))
		}
		if err := models.SavePlaidAccounts(userID, item.Id, accountRows); err != nil {
			return 0, 0, 0, err
		}
	}

	if err := models.UpsertPlaidTransactions(userID, append(added, modified...)); err != nil {
		return 0, 0, 0, err
	}
	if err := models.DeletePlaidTransactions(userID, removedIDs); err != nil {
		return 0, 0, 0, err
	}
	if err := models.SavePlaidCursor(item.Id, cursor); err != nil {
		return 0, 0, 0, err
	}

	return len(added), len(modified), len(removedIDs), nil
}

/**************************************************************************************
* mapTransaction converts a plaid-go Transaction into our storage row.
**************************************************************************************/
func mapTransaction(txn plaid.Transaction) models.PlaidTransactionRow {
	category := ""
	if txn.PersonalFinanceCategory.IsSet() {
		if pfc := txn.PersonalFinanceCategory.Get(); pfc != nil {
			category = pfc.GetPrimary()
		}
	}

	isoCurrencyCode := txn.GetIsoCurrencyCode()
	if isoCurrencyCode == "" {
		isoCurrencyCode = "USD"
	}

	return models.PlaidTransactionRow{
		TransactionID:   txn.GetTransactionId(),
		AccountID:       txn.GetAccountId(),
		Name:            txn.GetName(),
		MerchantName:    txn.GetMerchantName(),
		Amount:          txn.GetAmount(),
		IsoCurrencyCode: isoCurrencyCode,
		Date:            txn.GetDate(),
		Category:        category,
		Pending:         txn.GetPending(),
	}
}

/**************************************************************************************
* mapAccount converts a plaid-go AccountBase into our storage row.
**************************************************************************************/
func mapAccount(account plaid.AccountBase) models.PlaidAccountRow {
	balances := account.GetBalances()

	var availableBalance *float64
	if balances.Available.IsSet() {
		value := balances.GetAvailable()
		availableBalance = &value
	}

	currentBalance := 0.0
	if balances.Current.IsSet() {
		currentBalance = balances.GetCurrent()
	}

	isoCurrencyCode := balances.GetIsoCurrencyCode()
	if isoCurrencyCode == "" {
		isoCurrencyCode = "USD"
	}

	return models.PlaidAccountRow{
		AccountID:        account.GetAccountId(),
		Name:             account.GetName(),
		Mask:             account.GetMask(),
		Type:             string(account.GetType()),
		Subtype:          string(account.GetSubtype()),
		IsoCurrencyCode:  isoCurrencyCode,
		CurrentBalance:   currentBalance,
		AvailableBalance: availableBalance,
	}
}
