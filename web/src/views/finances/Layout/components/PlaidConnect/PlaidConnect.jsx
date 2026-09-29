import { useAppContext } from "../../../../context/appContextProvider";
import {
  API_POST_PLAID_LINK_TOKEN,
  API_POST_PLAID_EXCHANGE,
} from "@constants";
import { usePost } from "@utils";
import { useEffect } from "react";
import { Button } from "@ds";

// styles
import "./PlaidConnect.css";

/*********************************************************************************************************
 * Loads the Plaid Link script exactly once (module-level singleton promise). The script is pulled from
 * Plaid's CDN so no npm dependency is needed.
 * ******************************************************************************************************
 */
let plaidScriptPromise = null;

const loadPlaidScript = () => {
  if (window.Plaid) return Promise.resolve();
  if (plaidScriptPromise) return plaidScriptPromise;

  plaidScriptPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = "https://cdn.plaid.com/link/v2/stable/link-initialize.js";
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("Could not load Plaid Link"));
    document.body.appendChild(script);
  });

  return plaidScriptPromise;
};

/*********************************************************************************************************
 * "Connect bank" card — only rendered while no bank is linked. Clicking opens Plaid Link; on success
 * the public token is exchanged server-side (which also runs the initial transactions sync).
 * ******************************************************************************************************
 */
export const PlaidConnect = (props) => {
  const { onSynced } = props;
  const { showToast } = useAppContext();

  const exchange = usePost({
    url: API_POST_PLAID_EXCHANGE,
    callback: (data) => {
      if (!data) return;

      // The exchange endpoint already imported the transaction history.
      showToast({ message: "Bank connected — transactions imported", type: "success" });
      if (onSynced) onSynced();
    },
  });

  const linkToken = usePost({
    url: API_POST_PLAID_LINK_TOKEN,
    callback: async (data) => {
      if (!data?.link_token) return;

      try {
        await loadPlaidScript();
      } catch (error) {
        showToast({ message: error.message, type: "danger" });
        return;
      }

      const handler = window.Plaid.create({
        token: data.link_token,
        onSuccess: (publicToken, metadata) => {
          exchange.post({
            public_token: publicToken,
            institution_name: metadata?.institution?.name || "",
          });
        },
        onExit: (error) => {
          if (error) {
            showToast({
              message:
                error.display_message ||
                error.error_message ||
                "Bank connection could not be completed",
              type: "danger",
            });
          }
        },
      });

      handler.open();
    },
  });

  // surface request errors as toasts
  useEffect(() => {
    const message = linkToken.error || exchange.error;
    if (message) showToast({ message: String(message), type: "danger" });
  }, [linkToken.error, exchange.error]);

  const isLoading = linkToken.loading || exchange.loading;

  const handleClick = () => linkToken.post({});

  return (
    <section className='plaid-connect-3xk7'>
      <div className='plaid-connect-3xk7__info'>
        <p className='plaid-connect-3xk7__label'>Bank connection</p>
        <p className='plaid-connect-3xk7__hint'>
          Link your bank to track spending automatically.
        </p>
      </div>

      <Button primary onClick={handleClick} isLoading={isLoading}>
        Connect bank
      </Button>
    </section>
  );
};
