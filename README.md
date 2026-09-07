# Email-Job API

A production-ready REST API for transactional email delivery built with Go, successfully deployed as a real-world backend for a personal web profile. All API requests, data structures, and email workflows are aligned with the actual forms and requirements of the website, featuring RabbitMQ-based retries and Redis-backed daily rate limiting with automatic midnight reset.

---

## Architecture Overview

```
Client
  │
  ▼
Gin HTTP Server
  │
  ├── Admin Secret Middleware
  │     └── Header-based secret validation for protected endpoints
  │
  ├── Handlers → Services → Repositories
  │                             │
  │                         SQLite (Persistent email log & metadata store)
  │
  ├── Redis Cache
  │     └── Daily email counter (resets at 00:00 UTC)
  │           ├── Tracks total sent emails per day
  │           └── Enforces 130/day soft cap (of 150/day Mailtrap limit)
  │
  ├── Mailtrap SMTP / API
  │     └── Transactional email delivery (150 emails/day API Key limit)
  │
  └── RabbitMQ Worker
        └── Queues overflow emails (131+) → retries every 00:00 UTC on daily counter reset
```

---

## Tech Stack

| Layer | Technology | Reason |
| --- | --- | --- |
| Language | Go | High performance, lightweight goroutines for worker processes |
| Framework | Gin | Fast HTTP routing and clean middleware chaining |
| Database | SQLite | Relational datastore for persistent email logs and metadata |
| Cache Store | Redis | Sub-millisecond daily counter tracking with automatic TTL reset |
| Message Queue | RabbitMQ | Reliable overflow email retry queue triggered on daily limit reset |
| Mail Provider | Mailtrap | Transactional email delivery with 150 emails/day API Key quota |
| Containerization | Docker & Docker Compose | Consistent local and containerized deployments |

---

## Features

### Smart Daily Rate Limiting via Redis

* **Redis Counter:** Every outbound email increments a Redis-backed daily counter scoped to the current UTC date.
* **Soft Cap at 130/day:** Out of the 150 emails/day Mailtrap API Key limit, 130 are allocated to incoming requests, reserving the remaining quota for other operational needs.
* **Overflow Queuing:** When the daily counter reaches 130, any additional email requests are pushed to a RabbitMQ queue instead of being dropped.

### RabbitMQ Retry on Daily Reset

* **Queued Delivery:** Overflow emails (request #131 onwards) are held in a RabbitMQ queue without data loss.
* **Automatic Retry:** The worker monitors the daily reset at **00:00 UTC** — matching Mailtrap's quota reset cycle — and replays queued messages once the counter clears.

### Dual Email Workflow

* **Contact Message:** Guests can submit a contact form that delivers a direct email notification for personal web profile inquiries.
* **Live Demo Request:** Visitors can request a live demo session; admins are notified and can trigger a confirmation reply back to the requester via a dedicated ready endpoint.

### Admin Email Supervision

* **Header-Based Access Control:** Admin-only endpoints are protected by `AdminSecretMiddleware`, which validates a secret key from the request header — no session required.
* **Full Email Audit:** Admins can retrieve all emails, filter by date, look up by ID or originating IP address, and manage live demo request pipelines.

---

## API Endpoints

<details>
<summary><strong>Emails</strong> — 8 endpoints</summary>

<br>

<details>
<summary>&nbsp;&nbsp;&nbsp;&nbsp;<strong>Public</strong> — 3 endpoints</summary>

<br>

| Method | Endpoint | Access | Notes |
| --- | --- | --- | --- |
| POST | `/api/v1/emails/contact` | Public | Submit a contact message — delivers email notification |
| POST | `/api/v1/emails/live-demo` | Public | Submit a live demo request — queued if daily cap reached |
| POST | `/api/v1/emails/live-demo/:id/ready` | Admin Secret | Notify requester that their live demo session is ready |

</details>

<details>
<summary>&nbsp;&nbsp;&nbsp;&nbsp;<strong>Admin</strong> — 5 endpoints</summary>

<br>

| Method | Endpoint | Access | Notes |
| --- | --- | --- | --- |
| GET | `/api/v1/emails/live-demo` | Admin Secret | Retrieve all live demo requests |
| GET | `/api/v1/emails` | Admin Secret | Retrieve all email records across the system |
| GET | `/api/v1/emails/today` | Admin Secret | Retrieve all emails sent today (current UTC date) |
| GET | `/api/v1/emails/id/:id` | Admin Secret | Retrieve a specific email record by ID |
| GET | `/api/v1/emails/ip/:ip` | Admin Secret | Retrieve all email records from a specific IP address |

</details>

</details>

---

## Getting Started

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/) & Docker Compose
* Go 1.22+

### Run Locally

```bash
# Clone the repository
git clone https://github.com/Tyomaaans/email-job.git
cd email-job

# Copy environment variables
cp .env.example .env

# Start SQLite, Redis, and RabbitMQ containers
docker compose up -d
```

### Environment Variables

See `.env.example` for all required variables. Key configs:

```
env.example
# App
APP_PORT=

# Redis
REDIS_ADDR=
REDIS_PASSWORD=

# RabbitMQ
RABBITMQ_DSN=
RABBITMQ_USER=
RABBITMQ_PASS=
RABBITMQ_VHOST=

# MailTrap
MAILTRAP_NAME=
MAILTRAP_ROLE=
MAILTRAP_API_KEY=

# Mail Owner
MAIL_OWNER=

# Admin Secret Key
ADMIN_SECRET_KEY=
```

---

## Project Status

| Feature | Status | Notes |
| --- | --- | --- |
| Contact Email Endpoint (`POST /emails/contact`) | ✅ Done | Delivers direct email for web profile contact |
| Live Demo Request (`POST /emails/live-demo`) | ✅ Done | Queued via RabbitMQ if daily cap is reached |
| Live Demo Ready Notify (`POST /emails/live-demo/:id/ready`) | ✅ Done | Admin-triggered confirmation email to requester |
| Redis Daily Counter | ✅ Done | Tracks sent emails per UTC day, resets at 00:00 UTC |
| Daily Soft Cap (130/day) | ✅ Done | Reserves headroom from 150/day Mailtrap limit |
| RabbitMQ Overflow Queue | ✅ Done | Holds requests beyond cap, retries on daily reset |
| Admin Email Audit | ✅ Done | Filter by ID, IP, date, or list all records |
| Admin Secret Middleware | ✅ Done | Header-based secret validation for protected routes |