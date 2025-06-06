# API Documentation

## Overview
**Base URL**: `http://localhost:8080`

## Authentication

> **Development-Only Notice:**
> This documentation assumes you're running the instance in with **dummy** authentication, which ignores any authentication and the request is treated as coming from user `dev`.
>
> If you want to test this on your production server, you'll need to use your authentication headers & cookies to access the API.

## Endpoints

### 1. Upload — **`POST /upload`**

Uploads a brand-new file
| Header            | Required | Default | Description                                   |
|-------------------|:--------:|:-------:|-----------------------------------------------|
| `Upload-Complete` |   No     |  `1`    | Flag indicating whether this is the final chunk |

#### Request Example
```http
POST /upload HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary
Upload-Complete: 1

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="example.txt"
Content-Type: text/plain

(file content here)
------WebKitFormBoundary--
```

#### cURL
```bash
curl -i -X POST http://localhost:8080/upload \
     -H "Upload-Complete: 1" \
     -F "file=@example.txt"
```

#### Responses
| Status            | When                 | Headers                                        |
| ----------------- | -------------------- | -------------------------------------------------------------- |
| **303 See Other** | Final chunk received | `Location: /view/{userID}/{fileID}`                            |
| **202 Accepted**  | More chunks expected | `Location: /upload/{fileID}`<br>`Upload-Offset: <next-offset>` |
| **401 Unauthorized** | requester not authenticated |  |
| **413 Payload Too Large** | file exceeds server limit |  |
| **500 Internal Server Error** | unexpected failure while processing |  |

## Chunk Upload — **`PATCH /upload/{fileID}`**

Appends bytes to an existing in-progress upload.

| Header            | Required | Default | Description                                          |
| ----------------- | :------: | :-----: | ---------------------------------------------------- |
| `Upload-Complete` |    No    |   `1`   | Set to `1` if this is the final chunk                |
| `Upload-Offset`   |  **Yes** |    —    | Position (in bytes) at which this chunk should start (should be ℕ+, natural numbers only) |

#### Request Example
```http
PATCH /upload/uY3D4i7Uf5Mcocu2LCtMNw HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary
Upload-Complete: 1
Upload-Offset: 1118

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="example_continued.txt"
Content-Type: text/plain

(file content here)
------WebKitFormBoundary--
```

#### cURL
```bash
curl -i -X PATCH http://localhost:8080/upload/uY3D4i7Uf5Mcocu2LCtMNw \
     -H "Upload-Complete: 1" \
     -H "Upload-Offset: 1118" \
     -F "file=@example_continued.txt"
```

#### Responses
| Status            | When                 | Headers                                       |
| ----------------- | -------------------- | -------------------------------------------------------------- |
| **303 See Other** | Final chunk received | `Location: /view/{userID}/{fileID}`                            |
| **202 Accepted**  | More chunks expected | `Location: /upload/{fileID}`<br>`Upload-Offset: <next-offset>` |
| **401 Unauthorized** | requester not authenticated |  |
| **404 Not Found** | upload ID is missing or does not belong to the user |  |
| **413 Payload Too Large** | file exceeds server limit |  |
| **500 Internal Server Error** | unexpected failure while processing |  |

## Download — **`GET /download/{userID}/{fileID}`**

Streams the stored file to the client. Supports standard `Range` requests.

#### Simple Request
```http
GET /download/zoPFJGeHhHR20J7_R8wrdtQIVISjS0NHPTOf-veRkHQ/uY3D4i7Uf5Mcocu2LCtMNw HTTP/1.1
```` 

#### cURL
```bash
curl -i http://localhost:8080/download/zoPFJGeHhHR20J7_R8wrdtQIVISjS0NHPTOf-veRkHQ/uY3D4i7Uf5Mcocu2LCtMNw
```

#### 200 OK Response
```http
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Length: 2236
Content-Type: text/plain; charset=utf-8
Last-Modified: Thu, 05 Jun 2025 12:06:48 GMT
Date: Thu, 05 Jun 2025 12:07:45 GMT

(file content here)
```

> For partial transfers the server returns `206 Partial Content` with the appropriate `Content-Range` header.

**Errors**
- `404 Not Found` file does not exist or is inaccessible to the user.
- `416 Range Not Satisfiable` the requested byte range cannot be served.
- `400 Bad Request` malformed `Range` header.
- `500 Internal Server Error`
