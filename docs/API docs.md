# API Documentation

## Overview
**Base URL:** `http://localhost:8080/api/v1`
**Authentication:** Cookie (`session=YOUR_USER_ID`)

## Authenitcation
As of right now, the "authentication" is just dummy authentication, the `session` cookie has to be present, and is being read. But is not checked for anything..

## Endpoints

### Create transfer
**Endpoint:** `POST /transfers`
**Description:** Creates a transfer for the authenticated user.

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
| `expiry_date` | datetime | No | Expiry date in ISO8601 format, default: current date + 7 days |

**Response:**

Success:
```json
{
  "success": true,
  "data": {
    "transfer": {
      "id": "rJT7yu9R5SRjYlvwXcMI9w",
      "user_id": "dev",
      "file_count": 0,
      "total_byte_size": 0,
      "download_count": 0,
      "expiry_date": "2025-04-15T00:00:00Z",
      "creation_date": "2025-03-13T12:40:35.680625+01:00"
    }
  }
}
```

Errors:
- `401 Unauthorized` if the user is not authenticated.
- `400 Bad Request` if the expiry date is in the past or more than 30 days in the future.
- `500 Internal Server Error` if the transfer creation fails.

### Upload file
**Endpoint:** `POST /upload`
**Description:** Uploads a file to a specific transfer. The user must be authenticated, and the transfer ID must belong to them.

**Request:**
```http
POST /upload
Cookie: session=Hello, world!
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="transfer_id"

9
------WebKitFormBoundary
Content-Disposition: form-data; name="relative_path"

/images
------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="example.txt"
Content-Type: text/plain

(file content here)
------WebKitFormBoundary--
```

**Form Parameters:**
| Parameter | Type | Required | Description |
|--|--|--|--|
| `transfer_id` | integer | Yes | The ID of the transfer to upload the file to |
| `relative_path` | string | No | When uploading a directory, if it has child directories, what they should be called etc. |
| `file` | file | Yes | The file to be uploaded |

**Response:**

Success:
```json
{
  "success": true
}
```

Errors:
- `401 Unauthorized` if the user is not authenticated.
- `400 Bad Request` if no `transfer_id` is provided or it's not a valid integer.
- `413 Payload Too Large` if the file exceeds the allowed upload size.
- `404 Not Found` if the specified transfer does not exist.
- `401 Unauthorized` if the transfer does not belong to the user.
- `500 Internal Server Error` if the file processing fails.

## Error Handling

All errors return a standard JSON format.

**Example Error Response (400 Bad Request):
```json
{
  "success": false,
  "message": "Incorrect JSON format"
}
```


