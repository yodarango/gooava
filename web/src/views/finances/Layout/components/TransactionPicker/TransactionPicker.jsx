import { API_GET_PLAID_TRANSACTIONS } from "@constants";
import { Button, Input, Modal } from "@ds";
import { useMemo, useState } from "react";
import { useGet } from "@utils";

// styles
import "./TransactionPicker.css";

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

/*********************************************************************************************************
 * Modal for picking one of the synced Plaid transactions — used to link an expense (or a single
 * past-due bill) to the transaction that pays it. The filter matches merchant/transaction names
 * as text and amounts as a dollar substring ("12" or "12.99").
 *
 * Props: open, onClose, onSelect(txn), onUnlink (optional — renders an Unlink footer button),
 * title. Mount it only while open so the transactions refetch every time.
 * ******************************************************************************************************
 */
export const TransactionPicker = (props) => {
  const { open, onClose, onSelect, onUnlink, title = "Link a transaction" } = props;

  const [query, setQuery] = useState("");

  const transactions = useGet({ url: API_GET_PLAID_TRANSACTIONS + "?limit=200" });
  const list = useMemo(
    () => (Array.isArray(transactions.data) ? transactions.data : []),
    [transactions.data]
  );

  // text matches the merchant/transaction name; a number-ish query matches the dollar amount
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return list;

    const amountQuery = q.replace(/^\$/, "");
    return list.filter((txn) => {
      const name = `${txn.merchant_name || ""} ${txn.name || ""}`.toLowerCase();
      if (name.includes(q)) return true;
      return Math.abs(txn.amount).toFixed(2).includes(amountQuery);
    });
  }, [list, query]);

  return (
    <Modal open={open} onClose={onClose} title={title} zIndex={16}>
      <div className='transaction-picker-9qw1'>
        <Input
          onChange={(e) => setQuery(e.target.value)}
          className='transaction-picker-9qw1__search'
          placeholder='Filter by name or amount…'
          autoComplete='off'
          value={query}
          type='text'
          autoFocus
        />

        {transactions.loading && list.length === 0 ? (
          <p className='transaction-picker-9qw1__empty'>Loading…</p>
        ) : filtered.length === 0 ? (
          <p className='transaction-picker-9qw1__empty'>
            {list.length === 0
              ? "No transactions synced yet — connect a bank first."
              : "No transactions match."}
          </p>
        ) : (
          <ul className='transaction-picker-9qw1__list'>
            {filtered.map((txn) => {
              const isExpense = txn.amount > 0;

              return (
                <li key={txn.transaction_id}>
                  <button
                    className='transaction-picker-9qw1__row'
                    onClick={() => onSelect(txn)}
                    type='button'
                  >
                    <div className='transaction-picker-9qw1__details'>
                      <span className='transaction-picker-9qw1__name'>
                        {txn.merchant_name || txn.name}
                        {txn.pending && (
                          <span className='transaction-picker-9qw1__pending'>pending</span>
                        )}
                      </span>
                      <span className='transaction-picker-9qw1__meta'>
                        {txn.account_name}
                        {txn.account_mask ? ` ••${txn.account_mask}` : ""}
                        {" · "}
                        {formatDate(txn.date)}
                      </span>
                    </div>

                    <span
                      className={`transaction-picker-9qw1__amount ${
                        isExpense ? "expense" : "income"
                      }`}
                    >
                      {isExpense ? "-" : "+"}
                      {formatUSD(Math.abs(txn.amount))}
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
        )}

        {onUnlink && (
          <Button secondary onClick={onUnlink} className='w-100'>
            Unlink
          </Button>
        )}
      </div>
    </Modal>
  );
};
