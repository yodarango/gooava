package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"gooava/internal/utils"
	"net/http"
	"os"
	"time"
)

func Router () http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc(constants.ROUTE_GET_AUTH_SAMPLE, models.Authenticate(SampleAuth))
	mux.HandleFunc(constants.ROUTE_GET_PUBLIC_SAMPLE, SamplePub)

	// auth routes (no authentication required) — rate limited per client IP.
	// These endpoints guard every account: passwords and 6-digit verification
	// codes must be too slow to brute force.
	loginLimiter := newRateLimiter(10, time.Minute)
	signupLimiter := newRateLimiter(5, 15*time.Minute)
	forgotLimiter := newRateLimiter(5, 15*time.Minute)
	verifyLimiter := newRateLimiter(10, 15*time.Minute)

	mux.HandleFunc(constants.ROUTE_POST_CHANGE_PASSWORD, models.Authenticate(ChangePassword))
	mux.HandleFunc(constants.ROUTE_POST_UPDATE_PROFILE, models.Authenticate(UpdateProfile))
	mux.HandleFunc(constants.ROUTE_POST_FORGOT_PASSWORD, forgotLimiter.wrap(ForgotPassword))
	mux.HandleFunc(constants.ROUTE_POST_VERIFY_EMAIL, verifyLimiter.wrap(VerifyEmail))
	mux.HandleFunc(constants.ROUTE_POST_VERIFY_TWO_FA, verifyLimiter.wrap(VerifyTwoFA))
	mux.HandleFunc(constants.ROUTE_POST_SIGNUP, signupLimiter.wrap(Signup))
	mux.HandleFunc(constants.ROUTE_POST_LOGIN, loginLimiter.wrap(Login))

	// 2FA settings (authentication required)
	mux.HandleFunc(constants.ROUTE_POST_TWO_FA, models.Authenticate(ToggleTwoFA))

	// plaid routes (authentication required)
	mux.HandleFunc(constants.ROUTE_POST_PLAID_LINK_TOKEN, models.Authenticate(PlaidCreateLinkToken))
	mux.HandleFunc(constants.ROUTE_POST_PLAID_EXCHANGE, models.Authenticate(PlaidExchangePublicToken))
	mux.HandleFunc(constants.ROUTE_POST_PLAID_SYNC, models.Authenticate(PlaidSyncTransactions))
	mux.HandleFunc(constants.ROUTE_GET_PLAID_TRANSACTIONS, models.Authenticate(PlaidListTransactions))
	mux.HandleFunc(constants.ROUTE_GET_PLAID_ACCOUNTS, models.Authenticate(PlaidListAccounts))
	mux.HandleFunc(constants.ROUTE_GET_PLAID_ITEMS, models.Authenticate(PlaidListItems))
	mux.HandleFunc(constants.ROUTE_GET_PLAID_STATUS, models.Authenticate(PlaidStatus))

	// expense routes (authentication required)
	// GET and POST share some paths, so those are method-qualified patterns (Go 1.22+ mux)
	mux.HandleFunc("POST "+constants.ROUTE_POST_EXPENSES, models.Authenticate(CreateExpense))
	mux.HandleFunc("GET "+constants.ROUTE_GET_EXPENSES, models.Authenticate(ListExpenses))
	mux.HandleFunc(constants.ROUTE_POST_EXPENSES_UPDATE, models.Authenticate(UpdateExpense))
	mux.HandleFunc(constants.ROUTE_POST_EXPENSES_DELETE, models.Authenticate(DeleteExpense))
	mux.HandleFunc("GET "+constants.ROUTE_GET_EXPENSE_CATEGORIES, models.Authenticate(ListExpenseCategories))
	mux.HandleFunc("POST "+constants.ROUTE_POST_EXPENSE_CATEGORIES, models.Authenticate(CreateExpenseCategory))
	mux.HandleFunc(constants.ROUTE_POST_EXPENSE_OCCURRENCES_ENSURE, models.Authenticate(EnsureExpenseOccurrences))
	mux.HandleFunc(constants.ROUTE_POST_EXPENSE_OCCURRENCES_PAID, models.Authenticate(SetExpenseOccurrencePaid))

	// income routes (authentication required)
	// GET and POST share some paths, so those are method-qualified patterns (Go 1.22+ mux)
	mux.HandleFunc("POST "+constants.ROUTE_POST_INCOMES, models.Authenticate(CreateIncome))
	mux.HandleFunc("GET "+constants.ROUTE_GET_INCOMES, models.Authenticate(ListIncomes))
	mux.HandleFunc(constants.ROUTE_POST_INCOMES_UPDATE, models.Authenticate(UpdateIncome))
	mux.HandleFunc(constants.ROUTE_POST_INCOMES_DELETE, models.Authenticate(DeleteIncome))
	mux.HandleFunc("GET "+constants.ROUTE_GET_INCOME_ENTRIES, models.Authenticate(ListIncomeEntries))
	mux.HandleFunc(constants.ROUTE_POST_INCOME_ENTRIES_APPROVE, models.Authenticate(ApproveIncomeEntry))
	mux.HandleFunc(constants.ROUTE_POST_INCOME_ENTRIES_REJECT, models.Authenticate(RejectIncomeEntry))


	// Serve static files from the frontend build
	staticPath := os.Getenv("STATIC_PATH")
	if staticPath == "" {
		staticPath = "frontend/dist"
	}
	spa := spaHandler{staticPath: staticPath, indexPath: "index.html"}
	mux.Handle("/", spa)

	return utils.EnableCORS(mux)
}

// spaHandler implements the http.Handler interface for serving a SPA
type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := h.staticPath + r.URL.Path

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File does not exist, serve index.html
		http.ServeFile(w, r, h.staticPath+"/"+h.indexPath)
		return
	}

	// Serve the file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

/************************************************************************
* Simple Auth func
*********************************************************************/
func SampleAuth(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	// Get authenticated user
	_, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]interface{}{
		"message": "hello",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Simple Pub func
*********************************************************************/
func SamplePub(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse


	httpResponse.Data = "Hello"
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Handles login
*
* status: ✅
************************************************************************/
func Login(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	err := user.RequestToStruct(r.Body)

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	token, twoFARequired, err := user.Login()

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// 2FA accounts get no token yet — the emailed code must be confirmed first
	if twoFARequired {
		httpResponse.Data = map[string]interface{}{
			"TwoFARequired": true,
		}
		httpResponse.Success = true
		httpResponse.Error = nil
		httpResponse.Send(w)
		return
	}

	data := map[string]interface{}{
		"AuthToken": token,
		"User": user,
	}

	httpResponse.Data = data
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Handles signup
*
* status: ✅
************************************************************************/
func Signup(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	err := user.RequestToStruct(r.Body)

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	token, err := user.Signup()

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	data := map[string]interface{}{
		"AuthToken": token,
		"User": user,
	}


	httpResponse.Success = false
	httpResponse.Data = data
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Handles email verification
*
* status: ✅
************************************************************************/
func VerifyEmail(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	// Create struct to receive the code
	var requestBody struct {
		Code string `json:"code"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if len(requestBody.Code) == 0 {
		httpResponse.Error = "verification code is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	token, err := user.VerifyEmail(requestBody.Code)

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	data := map[string]interface{}{
		"AuthToken": token,
		"User": user,
	}

	httpResponse.Data = data
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Handles forgot password functionality by sending a temporary password
* to the user's email address. The user can then use this temporary
* password to log in and change their password.
*
* status: ✅
************************************************************************/
func ForgotPassword(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	err := user.ForgotPassword(r.Body)

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Password reset email sent successfully",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Handles password change functionality. Requires authentication and
* current password verification before allowing the user to set a new password.
*
* status: ✅
************************************************************************/
func ChangePassword(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	// Get authenticated user from context
	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	err := user.ChangePassword(r.Body, authUser.Id)

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Password changed successfully",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Updates user profile (first name and last name)
* status: ✅
************************************************************************/
func UpdateProfile(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	// Parse request body
	err := user.RequestToStruct(r.Body)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("Invalid request data: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// Get authenticated user from context
	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// Call the Update method from the user model
	err = user.Update(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// Profile data lives in the JWT claims, so mint a fresh token — otherwise
	// the session would keep showing the old profile until the next login
	updatedAuthUser := models.AuthUser{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Status:    authUser.Status,
		Id:        authUser.Id,
	}

	token, err := updatedAuthUser.GenerateJWT()
	if err != nil {
		httpResponse.Error = fmt.Sprintf("could not refresh session token: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message":   "Profile updated successfully",
		"AuthToken": token,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}


/************************************************************************
* Toggles two-factor authentication for the authenticated user. Returns a
* fresh token so the session carries the updated two_fa_enabled claim.
*
* status: ✅
************************************************************************/
func ToggleTwoFA(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	// Get authenticated user from context
	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var requestBody struct {
		Enabled bool `json:"enabled"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	err = user.SetTwoFA(authUser.Id, requestBody.Enabled)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// mint a fresh token so the session reflects the new 2FA setting
	updatedAuthUser := models.AuthUser{
		FirstName:    authUser.FirstName,
		LastName:     authUser.LastName,
		Username:     authUser.Username,
		Email:        authUser.Email,
		Avatar:       authUser.Avatar,
		Status:       authUser.Status,
		TwoFAEnabled: requestBody.Enabled,
		Id:           authUser.Id,
	}

	token, err := updatedAuthUser.GenerateJWT()
	if err != nil {
		httpResponse.Error = fmt.Sprintf("could not refresh session token: %v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message":   "Two-factor authentication updated",
		"AuthToken": token,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Confirms the emailed 6-digit 2FA code and completes the login.
*
* status: ✅
************************************************************************/
func VerifyTwoFA(w http.ResponseWriter, r *http.Request){
	var httpResponse models.HttpResponse
	var user models.User

	var requestBody struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if len(requestBody.Username) == 0 || len(requestBody.Code) == 0 {
		httpResponse.Error = "username and code are required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	token, err := user.VerifyTwoFA(requestBody.Username, requestBody.Code)

	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	data := map[string]interface{}{
		"AuthToken": token,
		"User": user,
	}

	httpResponse.Data = data
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
