import { usePlaidOAuthResume } from "./components/PlaidConnect/usePlaidConnect";
import { FinancesSettings } from "./components/FinancesSettings/FinancesSettings";
import { IncomeApprovals } from "./components/IncomeApprovals/IncomeApprovals";
import { UpcomingBills } from "./components/UpcomingBills/UpcomingBills";
import { BankBalance } from "./components/BankBalance/BankBalance";
import { PlaidConnect } from "./components/PlaidConnect/PlaidConnect";
import { Transactions } from "./components/Transactions/Transactions";
import { PaydayStrip } from "./components/PaydayStrip/PaydayStrip";
import { useAppContext } from "../../context/appContextProvider";
import {
  upcomingIncomeDates,
  upcomingPaydays,
  stripTime,
  useGet,
  usePost,
} from "@utils";
import {
  API_GET_BANK_BALANCE,
  API_POST_BANK_BALANCE,
  API_GET_PLAID_STATUS,
  API_GET_INCOMES,
} from "@constants";
import { Pencil, Check, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useState } from "react";

// styles
import "./Layout.css";

// Legacy per-user localStorage key. Balances saved before syncing existed are
// lifted up to the server once (see the balance fetch below), then cleared.
const storageKeyFor = (userId) => `gooava:finances:${userId}`;

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

const MS_PER_DAY = 1000 * 60 * 60 * 24;

export const Layout = () => {
  const { state, showToast } = useAppContext();
  const userId = state.user?.id;

  const [balance, setBalance] = useState(null);
  const [draft, setDraft] = useState("");
  const [editing, setEditing] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  // the current paycheck's leftover, reported by the plan card below
  const [planLeft, setPlanLeft] = useState(null);

  // Bank connection status — when connected, the manual balance card and the
  // connect card are swapped for live bank data.
  const status = useGet({
    url: API_GET_PLAID_STATUS,
    dependencies: [refreshKey],
  });
  const bankConnected = !!status.data?.connected;

  // Saves the balance to the server so it stays in sync across devices.
  const save = usePost({
    url: API_POST_BANK_BALANCE,
    callback: (data) => {
      if (data) showToast({ message: "Balance updated", type: "success" });
    },
  });

  // silent twin used for the one-time legacy lift below (no toast on page load)
  const lift = usePost({ url: API_POST_BANK_BALANCE });

  // surface save errors as toasts
  useEffect(() => {
    if (save.error) showToast({ message: String(save.error), type: "danger" });
  }, [save.error]);

  // The synced balance — the db is the source of truth, so every device agrees.
  useGet({
    url: API_GET_BANK_BALANCE,
    callback: (data) => {
      if (data && typeof data.balance === "number") {
        setBalance(data.balance);
        return;
      }

      // One-time lift: a balance saved to localStorage before syncing existed
      // is pushed up to the server, then cleared locally.
      if (!userId) return;

      try {
        const saved = JSON.parse(localStorage.getItem(storageKeyFor(userId)));
        if (saved && typeof saved.balance === "number") {
          setBalance(saved.balance);
          lift.post({ balance: saved.balance });
          localStorage.removeItem(storageKeyFor(userId));
        }
      } catch {
        // corrupted entry — start fresh
      }
    },
  });

  // resume the bank-link flow when an OAuth bank redirects back to /finances
  usePlaidOAuthResume(() => setRefreshKey((key) => key + 1));

  // Days until the next payday — derived from the user's fixed incomes when they
  // exist, otherwise the legacy every-other-Friday anchor (same as UpcomingBills).
  const incomes = useGet({ url: API_GET_INCOMES, dependencies: [refreshKey] });
  const incomeList = Array.isArray(incomes.data) ? incomes.data : [];
  const fixedIncomes = incomeList.filter((income) => income.recurring);

  const today = stripTime(new Date());
  const nextPayDate =
    fixedIncomes.length > 0
      ? upcomingIncomeDates(fixedIncomes, today, 1)[0]
      : upcomingPaydays(1, today)[0];
  const daysTillPayday = Math.round((nextPayDate - today) / MS_PER_DAY);

  const startEditing = () => {
    setDraft(balance === null ? "" : String(balance));
    setEditing(true);
  };

  const cancelEditing = () => setEditing(false);

  const saveBalance = () => {
    const parsed = parseFloat(draft);

    if (draft.trim() === "" || isNaN(parsed) || parsed < 0) {
      showToast({
        message: "Please enter a valid amount (0 or more)",
        type: "danger",
      });
      return;
    }

    const rounded = Math.round(parsed * 100) / 100;
    setBalance(rounded);
    setEditing(false);
    save.post({ balance: rounded });
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter") saveBalance();
    if (e.key === "Escape") cancelEditing();
  };

  return (
    <div className='finances-layout-4f8d'>
      {/* the page title lives in the app header — this row just hosts settings */}
      <div className='finances-layout-4f8d__header'>
        <FinancesSettings onChanged={() => setRefreshKey((key) => key + 1)} />
      </div>
      <p className='finances-layout-4f8d__today'>
        {daysTillPayday === 0
          ? "Payday is today"
          : `${daysTillPayday} day${daysTillPayday === 1 ? "" : "s"} till payday`}
        {planLeft !== null && (
          <span
            className={`finances-layout-4f8d__left ${planLeft < 0 ? "short" : ""}`}
            title='Balance after the current paycheck’s bills (and put-aways) come out'
          >
            {formatUSD(planLeft)} after bills
          </span>
        )}
      </p>

      <div className='finances-layout-4f8d__cards'>
        {/* Bank balance — pulled live once a bank is connected, editable manually otherwise */}
        {bankConnected ? (
          <BankBalance
            refreshKey={refreshKey}
            onSynced={() => setRefreshKey((key) => key + 1)}
          />
        ) : (
          <section className='finances-layout-4f8d__card'>
            <p className='finances-layout-4f8d__label'>In the bank</p>

            {editing ? (
              <div className='finances-layout-4f8d__edit'>
                <input
                  onChange={(e) => setDraft(e.target.value)}
                  onKeyDown={handleKeyDown}
                  className='finances-layout-4f8d__input'
                  placeholder='0.00'
                  inputMode='decimal'
                  value={draft}
                  type='number'
                  min='0'
                  step='0.01'
                  autoFocus
                />
                <button
                  className='finances-layout-4f8d__icon-btn save'
                  onClick={saveBalance}
                  aria-label='Save balance'
                  type='button'
                >
                  <MorphIcon icon={Check} label='Save' size={20} />
                </button>
                <button
                  className='finances-layout-4f8d__icon-btn cancel'
                  onClick={cancelEditing}
                  aria-label='Cancel'
                  type='button'
                >
                  <MorphIcon icon={X} label='Cancel' size={20} />
                </button>
              </div>
            ) : (
              <button
                className='finances-layout-4f8d__amount'
                onClick={startEditing}
                title='Click to edit'
                type='button'
              >
                <span>
                  {balance === null ? "— set amount —" : formatUSD(balance)}
                </span>
                <MorphIcon icon={Pencil} label='Edit balance' size={18} />
              </button>
            )}

            <p className='finances-layout-4f8d__hint'>
              Updated manually by you — no bank connection.
            </p>
          </section>
        )}
      </div>
      <PaydayStrip />

      {/* incomes awaiting the user's approval before they count toward the plan */}
      <IncomeApprovals onDecided={() => setRefreshKey((key) => key + 1)} />

      <UpcomingBills
        refreshKey={refreshKey}
        manualBalance={balance}
        onPlanComputed={setPlanLeft}
      />

      {/* The connect card goes away once a bank is linked (and stays hidden
          while the status is still loading to avoid a flash) */}
      {!status.loading && !bankConnected && (
        <PlaidConnect onSynced={() => setRefreshKey((key) => key + 1)} />
      )}

      <Transactions refreshKey={refreshKey} />
    </div>
  );
};
