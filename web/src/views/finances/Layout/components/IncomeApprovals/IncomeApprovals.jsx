import {
  API_POST_INCOME_ENTRIES_APPROVE,
  API_POST_INCOME_ENTRIES_REJECT,
  API_GET_INCOME_ENTRIES,
  API_GET_INCOMES,
} from "@constants";
import {
  pendingIncomeOccurrences,
  formatDateKey,
  useGet,
  usePost,
} from "@utils";
import { Check, CheckCheck, X, Ban, Pencil } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useMemo, useState } from "react";

// styles
import "./IncomeApprovals.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

/*********************************************************************************************************
 * Fixed-income occurrences that happened since the user's last approved/rejected income and still
 * need a decision. Shown on the Finances page on load. Each row can be approved (with the scheduled
 * or an edited amount) or rejected, individually or in bulk; approving creates an income_entries
 * record so the income is counted in the paycheck plan, and rejecting records it so it stops
 * showing up as pending.
 * ******************************************************************************************************
 */
export const IncomeApprovals = (props) => {
  const { onDecided } = props;

  const [refreshKey, setRefreshKey] = useState(0);
  const [drafts, setDrafts] = useState({}); // key -> edited amount string
  const [editingKey, setEditingKey] = useState(null); // which row's amount is being edited
  const [busyKeys, setBusyKeys] = useState({}); // key -> true while its request is in flight

  const incomes = useGet({ url: API_GET_INCOMES, dependencies: [refreshKey] });
  const entries = useGet({ url: API_GET_INCOME_ENTRIES, dependencies: [refreshKey] });

  const incomeList = useMemo(
    () => (Array.isArray(incomes.data) ? incomes.data : []),
    [incomes.data]
  );
  const entryList = useMemo(
    () => (Array.isArray(entries.data) ? entries.data : []),
    [entries.data]
  );

  // the most recent decision's date — pending scan resumes the day after it
  const lastEntryDate = useMemo(
    () =>
      entryList.reduce(
        (max, entry) => (entry.occurred_on > max ? entry.occurred_on : max),
        ""
      ) || null,
    [entryList]
  );

  const pending = useMemo(
    () => pendingIncomeOccurrences(incomeList, entryList, lastEntryDate, new Date()),
    [incomeList, entryList, lastEntryDate]
  );

  const refresh = () => {
    setRefreshKey((key) => key + 1);
    if (onDecided) onDecided();
  };

  const approve = usePost({
    url: API_POST_INCOME_ENTRIES_APPROVE,
    callback: (data) => {
      if (!data) return;
      refresh();
    },
  });

  const reject = usePost({
    url: API_POST_INCOME_ENTRIES_REJECT,
    callback: (data) => {
      if (!data) return;
      refresh();
    },
  });

  // the amount a row will be approved with — the edited draft when present, else the scheduled one
  const amountFor = (occurrence) => {
    const draft = drafts[occurrence.key];
    if (draft !== undefined && draft.trim() !== "" && !isNaN(parseFloat(draft))) {
      return Math.round(parseFloat(draft) * 100) / 100;
    }
    return occurrence.income.amount;
  };

  const setBusy = (key, value) =>
    setBusyKeys((prev) => ({ ...prev, [key]: value }));

  const decideOne = (occurrence, action) => {
    const payload = {
      income_id: occurrence.income.id,
      occurred_on: formatDateKey(occurrence.date),
      amount: amountFor(occurrence),
    };
    setBusy(occurrence.key, true);
    action.post(payload);
  };

  const decideAll = (action) => {
    pending.forEach((occurrence) => {
      if (!busyKeys[occurrence.key]) decideOne(occurrence, action);
    });
  };

  const startEdit = (occurrence) => {
    setEditingKey(occurrence.key);
    setDrafts((prev) => ({
      ...prev,
      [occurrence.key]: String(occurrence.income.amount),
    }));
  };

  const loading = incomes.loading || entries.loading;
  const anyBusy = Object.values(busyKeys).some(Boolean);

  // nothing to approve — render nothing at all
  if (!loading && pending.length === 0) return null;

  return (
    <section className='income-approvals-3vx8'>
      <div className='income-approvals-3vx8__head'>
        <p className='income-approvals-3vx8__label'>
          Income to approve{pending.length > 0 ? ` (${pending.length})` : ""}
        </p>

        {pending.length > 1 && (
          <div className='income-approvals-3vx8__bulk'>
            <button
              className='income-approvals-3vx8__bulk-btn approve'
              onClick={() => decideAll(approve)}
              disabled={anyBusy}
              type='button'
            >
              <MorphIcon icon={CheckCheck} label='Approve all' size={15} />
              Approve all
            </button>
            <button
              className='income-approvals-3vx8__bulk-btn reject'
              onClick={() => decideAll(reject)}
              disabled={anyBusy}
              type='button'
            >
              <MorphIcon icon={Ban} label='Reject all' size={15} />
              Reject all
            </button>
          </div>
        )}
      </div>

      <p className='income-approvals-3vx8__hint'>
        Approved income is counted in your paycheck plan; rejected income is skipped.
      </p>

      {loading ? (
        <p className='income-approvals-3vx8__loading'>Loading…</p>
      ) : (
        <ul className='income-approvals-3vx8__list'>
          {pending.map((occurrence) => {
            const isEditing = editingKey === occurrence.key;
            const shownAmount = amountFor(occurrence);
            const edited =
              drafts[occurrence.key] !== undefined &&
              parseFloat(drafts[occurrence.key]) !== occurrence.income.amount;

            return (
              <li key={occurrence.key} className='income-approvals-3vx8__row'>
                <div className='income-approvals-3vx8__details'>
                  <span className='income-approvals-3vx8__name'>
                    {occurrence.income.label}
                  </span>
                  <span className='income-approvals-3vx8__meta'>
                    {occurrence.date.toLocaleDateString("en-US", {
                      weekday: "short",
                      month: "short",
                      day: "numeric",
                    })}
                    {edited && (
                      <span className='income-approvals-3vx8__was'>
                        {" "}
                        (was {formatUSD(occurrence.income.amount)})
                      </span>
                    )}
                  </span>
                </div>

                {isEditing ? (
                  <input
                    onChange={(e) =>
                      setDrafts((prev) => ({
                        ...prev,
                        [occurrence.key]: e.target.value,
                      }))
                    }
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === "Escape") setEditingKey(null);
                    }}
                    className='income-approvals-3vx8__amount-input'
                    inputMode='decimal'
                    type='number'
                    step='0.01'
                    min='0'
                    value={drafts[occurrence.key]}
                    autoFocus
                  />
                ) : (
                  <button
                    className='income-approvals-3vx8__amount'
                    onClick={() => startEdit(occurrence)}
                    title='Click to edit the amount'
                    type='button'
                  >
                    <span>{formatUSD(shownAmount)}</span>
                    <MorphIcon icon={Pencil} label='Edit amount' size={13} />
                  </button>
                )}

                <div className='income-approvals-3vx8__actions'>
                  <button
                    className='income-approvals-3vx8__action approve'
                    onClick={() => decideOne(occurrence, approve)}
                    aria-label={`Approve ${occurrence.income.label}`}
                    disabled={busyKeys[occurrence.key]}
                    type='button'
                  >
                    <MorphIcon icon={Check} label='Approve' size={16} />
                  </button>
                  <button
                    className='income-approvals-3vx8__action reject'
                    onClick={() => decideOne(occurrence, reject)}
                    aria-label={`Reject ${occurrence.income.label}`}
                    disabled={busyKeys[occurrence.key]}
                    type='button'
                  >
                    <MorphIcon icon={X} label='Reject' size={16} />
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
};
