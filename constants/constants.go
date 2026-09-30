package constants

const ROUTE_API_BASE = "/api"

const ROUTE_POST_LOGIN = ROUTE_API_BASE + "/login"
const ROUTE_POST_SIGNUP = ROUTE_API_BASE + "/signup"
const ROUTE_POST_VERIFY_EMAIL = ROUTE_API_BASE + "/verify-email"
const ROUTE_POST_FORGOT_PASSWORD = ROUTE_API_BASE + "/forgot-password"
const ROUTE_POST_CHANGE_PASSWORD = ROUTE_API_BASE + "/change-password"
const ROUTE_POST_UPDATE_PROFILE = ROUTE_API_BASE + "/update-profile"

// sample
const ROUTE_GET_AUTH_SAMPLE = ROUTE_API_BASE + "/auth-sample"
const ROUTE_GET_PUBLIC_SAMPLE = ROUTE_API_BASE + "/public-sample"

// expenses
const ROUTE_POST_EXPENSES = ROUTE_API_BASE + "/expenses"
const ROUTE_GET_EXPENSES = ROUTE_API_BASE + "/expenses"
const ROUTE_POST_EXPENSES_UPDATE = ROUTE_API_BASE + "/expenses/update"
const ROUTE_POST_EXPENSES_DELETE = ROUTE_API_BASE + "/expenses/delete"
const ROUTE_GET_EXPENSE_CATEGORIES = ROUTE_API_BASE + "/expenses/categories"
const ROUTE_POST_EXPENSE_CATEGORIES = ROUTE_API_BASE + "/expenses/categories"

// plaid
const ROUTE_POST_PLAID_LINK_TOKEN = ROUTE_API_BASE + "/plaid/link-token/create"
const ROUTE_POST_PLAID_EXCHANGE = ROUTE_API_BASE + "/plaid/exchange"
const ROUTE_POST_PLAID_SYNC = ROUTE_API_BASE + "/plaid/sync"
const ROUTE_GET_PLAID_TRANSACTIONS = ROUTE_API_BASE + "/plaid/transactions"
const ROUTE_GET_PLAID_ACCOUNTS = ROUTE_API_BASE + "/plaid/accounts"
const ROUTE_GET_PLAID_ITEMS = ROUTE_API_BASE + "/plaid/items"
const ROUTE_GET_PLAID_STATUS = ROUTE_API_BASE + "/plaid/status"

// context must have a predefined custom context key type
type contextKey string
const USER_CONTEXT_AUTH_KEY contextKey = "currentUser"
