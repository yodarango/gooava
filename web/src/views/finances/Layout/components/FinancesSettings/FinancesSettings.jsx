import { API_GET_PLAID_ACCOUNTS, API_GET_PLAID_ITEMS } from "@constants";
import { usePlaidConnect } from "../PlaidConnect/usePlaidConnect";
import { Tabs, TabItem, TabContent, Button } from "@ds";
import { Settings, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { createPortal } from "react-dom";
import { useEffect, useState } from "react";
import { ExpensesForm } from "./ExpensesForm";
import { useGet } from "@utils";

// styles
import "./FinancesSettings.css";

/*********************************************************************************************************
 * The drawer body — mounted only while the drawer is open so the items/accounts are refetched
 * every time it opens.
 * ******************************************************************************************************
 */
const SettingsDrawer = (props) => {
  const { onClose, onChanged } = props;

  const items = useGet({ url: API_GET_PLAID_ITEMS });
  const accounts = useGet({ url: API_GET_PLAID_ACCOUNTS });

  const { connect, isLoading } = usePlaidConnect(() => {
    items.refetch();
    accounts.refetch();
    if (onChanged) onChanged();
  });

  // close on Escape + lock body scroll while open
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
    };
  }, [onClose]);

  const itemList = Array.isArray(items.data) ? items.data : [];
  const accountList = Array.isArray(accounts.data) ? accounts.data : [];

  return (
    <div className='finances-settings-8h3j' role='dialog' aria-modal='true' aria-label='Connected banks'>
      <div className='finances-settings-8h3j__backdrop' onClick={onClose} />

      <aside className='finances-settings-8h3j__drawer'>
        <div className='finances-settings-8h3j__header'>
          <h2 className='finances-settings-8h3j__title'>Connected banks</h2>
          <button
            className='finances-settings-8h3j__close'
            onClick={onClose}
            aria-label='Close'
            type='button'
          >
            <MorphIcon icon={X} label='Close' size={20} />
          </button>
        </div>

        <Tabs>
          <TabItem>Accounts</TabItem>
          <TabItem>Expenses</TabItem>

          <TabContent>
            <div className='finances-settings-8h3j__body'>
              {items.loading ? (
                <p className='finances-settings-8h3j__empty'>Loading…</p>
              ) : itemList.length === 0 ? (
                <p className='finances-settings-8h3j__empty'>No banks connected yet.</p>
              ) : (
                <ul className='finances-settings-8h3j__list'>
                  {itemList.map((item) => {
                    const itemAccounts = accountList.filter(
                      (account) => account.plaid_item_id === item.id
                    );

                    return (
                      <li key={item.id} className='finances-settings-8h3j__item'>
                        <p className='finances-settings-8h3j__institution'>
                          {item.institution_name || "Connected institution"}
                        </p>
                        {itemAccounts.length > 0 && (
                          <ul className='finances-settings-8h3j__accounts'>
                            {itemAccounts.map((account) => (
                              <li key={account.account_id}>
                                {account.name}
                                {account.mask ? ` ••${account.mask}` : ""}
                              </li>
                            ))}
                          </ul>
                        )}
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>

            <div className='finances-settings-8h3j__footer'>
              <Button primary onClick={connect} isLoading={isLoading} className='w-100'>
                Add bank
              </Button>
            </div>
          </TabContent>

          <TabContent>
            <div className='finances-settings-8h3j__body'>
              <ExpensesForm />
            </div>
          </TabContent>
        </Tabs>
      </aside>
    </div>
  );
};

/*********************************************************************************************************
 * Settings gear shown next to the Finances title. Opens a right-side drawer listing every connected
 * Plaid source (with its accounts) and an "Add bank" button to link another institution.
 * ******************************************************************************************************
 */
export const FinancesSettings = (props) => {
  const { onChanged } = props;
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        className='finances-settings-8h3j__trigger'
        onClick={() => setOpen(true)}
        aria-label='Bank settings'
        type='button'
      >
        <MorphIcon icon={Settings} label='Bank settings' size={22} />
      </button>

      {open &&
        createPortal(
          <SettingsDrawer onClose={() => setOpen(false)} onChanged={onChanged} />,
          document.body
        )}
    </>
  );
};
