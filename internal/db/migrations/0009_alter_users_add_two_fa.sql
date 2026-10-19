-- Two-factor authentication: when enabled, logging in requires a 6-digit code
-- emailed to the user. The code is stored in the existing verification_code
-- column (same type as the signup verification code), so no extra column is
-- needed for it — only for the on/off flag.

ALTER TABLE users
    ADD COLUMN two_fa_enabled TINYINT(1) NOT NULL DEFAULT 0 AFTER verification_code;
