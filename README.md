# URL Shortener

A small, API-only URL shortener written in Go, backed by SQLite. Originally made in my first year of college — there's still no UI, so everything happens through a few HTTP endpoints.

**Live demo:** https://url-shortener-h6yx.onrender.com/

> The demo runs on Render's free tier, so the first request after a period of inactivity can take a few seconds while the instance wakes up.

## Features

- **Short links** — any URL becomes a 6-character code
- **Custom codes** — pick your own code instead of a random one
- **Click tracking** — check how many times a link has been visited
- **Expiring links** — expired codes return `410 Gone`
- **Rate limiting** — per-IP token bucket, returns `429` when you run out
- **No setup** — SQLite via a pure-Go driver; no CGO, no database server

## Running locally

Requires Go 1.25.4 or newer.

```bash
git clone https://github.com/omkhanjodkar-dev/url-shortener
cd url-shortener
go run main.go
```

The server listens on port `8080`, or on `$PORT` if that's set. A `codes.db`
SQLite file is created next to the code on first run.

## API

| Method | Path       | Body                                 | Response                 |
| ------ | ---------- | ------------------------------------ | ------------------------ |
| `POST` | `/shorten` | `{"url": "...", "code": "optional"}` | short link as plain text |
| `GET`  | `/{code}`  | —                                    | `302` redirect to the URL |
| `POST` | `/clicks`  | `{"code": "..."}`                    | click count as plain text |

Responses are plain text, not JSON.

## Try it

Start the server with `go run main.go`, then in another terminal:

```bash
# shorten a URL
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

```
localhost:8080/9kQ2mZ
```

To choose the code yourself, add `"code":"mylink"` to the request body.

```bash
# follow the link (use the code from the response above)
curl -i http://localhost:8080/9kQ2mZ
```

```
HTTP/1.1 302 Found
Location: https://example.com
```

```bash
# check how many times it has been visited
curl -X POST http://localhost:8080/clicks \
  -H "Content-Type: application/json" \
  -d '{"code":"9kQ2mZ"}'
```

```
1
```

On Windows PowerShell, use `Invoke-RestMethod` instead — `curl` there is an
alias for `Invoke-WebRequest`, and the line continuations above are bash-only:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/shorten -ContentType application/json -Body '{"url":"https://example.com"}'
```

## Status codes

| Code  | Meaning                                                                                      |
| ----- | -------------------------------------------------------------------------------------------- |
| `200` | Request succeeded                                                                            |
| `302` | Redirect to the original URL                                                                 |
| `400` | Malformed body, missing field, custom code already taken, or unknown code passed to `/clicks` |
| `404` | No code matches that path                                                                    |
| `405` | Wrong HTTP method for the endpoint                                                            |
| `410` | The link has expired                                                                          |
| `429` | Rate limit exceeded — slow down and retry                                                     |
| `500` | Internal error, usually the database                                                          |

## Project layout

```
main.go        HTTP handlers, rate limiter, SQLite helpers
code_test.go   Unit tests
tester.py      Scratch script for poking the API by hand
codes.db       SQLite database
```
