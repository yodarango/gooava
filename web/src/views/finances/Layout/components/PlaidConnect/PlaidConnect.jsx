import { usePlaidConnect } from "./usePlaidConnect";
import { Button } from "@ds";

// styles
import "./PlaidConnect.css";

/*********************************************************************************************************
 * "Connect bank" card — only rendered while no bank is linked. Clicking opens Plaid Link; on success
 * the public token is exchanged server-side (which also runs the initial transactions sync).
 * ******************************************************************************************************
 */
export const PlaidConnect = (props) => {
  const { onSynced } = props;
  const { connect, isLoading } = usePlaidConnect(onSynced);

  return (
    <section className='plaid-connect-3xk7'>
      <div className='plaid-connect-3xk7__info'>
        <p className='plaid-connect-3xk7__label'>Bank connection</p>
        <p className='plaid-connect-3xk7__hint'>
          Link your bank to track spending automatically.
        </p>
      </div>

      <Button primary onClick={connect} isLoading={isLoading}>
        Connect bank
      </Button>
    </section>
  );
};
