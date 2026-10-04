import { API_GET_EXPENSES } from "@constants";
import {
  nextPaydayDate,
  totalUpcomingBills,
  upcomingBills,
  stripTime,
  useGet,
} from "@utils";

// styles
import "./UpcomingBills.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

/*********************************************************************************************************
 * Home page card: how much money is needed to cover the recurring bills (added via the Finances
 * settings drawer) that come due between today and the next payday. Each occurrence is listed with
 * its due date. When today is payday, the window runs until the following payday instead.
 * ******************************************************************************************************
 */
export const UpcomingBills = () => {
  const expenses = useGet({ url: API_GET_EXPENSES });

  // window: [today, next payday) — or until the following payday when today is payday
  const today = stripTime(new Date());
  let payday = nextPaydayDate(today);
  if (payday.getTime() <= today.getTime()) {
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);
    payday = nextPaydayDate(tomorrow);
  }

  const bills = upcomingBills(
    Array.isArray(expenses.data) ? expenses.data : [],
    today,
    payday
  );
  const total = totalUpcomingBills(bills);
  const loading = expenses.loading && !expenses.data;

  const paydayLabel = payday.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
  });

  return (
    <section className='upcoming-bills-5xk3'>
      <p className='upcoming-bills-5xk3__label'>Needed until payday</p>

      <p className='upcoming-bills-5xk3__total'>
        {loading ? "—" : formatUSD(total)}
      </p>

      <p className='upcoming-bills-5xk3__hint'>
        {loading
          ? "Loading…"
          : bills.length === 0
            ? `No bills due before ${paydayLabel} 🎉`
            : `${bills.length} bill${bills.length === 1 ? "" : "s"} due before ${paydayLabel}`}
      </p>

      {bills.length > 0 && (
        <ul className='upcoming-bills-5xk3__list'>
          {bills.map((bill) => (
            <li
              key={`${bill.expense.id}-${bill.date.getTime()}`}
              className='upcoming-bills-5xk3__row'
            >
              <div className='upcoming-bills-5xk3__details'>
                <span className='upcoming-bills-5xk3__name'>
                  {bill.expense.label}
                </span>
                <span className='upcoming-bills-5xk3__meta'>
                  {bill.expense.category}
                  {" · "}
                  {bill.date.getTime() === today.getTime()
                    ? "today"
                    : bill.date.toLocaleDateString("en-US", {
                        weekday: "short",
                        month: "short",
                        day: "numeric",
                      })}
                </span>
              </div>

              <span className='upcoming-bills-5xk3__amount'>
                {formatUSD(bill.expense.amount)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
};
