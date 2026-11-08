/*********************************************************************************************************
 * Record search predicate shared by the transaction lists (bank activity + billing
 * history). A record matches when the query appears in its title or in its amount
 * ("25" matches "$25.00" and "$125.99"; "sub" matches "Subway").
 * ******************************************************************************************************
 */
export function matchesRecordQuery(title, amount, query) {
  const q = String(query || "").trim().toLowerCase();
  if (!q) return true;

  const formatted = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  })
    .format(Math.abs(amount || 0))
    .toLowerCase();

  return (
    String(title || "").toLowerCase().includes(q) ||
    formatted.includes(q) ||
    String(Math.abs(amount || 0)).includes(q)
  );
}
