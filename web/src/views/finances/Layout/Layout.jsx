import { FinancesSettings } from "./components/FinancesSettings/FinancesSettings";
import { UpcomingBills } from "./components/UpcomingBills/UpcomingBills";
import { BankBalance } from "./components/BankBalance/BankBalance";
import { PlaidConnect } from "./components/PlaidConnect/PlaidConnect";
import { Transactions } from "./components/Transactions/Transactions";
import { PaydayStrip } from "./components/PaydayStrip/PaydayStrip";
import { useAppContext } from "../../context/appContextProvider";
import { daysUntilPayday, nextPaydayDate, useGet } from "@utils";
import { API_GET_PLAID_STATUS } from "@constants";
import { Pencil, Check, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useState } from "react";

// styles
import "./Layout.css";

// Per-user storage: everything is keyed by the logged-in user's id so
// multiple users on the same browser never see each other's data.
const storageKeyFor = (userId) => `gooava:finances:${userId}`;

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// e.g. "Oct. 29 08:10 AM"
const formatToday = (date) => {
  const month = date.toLocaleDateString("en-US", { month: "short" });
  const time = date.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
  });
  return `${month}. ${date.getDate()} ${time}`;
};

export const Layout = () => {
  const { state, showToast } = useAppContext();

  const [balance, setBalance] = useState(null);
  const [draft, setDraft] = useState("");
  const [editing, setEditing] = useState(false);
  const [now, setNow] = useState(() => new Date());
  const [refreshKey, setRefreshKey] = useState(0);

  // Bank connection status — when connected, the manual balance card and the
  // connect card are swapped for live bank data.
  const status = useGet({
    url: API_GET_PLAID_STATUS,
    dependencies: [refreshKey],
  });
  const bankConnected = !!status.data?.connected;

  // Live clock so the header date/time stays current
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(id);
  }, []);

  const userId = state.user?.id;
  const daysLeft = daysUntilPayday();
  const nextPayday = nextPaydayDate();

  // Load this user's balance once their identity is known
  useEffect(() => {
    if (!userId) return;

    try {
      const saved = JSON.parse(localStorage.getItem(storageKeyFor(userId)));
      if (saved && typeof saved.balance === "number") setBalance(saved.balance);
    } catch {
      // corrupted entry — start fresh
    }
  }, [userId]);

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
    localStorage.setItem(
      storageKeyFor(userId),
      JSON.stringify({ balance: rounded })
    );
    showToast({ message: "Balance updated", type: "success" });
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter") saveBalance();
    if (e.key === "Escape") cancelEditing();
  };

  return (
    <div className='finances-layout-4f8d'>
      <div className='finances-layout-4f8d__header'>
        <h1 className='finances-layout-4f8d__title'>Finances</h1>
        <FinancesSettings onChanged={() => setRefreshKey((key) => key + 1)} />
      </div>
      <p className='finances-layout-4f8d__today'>{formatToday(now)}</p>

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
              <span>{balance === null ? "— set amount —" : formatUSD(balance)}</span>
              <MorphIcon icon={Pencil} label='Edit balance' size={18} />
            </button>
          )}

          <p className='finances-layout-4f8d__hint'>
            Updated manually by you — no bank connection.
          </p>
        </section>
        )}

        {/* Payday countdown */}
        <section className='finances-layout-4f8d__card countdown'>
          <p className='finances-layout-4f8d__label'>Next payday</p>
          <p className='finances-layout-4f8d__days'>
            {daysLeft === 0 ? (
              "Payday is today! 🎉"
            ) : (
              <>
                <strong>{daysLeft}</strong>{" "}
                {daysLeft === 1 ? "day" : "days"} left
              </>
            )}
          </p>
          <p className='finances-layout-4f8d__hint'>
            {nextPayday.toLocaleDateString("en-US", {
              weekday: "long",
              month: "long",
              day: "numeric",
            })}
          </p>
        </section>
      </div>

      <UpcomingBills />

      <PaydayStrip />

      {/* The connect card goes away once a bank is linked (and stays hidden
          while the status is still loading to avoid a flash) */}
      {!status.loading && !bankConnected && (
        <PlaidConnect onSynced={() => setRefreshKey((key) => key + 1)} />
      )}

      <Transactions refreshKey={refreshKey} />
    </div>
  );
};
