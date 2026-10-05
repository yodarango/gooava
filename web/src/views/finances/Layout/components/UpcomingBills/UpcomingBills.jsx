import { useAppContext } from "../../../../context/appContextProvider";
import {
  API_POST_EXPENSE_OCCURRENCES_ENSURE,
  API_POST_EXPENSE_OCCURRENCES_PAID,
  API_GET_INCOME_ENTRIES,
  API_GET_PLAID_ACCOUNTS,
  API_GET_INCOMES,
  API_GET_EXPENSES,
} from "@constants";
import {
  upcomingIncomeDates,
  incomeOccurrences,
  upcomingPaydays,
  formatDateKey,
  billWindows,
  stripTime,
  useGet,
  usePost,
} from "@utils";
import { Pencil, Check, X, ChevronDown } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useRef, useState } from "react";

// styles
import "./UpcomingBills.css";

const paycheckKeyFor = (userId) => `gooava:paycheck:${userId}`;
// same key the Finances layout uses for the manually entered balance
const balanceKeyFor = (userId) => `gooava:finances:${userId}`;

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// e.g. "Oct 2"
const formatDay = (date) =>
  date.toLocaleDateString("en-US", { month: "short", day: "numeric" });

/*********************************************************************************************************
 * Paycheck plan card: splits the bills due across the next three paydays into per-period windows and
 * checks coverage — the current window against the money in the bank, later windows against the
 * paycheck amount (editable, saved per user). When a paycheck window comes up short, the headline
 * says how much to put away from this paycheck to cover it.
 * ******************************************************************************************************
 */
export const UpcomingBills = (props) => {
  const { refreshKey = 0 } = props;
  const { state, showToast } = useAppContext();
  const userId = state.user?.id;

  const expenses = useGet({ url: API_GET_EXPENSES });
  const accounts = useGet({ url: API_GET_PLAID_ACCOUNTS });
  const incomes = useGet({ url: API_GET_INCOMES, dependencies: [refreshKey] });
  const entries = useGet({ url: API_GET_INCOME_ENTRIES, dependencies: [refreshKey] });

  const [paycheck, setPaycheck] = useState(null);
  const [paycheckDraft, setPaycheckDraft] = useState("");
  const [editingPaycheck, setEditingPaycheck] = useState(false);
  const [manualBalance, setManualBalance] = useState(null);

  // occurrence entries keyed by "expenseId:YYYY-MM-DD" -> { id, paid, ... } from the db
  const [occurrenceMap, setOccurrenceMap] = useState({});
  // a token that re-runs the ensure pass after a manual (un)mark
  const [occurrenceRefresh, setOccurrenceRefresh] = useState(0);
  // guards against re-ensuring the same range on every render
  const ensuredRangeRef = useRef("");

  // which paycheck cards are open — all closed by default
  const [openWindows, setOpenWindows] = useState({});

  const toggleWindow = (index) =>
    setOpenWindows((prev) => ({ ...prev, [index]: !prev[index] }));

  // load the saved paycheck amount once the user is known
  useEffect(() => {
    if (!userId) return;

    try {
      const saved = JSON.parse(localStorage.getItem(paycheckKeyFor(userId)));
      if (saved && typeof saved.amount === "number") setPaycheck(saved.amount);
    } catch {
      // corrupted entry — start fresh
    }
  }, [userId]);

  // money available for the current period: the connected checking total wins,
  // otherwise fall back to the manually entered balance
  const accountList = Array.isArray(accounts.data) ? accounts.data : [];
  const hasBankBalance = accountList.length > 0;
  const checkingTotal = accountList
    .filter((account) => account.subtype === "checking")
    .reduce((sum, account) => sum + (account.current_balance || 0), 0);

  useEffect(() => {
    if (!userId || hasBankBalance) return;

    try {
      const saved = JSON.parse(localStorage.getItem(balanceKeyFor(userId)));
      if (saved && typeof saved.balance === "number") setManualBalance(saved.balance);
    } catch {
      // corrupted entry — start fresh
    }
  }, [userId, hasBankBalance]);

  const available = hasBankBalance ? checkingTotal : manualBalance;

  // paydays come from the fixed incomes (approved or not); fall back to the legacy
  // every-other-Friday anchor until the user has any fixed income
  const incomeList = Array.isArray(incomes.data) ? incomes.data : [];
  const entryList = Array.isArray(entries.data) ? entries.data : [];

  // rejected occurrences are the only ones excluded from the plan — everything else counts
  const rejectedKeySet = new Set(
    entryList
      .filter((entry) => entry.status === "rejected")
      .map((entry) => `${entry.income_id}:${entry.occurred_on}`)
  );
  // the schedule drives the plan as soon as any fixed income exists — approval only
  // records a *past* occurrence, it never gates whether upcoming ones count
  const fixedIncomes = incomeList.filter((income) => income.recurring);
  const hasIncomeSchedule = fixedIncomes.length > 0;

  // 3 windows ending at the next 3 paydays
  const today = stripTime(new Date());
  const ends = hasIncomeSchedule
    ? upcomingIncomeDates(fixedIncomes, today, 4)
    : upcomingPaydays(4, today);
  if (ends.length > 0 && ends[0].getTime() <= today.getTime()) ends.shift(); // today is payday — plan ahead
  const windowEnds = ends.slice(0, 3);

  const windows = billWindows(
    Array.isArray(expenses.data) ? expenses.data : [],
    today,
    windowEnds
  );

  // every occurrence across the whole rendered range — these become the DB entries.
  // One-time expenses have no recurrence dates, so their single occurrence is their created day;
  // recurring expenses expand into every due date in the window.
  const expenseList = Array.isArray(expenses.data) ? expenses.data : [];
  const rangeStart = today;
  const rangeEnd = windowEnds.length > 0 ? windowEnds[windowEnds.length - 1] : today;

  // the items to lazily ensure + the request range, as a stable signature so the effect only
  // re-fires when the actual occurrences change (not on every render)
  const occurrenceSignature = JSON.stringify([
    formatDateKey(rangeStart),
    formatDateKey(rangeEnd),
    expenseList.map((e) => e.id),
    occurrenceRefresh,
  ]);

  // Lazily create the occurrence entries for the rendered window, auto-pay any that match a
  // synced transaction by label, then load the resulting paid state. Re-runs when the range,
  // the expense set, or the manual-refresh token changes.
  useEffect(() => {
    if (!userId) return;
    if (windowEnds.length === 0) return;
    // skip if we already ensured this exact range + expense set
    if (ensuredRangeRef.current === occurrenceSignature) return;

    // build the occurrence items the interface is about to show
    const items = [];
    for (const expense of expenseList) {
      if (expense.recurring) {
        for (const window of windows) {
          for (const bill of window.bills) {
            if (bill.expense.id === expense.id) {
              items.push({ expense_id: expense.id, due_on: formatDateKey(bill.date) });
            }
          }
        }
      } else {
        // one-time expense: a single entry on its created date
        const created = expense.created_at
          ? formatDateKey(new Date(expense.created_at.replace(" ", "T")))
          : formatDateKey(today);
        items.push({ expense_id: expense.id, due_on: created });
      }
    }

    let cancelled = false;
    const ensure = async () => {
      try {
        const response = await fetch(API_POST_EXPENSE_OCCURRENCES_ENSURE, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: "Bearer " + localStorage.getItem("auth"),
          },
          body: JSON.stringify({
            from: formatDateKey(rangeStart),
            to: formatDateKey(rangeEnd),
            items,
          }),
        });
        const result = await response.json();
        if (cancelled || !result.success) return;

        const map = {};
        for (const o of result.data || []) {
          map[`${o.expense_id}:${o.due_on}`] = o;
        }
        ensuredRangeRef.current = occurrenceSignature;
        setOccurrenceMap(map);
      } catch {
        // network/parse failure — leave the plan rendering unpaid rather than break it
      }
    };
    ensure();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [occurrenceSignature, userId]);

  // merge the DB paid state onto each rendered bill. A bill is paid when its occurrence entry
  // is paid; paid bills are excluded from totals/put-away but still shown (crossed off).
  const occurrenceFor = (expenseId, date) =>
    occurrenceMap[`${expenseId}:${formatDateKey(date)}`] || null;

  const isBillPaid = (expenseId, date) => !!occurrenceFor(expenseId, date)?.paid;

  // manual (un)mark — posts then re-runs the ensure pass to refresh paid state
  const markPaid = usePost({
    url: API_POST_EXPENSE_OCCURRENCES_PAID,
    callback: (data) => {
      if (!data) return;
      ensuredRangeRef.current = ""; // force a re-ensure
      setOccurrenceRefresh((key) => key + 1);
    },
  });

  const togglePaid = (expenseId, date) => {
    const occurrence = occurrenceFor(expenseId, date);
    if (!occurrence) return; // entry not ensured yet
    markPaid.post({ id: occurrence.id, paid: !occurrence.paid });
  };

  // income arriving within each window (keyed by window index). Every scheduled occurrence
  // counts — including the upcoming paychecks that haven't been approved yet — unless the
  // user explicitly rejected it. An approved entry's edited amount wins over the scheduled one.
  const entryAmountByKey = new Map(
    entryList
      .filter((entry) => entry.status === "approved")
      .map((entry) => [`${entry.income_id}:${entry.occurred_on}`, entry.amount])
  );
  const incomeByWindow = windowEnds.map((end, index) => {
    const start = index === 0 ? today : windowEnds[index - 1];
    let total = 0;
    for (const income of fixedIncomes) {
      for (const date of incomeOccurrences(income, start, end)) {
        const key = `${income.id}:${formatDateKey(date)}`;
        if (rejectedKeySet.has(key)) continue; // user said this one didn't happen
        total += entryAmountByKey.has(key) ? entryAmountByKey.get(key) : income.amount;
      }
    }
    return total;
  });

  // each window's UNPAID total — paid bills are crossed off in the UI but no longer owed,
  // so they're removed from the plan's math
  const unpaidTotalByWindow = windows.map((window) =>
    window.bills.reduce(
      (sum, bill) =>
        sum + (isBillPaid(bill.expense.id, bill.date) ? 0 : bill.expense.amount || 0),
      0
    )
  );
  const unpaidTotalFor = (index) => unpaidTotalByWindow[index] ?? 0;

  // first window is covered by what's in the bank, later windows by the income scheduled
  // for them (approved or not — only rejected occurrences are excluded). Falls back to the
  // manually set paycheck only when there is no income schedule at all.
  const coverageFor = (index) =>
    index === 0
      ? available
      : hasIncomeSchedule
        ? incomeByWindow[index]
        : paycheck;

  // what later periods can't cover must be reserved from this one — per-window coverage,
  // so a window with less approved income than its bills shows the gap
  const putAway = windows.slice(1).reduce((sum, window, i) => {
    const coverage = coverageFor(i + 1);
    return coverage === null ? sum : sum + Math.max(0, unpaidTotalFor(i + 1) - coverage);
  }, 0);

  // a future window is short only when we actually know its coverage
  const hasFutureShortfall = windows
    .slice(1)
    .some((window, i) => {
      const coverage = coverageFor(i + 1);
      return coverage !== null && unpaidTotalFor(i + 1) > coverage;
    });

  const firstWindowVerdict =
    available === null || windows.length === 0
      ? null
      : available - unpaidTotalFor(0);

  const grandTotal = windows.reduce((sum, window, i) => sum + unpaidTotalFor(i), 0);
  const loading = expenses.loading && !expenses.data;

  const startEditingPaycheck = () => {
    setPaycheckDraft(paycheck === null ? "" : String(paycheck));
    setEditingPaycheck(true);
  };

  const savePaycheck = () => {
    const parsed = parseFloat(paycheckDraft);

    if (paycheckDraft.trim() === "" || isNaN(parsed) || parsed < 0) {
      showToast({ message: "Please enter a valid amount (0 or more)", type: "danger" });
      return;
    }

    const rounded = Math.round(parsed * 100) / 100;
    setPaycheck(rounded);
    setEditingPaycheck(false);
    localStorage.setItem(paycheckKeyFor(userId), JSON.stringify({ amount: rounded }));
    showToast({ message: "Paycheck updated", type: "success" });
  };

  const handlePaycheckKeyDown = (e) => {
    if (e.key === "Enter") savePaycheck();
    if (e.key === "Escape") setEditingPaycheck(false);
  };

  return (
    <section
      className={`upcoming-bills-5xk3 ${putAway > 0 ? "attention" : ""}`}
    >
      <div className='upcoming-bills-5xk3__head'>
        <p className='upcoming-bills-5xk3__label'>Paycheck plan</p>

        {/* the manual paycheck editor only applies until an income schedule exists —
            from then on each window is covered by its scheduled income */}
        {!hasIncomeSchedule && editingPaycheck ? (
          <div className='upcoming-bills-5xk3__paycheck-edit'>
            <input
              onChange={(e) => setPaycheckDraft(e.target.value)}
              onKeyDown={handlePaycheckKeyDown}
              className='upcoming-bills-5xk3__paycheck-input'
              placeholder='0.00'
              inputMode='decimal'
              value={paycheckDraft}
              type='number'
              min='0'
              step='0.01'
              autoFocus
            />
            <button
              className='upcoming-bills-5xk3__icon-btn save'
              onClick={savePaycheck}
              aria-label='Save paycheck'
              type='button'
            >
              <MorphIcon icon={Check} label='Save' size={16} />
            </button>
            <button
              className='upcoming-bills-5xk3__icon-btn cancel'
              onClick={() => setEditingPaycheck(false)}
              aria-label='Cancel'
              type='button'
            >
              <MorphIcon icon={X} label='Cancel' size={16} />
            </button>
          </div>
        ) : !hasIncomeSchedule ? (
          <button
            className='upcoming-bills-5xk3__paycheck'
            onClick={startEditingPaycheck}
            title='Click to edit'
            type='button'
          >
            <span>
              Paycheck: {paycheck === null ? "— set amount —" : formatUSD(paycheck)}
            </span>
            <MorphIcon icon={Pencil} label='Edit paycheck' size={14} />
          </button>
        ) : null}
      </div>

      {/* headline: the reserve needed from this paycheck, or the grand total when covered */}
      {loading ? (
        <p className='upcoming-bills-5xk3__hint'>Loading…</p>
      ) : hasFutureShortfall && putAway > 0 ? (
        <>
          <p className='upcoming-bills-5xk3__total short'>{formatUSD(putAway)}</p>
          <p className='upcoming-bills-5xk3__hint'>
            to put away from this paycheck — later periods come up short
          </p>
        </>
      ) : (
        <>
          <p className='upcoming-bills-5xk3__total'>{formatUSD(grandTotal)}</p>
          <p className='upcoming-bills-5xk3__hint'>
            {hasIncomeSchedule
              ? "due across the next 3 paydays — every period is covered ✓"
              : paycheck === null
                ? "due across the next 3 paydays — add income in settings to see if you're covered"
                : "due across the next 3 paydays — every period is covered ✓"}
          </p>
        </>
      )}

      {/* per-window breakdown */}
      {!loading &&
        windows.map((window, index) => {
          const coverage = coverageFor(index);
          const unpaidTotal = unpaidTotalFor(index);
          const paidTotal = window.total - unpaidTotal;
          const verdict = coverage === null ? null : coverage - unpaidTotal;

          const billCount = window.bills.length;
          const isOpen = !!openWindows[index];

          return (
            <div
              className={`upcoming-bills-5xk3__window ${isOpen ? "open" : ""}`}
              key={window.end.getTime()}
            >
              {/* collapsed header — always visible: range, verdict, bills total + count */}
              <button
                className='upcoming-bills-5xk3__window-toggle'
                onClick={() => toggleWindow(index)}
                aria-expanded={isOpen}
                type='button'
              >
                <div className='upcoming-bills-5xk3__window-title'>
                  <span className='upcoming-bills-5xk3__window-range'>
                    {index === 0 ? "Now" : formatDay(window.start)} →{" "}
                    {formatDay(window.end)}
                  </span>
                  <span className='upcoming-bills-5xk3__window-source'>
                    {index === 0
                      ? "from balance"
                      : hasIncomeSchedule
                        ? `${formatUSD(incomeByWindow[index])} scheduled`
                        : "from paycheck"}
                  </span>
                </div>

                <div className='upcoming-bills-5xk3__window-summary'>
                  {verdict !== null && (
                    <span
                      className={`upcoming-bills-5xk3__verdict ${verdict >= 0 ? "left" : "short"}`}
                    >
                      {verdict >= 0
                        ? `${formatUSD(verdict)} left`
                        : `${formatUSD(-verdict)} short`}
                    </span>
                  )}
                  <span className='upcoming-bills-5xk3__window-due'>
                    {formatUSD(unpaidTotal)} · {billCount} bill{billCount === 1 ? "" : "s"}
                  </span>
                  <span
                    className={`upcoming-bills-5xk3__chevron ${isOpen ? "open" : ""}`}
                  >
                    <MorphIcon
                      icon={ChevronDown}
                      label={isOpen ? "Collapse" : "Expand"}
                      size={18}
                    />
                  </span>
                </div>
              </button>

              {/* expanded body — the bill list, only when open */}
              {isOpen && (
                <div className='upcoming-bills-5xk3__window-body'>
                  <div className='upcoming-bills-5xk3__window-totals'>
                    <span>
                      {formatUSD(unpaidTotal)} due
                      {paidTotal > 0 && (
                        <span className='upcoming-bills-5xk3__paid-off'>
                          {" "}· {formatUSD(paidTotal)} paid ✓
                        </span>
                      )}
                    </span>
                    {verdict === null && (
                      <span className='upcoming-bills-5xk3__verdict unknown'>
                        {index === 0 ? "balance unknown" : "no income approved"}
                      </span>
                    )}
                  </div>

                  {window.bills.length === 0 ? (
                    <p className='upcoming-bills-5xk3__none'>No bills due</p>
                  ) : (
                    <ul className='upcoming-bills-5xk3__list'>
                      {window.bills.map((bill) => {
                        const occurrence = occurrenceFor(bill.expense.id, bill.date);
                        const paid = !!occurrence?.paid;

                        return (
                          <li
                            key={`${bill.expense.id}-${bill.date.getTime()}`}
                            className={`upcoming-bills-5xk3__row ${paid ? "paid" : ""}`}
                          >
                            <button
                              className={`upcoming-bills-5xk3__check ${paid ? "paid" : ""}`}
                              onClick={() => togglePaid(bill.expense.id, bill.date)}
                              aria-label={
                                paid
                                  ? `Mark ${bill.expense.label} unpaid`
                                  : `Mark ${bill.expense.label} paid`
                              }
                              title={
                                paid
                                  ? occurrence.paid_source === "auto"
                                    ? "Paid — matched a bank transaction"
                                    : "Paid — marked manually"
                                  : "Mark as paid"
                              }
                              disabled={!occurrence || markPaid.loading}
                              type='button'
                            >
                              {paid && <MorphIcon icon={Check} label='Paid' size={14} />}
                            </button>

                            <div className='upcoming-bills-5xk3__details'>
                              <span className='upcoming-bills-5xk3__name'>
                                {bill.expense.label}
                              </span>
                              <span className='upcoming-bills-5xk3__meta'>
                                {bill.expense.category}
                                {" · "}
                                {bill.date.getTime() === today.getTime()
                                  ? "today"
                                  : bill.date.toLocaleDateString("en-US", {
                                      weekday: "short",
                                      month: "short",
                                      day: "numeric",
                                    })}
                                {paid && " · paid ✓"}
                              </span>
                            </div>

                            <span className='upcoming-bills-5xk3__amount'>
                              {formatUSD(bill.expense.amount)}
                            </span>
                          </li>
                        );
                      })}
                    </ul>
                  )}
                </div>
              )}
            </div>
          );
        })}

      {/* current-period shortfall warning */}
      {firstWindowVerdict !== null && firstWindowVerdict < 0 && (
        <p className='upcoming-bills-5xk3__warning'>
          You&apos;re short {formatUSD(-firstWindowVerdict)} before{" "}
          {formatDay(windowEnds[0])}
        </p>
      )}
    </section>
  );
};
