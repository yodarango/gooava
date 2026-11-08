import { useAppContext } from "../../../../context/appContextProvider";
import {
  API_POST_EXPENSE_OCCURRENCES_ENSURE,
  API_POST_EXPENSE_OCCURRENCES_PAID,
  API_POST_EXPENSE_OCCURRENCE_AMOUNT,
  API_GET_FINANCE_PREFERENCES,
  API_POST_FINANCE_PREFERENCES,
  API_GET_INCOME_ENTRIES,
  API_GET_PLAID_ACCOUNTS,
  API_POST_EXPENSES,
  API_GET_INCOMES,
  API_GET_EXPENSES,
} from "@constants";
import {
  upcomingIncomeDates,
  incomeOccurrences,
  upcomingPaydays,
  incomePaydays,
  formatDateKey,
  billWindows,
  stripTime,
  isPayday,
  useGet,
  usePost,
} from "@utils";
import { Dropdown } from "@ds";
import { Pencil, Check, X, ChevronDown } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useRef, useState } from "react";

// styles
import "./UpcomingBills.css";

const paycheckKeyFor = (userId) => `gooava:paycheck:${userId}`;

// selectable sizes for the paycheck plan — saved per user on the server
const WINDOW_COUNT_OPTIONS = [3, 6, 9, 12, 24, 36];

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// e.g. "Oct 2"
const formatDay = (date) =>
  date.toLocaleDateString("en-US", { month: "short", day: "numeric" });

/*********************************************************************************************************
 * Paycheck plan card: splits the bills due across the next N paydays (user-selectable:
 * 3/6/9/12/24/36, synced per user) into per-period windows and
 * checks coverage — the current window against the money in the bank, later windows against the
 * paycheck amount (editable, saved per user). When a paycheck window comes up short, the headline
 * breaks down how much to put away from each paycheck to cover it (each period reserves what it
 * can spare for the ones after it; whatever's left falls on this paycheck).
 * ******************************************************************************************************
 */
export const UpcomingBills = (props) => {
  // manualBalance comes synced from the Layout (db-backed) — the connected
  // checking total still wins when a bank is linked
  const { refreshKey = 0, manualBalance = null, onPlanComputed } = props;
  const { state, showToast } = useAppContext();
  const userId = state.user?.id;

  const expenses = useGet({
    url: API_GET_EXPENSES,
    dependencies: [refreshKey],
  });
  const accounts = useGet({ url: API_GET_PLAID_ACCOUNTS });
  const incomes = useGet({ url: API_GET_INCOMES, dependencies: [refreshKey] });
  const entries = useGet({
    url: API_GET_INCOME_ENTRIES,
    dependencies: [refreshKey],
  });

  const [paycheck, setPaycheck] = useState(null);
  const [paycheckDraft, setPaycheckDraft] = useState("");
  const [editingPaycheck, setEditingPaycheck] = useState(false);

  // how many paycheck windows the plan shows — synced per user (default 6)
  const [windowCount, setWindowCount] = useState(6);

  const savePreferences = usePost({ url: API_POST_FINANCE_PREFERENCES });

  useGet({
    url: API_GET_FINANCE_PREFERENCES,
    callback: (data) => {
      if (data && WINDOW_COUNT_OPTIONS.includes(data.paycheck_windows)) {
        setWindowCount(data.paycheck_windows);
      }
    },
  });

  // surface preference save errors as toasts
  useEffect(() => {
    if (savePreferences.error) {
      showToast({ message: String(savePreferences.error), type: "danger" });
    }
  }, [savePreferences.error]);

  // occurrence entries keyed by "expenseId:YYYY-MM-DD" -> { id, paid, ... } from the db
  const [occurrenceMap, setOccurrenceMap] = useState({});
  // a token that re-runs the ensure pass after a manual (un)mark
  const [occurrenceRefresh, setOccurrenceRefresh] = useState(0);
  // guards against re-ensuring the same range on every render
  const ensuredRangeRef = useRef("");

  // which paycheck cards are open — all closed by default
  const [openWindows, setOpenWindows] = useState({});

  // flips true once the first ensure pass returns — past-due paychecks render only
  // after that, so paid history doesn't flash as unpaid on load
  const [occurrencesLoaded, setOccurrencesLoaded] = useState(false);

  // occurrence id being amount-edited (pencil) + its draft value
  const [editingAmountId, setEditingAmountId] = useState(null);
  const [amountDraft, setAmountDraft] = useState("");

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

  const available = hasBankBalance ? checkingTotal : manualBalance;

  // paydays come from the fixed incomes (approved or not); fall back to the legacy
  // every-other-Friday anchor until the user has any fixed income
  const incomeList = Array.isArray(incomes.data) ? incomes.data : [];
  const entryList = Array.isArray(entries.data) ? entries.data : [];

  // rejected occurrences are the only ones excluded from the plan — everything else counts
  const rejectedKeySet = new Set(
    entryList
      .filter((entry) => entry.status === "rejected")
      .map((entry) => `${entry.income_id}:${entry.occurred_on}`),
  );
  // the schedule drives the plan as soon as any fixed income exists — approval only
  // records a *past* occurrence, it never gates whether upcoming ones count
  const fixedIncomes = incomeList.filter((income) => income.recurring);
  const hasIncomeSchedule = fixedIncomes.length > 0;

  // `windowCount` windows ending at the next paydays (fetch one extra — when today
  // is payday the first end shifts ahead so the plan still shows `windowCount`)
  const today = stripTime(new Date());
  const ends = hasIncomeSchedule
    ? upcomingIncomeDates(fixedIncomes, today, windowCount + 1)
    : upcomingPaydays(windowCount + 1, today);
  if (ends.length > 0 && ends[0].getTime() <= today.getTime()) ends.shift(); // today is payday — plan ahead
  const windowEnds = ends.slice(0, windowCount);

  const expenseList = Array.isArray(expenses.data) ? expenses.data : [];

  // The plan owns the "put away" expense: its occurrences are materialized from the
  // plan's math (one per paycheck that owes a reserve), so it's excluded from the
  // schedule expansion — otherwise it would feed back into its own computation.
  // The oldest matching expense is canonical — same pick the server makes.
  const putAwayExpense =
    expenseList
      .filter((expense) => expense.category === "put away")
      .sort((a, b) => a.id - b.id)[0] || null;
  const planExpenses = expenseList.filter(
    (expense) => expense.category !== "put away",
  );

  const windows = billWindows(planExpenses, today, windowEnds);

  // Past paychecks stay in the plan while they hold an unpaid bill — boundaries are
  // the paydays before today (the income schedule, else the legacy every-other-Friday
  // anchor), at most 4 windows back.
  const pastPaydays = (() => {
    if (hasIncomeSchedule) {
      const scanFrom = new Date(today);
      scanFrom.setDate(scanFrom.getDate() - 75);
      return incomePaydays(fixedIncomes, scanFrom, today).slice(-5);
    }
    const found = [];
    const cursor = new Date(today);
    while (found.length < 5) {
      cursor.setDate(cursor.getDate() - 1);
      if (isPayday(cursor)) found.push(new Date(cursor));
    }
    return found.reverse(); // ascending
  })();

  // windows behind the current one, e.g. [p4,p5), [p5, today) — rendered only while
  // an unpaid bill keeps them alive
  const pastBoundaries = [...pastPaydays, today];
  const pastWindows =
    pastBoundaries.length > 1
      ? billWindows(planExpenses, pastBoundaries[0], pastBoundaries.slice(1))
      : [];

  // one-time expenses never expand into windows — an unpaid one created in a past
  // window belongs to that window
  for (const expense of planExpenses) {
    if (expense.recurring || !expense.created_at) continue;
    const created = stripTime(new Date(expense.created_at.replace(" ", "T")));
    if (created >= today) continue;
    const window = pastWindows.find(
      (w) => created >= w.start && created < w.end,
    );
    if (window) window.bills.push({ expense, date: created });
  }

  // every occurrence across the whole rendered range — these become the DB entries.
  // One-time expenses have no recurrence dates, so their single occurrence is their created day;
  // recurring expenses expand into every due date in the window.
  const rangeStart = pastBoundaries.length > 1 ? pastBoundaries[0] : today;
  const rangeEnd =
    windowEnds.length > 0 ? windowEnds[windowEnds.length - 1] : today;

  // create the plan-owned "put away" expense once — the reserve rows in each
  // paycheck are its occurrences (the real amounts live on those rows)
  const createPutAway = usePost({
    url: API_POST_EXPENSES,
    callback: () => expenses.refetch(),
  });
  const putAwayCreateAttemptedRef = useRef(false);

  useEffect(() => {
    if (expenses.loading || !Array.isArray(expenses.data)) return;
    if (putAwayExpense || putAwayCreateAttemptedRef.current) return;

    putAwayCreateAttemptedRef.current = true;
    createPutAway.post({
      label: "Put away",
      category: "put away",
      amount: 0.01, // placeholder — each occurrence carries the real reserve amount
      recurring: "",
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [expenses.loading, expenses.data, putAwayExpense]);

  // merge the DB paid state onto each rendered bill. A bill is paid when its occurrence entry
  // is paid; paid bills are excluded from totals/put-away but still shown (crossed off).
  const occurrenceFor = (expenseId, date) =>
    occurrenceMap[`${expenseId}:${formatDateKey(date)}`] || null;

  const isBillPaid = (expenseId, date) =>
    !!occurrenceFor(expenseId, date)?.paid;

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

  // the occurrence row's amount wins over the expense default — it carries the
  // snapshot as-ensured plus any pencil edit (planAmount = the put away row's
  // plan-computed amount before its occurrence exists)
  const amountFor = (bill) =>
    occurrenceFor(bill.expense.id, bill.date)?.amount ??
    bill.planAmount ??
    bill.expense.amount;

  // unpaid bills sitting in past windows — they stay in the plan until paid, and the
  // money on hand has to cover them before anything else
  const pastWindowUnpaid = (window) =>
    window.bills.reduce(
      (sum, bill) =>
        sum +
        (isBillPaid(bill.expense.id, bill.date) ||
        bill.expense.category === "milestone"
          ? 0
          : amountFor(bill) || 0),
      0,
    );
  const pastDueTotal = pastWindows.reduce(
    (sum, window) => sum + pastWindowUnpaid(window),
    0,
  );

  const saveAmount = usePost({
    url: API_POST_EXPENSE_OCCURRENCE_AMOUNT,
    callback: (data) => {
      if (!data) return;
      ensuredRangeRef.current = ""; // force a re-ensure so amounts reload
      setOccurrenceRefresh((key) => key + 1);
    },
  });

  // surface amount save errors as toasts
  useEffect(() => {
    if (saveAmount.error) {
      showToast({ message: String(saveAmount.error), type: "danger" });
    }
  }, [saveAmount.error]);

  const startEditingAmount = (occurrence) => {
    setEditingAmountId(occurrence.id);
    setAmountDraft(String(occurrence.amount));
  };

  const saveAmountEdit = () => {
    const parsed = parseFloat(amountDraft);

    if (amountDraft.trim() === "" || isNaN(parsed) || parsed < 0) {
      showToast({
        message: "Please enter a valid amount (0 or more)",
        type: "danger",
      });
      return;
    }

    saveAmount.post({
      id: editingAmountId,
      amount: Math.round(parsed * 100) / 100,
    });
    setEditingAmountId(null);
  };

  const handleAmountKeyDown = (e) => {
    if (e.key === "Enter") saveAmountEdit();
    if (e.key === "Escape") setEditingAmountId(null);
  };

  // income arriving within each window (keyed by window index). Every scheduled occurrence
  // counts — including the upcoming paychecks that haven't been approved yet — unless the
  // user explicitly rejected it. An approved entry's edited amount wins over the scheduled one.
  const entryAmountByKey = new Map(
    entryList
      .filter((entry) => entry.status === "approved")
      .map((entry) => [
        `${entry.income_id}:${entry.occurred_on}`,
        entry.amount,
      ]),
  );
  const incomeByWindow = windowEnds.map((end, index) => {
    const start = index === 0 ? today : windowEnds[index - 1];
    let total = 0;
    for (const income of fixedIncomes) {
      for (const date of incomeOccurrences(income, start, end)) {
        const key = `${income.id}:${formatDateKey(date)}`;
        if (rejectedKeySet.has(key)) continue; // user said this one didn't happen
        total += entryAmountByKey.has(key)
          ? entryAmountByKey.get(key)
          : income.amount;
      }
    }
    return total;
  });

  // each window's UNPAID total — paid bills are crossed off in the UI but no longer owed,
  // so they're removed from the plan's math
  const unpaidTotalByWindow = windows.map((window) =>
    window.bills.reduce(
      (sum, bill) =>
        sum +
        (isBillPaid(bill.expense.id, bill.date) ? 0 : amountFor(bill) || 0),
      0,
    ),
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

  // what each window pockets: with an income schedule it's the income scheduled in
  // that window; otherwise the manual paycheck (window 0 lives off the bank balance)
  const incomeForWindow = (index) =>
    hasIncomeSchedule ? incomeByWindow[index] : index === 0 ? 0 : paycheck;

  // what a short window receives from earlier paychecks' put-aways — its own
  // shortfall (the plan below covers every shortfall on paper, window 0 last resort)
  const reserveReceivedFor = (index) => {
    if (index === 0) return 0;
    const coverage = coverageFor(index);
    if (coverage === null) return 0;
    return Math.max(unpaidTotalFor(index) - coverage, 0);
  };

  // The savings plan, per paycheck. Walk the windows from the last one backwards: each
  // period first covers its own bills, then reserves whatever it can spare for the
  // periods after it; whatever it can't cover itself falls to the period before — with
  // "this paycheck" (the money available right now) as the last resort. putAwayByWindow[i]
  // is how much to set aside from period i's money.
  const putAwayByWindow = windows.map(() => 0);
  let laterNeed = 0;
  windows
    .map((window, index) => index)
    .reverse()
    .forEach((index) => {
      if (index === 0) {
        putAwayByWindow[0] = laterNeed; // last resort — everything later periods still need
        return;
      }
      const coverage = coverageFor(index);
      if (coverage === null) return; // coverage unknown — nothing can be reserved here
      const spare = coverage - unpaidTotalFor(index);
      putAwayByWindow[index] = Math.min(Math.max(spare, 0), laterNeed);
      laterNeed = laterNeed - putAwayByWindow[index] + Math.max(-spare, 0);
    });
  const putAway = putAwayByWindow.reduce((sum, amount) => sum + amount, 0);

  // The outstanding reserve a paycheck still owes — exactly like a bill: 0 once its
  // occurrence is paid (money set aside), the hand-edited amount when penciled,
  // otherwise the plan's computed amount.
  const putAwayOutstandingFor = (index) => {
    const planned = putAwayByWindow[index] || 0;
    if (planned <= 0 || !putAwayExpense) return 0;

    const due = new Date(windowEnds[index]);
    due.setDate(due.getDate() - 1);
    const occurrence = occurrenceFor(putAwayExpense.id, due);

    if (occurrence && occurrence.paid) return 0;
    return occurrence ? occurrence.amount : planned;
  };

  // the items to lazily ensure + the request range, as a stable signature so the effect only
  // re-fires when the actual occurrences change (not on every render). The put away amounts
  // are part of it — a plan change re-materializes those rows server-side.
  const occurrenceSignature = JSON.stringify([
    formatDateKey(rangeStart),
    formatDateKey(rangeEnd),
    planExpenses.map((e) => `${e.id}:${e.linked_label || ""}`),
    putAwayExpense?.id || 0,
    putAwayByWindow.map((amount) => Math.round(amount * 100)),
    occurrenceRefresh,
    refreshKey,
  ]);

  // Lazily create the occurrence entries for the rendered window (the put away expense
  // gets one row per paycheck that owes a reserve, due the day before its payday),
  // auto-pay the ones whose expense the user explicitly linked to a transaction, then
  // load the resulting paid state.
  useEffect(() => {
    if (!userId) return;
    if (windowEnds.length === 0) return;
    // skip if we already ensured this exact range + expense set + plan
    if (ensuredRangeRef.current === occurrenceSignature) return;

    // build the occurrence items the interface is about to show
    const items = [];
    for (const expense of planExpenses) {
      if (expense.recurring) {
        for (const window of [...pastWindows, ...windows]) {
          for (const bill of window.bills) {
            if (bill.expense.id === expense.id) {
              items.push({
                expense_id: expense.id,
                due_on: formatDateKey(bill.date),
              });
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
    // materialize the reserve each paycheck owes, due the day before its payday
    if (putAwayExpense) {
      windowEnds.forEach((end, index) => {
        const amount = Math.round(putAwayByWindow[index] * 100) / 100;
        if (amount > 0) {
          const due = new Date(end);
          due.setDate(due.getDate() - 1);
          items.push({
            expense_id: putAwayExpense.id,
            due_on: formatDateKey(due),
            amount,
          });
        }
      });
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
        setOccurrencesLoaded(true);
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

  // Cumulative running balance: start from what's in the bank; each window adds its
  // income (+ reserves received from earlier paychecks) and deducts its unpaid bills
  // (+ the reserves it puts away for later ones). Reserves net to zero across the
  // windows, so the final balance still equals balance + income − bills. An unknown
  // input (no balance set, no paycheck amount) makes the rest of the run unknown.
  const runningAfterByWindow = [];
  // past-due unpaid bills are owed out of the money on hand before any window runs
  let runningBalance = available === null ? null : available - pastDueTotal;
  windows.forEach((window, index) => {
    const income = incomeForWindow(index);
    if (runningBalance === null || income === null) {
      runningBalance = null;
      runningAfterByWindow.push(null);
      return;
    }
    runningBalance +=
      income +
      reserveReceivedFor(index) -
      unpaidTotalFor(index) -
      putAwayOutstandingFor(index);
    runningAfterByWindow.push(runningBalance);
  });

  // "final period finances" — where the user stands after the last window
  const finalBalance =
    runningAfterByWindow.length > 0
      ? runningAfterByWindow[runningAfterByWindow.length - 1]
      : null;

  // a future window is short only when we actually know its coverage
  const hasFutureShortfall = windows.slice(1).some((window, i) => {
    const coverage = coverageFor(i + 1);
    return coverage !== null && unpaidTotalFor(i + 1) > coverage;
  });

  // what the money on hand comes to after past-due bills, this window's bills, AND
  // the reserves it must put away for later short windows
  const firstWindowVerdict =
    available === null || windows.length === 0
      ? null
      : available - pastDueTotal - unpaidTotalFor(0) - putAwayOutstandingFor(0);

  // total still owed across the whole plan — unpaid bills plus outstanding put aways
  const grandTotal = windows.reduce(
    (sum, window, i) => sum + unpaidTotalFor(i) + putAwayOutstandingFor(i),
    0,
  );
  const loading = expenses.loading && !expenses.data;

  // report the current paycheck's leftover so the page heading can show it next to
  // the days-till-payday text
  useEffect(() => {
    if (onPlanComputed) onPlanComputed(firstWindowVerdict);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [firstWindowVerdict]);

  const startEditingPaycheck = () => {
    setPaycheckDraft(paycheck === null ? "" : String(paycheck));
    setEditingPaycheck(true);
  };

  const savePaycheck = () => {
    const parsed = parseFloat(paycheckDraft);

    if (paycheckDraft.trim() === "" || isNaN(parsed) || parsed < 0) {
      showToast({
        message: "Please enter a valid amount (0 or more)",
        type: "danger",
      });
      return;
    }

    const rounded = Math.round(parsed * 100) / 100;
    setPaycheck(rounded);
    setEditingPaycheck(false);
    localStorage.setItem(
      paycheckKeyFor(userId),
      JSON.stringify({ amount: rounded }),
    );
    showToast({ message: "Paycheck updated", type: "success" });
  };

  const changeWindowCount = (count) => {
    if (!WINDOW_COUNT_OPTIONS.includes(count)) return;

    setWindowCount(count);
    savePreferences.post({ paycheck_windows: count });
  };

  const handlePaycheckKeyDown = (e) => {
    if (e.key === "Enter") savePaycheck();
    if (e.key === "Escape") setEditingPaycheck(false);
  };

  // shared bill row — current/future windows and past-due paychecks. pastDue turns
  // the row red and adds the badge for bills that went unpaid past their due date.
  // Milestones are personal goals, so they are never flagged as past-due or late.
  const renderBillRow = (bill, pastDue = false) => {
    const occurrence = occurrenceFor(bill.expense.id, bill.date);
    const paid = !!occurrence?.paid;
    const isMilestone = bill.expense.category === "milestone";
    const showPastDue = pastDue && !paid && !isMilestone;

    return (
      <li
        key={`${bill.expense.id}-${bill.date.getTime()}`}
        className={`upcoming-bills-5xk3__row ${paid ? "paid" : ""} ${showPastDue ? "past-due" : ""} ${isMilestone ? "milestone" : ""}`}
      >
        <div className='d-flex justify-content-start gap-4 align-items-center'>
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
              {bill.expense.linked_label && (
                <>
                  {" "}
                  <span className='upcoming-bills-5xk3__auto-badge'>auto</span>
                </>
              )}
              {showPastDue && (
                <>
                  {" "}
                  <span className='upcoming-bills-5xk3__pastdue-badge'>
                    past due
                  </span>
                </>
              )}
              {isMilestone && (
                <>
                  {" "}
                  <span className='upcoming-bills-5xk3__milestone-badge'>
                    goal
                  </span>
                </>
              )}
            </span>
          </div>
        </div>

        {occurrence && !paid && editingAmountId === occurrence.id ? (
          <span className='upcoming-bills-5xk3__paycheck-edit'>
            <input
              onChange={(e) => setAmountDraft(e.target.value)}
              onKeyDown={handleAmountKeyDown}
              className='upcoming-bills-5xk3__paycheck-input'
              placeholder='0.00'
              inputMode='decimal'
              value={amountDraft}
              type='number'
              min='0'
              step='0.01'
              autoFocus
            />
            <button
              className='upcoming-bills-5xk3__icon-btn save'
              onClick={saveAmountEdit}
              aria-label='Save amount'
              type='button'
            >
              <MorphIcon icon={Check} label='Save' size={16} />
            </button>
            <button
              className='upcoming-bills-5xk3__icon-btn cancel'
              onClick={() => setEditingAmountId(null)}
              aria-label='Cancel'
              type='button'
            >
              <MorphIcon icon={X} label='Cancel' size={16} />
            </button>
          </span>
        ) : (
          <span className='upcoming-bills-5xk3__amount'>
            {formatUSD(amountFor(bill))}
            {occurrence && !paid && (
              <button
                className='upcoming-bills-5xk3__amount-edit'
                onClick={() => startEditingAmount(occurrence)}
                aria-label={`Edit the amount of ${bill.expense.label}`}
                title='Edit this bill’s amount'
                type='button'
              >
                <MorphIcon icon={Pencil} label='Edit amount' size={13} />
              </button>
            )}
          </span>
        )}
      </li>
    );
  };

  return (
    <section
      className={`upcoming-bills-5xk3 ${putAway > 0 ? "attention" : ""}`}
    >
      <div className='upcoming-bills-5xk3__head'>
        <p className='upcoming-bills-5xk3__label'>Paycheck plan</p>

        <div className='upcoming-bills-5xk3__controls'>
          {/* how many paychecks ahead the plan shows — saved per user */}
          <Dropdown
            secondary
            value={windowCount}
            options={WINDOW_COUNT_OPTIONS.map((count) => ({
              value: count,
              label: `${count}`,
            }))}
            onChange={changeWindowCount}
            ariaLabel='Paychecks to show'
          />

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
                Paycheck:{" "}
                {paycheck === null ? "— set amount —" : formatUSD(paycheck)}
              </span>
              <MorphIcon icon={Pencil} label='Edit paycheck' size={14} />
            </button>
          ) : null}
        </div>
      </div>

      {/* headline: total still owed across the whole plan (put away rows included) */}
      {loading ? (
        <p className='upcoming-bills-5xk3__hint'>Loading…</p>
      ) : (
        <>
          <p className='upcoming-bills-5xk3__total'>{formatUSD(grandTotal)}</p>
          <p className='upcoming-bills-5xk3__hint'>
            due across the next {windowCount} paydays
            {!hasIncomeSchedule && paycheck === null
              ? " — add income in settings to see if you're covered"
              : hasFutureShortfall
                ? " — a later period comes up short (see its put away rows)"
                : " — every period is covered ✓"}
          </p>
        </>
      )}

      {/* past paychecks stay in the plan while an unpaid bill keeps them alive —
          red border on the card, red rows with a past due badge */}
      {!loading &&
        occurrencesLoaded &&
        pastWindows.map((window) => {
          const unpaidBills = window.bills.filter(
            (bill) => !isBillPaid(bill.expense.id, bill.date),
          );
          if (unpaidBills.length === 0) return null;

          const pastKey = `past-${window.end.getTime()}`;
          const isOpen = !!openWindows[pastKey];
          const unpaidTotal = pastWindowUnpaid(window);

          return (
            <div
              className={`upcoming-bills-5xk3__window past-due ${isOpen ? "open" : ""}`}
              key={pastKey}
            >
              <button
                className='upcoming-bills-5xk3__window-toggle'
                onClick={() => toggleWindow(pastKey)}
                aria-expanded={isOpen}
                type='button'
              >
                <div className='upcoming-bills-5xk3__window-title'>
                  <span className='upcoming-bills-5xk3__window-range'>
                    {formatDay(window.start)} → {formatDay(window.end)}
                  </span>
                  <span className='upcoming-bills-5xk3__window-source'>
                    paycheck behind
                  </span>
                </div>

                <div className='upcoming-bills-5xk3__window-summary'>
                  <span className='upcoming-bills-5xk3__verdict short'>
                    past due
                  </span>
                  <span className='upcoming-bills-5xk3__window-due'>
                    {formatUSD(unpaidTotal)} · {unpaidBills.length} bill
                    {unpaidBills.length === 1 ? "" : "s"}
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

              {isOpen && (
                <div className='upcoming-bills-5xk3__window-body'>
                  <ul className='upcoming-bills-5xk3__list'>
                    {unpaidBills.map((bill) => renderBillRow(bill, true))}
                  </ul>
                </div>
              )}
            </div>
          );
        })}

      {/* per-window breakdown */}
      {!loading &&
        windows.map((window, index) => {
          const coverage = coverageFor(index);
          const unpaidTotal = unpaidTotalFor(index);
          const reserveReceived = reserveReceivedFor(index);
          const putAwayOutstanding = putAwayOutstandingFor(index);

          // the put away renders as a normal bill in this paycheck — due the day
          // before the payday that ends the window
          const putAwayDate = new Date(window.end);
          putAwayDate.setDate(putAwayDate.getDate() - 1);
          const displayBills =
            putAwayByWindow[index] > 0 && putAwayExpense
              ? [
                  ...window.bills,
                  {
                    expense: putAwayExpense,
                    date: putAwayDate,
                    planAmount: putAwayByWindow[index],
                  },
                ]
              : window.bills;

          // reserves fold into the verdict: putting money away lowers what this window
          // keeps; a short window is covered by what earlier paychecks reserved for it
          const verdict =
            coverage === null
              ? null
              : coverage + reserveReceived - unpaidTotal - putAwayOutstanding;
          const runningAfter = runningAfterByWindow[index] ?? null;

          const displayedDue = unpaidTotal + putAwayOutstanding;
          const billCount = displayBills.length;
          // open-body stat lines
          const scheduledTotal = displayBills.reduce(
            (sum, bill) => sum + (amountFor(bill) || 0),
            0,
          );
          const leftHere = index === 0 ? firstWindowVerdict : verdict;
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
                </div>

                <div className='upcoming-bills-5xk3__window-summary'>
                  {/* one badge: this window's left-over combined with the running
                      total accumulated across all paychecks */}
                  {runningAfter !== null && (
                    <span
                      className={`upcoming-bills-5xk3__verdict ${runningAfter >= 0 ? "left" : "short"}`}
                      title='Accumulated balance after this paycheck'
                    >
                      {formatUSD(runningAfter)}
                    </span>
                  )}
                  <span className='upcoming-bills-5xk3__window-due'>
                    {formatUSD(displayedDue)} · {billCount} bill
                    {billCount === 1 ? "" : "s"}
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
                  {/* the three stat lines: what this paycheck owes, what it keeps once
                      its bills (and past-due ones, for the current one) are paid, and
                      the running total accumulated across all paychecks */}
                  <div className='upcoming-bills-5xk3__window-stats'>
                    <div className='upcoming-bills-5xk3__stat'>
                      <span>Scheduled expenses</span>
                      <span>{formatUSD(scheduledTotal)}</span>
                    </div>
                    <div className='upcoming-bills-5xk3__stat'>
                      <span>Left from this paycheck</span>
                      <span
                        className={
                          leftHere !== null && leftHere < 0 ? "short" : ""
                        }
                      >
                        {leftHere === null
                          ? index === 0
                            ? "— balance unknown"
                            : "— no income approved"
                          : formatUSD(leftHere)}
                      </span>
                    </div>
                    <div className='upcoming-bills-5xk3__stat'>
                      <span>Accumulated balance</span>
                      <span
                        className={
                          runningAfter !== null && runningAfter < 0
                            ? "short"
                            : ""
                        }
                      >
                        {runningAfter === null ? "—" : formatUSD(runningAfter)}
                      </span>
                    </div>
                  </div>

                  {displayBills.length === 0 ? (
                    <p className='upcoming-bills-5xk3__none'>No bills due</p>
                  ) : (
                    <ul className='upcoming-bills-5xk3__list'>
                      {displayBills.map((bill) => renderBillRow(bill))}
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

      {/* final cumulative standing across the whole plan */}
      {!loading && (
        <div className='upcoming-bills-5xk3__final'>
          <p className='upcoming-bills-5xk3__label'>Final period finances</p>
          {finalBalance === null ? (
            <p className='upcoming-bills-5xk3__hint'>
              — set your balance{!hasIncomeSchedule ? " and paycheck" : ""} to
              see where you end up
            </p>
          ) : (
            <>
              <p
                className={`upcoming-bills-5xk3__total ${finalBalance < 0 ? "negative" : "positive"}`}
              >
                {formatUSD(finalBalance)}
              </p>
              <p className='upcoming-bills-5xk3__hint'>
                balance + income − bills across the next {windowCount} paychecks
              </p>
            </>
          )}
        </div>
      )}
    </section>
  );
};
