export const ROUTE_HOME = "/";
export const ROUTE_AUTH = "/auth";
export const ROUTE_AUTH_VERIFY = "/auth/verify";
export const ROUTE_FINANCES = "/finances";
export const ROUTE_CALENDAR = "/calendar";

// api base
export const API_BASE = import.meta.env.VITE_API_BASE || "/api";

// api auth routes
export const API_POST_LOGIN = API_BASE + "/login";
export const API_POST_SIGNUP = API_BASE + "/signup";
export const API_GET_VERIFY_EMAIL = API_BASE + "/verify-email";
export const API_POST_FORGOT_PASSWORD = API_BASE + "/forgot-password";
export const API_POST_CHANGE_PASSWORD = API_BASE + "/change-password";
export const API_POST_UPDATE_PROFILE = API_BASE + "/update-profile";

// api expense routes
export const API_POST_EXPENSES = API_BASE + "/expenses";
export const API_GET_EXPENSES = API_BASE + "/expenses";
export const API_POST_EXPENSES_UPDATE = API_BASE + "/expenses/update";
export const API_POST_EXPENSES_DELETE = API_BASE + "/expenses/delete";
export const API_GET_EXPENSE_CATEGORIES = API_BASE + "/expenses/categories";
export const API_POST_EXPENSE_CATEGORIES = API_BASE + "/expenses/categories";

// api plaid routes
export const API_POST_PLAID_LINK_TOKEN = API_BASE + "/plaid/link-token/create";
export const API_POST_PLAID_EXCHANGE = API_BASE + "/plaid/exchange";
export const API_POST_PLAID_SYNC = API_BASE + "/plaid/sync";
export const API_GET_PLAID_TRANSACTIONS = API_BASE + "/plaid/transactions";
export const API_GET_PLAID_ACCOUNTS = API_BASE + "/plaid/accounts";
export const API_GET_PLAID_ITEMS = API_BASE + "/plaid/items";
export const API_GET_PLAID_STATUS = API_BASE + "/plaid/status";

// user statuses
export const USER_STATUS_PENDING = "pending";
export const USER_STATUS_ACTIVE = "active";
export const USER_STATUS_DELETED = "deleted";

// expense recurring schedules
// values follow the JS Date convention: 0 = Sunday … 6 = Saturday
export const DAYS_OF_WEEK = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

export const MONTH_NAMES = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
];
