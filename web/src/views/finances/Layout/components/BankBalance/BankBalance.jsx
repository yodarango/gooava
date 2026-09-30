import { useAppContext } from "../../../../context/appContextProvider";
import { API_GET_PLAID_ACCOUNTS, API_POST_PLAID_SYNC } from "@constants";
import { useGet, usePost } from "@utils";
import { useEffect } from "react";
import { Button } from "@ds";

// styles
import "./BankBalance.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

/*********************************************************************************************************
 * Replaces the manual "In the bank" card once a bank is connected. Lists every connected checking
 * account with its balance in smaller text, then the total sum across all of them. Also hosts the
 * Sync button, since the bank connection card is hidden while connected.
 * ******************************************************************************************************
 */
export const BankBalance = (props) => {
  const { refreshKey, onSynced } = props;
  const { showToast } = useAppContext();

  const accounts = useGet({
    url: API_GET_PLAID_ACCOUNTS,
    dependencies: [refreshKey],
  });

  const sync = usePost({
    url: API_POST_PLAID_SYNC,
    callback: (data) => {
      if (!data) return;

      const failed = data.failed || 0;
      showToast({
        message:
          `Bank synced — ${data.added} new, ${data.modified} updated, ${data.removed} removed` +
          (failed > 0 ? ` — ${failed} source${failed > 1 ? "s" : ""} failed` : ""),
        type: failed > 0 ? "warning" : "success",
      });
      if (onSynced) onSynced();
    },
  });

  // surface sync errors as toasts
  useEffect(() => {
    if (sync.error) showToast({ message: String(sync.error), type: "danger" });
  }, [sync.error]);

  const checkingAccounts = (Array.isArray(accounts.data) ? accounts.data : []).filter(
    (account) => account.subtype === "checking"
  );

  const total = checkingAccounts.reduce(
    (sum, account) => sum + (account.current_balance || 0),
    0
  );

  return (
    <section className='bank-balance-6pw1'>
      <div className='bank-balance-6pw1__header'>
        <p className='bank-balance-6pw1__label'>In the bank</p>
        <Button secondary onClick={() => sync.post({})} isLoading={sync.loading}>
          Sync
        </Button>
      </div>

      {checkingAccounts.length > 0 && (
        <ul className='bank-balance-6pw1__accounts'>
          {checkingAccounts.map((account) => (
            <li key={account.account_id} className='bank-balance-6pw1__account'>
              <span className='bank-balance-6pw1__account-name'>
                {account.name}
                {account.mask ? ` ••${account.mask}` : ""}
              </span>
              <span className='bank-balance-6pw1__account-amount'>
                {formatUSD(account.current_balance || 0)}
              </span>
            </li>
          ))}
        </ul>
      )}

      <p className='bank-balance-6pw1__total'>{formatUSD(total)}</p>

      <p className='bank-balance-6pw1__hint'>
        Pulled from your connected bank — hit Sync to refresh.
      </p>
    </section>
  );
};
