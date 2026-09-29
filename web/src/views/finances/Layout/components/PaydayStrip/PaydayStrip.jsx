import { upcomingPaydays } from "@utils";

// styles
import "./PaydayStrip.css";

/*********************************************************************************************************
 * Horizontal, scrollable strip of upcoming paydays. Each entry shows the month (small black letters)
 * above a pink circle with the day — matching the PaydayCalendar's payday styling.
 * ******************************************************************************************************
 */
export const PaydayStrip = () => {
  const paydays = upcomingPaydays(12);

  return (
    <div className='payday-strip-9k2x'>
      {paydays.map((date) => (
        <div key={date.getTime()} className='payday-strip-9k2x__item'>
          <span className='payday-strip-9k2x__month'>
            {date.toLocaleDateString("en-US", { month: "short" })}
          </span>
          <span className='payday-strip-9k2x__circle'>{date.getDate()}</span>
        </div>
      ))}
    </div>
  );
};
