/*********************************************************************************************************
 * Payday helpers. Paydays land every other Friday; the anchor is the last known payday (Oct 2, 2026),
 * so every payday is anchor + 14k days. All math is done on local dates (time stripped).
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
