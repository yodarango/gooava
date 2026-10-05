import { useAppContext } from "../../../../context/appContextProvider";
import { PaydayCalendar } from "../PaydayCalendar/PaydayCalendar";
import {
  API_POST_INCOMES_UPDATE,
  API_POST_INCOMES,
  DAYS_OF_WEEK,
} from "@constants";
import { formatDateKey, parseDateKey } from "@utils";
import { useEffect, useState } from "react";
import { usePost } from "@utils";
import { Button, Input } from "@ds";

// styles
import "./IncomeForm.css";

const KIND_OPTIONS = [
  { value: "fixed", label: "Fixed (recurring)" },
  { value: "one_time", label: "One-time" },
];

const FREQUENCY_OPTIONS = [
  { value: "weekly", label: "weekly" },
  { value: "biweekly", label: "biweekly" },
  { value: "monthly", label: "monthly" },
  { value: "quarterly", label: "quarterly" },
  { value: "semiannually", label: "semiannually" },
  { value: "annually", label: "annually" },
];

// how many month days each frequency lets the user pick on the calendar
const PICK_LIMIT = { biweekly: 2, monthly: 1, quarterly: 1, semiannually: 1, annually: 1 };

/*********************************************************************************************************
 * Form to enter an income source. Fixed incomes ask for a frequency and a schedule picked on the
 * calendar; one-time incomes just ask for the date. Days of the month are saved as picked — when a
 * month is shorter than the picked day, the payday lands on that month's last day instead.
 *
 * Props: income (object) puts the form in edit mode and prefills every field, null means "add";
 * onDone is called after a successful save, onCancel renders a Cancel button.
 * ******************************************************************************************************
 */
export const IncomeForm = (props) => {
  const { income, onDone, onCancel } = props;
  const editing = !!income;

  const { showToast } = useAppContext();

  const [label, setLabel] = useState("");
  const [amount, setAmount] = useState("");
  const [kind, setKind] = useState("fixed"); // fixed | one_time
  const [recurring, setRecurring] = useState("biweekly");
  const [dayOfWeek, setDayOfWeek] = useState(5); // Friday (weekly)
  const [oneTimeDate, setOneTimeDate] = useState(null); // { year, month, day }
  const [selectedDays, setSelectedDays] = useState([]); // calendar picks (1-31)
  const [selectedDateMonth, setSelectedDateMonth] = useState(null); // annually: month of the pick (1-12)
  const [anchorDate, setAnchorDate] = useState(null); // biweekly every-other-week anchor { year, month, day }
  const [biweeklyMode, setBiweeklyMode] = useState("every_other_week"); // every_other_week | days_of_month

  // editing mode: prefill the form from the income (add mode starts blank)
  useEffect(() => {
    setDayOfWeek(5);
    setSelectedDays([]);
    setAnchorDate(null);
    setOneTimeDate(null);

    if (!income) {
      setLabel("");
      setAmount("");
      setKind("fixed");
      setRecurring("biweekly");
      return;
    }

    const rule = income.recur_rule || {};

    setLabel(income.label || "");
    setAmount(income.amount != null ? String(income.amount) : "");
    setKind(income.recurring ? "fixed" : "one_time");
    setRecurring(income.recurring || "biweekly");

    if (!income.recurring) {
      if (rule.date) {
        const d = parseDateKey(rule.date);
        setOneTimeDate({ year: d.getFullYear(), month: d.getMonth() + 1, day: d.getDate() });
      }
      return;
    }

    switch (income.recurring) {
      case "weekly":
        setDayOfWeek(rule.day_of_week ?? 5);
        break;
      case "biweekly":
        if (rule.mode === "days_of_month") {
          setBiweeklyMode("days_of_month");
          setSelectedDays(rule.days_of_month || []);
        } else if (rule.mode === "every_other_week" && rule.anchor) {
          setBiweeklyMode("every_other_week");
          const d = parseDateKey(rule.anchor);
          setAnchorDate({ year: d.getFullYear(), month: d.getMonth() + 1, day: d.getDate() });
        }
        break;
      case "annually":
        // the picked calendar month is part of the schedule
        if (rule.day_of_month != null && rule.month != null) {
          setSelectedDays([rule.day_of_month]);
          setSelectedDateMonth(rule.month);
        }
        break;
      default: // monthly, quarterly, semiannually
        setSelectedDays(rule.day_of_month ? [rule.day_of_month] : []);
    }
  }, [income]);

  const handleSaved = (message) => (data) => {
    if (!data) return;
    showToast({ message, type: "success" });
    if (onDone) onDone();
  };

  const create = usePost({ url: API_POST_INCOMES, callback: handleSaved("Income added") });
  const update = usePost({ url: API_POST_INCOMES_UPDATE, callback: handleSaved("Income updated") });

  const isLoading = create.loading || update.loading;

  // surface request errors as toasts
  useEffect(() => {
    const message = create.error || update.error;
    if (message) showToast({ message: String(message), type: "danger" });
  }, [create.error, update.error]);

  // switching frequency clears any schedule picked for the previous one
  const handleRecurringChange = (value) => {
    setRecurring(value);
    setSelectedDays([]);
    setAnchorDate(null);
  };

  // calendar pick for the month-day based frequencies (cap per frequency, toggle off on re-click)
  const handleDayOfMonthSelect = ({ month, day }) => {
    const limit = PICK_LIMIT[recurring] || 1;
    if (recurring === "annually") setSelectedDateMonth(month);
    setSelectedDays((prev) =>
      prev.includes(day)
        ? prev.filter((d) => d !== day)
        : prev.length >= limit
          ? [...prev.slice(1), day] // full — drop the oldest pick
          : [...prev, day]
    );
  };

  // biweekly every-other-week: the picked date is the anchor payday
  const handleAnchorSelect = ({ year, month, day }) => {
    setAnchorDate({ year, month, day });
  };

  const handleOneTimeSelect = ({ year, month, day }) => {
    setOneTimeDate({ year, month, day });
  };

  // builds { recurring, rule } ready to send, or { error } when the schedule is incomplete
  const buildPayload = () => {
    if (kind === "one_time") {
      if (!oneTimeDate) return { error: "Pick the date of the payment" };
      const date = formatDateKey(new Date(oneTimeDate.year, oneTimeDate.month - 1, oneTimeDate.day));
      return { recurring: "", rule: { date } };
    }

    switch (recurring) {
      case "weekly":
        return { recurring, rule: { day_of_week: dayOfWeek } };
      case "biweekly":
        if (biweeklyMode === "every_other_week") {
          if (!anchorDate) return { error: "Pick the payday date (it repeats every other week)" };
          const anchor = formatDateKey(new Date(anchorDate.year, anchorDate.month - 1, anchorDate.day));
          return { recurring, rule: { mode: "every_other_week", anchor } };
        }
        if (selectedDays.length !== 2) {
          return { error: "Pick the 2 days of the month you get paid" };
        }
        return {
          recurring,
          rule: { mode: "days_of_month", days_of_month: [...selectedDays].sort((a, b) => a - b) },
        };
      case "monthly":
      case "quarterly":
      case "semiannually":
        if (selectedDays.length !== 1) return { error: "Pick the day of the month" };
        return { recurring, rule: { day_of_month: selectedDays[0] } };
      case "annually": {
        if (selectedDays.length !== 1 || selectedDateMonth == null) {
          return { error: "Pick the month and day on the calendar" };
        }
        return { recurring, rule: { month: selectedDateMonth, day_of_month: selectedDays[0] } };
      }
      default:
        return { error: "Pick a frequency" };
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

    const { recurring: recurringValue, rule, error } = buildPayload();
    if (error) {
      showToast({ message: error, type: "danger" });
      return;
    }

    const payload = {
      amount: Math.round(parsedAmount * 100) / 100,
      label: trimmedLabel,
      recurring: recurringValue,
      ...(rule ? { recur_rule: rule } : {}),
    };

    if (editing) {
      update.post({ id: income.id, ...payload });
    } else {
      create.post(payload);
    }
  };

  const pickedSummary = [...selectedDays].sort((a, b) => a - b).join(" & ");

  return (
    <form className='income-form-7qw2' onSubmit={handleSubmit}>
      <div className='income-form-7qw2__field'>
        <label className='income-form-7qw2__label' htmlFor='income-label'>
          Label
        </label>
        <Input
          onChange={(e) => setLabel(e.target.value)}
          className='income-form-7qw2__input'
          placeholder='e.g. Paycheck'
          autoComplete='off'
          id='income-label'
          value={label}
          type='text'
          required
        />
      </div>

      <div className='income-form-7qw2__field'>
        <label className='income-form-7qw2__label' htmlFor='income-amount'>
          Amount
        </label>
        <Input
          onChange={(e) => setAmount(e.target.value)}
          className='income-form-7qw2__input'
          placeholder='0.00'
          id='income-amount'
          inputMode='decimal'
          value={amount}
          type='number'
          step='0.01'
          min='0'
          required
        />
      </div>

      <div className='income-form-7qw2__field'>
        <label className='income-form-7qw2__label' htmlFor='income-kind'>
          Type
        </label>
        <select
          onChange={(e) => setKind(e.target.value)}
          className='income-form-7qw2__select'
          id='income-kind'
          value={kind}
        >
          {KIND_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {kind === "one_time" && (
        <div className='income-form-7qw2__field'>
          <p className='income-form-7qw2__label'>Pick the date of the payment</p>
          <PaydayCalendar onDaySelect={handleOneTimeSelect} selectedDate={oneTimeDate} />
        </div>
      )}

      {kind === "fixed" && (
        <div className='income-form-7qw2__field'>
          <label className='income-form-7qw2__label' htmlFor='income-recurring'>
            Frequency
          </label>
          <select
            onChange={(e) => handleRecurringChange(e.target.value)}
            className='income-form-7qw2__select'
            id='income-recurring'
            value={recurring}
          >
            {FREQUENCY_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
      )}

      {kind === "fixed" && recurring === "weekly" && (
        <div className='income-form-7qw2__field'>
          <label className='income-form-7qw2__label' htmlFor='income-day-of-week'>
            Day of the week
          </label>
          <select
            onChange={(e) => setDayOfWeek(Number(e.target.value))}
            className='income-form-7qw2__select'
            id='income-day-of-week'
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

      {kind === "fixed" && recurring === "biweekly" && (
        <div className='income-form-7qw2__field'>
          <div className='income-form-7qw2__toggle'>
            <button
              className={biweeklyMode === "every_other_week" ? "active" : ""}
              onClick={() => setBiweeklyMode("every_other_week")}
              type='button'
            >
              Every other week
            </button>
            <button
              className={biweeklyMode === "days_of_month" ? "active" : ""}
              onClick={() => setBiweeklyMode("days_of_month")}
              type='button'
            >
              Two month days
            </button>
          </div>

          {biweeklyMode === "every_other_week" ? (
            <>
              <p className='income-form-7qw2__label'>
                Pick one payday — it repeats every other week on the same weekday
              </p>
              <PaydayCalendar onDaySelect={handleAnchorSelect} selectedDate={anchorDate} />
            </>
          ) : (
            <>
              <p className='income-form-7qw2__label'>
                Pick the 2 days of the month you get paid
                {pickedSummary && (
                  <span className='income-form-7qw2__picks'> — {pickedSummary}</span>
                )}
              </p>
              <PaydayCalendar
                onDaySelect={handleDayOfMonthSelect}
                selectedDays={selectedDays}
              />
            </>
          )}
        </div>
      )}

      {kind === "fixed" &&
        ["monthly", "quarterly", "semiannually"].includes(recurring) && (
          <div className='income-form-7qw2__field'>
            <p className='income-form-7qw2__label'>
              Pick the day of the month
              {pickedSummary && (
                <span className='income-form-7qw2__picks'> — {pickedSummary}</span>
              )}
            </p>
            <PaydayCalendar
              onDaySelect={handleDayOfMonthSelect}
              selectedDays={selectedDays}
            />
            <p className='income-form-7qw2__hint'>
              Months shorter than the picked day pay on their last day instead.
            </p>
          </div>
        )}

      {kind === "fixed" && recurring === "annually" && (
        <div className='income-form-7qw2__field'>
          <p className='income-form-7qw2__label'>Pick the month and day</p>
          <PaydayCalendar
            onDaySelect={handleDayOfMonthSelect}
            selectedDays={selectedDays}
            selectedDate={
              selectedDateMonth != null && selectedDays.length === 1
                ? { month: selectedDateMonth, day: selectedDays[0] }
                : null
            }
          />
        </div>
      )}

      <div className='income-form-7qw2__actions'>
        <Button primary type='submit' isLoading={isLoading} className='w-100'>
          {editing ? "Save changes" : "Add income"}
        </Button>
        {onCancel && (
          <Button secondary type='button' onClick={onCancel}>
            Cancel
          </Button>
        )}
      </div>
    </form>
  );
};
