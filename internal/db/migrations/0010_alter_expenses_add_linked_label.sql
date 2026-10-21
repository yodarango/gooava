-- The merchant/name string of the Plaid transaction the user manually linked to the expense
-- (via the link button in the expenses tab or the past-due card). Once set, occurrences are
-- auto-marked paid when a synced transaction matches this label — unlinked expenses whose
-- label merely string-matches a transaction go through the past-due review card instead.

ALTER TABLE expenses
    ADD COLUMN linked_label VARCHAR(512) DEFAULT NULL AFTER label;
