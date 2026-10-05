/*********************************************************************************************************
 * Payday helpers. Paydays land every other Friday; the anchor is the last known payday (Oct 2, 2026),
 * so every payday is anchor + 14k days. All math is done on local dates (time stripped).
 *
 * This is the FALLBACK schedule used until the user has approved income — once they do, the
 * paycheck plan derives its paydays from the income records (see utils/income.js). The calendar
 * and strip still use this anchor purely as a visual reference.
 * ******************************************************************************************************
 */

export const PAYDAY_ANCHOR = new Date(2026, 9, 2); // Oct 2, 2026 — a Friday
export const PAY_PERIOD_DAYS = 14;

const MS_PER_DAY = 1000 * 60 * 60 * 24;

export function stripTime(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function daysSinceAnchor(date) {
  return Math.round((stripTime(date) - stripTime(PAYDAY_ANCHOR)) / MS_PER_DAY);
}

export function isPayday(date) {
  const diff = daysSinceAnchor(date);
  return ((diff % PAY_PERIOD_DAYS) + PAY_PERIOD_DAYS) % PAY_PERIOD_DAYS === 0;
}

// 0 when fromDate itself is a payday
export function daysUntilPayday(fromDate = new Date()) {
  const diff = daysSinceAnchor(fromDate);
  const sinceLast = ((diff % PAY_PERIOD_DAYS) + PAY_PERIOD_DAYS) % PAY_PERIOD_DAYS;
  return (PAY_PERIOD_DAYS - sinceLast) % PAY_PERIOD_DAYS;
}

export function nextPaydayDate(fromDate = new Date()) {
  const date = stripTime(fromDate);
  date.setDate(date.getDate() + daysUntilPayday(date));
  return date;
}

// The next `count` paydays (starting with the upcoming one), every other Friday
export function upcomingPaydays(count = 12, fromDate = new Date()) {
  const first = nextPaydayDate(fromDate);
  const paydays = [];
  for (let i = 0; i < count; i++) {
    const date = new Date(first);
    date.setDate(first.getDate() + i * PAY_PERIOD_DAYS);
    paydays.push(date);
  }
  return paydays;
}
