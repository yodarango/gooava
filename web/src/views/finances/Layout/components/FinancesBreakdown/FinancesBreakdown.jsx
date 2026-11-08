import { API_GET_SPENDING_BREAKDOWN } from "@constants";
import { ChartColumnIncreasing, ChevronDown, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { Loading } from "@ds";
import { createPortal } from "react-dom";
import { useEffect, useMemo, useRef, useState } from "react";
import { useGet } from "@utils";

// styles
import "./FinancesBreakdown.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value || 0);

/*********************************************************************************************************
 * The drawer body — fetches the 36-month breakdown and category totals on open.
 * Shows horizontal-scroll stacked bars and bottom category totals with income percentages.
 * Clicking bars cumulatively filters the breakdown for those months.
 * Categories are sorted biggest to smallest, collapsible, and closed by default.
 * ******************************************************************************************************
 */
const BreakdownDrawer = (props) => {
  const { onClose } = props;

  // close on Escape + lock body scroll while open
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
    };
  }, [onClose]);

  const breakdown = useGet({ url: API_GET_SPENDING_BREAKDOWN });
  const scrollRef = useRef(null);

  // multiple cumulative selected months; empty set = all-time
  const [selectedMonths, setSelectedMonths] = useState(() => new Set());
  // collapsible category keys; empty set = all closed by default
  const [expandedCategories, setExpandedCategories] = useState(
    () => new Set(),
  );

  const data = breakdown.data;
  const months = useMemo(
    () => (Array.isArray(data?.months) ? data.months : []),
    [data?.months],
  );
  const totalsByCategory = useMemo(
    () =>
      Array.isArray(data?.totals_by_category) ? data.totals_by_category : [],
    [data?.totals_by_category],
  );

  // find max monthly total across the 36 months for relative bar heights
  const maxMonthTotal = useMemo(() => {
    if (!months.length) return 1;
    const max = Math.max(...months.map((m) => m.total || 0));
    return max > 0 ? max : 1;
  }, [months]);

  // auto-scroll to the rightmost (most recent) month once data lands
  useEffect(() => {
    if (scrollRef.current && months.length > 0) {
      scrollRef.current.scrollLeft = scrollRef.current.scrollWidth;
    }
  }, [months.length]);

  const toggleMonth = (monthKey) => {
    setSelectedMonths((prev) => {
      const next = new Set(prev);
      if (next.has(monthKey)) {
        next.delete(monthKey);
      } else {
        next.add(monthKey);
      }
      return next;
    });
  };

  const toggleCategory = (categoryKey) => {
    setExpandedCategories((prev) => {
      const next = new Set(prev);
      if (next.has(categoryKey)) {
        next.delete(categoryKey);
      } else {
        next.add(categoryKey);
      }
      return next;
    });
  };

  // replace bottom breakdown with data for the selected month/s (cumulative)
  // or show all-time data when no months are selected
  const activeBreakdown = useMemo(() => {
    if (!data) {
      return { categories: [], totalSpent: 0, income: 0, isFiltered: false };
    }

    const isFiltered = selectedMonths.size > 0;
    if (!isFiltered) {
      // all-time: sort categories from biggest to smallest
      const sorted = [...totalsByCategory].sort(
        (a, b) => b.total_spent - a.total_spent,
      );
      return {
        categories: sorted,
        totalSpent: data.total_spent_to_date || 0,
        income: data.total_income_to_date || 0,
        isFiltered: false,
      };
    }

    // cumulative aggregation across the selected months
    const filteredMonths = months.filter((m) => selectedMonths.has(m.month));
    let selectedSpent = 0;
    let selectedIncome = 0;
    const catMap = new Map();

    for (const m of filteredMonths) {
      selectedSpent += m.total || 0;
      selectedIncome += m.income || 0;
      for (const item of m.breakdown || []) {
        if (!catMap.has(item.category)) {
          catMap.set(item.category, {
            category: item.category,
            color: item.color,
            total_spent: 0,
            items: [],
          });
        }
        const entry = catMap.get(item.category);
        entry.total_spent += item.amount;
        if (Array.isArray(item.items)) {
          entry.items.push(...item.items);
        }
      }
    }

    const incomeBasis =
      selectedIncome > 0 ? selectedIncome : data.total_income_to_date || 0;

    const categories = Array.from(catMap.values())
      .map((c) => ({
        ...c,
        total_spent: Math.round(c.total_spent * 100) / 100,
        items: [...c.items].sort((a, b) => b.amount - a.amount),
        percentage_of_income:
          incomeBasis > 0
            ? Math.round((c.total_spent / incomeBasis) * 1000) / 10
            : null,
      }))
      // organize categories from biggest to smallest
      .sort((a, b) => b.total_spent - a.total_spent);

    return {
      categories,
      totalSpent: Math.round(selectedSpent * 100) / 100,
      income: incomeBasis,
      isFiltered: true,
      monthCount: selectedMonths.size,
    };
  }, [data, months, totalsByCategory, selectedMonths]);

  return (
    <div
      className='finances-breakdown-5m1k'
      role='dialog'
      aria-modal='true'
      aria-label='Spending breakdown'
    >
      <div className='finances-breakdown-5m1k__backdrop' onClick={onClose} />

      <aside className='finances-breakdown-5m1k__drawer'>
        <div className='finances-breakdown-5m1k__header'>
          <h2 className='finances-breakdown-5m1k__title'>Spending breakdown</h2>
          <button
            className='finances-breakdown-5m1k__close'
            onClick={onClose}
            aria-label='Close'
            type='button'
          >
            <MorphIcon icon={X} label='Close' size={20} />
          </button>
        </div>

        <div className='finances-breakdown-5m1k__body'>
          {breakdown.loading ? (
            <div className='finances-breakdown-5m1k__loading'>
              <Loading size={40} />
            </div>
          ) : months.length === 0 ? (
            <p className='finances-breakdown-5m1k__empty'>
              No spending recorded yet.
            </p>
          ) : (
            <div className='finances-breakdown-5m1k__content'>
              {/* 36-month horizontally scrollable stacked bar chart */}
              <div className='finances-breakdown-5m1k__chart-card'>
                <div
                  className='finances-breakdown-5m1k__scroll-track'
                  ref={scrollRef}
                >
                  <div className='finances-breakdown-5m1k__bars-row'>
                    {months.map((month) => {
                      const isSelected = selectedMonths.has(month.month);
                      const isFiltered = selectedMonths.size > 0;
                      const hasSpending = month.total > 0;
                      const barHeightPercent = hasSpending
                        ? Math.max((month.total / maxMonthTotal) * 100, 4)
                        : 0;

                      return (
                        <div
                          key={month.month}
                          className={`finances-breakdown-5m1k__bar-col ${isSelected ? "selected" : ""} ${isFiltered && !isSelected ? "dimmed" : ""}`}
                          onClick={() => toggleMonth(month.month)}
                          role='button'
                          tabIndex={0}
                          aria-label={`${month.label}: ${formatUSD(month.total)} (click to toggle)`}
                          title={`${month.label}: ${formatUSD(month.total)} (click to toggle)`}
                        >
                          <div className='finances-breakdown-5m1k__bar-slot'>
                            {hasSpending ? (
                              <div
                                className={`finances-breakdown-5m1k__stacked-bar ${isSelected ? "selected" : ""}`}
                                style={{ height: `${barHeightPercent}%` }}
                              >
                                {month.breakdown.map((item) => {
                                  const segPercent =
                                    month.total > 0
                                      ? (item.amount / month.total) * 100
                                      : 0;
                                  return (
                                    <div
                                      key={item.category}
                                      className='finances-breakdown-5m1k__segment'
                                      style={{
                                        height: `${segPercent}%`,
                                        backgroundColor: item.color,
                                      }}
                                      title={`${item.category}: ${formatUSD(item.amount)}`}
                                    />
                                  );
                                })}
                              </div>
                            ) : (
                              <div className='finances-breakdown-5m1k__zero-tick' />
                            )}
                          </div>
                          <span
                            className={`finances-breakdown-5m1k__month-label ${isSelected ? "selected" : ""}`}
                          >
                            {month.label}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>

              {/* bottom category totals with cumulative selected months support */}
              <div className='finances-breakdown-5m1k__summary'>
                <div className='finances-breakdown-5m1k__summary-header'>
                  <div className='finances-breakdown-5m1k__summary-titles'>
                    <h3 className='finances-breakdown-5m1k__summary-title'>
                      Total spent by category
                    </h3>
                    <span className='finances-breakdown-5m1k__summary-subtitle'>
                      {activeBreakdown.isFiltered ? (
                        <>
                          {activeBreakdown.monthCount} month
                          {activeBreakdown.monthCount > 1 ? "s" : ""} selected ·{" "}
                          <button
                            className='finances-breakdown-5m1k__clear-btn'
                            onClick={() => setSelectedMonths(new Set())}
                            type='button'
                          >
                            Show all time
                          </button>
                        </>
                      ) : (
                        "All time · Tap month bars above to filter"
                      )}
                    </span>
                  </div>

                  <div className='finances-breakdown-5m1k__summary-totals'>
                    <span className='finances-breakdown-5m1k__summary-spent'>
                      {formatUSD(activeBreakdown.totalSpent)}
                    </span>
                    {activeBreakdown.income > 0 && (
                      <span className='finances-breakdown-5m1k__summary-income'>
                        Income: {formatUSD(activeBreakdown.income)}
                      </span>
                    )}
                  </div>
                </div>

                {activeBreakdown.categories.length === 0 ? (
                  <p className='finances-breakdown-5m1k__category-empty'>
                    No expenses recorded for the selected period.
                  </p>
                ) : (
                  <ul className='finances-breakdown-5m1k__category-list'>
                    {activeBreakdown.categories.map((cat) => {
                      const isExpanded = expandedCategories.has(cat.category);
                      const items = Array.isArray(cat.items) ? cat.items : [];

                      return (
                        <li
                          key={cat.category}
                          className={`finances-breakdown-5m1k__category-card ${isExpanded ? "expanded" : ""}`}
                        >
                          <button
                            className='finances-breakdown-5m1k__category-header'
                            onClick={() => toggleCategory(cat.category)}
                            aria-expanded={isExpanded}
                            type='button'
                          >
                            <div className='finances-breakdown-5m1k__category-left'>
                              <span
                                className='finances-breakdown-5m1k__cat-dot'
                                style={{ backgroundColor: cat.color }}
                              />
                              <span className='finances-breakdown-5m1k__cat-name'>
                                {cat.category}
                              </span>
                              <span className='finances-breakdown-5m1k__cat-count'>
                                ({items.length})
                              </span>
                            </div>
                            <div className='finances-breakdown-5m1k__category-right'>
                              <span className='finances-breakdown-5m1k__cat-spent'>
                                {formatUSD(cat.total_spent)}
                              </span>
                              <span className='finances-breakdown-5m1k__cat-pct'>
                                (
                                {cat.percentage_of_income !== null &&
                                cat.percentage_of_income !== undefined
                                  ? `${cat.percentage_of_income.toFixed(1)}%`
                                  : "0.0%"}
                                )
                              </span>
                              <span
                                className={`finances-breakdown-5m1k__chevron ${isExpanded ? "open" : ""}`}
                              >
                                <MorphIcon
                                  icon={ChevronDown}
                                  label='Toggle details'
                                  size={16}
                                />
                              </span>
                            </div>
                          </button>

                          {isExpanded && (
                            <div className='finances-breakdown-5m1k__category-items'>
                              {items.length === 0 ? (
                                <p className='finances-breakdown-5m1k__item-empty'>
                                  No expense details found
                                </p>
                              ) : (
                                <ul className='finances-breakdown-5m1k__item-list'>
                                  {items.map((item, idx) => (
                                    <li
                                      key={`${item.label}-${item.date}-${idx}`}
                                      className='finances-breakdown-5m1k__item-row'
                                    >
                                      <div className='finances-breakdown-5m1k__item-left'>
                                        <span className='finances-breakdown-5m1k__item-label'>
                                          {item.label}
                                        </span>
                                        {item.date && (
                                          <span className='finances-breakdown-5m1k__item-date'>
                                            {item.date}
                                          </span>
                                        )}
                                      </div>
                                      <span className='finances-breakdown-5m1k__item-amount'>
                                        {formatUSD(item.amount)}
                                      </span>
                                    </li>
                                  ))}
                                </ul>
                              )}
                            </div>
                          )}
                        </li>
                      );
                    })}
                  </ul>
                )}
              </div>
            </div>
          )}
        </div>
      </aside>
    </div>
  );
};

/*********************************************************************************************************
 * Graph icon button shown next to the settings gear. Opens the spending breakdown drawer.
 * ******************************************************************************************************
 */
export const FinancesBreakdown = () => {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        className='finances-breakdown-5m1k__trigger'
        onClick={() => setOpen(true)}
        aria-label='Spending breakdown'
        title='Spending breakdown'
        type='button'
      >
        <MorphIcon
          icon={ChartColumnIncreasing}
          label='Spending breakdown'
          size={22}
        />
      </button>

      {open &&
        createPortal(
          <BreakdownDrawer onClose={() => setOpen(false)} />,
          document.body,
        )}
    </>
  );
};
