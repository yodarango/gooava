import { stripTime } from "./payday";

/* *************************************************************************************************
 * Income schedules. Each income carries `recurring` (one of "", weekly, biweekly, monthly,
 * quarterly, semiannually, annually) and a `recur_rule` JSON blob built by the income form:
 *   "" (one-time): { date: "YYYY-MM-DD" }
 *   weekly:        { day_of_week: 0-6 }                    (0 = Sunday)
 *   biweekly:      { mode: "every_other_week", anchor: "YYYY-MM-DD" }  (anchor is a real payday)
 *               or { mode: "days_of_month", days_of_month: [d1, d2] }
 *   monthly:       { day_of_month: d }
 *   quarterly:     { day_of_month: d }                     (every 3rd month from the anchor month)
 *   semiannually:  { day_of_month: d }                     (every 6th month from the anchor month)
 *   annually:      { month: m, day_of_month: d }           (1-12, 1-31)
 *
 * Days of the month are stored as picked; months shorter than the picked day pay on their last
 * day instead (e.g. the 31st becomes Feb 28). The anchor month for quarterly/semiannually is
 * captured when the rule is first expanded (first occurrence defines the cycle).
 * *************************************************************************************************
 */

const MS_PER_DAY = 1000 * 60 * 60 * 24;

// "YYYY-MM-DD" <-> local Date (time stripped). String parsing avoids timezone surprises.
export function formatDateKey(date) {
  const d = stripTime(new Date(date));
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${month}-${day}`;
}

export function parseDateKey(key) {
  const [year, month, day] = String(key).split("-").map(Number);
  return new Date(year, (month || 1) - 1, day || 1);
}

// clamp a 1-31 day-of-month to the month's real length (e.g. the 31st becomes Feb 28)
const clampToMonth = (year, monthIndex, day) =>
  Math.min(day, new Date(year, monthIndex + 1, 0).getDate());

const isDayOfMonth = (date, day) =>
  date.getDate() === clampToMonth(date.getFullYear(), date.getMonth(), day);

/*********************************************************************************************************
 * True when the income occurs on the given date. For quarterly/semiannually the cycle is anchored
 * to the month of `anchorDate` (the income's first occurrence — pass null to match every month,
 * which callers avoid by always supplying the anchor once known).
 * ******************************************************************************************************
 */
export function incomeOccursOn(income, date, anchorDate = null) {
  const rule = income.recur_rule || {};
  const target = stripTime(new Date(date));

  switch (income.recurring) {
    case "weekly":
      return target.getDay() === rule.day_of_week;

    case "biweekly":
      if (rule.mode === "days_of_month") {
        return (rule.days_of_month || []).some((day) => isDayOfMonth(target, day));
      }
      if (rule.mode === "every_other_week" && rule.anchor) {
        const diff = Math.round((target - parseDateKey(rule.anchor)) / MS_PER_DAY);
        return ((diff % 14) + 14) % 14 === 0;
      }
      return false;

    case "monthly":
      return rule.day_of_month != null && isDayOfMonth(target, rule.day_of_month);

    case "quarterly":
    case "semiannually": {
      if (rule.day_of_month == null || !isDayOfMonth(target, rule.day_of_month)) return false;
      if (!anchorDate) return true; // no anchor yet — any month matches
      const step = income.recurring === "quarterly" ? 3 : 6;
      const anchor = stripTime(new Date(anchorDate));
      const monthDiff =
        (target.getFullYear() - anchor.getFullYear()) * 12 +
        (target.getMonth() - anchor.getMonth());
      return ((monthDiff % step) + step) % step === 0;
    }

    case "annually":
      return (
        rule.month === target.getMonth() + 1 &&
        rule.day_of_month != null &&
        isDayOfMonth(target, rule.day_of_month)
      );

    default:
      return false;
  }
}

/*********************************************************************************************************
 * The first date on/after fromDate the income occurs on, or null for one-time incomes whose date
 * has passed. Used to anchor quarterly/semiannually cycles.
 * ******************************************************************************************************
 */
export function firstIncomeOccurrence(income, fromDate) {
  const rule = income.recur_rule || {};
  const from = stripTime(new Date(fromDate));

  if (!income.recurring) {
    if (!rule.date) return null;
    const date = parseDateKey(rule.date);
    return date >= from ? date : null;
  }

  const cursor = new Date(from);
  // a 2-year scan always finds the next occurrence (annually included)
  const limit = new Date(from);
  limit.setFullYear(limit.getFullYear() + 2);

  while (cursor < limit) {
    // match with no anchor so the first hit becomes the anchor itself
    if (incomeOccursOn(income, cursor, null)) return new Date(cursor);
    cursor.setDate(cursor.getDate() + 1);
  }

  return null;
}


/*********************************************************************************************************
 * Every date the income occurs on in [fromDate, toDate). One-time incomes occur at most once.
 * ******************************************************************************************************
 */
export function incomeOccurrences(income, fromDate, toDate) {
  const from = stripTime(new Date(fromDate));
  const end = stripTime(new Date(toDate));
  const rule = income.recur_rule || {};

  if (!income.recurring) {
    if (!rule.date) return [];
    const date = parseDateKey(rule.date);
    return date >= from && date < end ? [date] : [];
  }

  const anchor = firstIncomeOccurrence(income, from);
  const dates = [];
  const cursor = new Date(from);

  while (cursor < end) {
    if (incomeOccursOn(income, cursor, anchor)) dates.push(new Date(cursor));
    cursor.setDate(cursor.getDate() + 1);
  }

  return dates;
}

/*********************************************************************************************************
 * Payday anchors from the user's approved fixed incomes. Collects every occurrence of every fixed
 * income in [from, horizonEnd] and returns the distinct sorted dates — these replace the old
 * hardcoded every-other-Friday anchor. Returns [] when nothing is approved, so callers can fall
 * back to the legacy anchor.
 * ******************************************************************************************************
 */
export function incomePaydays(incomes, fromDate, horizonEnd) {
  const from = stripTime(new Date(fromDate));
  const seen = new Set();

  for (const income of incomes) {
    if (!income.recurring) continue; // one-time incomes don't define the pay cycle
    for (const date of incomeOccurrences(income, from, horizonEnd)) {
      seen.add(stripTime(date).getTime());
    }
  }

  return [...seen].sort((a, b) => a - b).map((time) => new Date(time));
}

/*********************************************************************************************************
 * Fixed-income occurrences in (afterEntryDate, throughDate] that still need a decision: no
 * approved or rejected entry exists for that income+date yet. afterEntryDate is the occurred_on
 * of the user's most recent entry (any status), or null to scan from scratch. Returns
 * [{ income, date, key }] sorted by date; key is "incomeId:YYYY-MM-DD".
 * ******************************************************************************************************
 */
export function pendingIncomeOccurrences(incomes, entries, afterEntryDate, throughDate) {
  const through = stripTime(new Date(throughDate));
  const decided = new Set(
    (entries || []).map((entry) => `${entry.income_id}:${entry.occurred_on}`)
  );

  // scan from the day after the last decided occurrence; with no entries yet, only look
  // back ~5 weeks so a brand-new user isn't buried in a year of historical occurrences
  const from = afterEntryDate
    ? (() => {
        const date = parseDateKey(afterEntryDate);
        date.setDate(date.getDate() + 1);
        return date;
      })()
    : (() => {
        const date = new Date(through);
        date.setDate(date.getDate() - 35);
        return date;
      })();

  const pending = [];
  for (const income of incomes) {
    if (!income.recurring) continue; // one-time incomes don't need approval
    for (const date of incomeOccurrences(income, from, new Date(through.getTime() + MS_PER_DAY))) {
      const key = `${income.id}:${formatDateKey(date)}`;
      if (!decided.has(key)) pending.push({ income, date, key });
    }
  }

  return pending.sort((a, b) => a.date - b.date);
}

/*********************************************************************************************************
 * Union of the next `count` occurrences of every fixed income on/after fromDate, sorted.
 * ******************************************************************************************************
 */
export function upcomingIncomeDates(incomes, fromDate, count = 4) {
  const from = stripTime(new Date(fromDate));
  const horizon = new Date(from);
  horizon.setFullYear(horizon.getFullYear() + 1);

  return incomePaydays(incomes, from, horizon).slice(0, count);
}