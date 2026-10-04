-- Free-text notes on expenses (e.g. payment method / recipient info from the bills import).
-- Read-only for now: not settable through the API, preserved on update.

ALTER TABLE expenses
    ADD COLUMN notes TEXT DEFAULT NULL AFTER recur_rule;
