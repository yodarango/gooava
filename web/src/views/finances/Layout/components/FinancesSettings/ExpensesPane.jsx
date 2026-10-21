import { useAppContext } from "../../../../context/appContextProvider";
import { TransactionPicker } from "../TransactionPicker/TransactionPicker";
import { ExpensesForm } from "./ExpensesForm";
import {
  API_GET_EXPENSE_CATEGORIES,
  API_POST_EXPENSES_DELETE,
  API_POST_EXPENSES_LINK,
  DAYS_OF_WEEK,
  API_GET_EXPENSES,
  MONTH_NAMES,
} from "@constants";
import { Link2, Pencil, Plus, Trash2, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useMemo, useState } from "react";
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
 * saved expenses. Each row has link (attach the transaction that pays it — future matching bills
 * then auto-mark paid), edit (opens the form prefilled, inline in the expense's card) and delete
 * (with confirm) actions.
 * ******************************************************************************************************
 */
export const ExpensesPane = (props) => {
  const { onChanged } = props;
  const { showToast } = useAppContext();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState(null); // expense being edited, null = add mode
  const [linkingFor, setLinkingFor] = useState(null); // expense being linked to a transaction
  const [refreshKey, setRefreshKey] = useState(0);

  const expenses = useGet({ url: API_GET_EXPENSES, dependencies: [refreshKey] });

  // category label -> color, for the dot next to each expense title
  const categories = useGet({
    url: API_GET_EXPENSE_CATEGORIES,
    dependencies: [refreshKey],
  });

  const colorByLabel = useMemo(() => {
    const map = {};
    const list = Array.isArray(categories.data) ? categories.data : [];
    for (const category of list) map[category.label] = category.color;
    return map;
  }, [categories.data]);

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

  const linkExpense = usePost({
    url: API_POST_EXPENSES_LINK,
    callback: (data) => {
      if (!data) return;
      showToast({
        message: data.linked_label
          ? `Linked to "${data.linked_label}" — future matching bills will be marked paid automatically`
          : "Link removed",
        type: "success",
      });
      setLinkingFor(null);
      setRefreshKey((key) => key + 1);
      if (onChanged) onChanged();
    },
  });

  // surface link errors as toasts
  useEffect(() => {
    if (linkExpense.error) {
      showToast({ message: String(linkExpense.error), type: "danger" });
    }
  }, [linkExpense.error]);

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

      {/* add form stays at the top; the edit form opens inside the edited expense's card */}
      {formOpen && !editing && (
        <div className='expenses-pane-4rt8__form'>
          <ExpensesForm
            onCancel={handleFormDone}
            onDone={handleFormDone}
            expense={null}
            key='new'
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
            {expenseList.map((expense) => {
              const isEditing = formOpen && editing?.id === expense.id;

              return (
                <li
                  key={expense.id}
                  className={`expenses-pane-4rt8__row ${isEditing ? "editing" : ""}`}
                >
                  {isEditing ? (
                    <ExpensesForm
                      onCancel={handleFormDone}
                      onDone={handleFormDone}
                      expense={editing}
                      key={editing.id}
                    />
                  ) : (
                    <>
                      <div className='expenses-pane-4rt8__details'>
                        <span className='expenses-pane-4rt8__label'>
                          <span
                            className='expenses-pane-4rt8__dot'
                            style={{
                              backgroundColor:
                                colorByLabel[expense.category] ||
                                "rgba(var(--dr-beta-rgb), 0.35)",
                            }}
                          />
                          <span className='expenses-pane-4rt8__label-text'>
                            {expense.label}
                          </span>
                        </span>
                        <span className='expenses-pane-4rt8__meta'>
                          {expense.category} · {recurringSummary(expense)}
                          {expense.linked_label ? ` · auto-pay: ${expense.linked_label}` : ""}
                          {expense.notes ? ` · ${expense.notes}` : ""}
                        </span>
                      </div>

                      <span className='expenses-pane-4rt8__amount'>
                        -{formatUSD(expense.amount)}
                      </span>

                      <div className='expenses-pane-4rt8__actions'>
                        <button
                          className={`expenses-pane-4rt8__action link ${expense.linked_label ? "linked" : ""}`}
                          onClick={() => setLinkingFor(expense)}
                          aria-label={`Link ${expense.label} to a transaction`}
                          title={
                            expense.linked_label
                              ? `Linked to ${expense.linked_label} — future matching bills are marked paid automatically`
                              : "Link to a transaction"
                          }
                          type='button'
                        >
                          <MorphIcon icon={Link2} label='Link transaction' size={16} />
                        </button>
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
                    </>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {linkingFor && (
        <TransactionPicker
          open
          title={`Link "${linkingFor.label}" to a transaction`}
          onClose={() => setLinkingFor(null)}
          onSelect={(txn) =>
            linkExpense.post({ id: linkingFor.id, transaction_id: txn.transaction_id })
          }
          onUnlink={
            linkingFor.linked_label
              ? () => linkExpense.post({ id: linkingFor.id, transaction_id: "" })
              : undefined
          }
        />
      )}
    </div>
  );
};
