import { useAppContext } from "../../../../context/appContextProvider";
import { IncomeForm } from "./IncomeForm";
import {
  API_POST_INCOMES_DELETE,
  API_GET_INCOMES,
  DAYS_OF_WEEK,
  MONTH_NAMES,
} from "@constants";
import { Pencil, Plus, Trash2, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useState } from "react";
import { useGet, usePost } from "@utils";

// styles
import "./IncomePane.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// e.g. "weekly · Fridays", "biweekly · every other week", "one-time · Oct 20"
const scheduleSummary = (income) => {
  const rule = income.recur_rule || {};

  if (!income.recurring) {
    if (!rule.date) return "one-time";
    const date = new Date(rule.date + "T00:00:00");
    return `one-time · ${date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}`;
  }

  switch (income.recurring) {
    case "weekly":
      return `weekly · ${DAYS_OF_WEEK[rule.day_of_week ?? 0]}s`;
    case "biweekly":
      return rule.mode === "days_of_month"
        ? `biweekly · days ${(rule.days_of_month || []).join(" & ")}`
        : "biweekly · every other week";
    case "monthly":
      return `monthly · day ${rule.day_of_month}`;
    case "quarterly":
      return `quarterly · day ${rule.day_of_month}`;
    case "semiannually":
      return `semiannually · day ${rule.day_of_month}`;
    case "annually":
      return `annually · ${MONTH_NAMES[(rule.month || 1) - 1]} ${rule.day_of_month}`;
    default:
      return income.recurring;
  }
};

/*********************************************************************************************************
 * The Income tab: a title row ("Income" + plus icon that opens the add form) above the list of
 * saved income sources. Each row has edit (opens the form prefilled) and delete (with confirm).
 * ******************************************************************************************************
 */
export const IncomePane = () => {
  const { showToast } = useAppContext();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState(null); // income being edited, null = add mode
  const [refreshKey, setRefreshKey] = useState(0);

  const incomes = useGet({ url: API_GET_INCOMES, dependencies: [refreshKey] });

  const deleteIncome = usePost({
    url: API_POST_INCOMES_DELETE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Income deleted", type: "success" });
      setRefreshKey((key) => key + 1);
    },
  });

  // surface delete errors as toasts
  useEffect(() => {
    if (deleteIncome.error) {
      showToast({ message: String(deleteIncome.error), type: "danger" });
    }
  }, [deleteIncome.error]);

  // plus opens the add form (or closes it if already open in add mode)
  const handlePlus = () => {
    if (formOpen && !editing) {
      setFormOpen(false);
      return;
    }
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (income) => {
    setEditing(income);
    setFormOpen(true);
  };

  const handleDelete = (income) => {
    if (!window.confirm(`Delete "${income.label}"?`)) return;
    deleteIncome.post({ id: income.id });
  };

  const handleFormDone = () => {
    setFormOpen(false);
    setEditing(null);
    setRefreshKey((key) => key + 1);
  };

  const incomeList = Array.isArray(incomes.data) ? incomes.data : [];

  return (
    <div className='income-pane-6yt3'>
      <div className='income-pane-6yt3__header'>
        <h3 className='income-pane-6yt3__title'>Income</h3>
        <button
          className='income-pane-6yt3__add'
          onClick={handlePlus}
          aria-label={formOpen && !editing ? "Close form" : "Add income"}
          type='button'
        >
          <MorphIcon
            icon={formOpen && !editing ? X : Plus}
            label='Add income'
            size={20}
          />
        </button>
      </div>

      {formOpen && (
        <div className='income-pane-6yt3__form'>
          <IncomeForm
            onCancel={handleFormDone}
            onDone={handleFormDone}
            income={editing}
            key={editing ? editing.id : "new"}
          />
        </div>
      )}

      <div className='income-pane-6yt3__list'>
        {incomes.loading && incomeList.length === 0 ? (
          <p className='income-pane-6yt3__empty'>Loading…</p>
        ) : incomeList.length === 0 ? (
          <p className='income-pane-6yt3__empty'>No income yet — tap + to add one.</p>
        ) : (
          <ul>
            {incomeList.map((income) => (
              <li key={income.id} className='income-pane-6yt3__row'>
                <div className='income-pane-6yt3__details'>
                  <span className='income-pane-6yt3__label'>{income.label}</span>
                  <span className='income-pane-6yt3__meta'>
                    {scheduleSummary(income)}
                  </span>
                </div>

                <span className='income-pane-6yt3__amount'>
                  +{formatUSD(income.amount)}
                </span>

                <div className='income-pane-6yt3__actions'>
                  <button
                    className='income-pane-6yt3__action edit'
                    onClick={() => handleEdit(income)}
                    aria-label={`Edit ${income.label}`}
                    type='button'
                  >
                    <MorphIcon icon={Pencil} label='Edit' size={16} />
                  </button>
                  <button
                    className='income-pane-6yt3__action delete'
                    onClick={() => handleDelete(income)}
                    aria-label={`Delete ${income.label}`}
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
