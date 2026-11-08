import { BillingHistory } from "../BillingHistory/BillingHistory";
import { API_GET_PLAID_TRANSACTIONS } from "@constants";
import { matchesRecordQuery, useGet } from "@utils";
import { useState } from "react";
import { Input } from "@ds";

// styles
import "./Transactions.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// "2026-09-28" -> "Sep 28" (parsed as local time so the day never shifts)
const formatDate = (dateString) => {
  const date = new Date(`${dateString}T00:00:00`);
  return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
};

// "FOOD_AND_DRINK" -> "food and drink"
const formatCategory = (category) =>
  category ? category.toLowerCase().replace(/_/g, " ") : "";

/*********************************************************************************************************
 * The activity card, split into two tabs: "Bank Activity" lists the synced Plaid transactions,
 * newest first (Plaid's sign convention: amount > 0 means money left the account); "Historical
 * billing" lists every bill ever marked paid, loaded one month at a time on scroll.
 * The history tab mounts on first open and stays mounted so its loaded pages survive tab switches.
 * ******************************************************************************************************
 */
export const Transactions = (props) => {
  const { refreshKey } = props;

  const [tab, setTab] = useState("activity"); // "activity" | "history"
  const [historyMounted, setHistoryMounted] = useState(false);
  // one search bar filters whichever tab is showing — by title or by amount
  const [query, setQuery] = useState("");

  const { data, loading } = useGet({
    url: API_GET_PLAID_TRANSACTIONS + "?limit=50",
    dependencies: [refreshKey],
  });

  const allTransactions = Array.isArray(data) ? data : [];
  const transactions = allTransactions.filter((txn) =>
    matchesRecordQuery(txn.merchant_name || txn.name, txn.amount, query)
  );

  const showTab = (next) => {
    setTab(next);
    if (next === "history") setHistoryMounted(true);
  };

  return (
    <section className='transactions-7qm2'>
      <div className='transactions-7qm2__tabs' role='tablist' aria-label='Activity'>
        <button
          className={`transactions-7qm2__tab ${tab === "activity" ? "active" : ""}`}
          onClick={() => showTab("activity")}
          aria-selected={tab === "activity"}
          role='tab'
          type='button'
        >
          Bank Activity
        </button>
        <button
          className={`transactions-7qm2__tab ${tab === "history" ? "active" : ""}`}
          onClick={() => showTab("history")}
          aria-selected={tab === "history"}
          role='tab'
          type='button'
        >
          Historical billing
        </button>
      </div>

      {/* one bar filters both tabs — by title or by amount */}
      <div className='transactions-7qm2__search'>
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder='Search by title or amount…'
          aria-label='Search transactions by title or amount'
        />
      </div>

      <div hidden={tab !== "activity"} role='tabpanel'>
        {loading && allTransactions.length === 0 ? (
          <p className='transactions-7qm2__empty'>Loading transactions…</p>
        ) : allTransactions.length === 0 ? (
          <p className='transactions-7qm2__empty'>
            No transactions yet — connect your bank.
          </p>
        ) : transactions.length === 0 ? (
          <p className='transactions-7qm2__empty'>No matches for “{query}”.</p>
        ) : (
          <ul className='transactions-7qm2__list'>
            {transactions.map((txn) => {
              const isExpense = txn.amount > 0;

              return (
                <li key={txn.transaction_id} className='transactions-7qm2__row'>
                  <div className='transactions-7qm2__details'>
                    <span className='transactions-7qm2__name'>
                      {txn.merchant_name || txn.name}
                      {txn.pending && (
                        <span className='transactions-7qm2__pending'>pending</span>
                      )}
                    </span>
                    <span className='transactions-7qm2__meta'>
                      {formatCategory(txn.category)}
                      {txn.category ? " · " : ""}
                      {txn.account_name}
                      {txn.account_mask ? ` ••${txn.account_mask}` : ""}
                      {" · "}
                      {formatDate(txn.date)}
                    </span>
                  </div>

                  <span
                    className={`transactions-7qm2__amount ${
                      isExpense ? "expense" : "income"
                    }`}
                  >
                    {isExpense ? "-" : "+"}
                    {formatUSD(Math.abs(txn.amount))}
                  </span>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {historyMounted && (
        <div hidden={tab !== "history"} role='tabpanel'>
          <BillingHistory query={query} />
        </div>
      )}
    </section>
  );
};
