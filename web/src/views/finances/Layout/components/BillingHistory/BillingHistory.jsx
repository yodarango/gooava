import { API_GET_EXPENSE_OCCURRENCES_HISTORY } from "@constants";
import { useEffect, useRef, useState } from "react";

// styles
import "./BillingHistory.css";

const formatUSD = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

// "2026-09-28" -> "Mon, Sep 28" (parsed as local time so the day never shifts)
const formatDay = (key) =>
  new Date(`${key}T00:00:00`).toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
  });

// "2026-09-28" -> "September 2026"
const formatMonth = (key) =>
  new Date(`${key}T00:00:00`).toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
  });

const monthKeyOf = (date) =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;

/*********************************************************************************************************
 * "Historical billing" tab: every bill ever marked paid, newest due date first, loaded one month
 * at a time as the user scrolls — a sentinel at the bottom of the list triggers the next month's
 * fetch. Months without a single paid bill are skipped, so every fetch that returns adds rows.
 * ******************************************************************************************************
 */
export const BillingHistory = () => {
  const [items, setItems] = useState([]);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);

  // cursor: the month to fetch next (first day). Refs guard against overlapping fetches
  // and the observer firing while a request is in flight.
  const cursorRef = useRef(new Date(new Date().getFullYear(), new Date().getMonth(), 1));
  const loadingRef = useRef(false);
  const hasMoreRef = useRef(true);
  const startedRef = useRef(false);
  const sentinelRef = useRef(null);

  const loadMore = async () => {
    if (loadingRef.current || !hasMoreRef.current) return;
    loadingRef.current = true;
    setLoading(true);

    try {
      // walk back month by month until a month with rows comes back (or history runs out)
      let keepGoing = true;
      while (keepGoing) {
        const cursor = cursorRef.current;
        const response = await fetch(
          `${API_GET_EXPENSE_OCCURRENCES_HISTORY}?month=${monthKeyOf(cursor)}`,
          { headers: { Authorization: "Bearer " + localStorage.getItem("auth") } }
        );
        const result = await response.json();
        if (!result.success) {
          setFailed(true);
          break;
        }

        const page = result.data?.items || [];
        setItems((prev) => [...prev, ...page]);
        hasMoreRef.current = !!result.data?.has_more;
        setHasMore(hasMoreRef.current);
        cursorRef.current = new Date(cursor.getFullYear(), cursor.getMonth() - 1, 1);
        keepGoing = page.length === 0 && hasMoreRef.current;
      }
    } catch {
      setFailed(true);
    } finally {
      loadingRef.current = false;
      setLoading(false);
    }
  };

  // first page on mount (guarded for StrictMode's double effect)
  useEffect(() => {
    if (startedRef.current) return;
    startedRef.current = true;
    loadMore();
  }, []);

  // infinite scroll: the next month loads whenever the sentinel enters the viewport
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) loadMore();
    });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, []);

  if (items.length === 0 && !loading && !hasMore) {
    return (
      <p className='billing-history-2hf8__empty'>
        No bills paid yet — they&apos;ll show up here once you mark them paid.
      </p>
    );
  }

  return (
    <div className='billing-history-2hf8'>
      <ul className='billing-history-2hf8__list'>
        {items.map((occurrence, index) => {
          // month group header whenever the month changes between consecutive rows
          const showMonth =
            index === 0 || formatMonth(items[index - 1].due_on) !== formatMonth(occurrence.due_on);

          return (
            <li key={occurrence.id}>
              {showMonth && (
                <p className='billing-history-2hf8__month'>{formatMonth(occurrence.due_on)}</p>
              )}
              <div className='billing-history-2hf8__row'>
                <div className='billing-history-2hf8__details'>
                  <span className='billing-history-2hf8__name'>{occurrence.expense_label}</span>
                  <span className='billing-history-2hf8__meta'>
                    due {formatDay(occurrence.due_on)}
                    {occurrence.paid_at &&
                      ` · paid ${formatDay(occurrence.paid_at.split(" ")[0])}`}
                    {occurrence.paid_source === "auto" && " · auto ✓"}
                  </span>
                </div>

                <span className='billing-history-2hf8__amount'>
                  -{formatUSD(occurrence.amount)}
                </span>
              </div>
            </li>
          );
        })}
      </ul>

      {loading && <p className='billing-history-2hf8__status'>Loading…</p>}

      {failed && !loading && (
        <button
          className='billing-history-2hf8__retry'
          onClick={() => {
            setFailed(false);
            loadMore();
          }}
          type='button'
        >
          Couldn&apos;t load — tap to retry
        </button>
      )}

      {!hasMore && items.length > 0 && (
        <p className='billing-history-2hf8__status'>That&apos;s everything ✓</p>
      )}

      {hasMore && !failed && (
        <div ref={sentinelRef} className='billing-history-2hf8__sentinel' />
      )}
    </div>
  );
};
