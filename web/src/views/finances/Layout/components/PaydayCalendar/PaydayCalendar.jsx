import { ChevronLeft, ChevronRight } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { isPayday, stripTime } from "@utils";
import { useState } from "react";

// styles
import "./PaydayCalendar.css";

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

// Month grid cells: leading blanks (week starts Sunday) followed by day numbers
function buildCells(year, month) {
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const leadingBlanks = new Date(year, month, 1).getDay();

  const cells = [];
  for (let i = 0; i < leadingBlanks; i++) cells.push(null);
  for (let day = 1; day <= daysInMonth; day++) cells.push(day);
  return cells;
}

/*********************************************************************************************************
 * Custom-made calendar (no library) that highlights paydays — every other Friday, anchored to the
 * Oct 2, 2026 payday — and today's date.
 * ******************************************************************************************************
 */
export const PaydayCalendar = () => {
  const today = stripTime(new Date());
  const [viewed, setViewed] = useState(() => ({
    year: today.getFullYear(),
    month: today.getMonth(),
  }));

  const shiftMonth = (delta) =>
    setViewed((prev) => {
      const date = new Date(prev.year, prev.month + delta, 1);
      return { year: date.getFullYear(), month: date.getMonth() };
    });

  const monthLabel = new Date(viewed.year, viewed.month, 1).toLocaleDateString(
    "en-US",
    { month: "long", year: "numeric" }
  );

  const cells = buildCells(viewed.year, viewed.month);

  return (
    <div className='payday-calendar-7m3q'>
      <div className='payday-calendar-7m3q__header'>
        <button
          onClick={() => shiftMonth(-1)}
          aria-label='Previous month'
          type='button'
        >
          <MorphIcon icon={ChevronLeft} label='Previous month' size={20} />
        </button>
        <h3>{monthLabel}</h3>
        <button
          onClick={() => shiftMonth(1)}
          aria-label='Next month'
          type='button'
        >
          <MorphIcon icon={ChevronRight} label='Next month' size={20} />
        </button>
      </div>

      <div className='payday-calendar-7m3q__grid'>
        {WEEKDAYS.map((weekday) => (
          <span key={weekday} className='payday-calendar-7m3q__weekday'>
            {weekday}
          </span>
        ))}

        {cells.map((day, index) => {
          if (day === null) {
            return <span key={`blank-${index}`} />;
          }

          const date = new Date(viewed.year, viewed.month, day);
          const payday = isPayday(date);
          const isToday = stripTime(date).getTime() === today.getTime();

          const className = [
            "payday-calendar-7m3q__day",
            payday ? "payday" : "",
            isToday ? "today" : "",
          ]
            .filter(Boolean)
            .join(" ");

          return (
            <span
              key={day}
              className={className}
              title={payday ? "Payday 💸" : undefined}
            >
              {day}
            </span>
          );
        })}
      </div>

      <p className='payday-calendar-7m3q__legend'>
        <span className='payday-calendar-7m3q__legend-payday' /> Payday (every
        other Friday)
      </p>
    </div>
  );
};
