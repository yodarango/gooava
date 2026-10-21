import { useAppContext } from "../../../../context/appContextProvider";
import { TransactionPicker } from "../TransactionPicker/TransactionPicker";
import {
  API_POST_EXPENSE_OCCURRENCES_CONFIRM,
  API_POST_EXPENSE_OCCURRENCES_REVIEW,
  API_POST_EXPENSE_OCCURRENCES_PAID,
  API_POST_EXPENSES_LINK,
  API_GET_INCOMES,
  API_GET_EXPENSES,
} from "@constants";
import {
  billOccurrences,
  formatDateKey,
  incomePaydays,
  isPayday,
  stripTime,
  useGet,
  usePost,
} from "@utils";
import { Check, Link2 } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useRef, useState } from "react";
import { Button } from "@ds";

// styles
import "./PastDueBills.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// "2026-09-28" -> "Mon, Sep 28" (parsed as local time so the day never shifts)
const formatDay = (key) =>
  new Date(`${key}T00:00:00`).toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
  });

const authHeaders = () => ({
  "Content-Type": "application/json",
  Authorization: "Bearer " + localStorage.getItem("auth"),
});

/*********************************************************************************************************
 * Past-due bills review, run lazily every time the app opens. Covers bills due since the second
 * most recent paycheck ("the past two paychecks"). Linked expenses pay themselves on the server;
 * what's left comes back split by whether a transaction string-matches:
 *  - matched: "the following bills found matching transactions — should I mark them paid?"
 *    every row starts selected, the user can unselect any, and Save marks the rest paid.
 *  - unmatched: must be marked paid manually, or linked to a transaction (which also pays this
 *    occurrence and auto-pays future ones with the same label).
 * ******************************************************************************************************
 */
export const PastDueBills = (props) => {
  const { refreshKey = 0, onChanged } = props;
  const { state, showToast } = useAppContext();
  const userId = state.user?.id;

  const expenses = useGet({ url: API_GET_EXPENSES });
  const incomes = useGet({ url: API_GET_INCOMES });

  const [review, setReview] = useState(null); // occurrences from the review endpoint
  const [checking, setChecking] = useState(true); // the lazy review is in flight
  const [checked, setChecked] = useState({}); // occurrence id -> included in the save
  const [saving, setSaving] = useState(false);
  const [pickerFor, setPickerFor] = useState(null); // occurrence being linked
  const [reviewRefresh, setReviewRefresh] = useState(0);
  const reviewedSignatureRef = useRef("");

  const expenseList = Array.isArray(expenses.data) ? expenses.data : [];
  const incomeList = Array.isArray(incomes.data) ? incomes.data : [];
  const fixedIncomes = incomeList.filter((income) => income.recurring);

  const today = stripTime(new Date());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  // lookback start: the second most recent paycheck, so the range covers the bills of
  // the past two paychecks (falls back to the every-other-Friday anchor, then 28 days)
  const scanFrom = new Date(today);
  scanFrom.setDate(scanFrom.getDate() - 75);
  let rangeStart = null;
  if (fixedIncomes.length > 0) {
    const through = new Date(today);
    through.setDate(through.getDate() + 1);
    const paydays = incomePaydays(fixedIncomes, scanFrom, through);
    rangeStart =
      paydays.length >= 2 ? paydays[paydays.length - 2] : paydays[paydays.length - 1];
  } else {
    const paydays = [];
    const cursor = new Date(today);
    while (paydays.length < 2 && cursor >= scanFrom) {
      if (isPayday(cursor)) paydays.push(new Date(cursor));
      cursor.setDate(cursor.getDate() - 1);
    }
    rangeStart = paydays.length === 2 ? paydays[1] : paydays[0] || null;
  }
  if (!rangeStart) {
    rangeStart = new Date(today);
    rangeStart.setDate(rangeStart.getDate() - 28);
  }

  // the range + expense set as a stable signature so the review only re-runs on change
  const reviewSignature = JSON.stringify([
    formatDateKey(rangeStart),
    formatDateKey(yesterday),
    expenseList.map((e) => `${e.id}:${e.linked_label || ""}:${e.created_at || ""}`),
    reviewRefresh,
    refreshKey,
  ]);

  useEffect(() => {
    if (!userId) return;
    if (expenses.loading || incomes.loading) return;
    if (rangeStart > yesterday) {
      setReview([]);
      setChecking(false);
      return;
    }
    if (reviewedSignatureRef.current === reviewSignature) return;

    // expand each expense into its past-due occurrences in the range
    const items = [];
    for (const expense of expenseList) {
      if (expense.recurring) {
        for (const date of billOccurrences(expense, rangeStart, today)) {
          items.push({ expense_id: expense.id, due_on: formatDateKey(date) });
        }
      } else if (expense.created_at) {
        // one-time expense: a single occurrence on its created day
        const created = stripTime(new Date(expense.created_at.replace(" ", "T")));
        if (created >= rangeStart && created < today) {
          items.push({ expense_id: expense.id, due_on: formatDateKey(created) });
        }
      }
    }

    let cancelled = false;
    const runReview = async () => {
      setChecking(true);
      try {
        const response = await fetch(API_POST_EXPENSE_OCCURRENCES_REVIEW, {
          method: "POST",
          headers: authHeaders(),
          body: JSON.stringify({
            from: formatDateKey(rangeStart),
            to: formatDateKey(yesterday),
            items,
          }),
        });
        const result = await response.json();
        if (cancelled || !result.success) return;

        const list = Array.isArray(result.data) ? result.data : [];
        setReview(list);
        // every matched bill starts selected; keep any selections the user already made
        setChecked((prev) => {
          const next = { ...prev };
          for (const occurrence of list) {
            if (occurrence.match && next[occurrence.id] === undefined) {
              next[occurrence.id] = true;
            }
          }
          return next;
        });
        reviewedSignatureRef.current = reviewSignature;
      } catch {
        // network/parse failure — stay silent, the cards just don't render
      } finally {
        if (!cancelled) setChecking(false);
      }
    };
    runReview();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reviewSignature, userId, expenses.loading, incomes.loading]);

  // re-run the review + let the rest of the page refresh (plan, transactions…)
  const refresh = () => {
    reviewedSignatureRef.current = "";
    setReviewRefresh((key) => key + 1);
    if (onChanged) onChanged();
  };

  const markPaid = usePost({
    url: API_POST_EXPENSE_OCCURRENCES_PAID,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Marked as paid", type: "success" });
      refresh();
    },
  });

  const occurrences = review || [];
  const matched = occurrences.filter((occurrence) => occurrence.match);
  const unmatched = occurrences.filter((occurrence) => !occurrence.match);
  const selectedMatched = matched.filter((occurrence) => checked[occurrence.id]);

  // the confirmation save — mark every still-selected matched bill paid with its transaction
  const saveMatched = async () => {
    if (selectedMatched.length === 0 || saving) return;
    setSaving(true);
    try {
      const response = await fetch(API_POST_EXPENSE_OCCURRENCES_CONFIRM, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({
          items: selectedMatched.map((occurrence) => ({
            id: occurrence.id,
            transaction_id: occurrence.match.transaction_id,
          })),
        }),
      });
      const result = await response.json();
      if (result.success) {
        showToast({
          message: `Marked ${selectedMatched.length} bill${selectedMatched.length === 1 ? "" : "s"} as paid`,
          type: "success",
        });
        refresh();
      } else {
        showToast({ message: result.error || "Could not mark bills as paid", type: "danger" });
      }
    } catch {
      showToast({ message: "Could not mark bills as paid", type: "danger" });
    } finally {
      setSaving(false);
    }
  };

  // link from the manual card: link the expense (future occurrences auto-pay), then mark
  // this occurrence paid with the picked transaction
  const linkAndPay = async (occurrence, txn) => {
    setPickerFor(null);
    try {
      const linkResponse = await fetch(API_POST_EXPENSES_LINK, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({ id: occurrence.expense_id, transaction_id: txn.transaction_id }),
      });
      const linkResult = await linkResponse.json();
      if (!linkResult.success) {
        showToast({ message: linkResult.error || "Could not link the transaction", type: "danger" });
        return;
      }

      await fetch(API_POST_EXPENSE_OCCURRENCES_CONFIRM, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({
          items: [{ id: occurrence.id, transaction_id: txn.transaction_id }],
        }),
      });

      showToast({
        message: `Linked to "${linkResult.data?.linked_label}" — future matching bills will be marked paid automatically`,
        type: "success",
      });
      refresh();
    } catch {
      showToast({ message: "Could not link the transaction", type: "danger" });
    }
  };

  if ((checking && review === null) || occurrences.length === 0) return null;

  return (
    <>
      {matched.length > 0 && (
        <section className='past-due-bills-6t4r matched'>
          <p className='past-due-bills-6t4r__label'>Bills matched to transactions</p>
          <p className='past-due-bills-6t4r__hint'>
            The following bills found matching transactions — should I mark them paid?
          </p>

          <ul className='past-due-bills-6t4r__list'>
            {matched.map((occurrence) => {
              const isChecked = !!checked[occurrence.id];

              return (
                <li key={occurrence.id} className='past-due-bills-6t4r__row'>
                  <button
                    className={`past-due-bills-6t4r__check ${isChecked ? "checked" : ""}`}
                    onClick={() =>
                      setChecked((prev) => ({ ...prev, [occurrence.id]: !isChecked }))
                    }
                    aria-label={
                      isChecked
                        ? `Skip ${occurrence.expense_label}`
                        : `Include ${occurrence.expense_label}`
                    }
                    title={isChecked ? "Will be marked paid — tap to skip" : "Skipped — tap to include"}
                    type='button'
                  >
                    {isChecked && <MorphIcon icon={Check} label='Selected' size={14} />}
                  </button>

                  <div className='past-due-bills-6t4r__details'>
                    <span className='past-due-bills-6t4r__name'>{occurrence.expense_label}</span>
                    <span className='past-due-bills-6t4r__meta'>
                      due {formatDay(occurrence.due_on)} · matched{" "}
                      {occurrence.match.merchant_name || occurrence.match.name} on{" "}
                      {formatDay(occurrence.match.date)}
                    </span>
                  </div>

                  <span className='past-due-bills-6t4r__amount'>
                    {formatUSD(occurrence.amount)}
                  </span>
                </li>
              );
            })}
          </ul>

          <Button
            primary
            onClick={saveMatched}
            isLoading={saving}
            disabled={selectedMatched.length === 0}
            className='w-100'
          >
            {selectedMatched.length === 0
              ? "Nothing selected"
              : `Mark ${selectedMatched.length} as paid`}
          </Button>
        </section>
      )}

      {unmatched.length > 0 && (
        <section className='past-due-bills-6t4r overdue'>
          <p className='past-due-bills-6t4r__label'>Past due</p>
          <p className='past-due-bills-6t4r__hint'>
            No matching transaction found — mark them paid, or link the transaction that
            paid them so future ones are handled automatically.
          </p>

          <ul className='past-due-bills-6t4r__list'>
            {unmatched.map((occurrence) => (
              <li key={occurrence.id} className='past-due-bills-6t4r__row'>
                <div className='past-due-bills-6t4r__details'>
                  <span className='past-due-bills-6t4r__name'>{occurrence.expense_label}</span>
                  <span className='past-due-bills-6t4r__meta'>
                    due {formatDay(occurrence.due_on)}
                  </span>
                </div>

                <span className='past-due-bills-6t4r__amount'>
                  {formatUSD(occurrence.amount)}
                </span>

                <div className='past-due-bills-6t4r__actions'>
                  <button
                    className='past-due-bills-6t4r__action link'
                    onClick={() => setPickerFor(occurrence)}
                    aria-label={`Link ${occurrence.expense_label} to a transaction`}
                    title='Link to a transaction'
                    type='button'
                  >
                    <MorphIcon icon={Link2} label='Link' size={16} />
                  </button>
                  <button
                    className='past-due-bills-6t4r__action paid'
                    onClick={() => markPaid.post({ id: occurrence.id, paid: true })}
                    aria-label={`Mark ${occurrence.expense_label} paid`}
                    disabled={markPaid.loading}
                    title='Mark as paid'
                    type='button'
                  >
                    <MorphIcon icon={Check} label='Mark paid' size={16} />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        </section>
      )}

      {pickerFor && (
        <TransactionPicker
          open
          title={`Link "${pickerFor.expense_label}" to a transaction`}
          onClose={() => setPickerFor(null)}
          onSelect={(txn) => linkAndPay(pickerFor, txn)}
        />
      )}
    </>
  );
};
