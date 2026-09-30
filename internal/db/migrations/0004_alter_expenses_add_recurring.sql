-- Recurring expense schedule: frequency + a JSON blob with the recurrence details
-- (e.g. {"day_of_week":5}, {"days_of_month":[1,15]}, {"day":15,"months":[1,4,7,10]})

ALTER TABLE expenses
    ADD COLUMN recurring VARCHAR(16) DEFAULT NULL AFTER label,
    ADD COLUMN recur_rule JSON DEFAULT NULL AFTER recurring;
