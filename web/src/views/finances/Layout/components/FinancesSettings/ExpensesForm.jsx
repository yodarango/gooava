import { useAppContext } from "../../../../context/appContextProvider";
import { PaydayCalendar } from "../PaydayCalendar/PaydayCalendar";
import { API_GET_PLAID_TRANSACTIONS, API_POST_EXPENSES } from "@constants";
import { useEffect, useMemo, useState } from "react";
import { useGet, usePost } from "@utils";
import { Button, Input } from "@ds";

// styles
import "./ExpensesForm.css";

// Until a categories section exists, "misc" is the only option.
const EXPENSE_CATEGORIES = ["misc"];

const RECURRING_OPTIONS = [
  { value: "", label: "never" },
  { value: "daily", label: "daily" },
  { value: "weekly", label: "weekly" },
  { value: "biweekly", label: "biweekly" },
  { value: "monthly", label: "monthly" },
  { value: "quarterly", label: "quarterly" },
  { value: "yearly", label: "yearly" },
];

// values follow the JS Date convention: 0 = Sunday … 6 = Saturday
const DAYS_OF_WEEK = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

const MONTH_NAMES = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
];

/*********************************************************************************************************
 * Form to enter an expense manually. The label field autocompletes from the labels of synced Plaid
 * transactions (merchant name, falling back to the transaction name) as a typing aid only — the label
 * is always saved as a plain string, so expenses and transactions can later be linked by string
 * equivalence.
 * ******************************************************************************************************
 */
export const ExpensesForm = () => {
  const { showToast } = useAppContext();

  const [category, setCategory] = useState(EXPENSE_CATEGORIES[0]);
  const [amount, setAmount] = useState("");
  const [label, setLabel] = useState("");
  const [showSuggestions, setShowSuggestions] = useState(false);

  // recurring schedule state
  const [recurring, setRecurring] = useState("");
  const [dayOfWeek, setDayOfWeek] = useState(5); // Friday
  const [selectedDays, setSelectedDays] = useState([]); // biweekly (2) / monthly (1)
  const [selectedDate, setSelectedDate] = useState(null); // { month, day } — quarterly / yearly
  const [selectedMonths, setSelectedMonths] = useState([]); // quarterly: 3 extra months

  // source for the label autocomplete
  const transactions = useGet({ url: API_GET_PLAID_TRANSACTIONS + "?limit=200" });

  const expense = usePost({
    url: API_POST_EXPENSES,
    callback: (data) => {
      if (!data) return;

      showToast({ message: "Expense added", type: "success" });
      setCategory(EXPENSE_CATEGORIES[0]);
      setAmount("");
      setLabel("");
      setRecurring("");
      setDayOfWeek(5);
      setSelectedDays([]);
      setSelectedDate(null);
      setSelectedMonths([]);
    },
  });

  // surface request errors as toasts
  useEffect(() => {
    if (expense.error) showToast({ message: String(expense.error), type: "danger" });
  }, [expense.error]);

  // unique transaction labels, merchant name preferred
  const knownLabels = useMemo(() => {
    const list = Array.isArray(transactions.data) ? transactions.data : [];
    const seen = new Set();
    const labels = [];

    for (const txn of list) {
      const value = (txn.merchant_name || txn.name || "").trim();
      const key = value.toLowerCase();
      if (value && !seen.has(key)) {
        seen.add(key);
        labels.push(value);
      }
    }

    return labels;
  }, [transactions.data]);

  // matches for what the user is typing — prefix matches rank first
  const suggestions = useMemo(() => {
    const query = label.trim().toLowerCase();
    if (!query) return [];

    return knownLabels
      .filter((known) => {
        const lower = known.toLowerCase();
        return lower !== query && lower.includes(query);
      })
      .sort((a, b) => {
        const aStarts = a.toLowerCase().startsWith(query) ? 0 : 1;
        const bStarts = b.toLowerCase().startsWith(query) ? 0 : 1;
        return aStarts - bStarts || a.localeCompare(b);
      })
      .slice(0, 6);
  }, [label, knownLabels]);

  const pickSuggestion = (value) => {
    setLabel(value);
    setShowSuggestions(false);
  };

  // switching frequency clears any schedule picked for the previous one
  const handleRecurringChange = (value) => {
    setRecurring(value);
    setSelectedDays([]);
    setSelectedDate(null);
    setSelectedMonths([]);
  };

  // biweekly / monthly: only the day-of-month number matters, so a pick shows in every month view
  const handleDayOfMonthSelect = ({ day }) => {
    if (recurring === "biweekly") {
      setSelectedDays((prev) =>
        prev.includes(day)
          ? prev.filter((d) => d !== day)
          : prev.length >= 2
            ? [...prev.slice(1), day] // already 2 — drop the oldest pick
            : [...prev, day]
      );
    } else if (recurring === "monthly") {
      setSelectedDays((prev) => (prev[0] === day ? [] : [day]));
    }
  };

  // quarterly / yearly: the specific month + day matters
  const handleDateSelect = ({ month, day }) => {
    if (recurring === "quarterly" && selectedDate && selectedDate.month !== month) {
      setSelectedMonths([]); // anchor month changed — clear the extra picks
    }
    setSelectedDate({ month, day });
  };

  // quarterly: the calendar's month is the anchor (locked); pick 3 more
  const toggleQuarterMonth = (month) => {
    if (selectedDate && month === selectedDate.month) return;
    setSelectedMonths((prev) =>
      prev.includes(month)
        ? prev.filter((m) => m !== month)
        : prev.length >= 3
          ? [...prev.slice(1), month] // already 3 — drop the oldest pick
          : [...prev, month]
    );
  };

  // builds { rule } ready to send, or { error } when the schedule is incomplete
  const buildRecurRule = () => {
    switch (recurring) {
      case "":
      case "daily":
        return {};
      case "weekly":
        return { rule: { day_of_week: dayOfWeek } };
      case "biweekly":
        if (selectedDays.length !== 2) return { error: "Pick 2 days of the month" };
        return { rule: { days_of_month: [...selectedDays].sort((a, b) => a - b) } };
      case "monthly":
        if (selectedDays.length !== 1) return { error: "Pick a day of the month" };
        return { rule: { day_of_month: selectedDays[0] } };
      case "quarterly":
        if (!selectedDate) return { error: "Pick the date of the first occurrence" };
        if (selectedMonths.length !== 3) {
          return { error: "Pick 3 more months for the quarterly schedule" };
        }
        return {
          rule: {
            day: selectedDate.day,
            months: [selectedDate.month, ...selectedMonths].sort((a, b) => a - b),
          },
        };
      case "yearly":
        if (!selectedDate) return { error: "Pick the month and day" };
        return { rule: { month: selectedDate.month, day: selectedDate.day } };
      default:
        return {};
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    const trimmedLabel = label.trim();
    const parsedAmount = parseFloat(amount);

    if (trimmedLabel === "") {
      showToast({ message: "Please enter a label", type: "danger" });
      return;
    }

    if (amount.trim() === "" || isNaN(parsedAmount) || parsedAmount <= 0) {
      showToast({ message: "Please enter a valid amount (more than 0)", type: "danger" });
      return;
    }

    const { rule, error } = buildRecurRule();
    if (error) {
      showToast({ message: error, type: "danger" });
      return;
    }

    expense.post({
      category,
      amount: Math.round(parsedAmount * 100) / 100,
      label: trimmedLabel,
      recurring,
      ...(rule ? { recur_rule: rule } : {}),
    });
  };

  return (
    <form className='expenses-form-2nb5' onSubmit={handleSubmit}>
      <div className='expenses-form-2nb5__field'>
        <label className='expenses-form-2nb5__label' htmlFor='expense-category'>
          Category
        </label>
        <select
          className='expenses-form-2nb5__select'
          onChange={(e) => setCategory(e.target.value)}
          id='expense-category'
          value={category}
        >
          {EXPENSE_CATEGORIES.map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
      </div>

      <div className='expenses-form-2nb5__field'>
        <label className='expenses-form-2nb5__label' htmlFor='expense-amount'>
          Amount
        </label>
        <Input
          onChange={(e) => setAmount(e.target.value)}
          className='expenses-form-2nb5__input'
          placeholder='0.00'
          id='expense-amount'
          inputMode='decimal'
          value={amount}
          type='number'
          step='0.01'
          min='0'
          required
        />
      </div>

      <div className='expenses-form-2nb5__field expenses-form-2nb5__label-field'>
        <label className='expenses-form-2nb5__label' htmlFor='expense-label'>
          Label
        </label>
        <Input
          onChange={(e) => {
            setLabel(e.target.value);
            setShowSuggestions(true);
          }}
          onFocus={() => setShowSuggestions(true)}
          onBlur={() => setTimeout(() => setShowSuggestions(false), 150)}
          className='expenses-form-2nb5__input'
          placeholder='e.g. Netflix'
          autoComplete='off'
          id='expense-label'
          value={label}
          type='text'
          required
        />

        {showSuggestions && suggestions.length > 0 && (
          <ul className='expenses-form-2nb5__suggestions'>
            {suggestions.map((suggestion) => (
              <li key={suggestion}>
                <button
                  onMouseDown={(e) => {
                    // mousedown fires before the input's blur — pick the suggestion first
                    e.preventDefault();
                    pickSuggestion(suggestion);
                  }}
                  className='expenses-form-2nb5__suggestion'
                  type='button'
                >
                  {suggestion}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className='expenses-form-2nb5__field'>
        <label className='expenses-form-2nb5__label' htmlFor='expense-recurring'>
          Recurring
        </label>
        <select
          onChange={(e) => handleRecurringChange(e.target.value)}
          className='expenses-form-2nb5__select'
          id='expense-recurring'
          value={recurring}
        >
          {RECURRING_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {recurring === "weekly" && (
        <div className='expenses-form-2nb5__field'>
          <label className='expenses-form-2nb5__label' htmlFor='expense-day-of-week'>
            Day of the week
          </label>
          <select
            onChange={(e) => setDayOfWeek(Number(e.target.value))}
            className='expenses-form-2nb5__select'
            id='expense-day-of-week'
            value={dayOfWeek}
          >
            {DAYS_OF_WEEK.map((name, index) => (
              <option key={name} value={index}>
                {name}
              </option>
            ))}
          </select>
        </div>
      )}

      {(recurring === "biweekly" || recurring === "monthly") && (
        <div className='expenses-form-2nb5__field'>
          <p className='expenses-form-2nb5__label'>
            {recurring === "biweekly"
              ? "Pick 2 days of the month"
              : "Pick the day of the month"}
            {selectedDays.length > 0 && (
              <span className='expenses-form-2nb5__picks'>
                {" "}— {[...selectedDays].sort((a, b) => a - b).join(" & ")}
              </span>
            )}
          </p>
          <PaydayCalendar
            onDaySelect={handleDayOfMonthSelect}
            selectedDays={selectedDays}
          />
        </div>
      )}

      {recurring === "quarterly" && (
        <div className='expenses-form-2nb5__field'>
          <p className='expenses-form-2nb5__label'>Pick the date of the first occurrence</p>
          <PaydayCalendar
            onDaySelect={handleDateSelect}
            selectedDate={selectedDate}
          />

          {selectedDate && (
            <>
              <p className='expenses-form-2nb5__label'>
                {MONTH_NAMES[selectedDate.month - 1]} is the first month — pick 3 more
              </p>
              <div className='expenses-form-2nb5__months'>
                {MONTH_NAMES.map((name, index) => {
                  const month = index + 1;
                  const isAnchor = selectedDate.month === month;
                  const isPicked = selectedMonths.includes(month);

                  const className = [
                    "expenses-form-2nb5__month",
                    isAnchor ? "anchor" : "",
                    isPicked ? "picked" : "",
                  ]
                    .filter(Boolean)
                    .join(" ");

                  return (
                    <button
                      onClick={() => toggleQuarterMonth(month)}
                      aria-pressed={isAnchor || isPicked}
                      title={isAnchor ? "First month (from the date above)" : undefined}
                      className={className}
                      disabled={isAnchor}
                      type='button'
                      key={month}
                    >
                      {name}
                    </button>
                  );
                })}
              </div>
            </>
          )}
        </div>
      )}

      {recurring === "yearly" && (
        <div className='expenses-form-2nb5__field'>
          <p className='expenses-form-2nb5__label'>Pick the month and day</p>
          <PaydayCalendar
            onDaySelect={handleDateSelect}
            selectedDate={selectedDate}
          />
        </div>
      )}

      <Button primary type='submit' isLoading={expense.loading} className='w-100'>
        Add expense
      </Button>
    </form>
  );
};
