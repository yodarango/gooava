import { useAppContext } from "../../../../context/appContextProvider";
import {
  API_GET_MILESTONES,
  API_POST_MILESTONES,
  API_POST_MILESTONES_UPDATE,
  API_POST_MILESTONES_DELETE,
} from "@constants";
import { Button, Calendar, Input } from "@ds";
import { Pencil, Plus, Trash2, X } from "lucide";
import { MorphIcon } from "morphicons/react";
import { useEffect, useMemo, useState } from "react";
import { useGet, usePost } from "@utils";

// styles
import "./MilestonesPane.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value || 0);

const formatDate = (dateStr) => {
  if (!dateStr) return "—";
  const [y, m, d] = String(dateStr).split("-").map(Number);
  return new Date(y, (m || 1) - 1, d || 1).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const parseDollarAmount = (val) => {
  if (typeof val !== "string") val = String(val ?? "");
  const cleaned = val.replace(/[$,]/g, "").trim();
  if (!cleaned) return null;
  const num = Number(cleaned);
  if (isNaN(num) || num <= 0) return null;
  return Math.round(num * 100) / 100;
};

/*********************************************************************************************************
 * MilestoneForm: Add or edit a savings goal. Computes the estimated per-paycheck contribution live.
 * ******************************************************************************************************
 */
const MilestoneForm = ({ milestone, onDone, onCancel }) => {
  const editing = !!milestone;
  const { showToast } = useAppContext();

  const [title, setTitle] = useState(milestone?.title || "");
  const [targetAmount, setTargetAmount] = useState(
    milestone?.target_amount ? String(milestone.target_amount) : "",
  );
  const [targetDate, setTargetDate] = useState(milestone?.target_date || "");

  const create = usePost({
    url: API_POST_MILESTONES,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Milestone goal created!", type: "success" });
      if (onDone) onDone();
    },
  });

  const update = usePost({
    url: API_POST_MILESTONES_UPDATE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Milestone updated!", type: "success" });
      if (onDone) onDone();
    },
  });

  const request = editing ? update : create;
  useEffect(() => {
    if (request.error) {
      showToast({ message: String(request.error), type: "danger" });
    }
  }, [request.error, showToast]);

  const selectedCalendarDate = useMemo(() => {
    if (!targetDate) return null;
    const [y, m, d] = targetDate.split("-").map(Number);
    return { year: y, month: m, day: d };
  }, [targetDate]);

  const handleCalendarSelect = ({ year, month, day }) => {
    const yStr = String(year);
    const mStr = String(month).padStart(2, "0");
    const dStr = String(day).padStart(2, "0");
    setTargetDate(`${yStr}-${mStr}-${dStr}`);
  };

  // live financial estimate of per-paycheck breakdown
  const estimate = useMemo(() => {
    const amount = parseDollarAmount(targetAmount);
    if (!amount || !targetDate) return null;

    const today = new Date();
    const [y, m, d] = targetDate.split("-").map(Number);
    const target = new Date(y, (m || 1) - 1, d || 1);
    const diffDays = Math.ceil((target - today) / (1000 * 60 * 60 * 24));
    if (diffDays <= 0) return null;

    // assume biweekly (every 14 days) standard pay period
    const approxPaychecks = Math.max(1, Math.round(diffDays / 14));
    const perPaycheck = Math.round((amount / approxPaychecks) * 100) / 100;
    return { approxPaychecks, perPaycheck };
  }, [targetAmount, targetDate]);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!title.trim()) {
      showToast({ message: "Please enter a title", type: "danger" });
      return;
    }
    const amount = parseDollarAmount(targetAmount);
    if (!amount) {
      showToast({
        message: "Please enter a valid dollar amount (e.g. 2500 or 2,500.00)",
        type: "danger",
      });
      return;
    }
    if (!targetDate) {
      showToast({
        message: "Please pick a target completion date on the calendar below",
        type: "danger",
      });
      return;
    }

    const payload = {
      title: title.trim(),
      target_amount: amount,
      target_date: targetDate,
    };
    if (editing) payload.id = milestone.id;

    request.post(payload);
  };

  return (
    <form className='milestones-pane-3k8v__form' onSubmit={handleSubmit}>
      <h4 className='milestones-pane-3k8v__form-title'>
        {editing ? `Edit “${milestone.title}”` : "New savings milestone"}
      </h4>

      <div className='milestones-pane-3k8v__field'>
        <label className='milestones-pane-3k8v__label'>Goal name</label>
        <Input
          placeholder='e.g. Emergency Fund, New Car, Vacation'
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>

      <div className='milestones-pane-3k8v__field'>
        <label className='milestones-pane-3k8v__label'>Target amount ($)</label>
        <Input
          type='text'
          placeholder='e.g. 5,000'
          value={targetAmount}
          onChange={(e) => setTargetAmount(e.target.value)}
          required
        />
      </div>

      <div className='milestones-pane-3k8v__field'>
        <p className='milestones-pane-3k8v__label'>
          Target completion date:{" "}
          <span className='milestones-pane-3k8v__date-highlight'>
            {targetDate ? formatDate(targetDate) : "Pick on calendar below"}
          </span>
        </p>
        <div className='milestones-pane-3k8v__calendar-box'>
          <Calendar
            onDaySelect={handleCalendarSelect}
            selectedDate={selectedCalendarDate}
          />
        </div>
      </div>

      {estimate && (
        <div className='milestones-pane-3k8v__estimate'>
          <span className='milestones-pane-3k8v__estimate-amount'>
            ~{formatUSD(estimate.perPaycheck)} per paycheck
          </span>
          <span className='milestones-pane-3k8v__estimate-desc'>
            across approx. {estimate.approxPaychecks} paychecks to reach your goal
          </span>
        </div>
      )}

      <div className='milestones-pane-3k8v__form-actions'>
        <Button secondary type='button' onClick={onCancel} disabled={request.loading}>
          Cancel
        </Button>
        <Button primary type='submit' isLoading={request.loading}>
          {editing ? "Save changes" : "Create milestone"}
        </Button>
      </div>
    </form>
  );
};

/*********************************************************************************************************
 * The Milestones tab: displays savings goals with progress bars, original vs expected dates,
 * per-paycheck breakdowns, and add/edit/delete actions.
 * ******************************************************************************************************
 */
export const MilestonesPane = (props) => {
  const { onChanged } = props;
  const { showToast } = useAppContext();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState(null); // milestone being edited, null = add mode
  const [refreshKey, setRefreshKey] = useState(0);

  const milestones = useGet({
    url: API_GET_MILESTONES,
    dependencies: [refreshKey],
  });

  const deleteMilestone = usePost({
    url: API_POST_MILESTONES_DELETE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Milestone deleted", type: "success" });
      setRefreshKey((key) => key + 1);
      if (onChanged) onChanged();
    },
  });

  useEffect(() => {
    if (deleteMilestone.error) {
      showToast({ message: String(deleteMilestone.error), type: "danger" });
    }
  }, [deleteMilestone.error, showToast]);

  const handlePlus = () => {
    if (formOpen && !editing) {
      setFormOpen(false);
      return;
    }
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (item) => {
    setEditing(item);
    setFormOpen(true);
  };

  const handleDelete = (item) => {
    if (!window.confirm(`Delete the “${item.title}” milestone and its scheduled payments?`)) {
      return;
    }
    deleteMilestone.post({ id: item.id });
  };

  const handleFormDone = () => {
    setFormOpen(false);
    setEditing(null);
    setRefreshKey((key) => key + 1);
    if (onChanged) onChanged();
  };

  const list = Array.isArray(milestones.data) ? milestones.data : [];

  return (
    <div className='milestones-pane-3k8v'>
      <div className='milestones-pane-3k8v__header'>
        <h3 className='milestones-pane-3k8v__title'>Milestones</h3>
        <button
          className='milestones-pane-3k8v__add'
          onClick={handlePlus}
          aria-label={formOpen ? "Close form" : "Add milestone"}
          type='button'
        >
          <MorphIcon
            icon={formOpen ? X : Plus}
            label='Add milestone'
            size={20}
          />
        </button>
      </div>

      {formOpen && !editing && (
        <div className='milestones-pane-3k8v__form-container'>
          <MilestoneForm
            milestone={null}
            onDone={handleFormDone}
            onCancel={handleFormDone}
          />
        </div>
      )}

      {milestones.loading ? (
        <p className='milestones-pane-3k8v__empty'>Loading milestones…</p>
      ) : list.length === 0 ? (
        <div className='milestones-pane-3k8v__empty-state'>
          <p className='milestones-pane-3k8v__empty-title'>No milestones yet</p>
          <p className='milestones-pane-3k8v__empty-desc'>
            Add a savings goal (e.g. emergency fund, vacation, or down payment) to
            automatically break down contributions per paycheck and track your progress.
          </p>
        </div>
      ) : (
        <ul className='milestones-pane-3k8v__list'>
          {list.map((item) => {
            const isEditing = editing?.id === item.id;

            if (isEditing) {
              return (
                <li key={item.id} className='milestones-pane-3k8v__card editing'>
                  <MilestoneForm
                    milestone={editing}
                    onDone={handleFormDone}
                    onCancel={handleFormDone}
                  />
                </li>
              );
            }

            const isDelayed = item.paychecks_behind > 0;
            const isComplete = item.is_completed;

            return (
              <li key={item.id} className={`milestones-pane-3k8v__card ${isComplete ? "complete" : ""}`}>
                <div className='milestones-pane-3k8v__card-head'>
                  <div className='milestones-pane-3k8v__title-group'>
                    <h4 className='milestones-pane-3k8v__card-title'>{item.title}</h4>
                    <span
                      className={`milestones-pane-3k8v__badge ${
                        isComplete
                          ? "complete"
                          : isDelayed
                            ? "delayed"
                            : "ontrack"
                      }`}
                    >
                      {isComplete
                        ? "Completed 🎉"
                        : isDelayed
                          ? `Delayed by ${item.paychecks_behind} paycheck${item.paychecks_behind > 1 ? "s" : ""}`
                          : "On track ✓"}
                    </span>
                  </div>

                  <div className='milestones-pane-3k8v__actions'>
                    <button
                      className='milestones-pane-3k8v__action-btn'
                      onClick={() => handleEdit(item)}
                      aria-label={`Edit ${item.title}`}
                      type='button'
                    >
                      <MorphIcon icon={Pencil} label='Edit' size={15} />
                    </button>
                    <button
                      className='milestones-pane-3k8v__action-btn delete'
                      onClick={() => handleDelete(item)}
                      aria-label={`Delete ${item.title}`}
                      type='button'
                    >
                      <MorphIcon icon={Trash2} label='Delete' size={15} />
                    </button>
                  </div>
                </div>

                {/* Progress bar */}
                <div className='milestones-pane-3k8v__progress-section'>
                  <div className='milestones-pane-3k8v__progress-labels'>
                    <span className='milestones-pane-3k8v__saved'>
                      {formatUSD(item.total_saved)}{" "}
                      <span className='milestones-pane-3k8v__target'>
                        of {formatUSD(item.target_amount)}
                      </span>
                    </span>
                    <span className='milestones-pane-3k8v__percentage'>
                      {item.percentage.toFixed(1)}%
                    </span>
                  </div>
                  <div className='milestones-pane-3k8v__progress-track'>
                    <div
                      className={`milestones-pane-3k8v__progress-bar ${isComplete ? "complete" : ""}`}
                      style={{ width: `${Math.min(100, item.percentage)}%` }}
                    />
                  </div>
                </div>

                {/* Financial breakdown metrics */}
                <div className='milestones-pane-3k8v__metrics'>
                  <div className='milestones-pane-3k8v__metric'>
                    <span className='milestones-pane-3k8v__metric-label'>Per paycheck</span>
                    <span className='milestones-pane-3k8v__metric-value'>
                      {formatUSD(item.amount_per_paycheck)}
                    </span>
                  </div>

                  <div className='milestones-pane-3k8v__metric'>
                    <span className='milestones-pane-3k8v__metric-label'>Original date</span>
                    <span className='milestones-pane-3k8v__metric-value'>
                      {formatDate(item.target_date)}
                    </span>
                  </div>

                  <div className='milestones-pane-3k8v__metric'>
                    <span className='milestones-pane-3k8v__metric-label'>
                      {isComplete ? "Completed on" : "Expected date"}
                    </span>
                    <span
                      className={`milestones-pane-3k8v__metric-value ${
                        isDelayed && !isComplete ? "delayed-text" : ""
                      }`}
                    >
                      {formatDate(item.expected_completion_date)}
                    </span>
                  </div>

                  <div className='milestones-pane-3k8v__metric'>
                    <span className='milestones-pane-3k8v__metric-label'>Remaining</span>
                    <span className='milestones-pane-3k8v__metric-value'>
                      {formatUSD(item.remaining_amount)}
                    </span>
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
};

