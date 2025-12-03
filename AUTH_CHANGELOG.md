# Auth Changelog

Date: 2025-12-03

## Login issue – root cause and fix
- Symptom: `/login` intermittently failed / behaved inconsistently when using GORM, and worked only with a raw SQL query.
- Root cause: Mismatch in `internal/models/user.go` model mapping caused GORM lookups to not behave as expected. Specifically:
  - GORM `gorm.Model` soft-delete (`DeletedAt`) default scope was intended, but the model/DB mapping was not fully aligned, leading to query differences versus raw SQL.
  - Field/type mapping in `User` (including the custom `Role` type) required correct tags and Scanner/Valuer implementations to scan selected columns reliably.
- Fix:
  - Corrected `User` model mapping in `internal/models/user.go` so it matches the DB schema (tags and types), preserving GORM default scopes.
  - Switched login lookup to GORM with an explicit column list and normalized case matching: `LOWER(email)`.
  - Kept the 2FA check, metrics, and activity logging unchanged.

## Implementation notes
- GORM query used in `LoginUser`:
  - `Model(&models.User{}).Select("id", "account_id", "email", "password", "role", "is_active", "email_verified").Where("LOWER(email) = ?", normalizedEmail).Take(&user)`
  - Benefits: respects soft-deletes via default scope; fetches only required columns.
- Request normalization: email is lowercased and trimmed before querying to ensure consistent matches.

## Verification
- Manual test: POST `/login` with a valid user returns a token and updates `refresh_token`, `refresh_token_exp`, and `last_login_at`.
- Observed GORM SQL shows `WHERE "users"."deleted_at" IS NULL` automatically, as intended.

## Takeaways / Gotchas
- Ensure `models.User` fields and GORM tags map 1:1 to the DB: mismatches silently change GORM query behavior vs raw SQL.
- Custom types (like `Role`) must implement both `sql.Scanner` and `driver.Valuer` correctly.
- Prefer GORM for consistency with default scopes; if using raw SQL, remember to include `deleted_at IS NULL` and keep column casing/types aligned.
