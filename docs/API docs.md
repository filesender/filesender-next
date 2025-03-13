# API Documentation

## Overview
**Base URL:** `http://localhost:8080/api/v1`
**Authentication:** Cookie (`session=YOUR_USER_ID`)

## Authenitcation
As of right now, the "authentication" is just dummy authentication, the `session` cookie has to be present, and is being read. But is not checked for anything..

## Endpoints

### Create transfer
**Endpoint:** `POST /transfers`
**Description:** Creates a transfer for the authenticated user

**Request:**
```http
POST /transfers
Cookie: session=Hello, world!

{
  "subject": "Here are my files!",
  "message": "Thank you for giving me additional time on the assignment",
  "expiry_date": "2025-04-15T00:00:00Z"
}
```

**JSON Body:**

| Parameter | Type | Required | Description |
|--|--|--|--|
| `subject` | string | No | Subject of transfer |
| `message` | string | No | Message of transfer |
| `expiry_date` | datetime | No | Expiry date in ISO8601 format, default: current date + 7 days |

**Response:**

```json
{
  "success": true,
  "data": {
    "transfer": {
      "id": 9,
      "user_id": "Hello, world!",
      "guest_voucher_id": 0,
      "file_count": 0,
      "total_byte_size": 0,
      "subject": "Here are my files!",
      "message": "Thank you for giving me additional time on the assignment",
      "download_count": 0,
      "expiry_date": "2025-04-15T00:00:00Z",
      "creation_date": "2025-03-13T12:40:35.680625+01:00"
    }
  }
}
```

## Error Handling

All errors return a standard JSON format.

**Example Error Response (400 Bad Request):
```json
{
  "success": false,
  "message": "Incorrect JSON format"
}
```


