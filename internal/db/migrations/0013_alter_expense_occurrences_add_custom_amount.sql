-- Hand-edited occurrence amounts: amount_custom = 1 marks rows whose amount was set
-- individually (pencil edit). Expense-level amount changes cascade only onto unpaid,
-- non-custom rows, so a hand edit is never silently overwritten — and paid rows
-- (history) are never rewritten either.

ALTER TABLE expense_occurrences
    ADD COLUMN amount_custom TINYINT(1) NOT NULL DEFAULT 0 AFTER amount;
