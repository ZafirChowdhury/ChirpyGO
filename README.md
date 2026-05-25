# ChirpyGO
A Twitter-like REST API built in Go.

---

## What it does
- Register users and authenticate with JWT access tokens and refresh tokens
- Post, retrieve, and delete short messages called "chirps" (max 140 characters)
- Automatic profanity filtering on chirp content
- Token refresh flow — exchange a refresh token for a new access token
- Upgrade users to "Chirpy Red" via a webhook
- Filter and sort chirps by author or date

---

## API Reference

### Users

#### `POST /api/users`
Register a new user.

Request:
```json
{
  "email": "lane@example.com",
  "password": "secret123"
}
```

Response `201`:
```json
{
  "id": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "email": "lane@example.com",
  "is_chirpy_red": false
}
```

---

#### `POST /api/login`
Log in and receive an access token and refresh token.

Request:
```json
{
  "email": "lane@example.com",
  "password": "secret123",
  "expires_in_seconds": 3600
}
```
`expires_in_seconds` is optional — defaults to 1 hour. Capped at 1 hour.

Response `200`:
```json
{
  "id": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "email": "lane@example.com",
  "is_chirpy_red": false,
  "token": "<access_jwt>",
  "refresh_token": "<refresh_token>"
}
```

---

#### `PUT /api/users`
Update the authenticated user's email and password.

Headers:
```
Authorization: Bearer <access_token>
```

Request:
```json
{
  "email": "new@example.com",
  "password": "newpassword"
}
```

Response `200`:
```json
{
  "id": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "email": "new@example.com",
  "is_chirpy_red": false
}
```

---

### Authentication

#### `POST /api/refresh`
Exchange a valid refresh token for a new short-lived access token.

Headers:
```
Authorization: Bearer <refresh_token>
```

Response `200`:
```json
{
  "token": "<new_access_jwt>"
}
```

Returns `401` if the refresh token is expired or revoked.

---

#### `POST /api/revoke`
Revoke a refresh token — effectively logging the user out.

Headers:
```
Authorization: Bearer <refresh_token>
```

Response `204` — no body.

---

### Chirps

#### `POST /api/chirps`
Create a new chirp. Requires authentication. Max 140 characters. Profanity is automatically filtered.

Headers:
```
Authorization: Bearer <access_token>
```

Request:
```json
{
  "body": "Hello world!"
}
```

Response `201`:
```json
{
  "id": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "body": "Hello world!",
  "user_id": "uuid"
}
```

---

#### `GET /api/chirps`
Get all chirps. Supports optional query parameters.

| Parameter   | Values         | Default | Description                        |
|-------------|----------------|---------|------------------------------------|
| `author_id` | any user UUID  | —       | Filter chirps by a specific author |
| `sort`      | `asc` / `desc` | `asc`   | Sort by `created_at`               |

Example:
```
GET /api/chirps?author_id=uuid&sort=desc
```

Response `200`:
```json
[
  {
    "id": "uuid",
    "created_at": "...",
    "updated_at": "...",
    "body": "Hello world!",
    "user_id": "uuid"
  }
]
```

---

#### `GET /api/chirps/{chirpID}`
Get a single chirp by ID.

Response `200`:
```json
{
  "id": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "body": "Hello world!",
  "user_id": "uuid"
}
```

Returns `404` if not found.

---

#### `DELETE /api/chirps/{chirpID}`
Delete a chirp. Only the author can delete their own chirp.

Headers:
```
Authorization: Bearer <access_token>
```

Response `204` — no body.

Returns `403` if the authenticated user is not the author.

---

### Webhooks

#### `POST /api/polka/webhooks`
Upgrade a user to Chirpy Red. Authenticated via API key.

Headers:
```
Authorization: ApiKey <polka_api_key>
```

Request:
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
```

Response `204` — no body. Events other than `user.upgraded` are ignored with a `204`.

Returns `401` if the API key is missing or invalid.

---

### Admin

#### `GET /admin/metrics`
Returns an HTML page showing how many times the file server has been visited.

#### `POST /admin/reset`
Resets the hit counter and clears all users from the database. Only available when `PLATFORM=dev`.

Returns `403` in non-dev environments.

---

## Getting started

### Prerequisites
- [Go](https://go.dev/dl/) (1.21+)
- [PostgreSQL](https://www.postgresql.org/download/) running locally

### Install
```bash
git clone https://github.com/ZafirChowdhury/ChirpyGO
cd ChirpyGO
go mod download
```

### Config
Create a `.env` file in the project root:
```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET_KEY=your-jwt-secret
POLKA_KEY=your-polka-api-key
```

Create the database in psql first:
```sql
CREATE DATABASE chirpy;
```

### Run migrations
The project uses `goose` for migrations. From the repo root:
```bash
goose -dir sql/schema postgres "your-connection-string" up
```

### Run the server
```bash
go run .
```
Server starts on port `8080`.

---

## What I learned
- Building a REST API in Go using only the standard `net/http` package — no frameworks
- JWT authentication — creating and validating signed tokens with expiry, and the difference between access tokens and refresh tokens
- Refresh token rotation — storing tokens in the database, checking expiry and revocation, and issuing new access tokens
- Password hashing with Argon2id — a modern, memory-hard hashing algorithm designed to resist brute-force attacks
- Using `sqlc` to generate type-safe Go from raw SQL, keeping the database layer explicit with no ORM
- Writing and running PostgreSQL migrations with `goose`
- Handling nullable SQL columns in Go with `sql.NullTime`
- API key authentication as an alternative auth scheme for webhooks
- Structuring a Go project — separating concerns across handlers, auth, and database packages
- Why sensitive config (secrets, API keys, DB URLs) must never be committed to source control and how environment variables solve that