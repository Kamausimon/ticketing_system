# Account Preferences System

## Overview
This system provides user account preferences for timezones, currencies, and date/datetime formats. These preferences allow users to customize how they view dates, times, and monetary values throughout the application.

## Database Tables

### Timezones
Stores available timezone options with IANA timezone names.

**Fields:**
- `id` - Primary key
- `name` - Short name (e.g., "EAT", "UTC", "EST")
- `display_name` - Full descriptive name
- `offset` - UTC offset (e.g., "+03:00", "-05:00")
- `iana_name` - Standard IANA timezone identifier (e.g., "Africa/Nairobi")
- `is_active` - Boolean flag to enable/disable

### Currencies
Stores available currency options.

**Fields:**
- `id` - Primary key
- `code` - 3-letter ISO currency code (e.g., "USD", "KSH", "EUR")
- `name` - Full currency name (e.g., "US Dollar")
- `symbol` - Currency symbol (e.g., "$", "KSh", "€")
- `is_active` - Boolean flag to enable/disable

### DateFormats
Stores available date format options.

**Fields:**
- `id` - Primary key
- `format` - Format pattern (e.g., "YYYY-MM-DD", "DD/MM/YYYY")
- `example` - Example date string (e.g., "2024-12-25")
- `is_active` - Boolean flag to enable/disable

### DateTimeFormats
Stores available datetime format options.

**Fields:**
- `id` - Primary key
- `format` - Format pattern (e.g., "YYYY-MM-DD HH:mm")
- `example` - Example datetime string (e.g., "2024-12-25 14:30")
- `is_active` - Boolean flag to enable/disable

## API Endpoints

### Get Available Options

#### Get Timezones
```
GET /account/timezones
```

**Response:**
```json
{
  "timezones": [
    {
      "id": 1,
      "name": "UTC",
      "display_name": "UTC - Coordinated Universal Time",
      "offset": "+00:00",
      "iana_name": "UTC"
    },
    {
      "id": 2,
      "name": "EAT",
      "display_name": "East Africa Time (Nairobi, Kampala, Dar es Salaam)",
      "offset": "+03:00",
      "iana_name": "Africa/Nairobi"
    }
  ]
}
```

#### Get Currencies
```
GET /account/currencies
```

**Response:**
```json
{
  "currencies": [
    {
      "id": 1,
      "code": "USD",
      "name": "US Dollar",
      "symbol": "$"
    },
    {
      "id": 2,
      "code": "KSH",
      "name": "Kenyan Shilling",
      "symbol": "KSh"
    }
  ]
}
```

#### Get Date Formats
```
GET /account/date-formats
```

**Response:**
```json
{
  "date_formats": [
    {
      "id": 1,
      "format": "YYYY-MM-DD",
      "example": "2024-12-25"
    },
    {
      "id": 2,
      "format": "DD/MM/YYYY",
      "example": "25/12/2024"
    }
  ]
}
```

#### Get DateTime Formats
```
GET /account/datetime-formats
```

**Response:**
```json
{
  "datetime_formats": [
    {
      "id": 1,
      "format": "YYYY-MM-DD HH:mm",
      "example": "2024-12-25 14:30"
    },
    {
      "id": 2,
      "format": "DD/MM/YYYY HH:mm",
      "example": "25/12/2024 14:30"
    }
  ]
}
```

### Get User Preferences
```
GET /account/preferences
Authorization: Bearer <token>
```

**Response:**
```json
{
  "timezone_id": 2,
  "date_format_id": 1,
  "date_time_format_id": 1,
  "currency_id": 2,
  "email_notifications": true,
  "sms_notifications": false
}
```

### Update User Preferences
```
PUT /account/preferences
Authorization: Bearer <token>
Content-Type: application/json

{
  "timezone_id": 2,
  "date_format_id": 1,
  "date_time_format_id": 1,
  "currency_id": 2
}
```

**Response:**
```json
{
  "message": "Preferences updated successfully",
  "preferences": {
    "timezone_id": 2,
    "date_format_id": 1,
    "date_time_format_id": 1,
    "currency_id": 2
  }
}
```

## Seeded Data

### Timezones (16 total)
- UTC
- East Africa Time (EAT)
- West Africa Time (WAT)
- Central Africa Time (CAT)
- Eastern/Central/Mountain/Pacific Standard Time (US)
- GMT/CET/EET (Europe)
- IST (India), JST (Japan), CST (China)
- AEST (Australia), NZST (New Zealand)

### Currencies (14 total)
- USD (US Dollar)
- KSH (Kenyan Shilling)
- EUR (Euro)
- GBP (British Pound)
- NGN (Nigerian Naira)
- ZAR (South African Rand)
- GHS (Ghanaian Cedi)
- UGX (Ugandan Shilling)
- TZS (Tanzanian Shilling)
- CAD/AUD (Canadian/Australian Dollar)
- INR (Indian Rupee)
- JPY/CNY (Japanese Yen/Chinese Yuan)

### Date Formats (8 total)
- YYYY-MM-DD
- DD/MM/YYYY
- MM/DD/YYYY
- DD-MM-YYYY
- MMM DD, YYYY
- DD MMM YYYY
- YYYY/MM/DD
- DD.MM.YYYY

### DateTime Formats (8 total)
- YYYY-MM-DD HH:mm
- DD/MM/YYYY HH:mm
- MM/DD/YYYY hh:mm A
- DD-MM-YYYY HH:mm
- MMM DD, YYYY hh:mm A
- DD MMM YYYY HH:mm
- YYYY/MM/DD HH:mm
- DD.MM.YYYY HH:mm

## Implementation Notes

1. **Automatic Seeding**: The seed data is automatically loaded when the server starts. It only creates records that don't already exist.

2. **Nullable Fields**: All preference fields in the Account model are nullable (`*int`), allowing users to have no preferences set (defaults will be used).

3. **Active Flag**: Each preference table has an `is_active` flag, allowing administrators to disable options without deleting them.

4. **Validation**: When updating preferences, the IDs should be validated to ensure they exist in the respective tables.

## Frontend Integration

1. **On Page Load**: Fetch available options for each preference type
2. **Display Dropdowns**: Show user-friendly names with examples
3. **Save Preferences**: Send only the IDs in the update request
4. **Handle Nulls**: If a preference is null, use system defaults

## Example Frontend Flow

```javascript
// Fetch options
const timezones = await fetch('/account/timezones');
const currencies = await fetch('/account/currencies');
const dateFormats = await fetch('/account/date-formats');
const datetimeFormats = await fetch('/account/datetime-formats');

// Get current preferences
const prefs = await fetch('/account/preferences');

// Update preferences
await fetch('/account/preferences', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    timezone_id: 2,      // EAT
    currency_id: 2,      // KSH
    date_format_id: 1,   // YYYY-MM-DD
    date_time_format_id: 1  // YYYY-MM-DD HH:mm
  })
});
```
