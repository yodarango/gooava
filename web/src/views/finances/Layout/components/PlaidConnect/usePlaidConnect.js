import { useAppContext } from "../../../../context/appContextProvider";
import { API_POST_PLAID_LINK_TOKEN, API_POST_PLAID_EXCHANGE } from "@constants";
import { usePost } from "@utils";
import { useEffect } from "react";

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

// In-flight link token, stashed so the bank's OAuth redirect back to /finances
// can resume the exact same Link session (the page reloads fresh, the original
// handler is gone).
const PLAID_LINK_TOKEN_KEY = "gooava:plaid_link_token";

// The finances page mounts the flow twice (connect card + settings drawer) —
// the OAuth resume below must run exactly once per page load.
let plaidOAuthResumeHandled = false;

/*********************************************************************************************************
 * Shared pieces of the Link flow used by both hooks: exchange a public token server-side (which also
 * runs the initial transactions sync), create + open the Link handler, toast on errors.
 *
 * onConnected is called after a successful exchange.
 * ******************************************************************************************************
 */
const usePlaidLinkFlow = (onConnected) => {
  const { showToast } = useAppContext();

  const exchange = usePost({
    url: API_POST_PLAID_EXCHANGE,
    callback: (data) => {
      if (!data) return;

      // The exchange endpoint already imported the transaction history.
      showToast({ message: "Bank connected — transactions imported", type: "success" });
      if (onConnected) onConnected();
    },
  });

  const openLink = async (token) => {
    try {
      await loadPlaidScript();
    } catch (error) {
      showToast({ message: error.message, type: "danger" });
      return;
    }

    const handler = window.Plaid.create({
      token,
      onSuccess: (publicToken, metadata) => {
        sessionStorage.removeItem(PLAID_LINK_TOKEN_KEY);
        exchange.post({
          public_token: publicToken,
          institution_name: metadata?.institution?.name || "",
        });
      },
      onExit: (error) => {
        if (error) {
          sessionStorage.removeItem(PLAID_LINK_TOKEN_KEY);
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
  };

  return { openLink, exchange };
};

/*********************************************************************************************************
 * "Connect a bank" flow: fetch a link token, stash it (OAuth redirect insurance), open Plaid Link.
 * Used by the PlaidConnect card and the Finances settings drawer ("Add bank").
 * ******************************************************************************************************
 */
export const usePlaidConnect = (onConnected) => {
  const { showToast } = useAppContext();
  const { openLink, exchange } = usePlaidLinkFlow(onConnected);

  const linkToken = usePost({
    url: API_POST_PLAID_LINK_TOKEN,
    callback: (data) => {
      if (!data?.link_token) return;

      // stash the token so an OAuth bank's redirect back to /finances can
      // resume this exact Link session (see usePlaidOAuthResume)
      sessionStorage.setItem(PLAID_LINK_TOKEN_KEY, data.link_token);

      openLink(data.link_token);
    },
  });

  // surface request errors as toasts
  useEffect(() => {
    const message = linkToken.error || exchange.error;
    if (message) showToast({ message: String(message), type: "danger" });
  }, [linkToken.error, exchange.error]);

  return {
    connect: () => linkToken.post({}),
    isLoading: linkToken.loading || exchange.loading,
  };
};

/*********************************************************************************************************
 * Resumes an interrupted OAuth bank flow. OAuth banks send the user to the bank's own site; the bank
 * then redirects back to /finances?oauth_state_id=... Re-creating Link with the SAME stashed link
 * token lets the SDK detect that state and finish the flow (onSuccess then exchanges as usual).
 *
 * Mounted once by the finances Layout.
 * ******************************************************************************************************
 */
export const usePlaidOAuthResume = (onConnected) => {
  const { showToast } = useAppContext();
  const { openLink, exchange } = usePlaidLinkFlow(onConnected);

  useEffect(() => {
    if (plaidOAuthResumeHandled) return;

    const params = new URLSearchParams(window.location.search);
    const oauthStateId = params.get("oauth_state_id");
    const storedToken = sessionStorage.getItem(PLAID_LINK_TOKEN_KEY);

    if (!oauthStateId || !storedToken) return;
    plaidOAuthResumeHandled = true;

    // strip the oauth params from the address bar
    window.history.replaceState({}, "", window.location.pathname);

    openLink(storedToken);
  }, []);

  // surface exchange errors as toasts
  useEffect(() => {
    if (exchange.error) showToast({ message: String(exchange.error), type: "danger" });
  }, [exchange.error]);
};
