import { useAppContext } from "../../../../context/appContextProvider";
import { ExpensesForm } from "./ExpensesForm";
import {
  API_POST_EXPENSES_DELETE,
  DAYS_OF_WEEK,
  API_GET_EXPENSES,
  MONTH_NAMES,
} from "@constants";
import { Pencil, Plus, Trash2, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useState } from "react";
import { useGet, usePost } from "@utils";

// styles
import "./ExpensesPane.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// e.g. "weekly · Fridays", "quarterly · day 15 of Jan, Apr, Jul, Oct"
const recurringSummary = (expense) => {
  const rule = expense.recur_rule || {};

  switch (expense.recurring) {
    case "daily":
      return "daily";
    case "weekly":
      return `weekly · ${DAYS_OF_WEEK[rule.day_of_week ?? 0]}s`;
    case "biweekly":
      return `biweekly · days ${(rule.days_of_month || []).join(" & ")}`;
    case "monthly":
      return `monthly · day ${rule.day_of_month}`;
    case "quarterly":
      return `quarterly · day ${rule.day} of ${(rule.months || [])
        .map((month) => MONTH_NAMES[month - 1])
        .join(", ")}`;
    case "yearly":
      return `yearly · ${MONTH_NAMES[(rule.month || 1) - 1]} ${rule.day}`;
    default:
      return "one-time";
  }
};

/*********************************************************************************************************
 * The Expenses tab: a title row ("Expenses" + plus icon that opens the add form) above the list of
 * saved expenses. Each row has edit (opens the form prefilled) and delete (with confirm) actions.
 * ******************************************************************************************************
 */
export const ExpensesPane = () => {
  const { showToast } = useAppContext();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState(null); // expense being edited, null = add mode
  const [refreshKey, setRefreshKey] = useState(0);

  const expenses = useGet({ url: API_GET_EXPENSES, dependencies: [refreshKey] });

  const deleteExpense = usePost({
    url: API_POST_EXPENSES_DELETE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Expense deleted", type: "success" });
      setRefreshKey((key) => key + 1);
    },
  });

  // surface delete errors as toasts
  useEffect(() => {
    if (deleteExpense.error) {
      showToast({ message: String(deleteExpense.error), type: "danger" });
    }
  }, [deleteExpense.error]);

  // plus opens the add form (or closes it if already open in add mode)
  const handlePlus = () => {
    if (formOpen && !editing) {
      setFormOpen(false);
      return;
    }
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (expense) => {
    setEditing(expense);
    setFormOpen(true);
  };

  const handleDelete = (expense) => {
    if (!window.confirm(`Delete "${expense.label}"?`)) return;
    deleteExpense.post({ id: expense.id });
  };

  const handleFormDone = () => {
    setFormOpen(false);
    setEditing(null);
    setRefreshKey((key) => key + 1);
  };

  const expenseList = Array.isArray(expenses.data) ? expenses.data : [];

  return (
    <div className='expenses-pane-4rt8'>
      <div className='expenses-pane-4rt8__header'>
        <h3 className='expenses-pane-4rt8__title'>Expenses</h3>
        <button
          className='expenses-pane-4rt8__add'
          onClick={handlePlus}
          aria-label={formOpen && !editing ? "Close form" : "Add expense"}
          type='button'
        >
          <MorphIcon
            icon={formOpen && !editing ? X : Plus}
            label='Add expense'
            size={20}
          />
        </button>
      </div>

      {formOpen && (
        <div className='expenses-pane-4rt8__form'>
          <ExpensesForm
            onCancel={handleFormDone}
            onDone={handleFormDone}
            expense={editing}
            key={editing ? editing.id : "new"}
          />
        </div>
      )}

      <div className='expenses-pane-4rt8__list'>
        {expenses.loading && expenseList.length === 0 ? (
          <p className='expenses-pane-4rt8__empty'>Loading…</p>
        ) : expenseList.length === 0 ? (
          <p className='expenses-pane-4rt8__empty'>No expenses yet — tap + to add one.</p>
        ) : (
          <ul>
            {expenseList.map((expense) => (
              <li key={expense.id} className='expenses-pane-4rt8__row'>
                <div className='expenses-pane-4rt8__details'>
                  <span className='expenses-pane-4rt8__label'>{expense.label}</span>
                  <span className='expenses-pane-4rt8__meta'>
                    {expense.category} · {recurringSummary(expense)}
                  </span>
                </div>

                <span className='expenses-pane-4rt8__amount'>
                  -{formatUSD(expense.amount)}
                </span>

                <div className='expenses-pane-4rt8__actions'>
                  <button
                    className='expenses-pane-4rt8__action edit'
                    onClick={() => handleEdit(expense)}
                    aria-label={`Edit ${expense.label}`}
                    type='button'
                  >
                    <MorphIcon icon={Pencil} label='Edit' size={16} />
                  </button>
                  <button
                    className='expenses-pane-4rt8__action delete'
                    onClick={() => handleDelete(expense)}
                    aria-label={`Delete ${expense.label}`}
                    type='button'
                  >
                    <MorphIcon icon={Trash2} label='Delete' size={16} />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};
