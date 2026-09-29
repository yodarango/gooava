import { API_GET_PLAID_TRANSACTIONS } from "@constants";
import { useGet } from "@utils";

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
 * "Recent activity" card listing the synced Plaid transactions, newest first. Plaid's sign convention:
 * amount > 0 means money left the account (expense), amount < 0 means money came in.
 * ******************************************************************************************************
 */
export const Transactions = (props) => {
  const { refreshKey } = props;

  const { data, loading } = useGet({
    url: API_GET_PLAID_TRANSACTIONS + "?limit=50",
    dependencies: [refreshKey],
  });

  const transactions = Array.isArray(data) ? data : [];

  return (
    <section className='transactions-7qm2'>
      <p className='transactions-7qm2__label'>Recent activity</p>

      {loading && transactions.length === 0 ? (
        <p className='transactions-7qm2__empty'>Loading transactions…</p>
      ) : transactions.length === 0 ? (
        <p className='transactions-7qm2__empty'>
          No transactions yet — connect your bank.
        </p>
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
    </section>
  );
};
