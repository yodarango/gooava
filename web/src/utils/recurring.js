import { stripTime } from "./payday";

/* *************************************************************************************************
 * Recurrence expansion for expenses. Each expense carries `recurring` (one of "", daily, weekly,
 * biweekly, monthly, quarterly, yearly) and a `recur_rule` JSON blob built by the expense form:
 *   weekly:    { day_of_week: 0-6 }          (0 = Sunday)
 *   biweekly:  { days_of_month: [d1, d2] }
 *   monthly:   { day_of_month: d }
 *   quarterly: { day: d, months: [m1..m4] }  (1-12)
 *   yearly:    { month: m, day: d }          (1-12, 1-31)
 * *************************************************************************************************
 */

// clamp a 1-31 day-of-month to the month's real length (e.g. the 31st becomes Feb 28)
const clampToMonth = (year, monthIndex, day) =>
  Math.min(day, new Date(year, monthIndex + 1, 0).getDate());

const isDayOfMonth = (date, day) =>
  date.getDate() === clampToMonth(date.getFullYear(), date.getMonth(), day);

const matchesDay = (recurring, rule, date) => {
  switch (recurring) {
    case "daily":
      return true;
    case "weekly":
      return date.getDay() === rule.day_of_week;
    case "biweekly":
      return (rule.days_of_month || []).some((day) => isDayOfMonth(date, day));
    case "monthly":
      return rule.day_of_month != null && isDayOfMonth(date, rule.day_of_month);
    case "quarterly":
      return (
        (rule.months || []).includes(date.getMonth() + 1) &&
        rule.day != null &&
        isDayOfMonth(date, rule.day)
      );
    case "yearly":
      return (
        rule.month === date.getMonth() + 1 &&
        rule.day != null &&
        isDayOfMonth(date, rule.day)
      );
    default:
      return false;
  }
};

/*********************************************************************************************************
 * Every date a recurring expense occurs on in [fromDate, toDate). One-time expenses ("never") have
 * no due dates, so they never produce occurrences.
 * ******************************************************************************************************
 */
export function billOccurrences(expense, fromDate, toDate) {
  if (!expense.recurring) return [];

  const rule = expense.recur_rule || {};
  const dates = [];
  const cursor = stripTime(new Date(fromDate));
  const end = stripTime(new Date(toDate));

  while (cursor < end) {
    if (matchesDay(expense.recurring, rule, cursor)) {
      dates.push(new Date(cursor));
    }
    cursor.setDate(cursor.getDate() + 1);
  }

  return dates;
}

/*********************************************************************************************************
 * Flattens all expenses into [{ expense, date }] occurrences in [fromDate, toDate), sorted by date.
 * ******************************************************************************************************
 */
export function upcomingBills(expenses, fromDate, toDate) {
  return expenses
    .flatMap((expense) =>
      billOccurrences(expense, fromDate, toDate).map((date) => ({ expense, date }))
    )
    .sort((a, b) => a.date - b.date);
}

/*********************************************************************************************************
 * Total amount of a bill occurrence list.
 * ******************************************************************************************************
 */
export function totalUpcomingBills(bills) {
  return bills.reduce((sum, bill) => sum + (bill.expense.amount || 0), 0);
}
