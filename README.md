<div align="center">

# 🌍 Roamify

**The travel social platform built for explorers who actually go.**

Plan trips with AI, link up with your squad, track every stamp in your passport, and discover where the world is going next — all in one place.

[Live App](https://roamify-zeta.vercel.app) · [API Health](https://roamify-f9v5.onrender.com/health) · [API Docs](https://roamify-f9v5.onrender.com/swagger/#/)

</div>

---

## What is Roamify?

Roamify is a full-stack travel social app. It combines the trip-planning utility of a travel tool with the social layer of a community platform — think group itineraries, live squad chat, vibe-matched discovery, gamified passport stamps, and an AI travel assistant, all in one backend.

---

## Tech stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25 · Gin |
| **Database** | PostgreSQL (GORM, AutoMigrate) |
| **Auth** | JWT (HS256) + Social OAuth (Google, TikTok, Apple) |
| **AI** | Groq API (itinerary generation, vibe search, travel assistant) |
| **Image storage** | Cloudinary |
| **Email** | SMTP + Resend |
| **Hosting** | Render (API) · Vercel (frontend) |

---

## Project structure

```
Roamify/
├── cmd/
│   ├── main.go                     ← server entrypoint
│   └── seed/
│       └── main.go                 ← database seeder
│
├── config/
│   ├── config.go                   ← env loading, Config struct
│   ├── database.go                 ← GORM connection
│   ├── migrate.go                  ← AutoMigrate (all models)
│   └── seed.go                     ← admin role seeder
│
├── internal/
│   ├── modules/
│   │   ├── auth/                   ← register, login, social auth, password reset
│   │   ├── users/                  ← profiles, vibe profiles, follow graph, privacy
│   │   ├── trips/                  ← trips, members, itinerary, expenses, squad chat
│   │   ├── posts/                  ← social feed, comments, likes
│   │   ├── discovery/              ← home dashboard, vibe search, atlas, AI assistant
│   │   ├── challenges/             ← gamification, trivia, leaderboard
│   │   ├── notifications/          ← in-app inbox + notification preferences
│   │   ├── passport/               ← encrypted passport vault + country stamps
│   │   ├── wishlist/               ← saved places, collections
│   │   ├── reports/                ← content moderation reports
│   │   ├── admin/                  ← admin dashboard (users, posts, trips, stats)
│   │   └── upload/                 ← Cloudinary image upload
│   │
│   ├── seed/                       ← modular demo data seeders
│   │   ├── helpers.go
│   │   ├── users.go
│   │   ├── trips.go
│   │   ├── squads.go
│   │   ├── comments.go
│   │   └── likes.go
│   │
│   └── services/
│       ├── cloudinary.go           ← image upload service
│       └── email_service.go        ← email sending (SMTP + Resend)
│
└── pkg/
    ├── jwt/                        ← token generation & parsing
    ├── middleware/                 ← auth, admin guard, CORS, logger
    ├── response/                   ← standardised JSON responses
    └── validator/                  ← request validation helpers
```

---

## Getting started

### 1. Clone and install

```bash
git clone https://github.com/khadijayo/roamify.git
cd Roamify
go mod download
```

### 2. Configure environment

Copy `.env.example` (or create `.env`) with the following variables:

```env
# ── Server ────────────────────────────────────────────────
PORT=8080
APP_ENV=development          # development | production
APP_BASE_URL=http://localhost:8080

# ── Database ──────────────────────────────────────────────
# Option A: full DSN (Render, Supabase, Railway, etc.)
DATABASE_URL=postgres://user:password@host:5432/roamify?sslmode=require

# Option B: individual params (local dev)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=roamify
DB_SSLMODE=disable

# ── Auth ──────────────────────────────────────────────────
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=72

# ── AI (Groq) ─────────────────────────────────────────────
GROQ_API_KEY=gsk_xxxxxxxxxxxxxxxxxxxx

# ── Cloudinary ────────────────────────────────────────────
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name

# ── Email ─────────────────────────────────────────────────
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your@email.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@roamify.app
SMTP_FROM_NAME=Roamify

# ── Admin ─────────────────────────────────────────────────
ADMIN_EMAILS=admin@yourdomain.com
```

### 3. Run the server

```bash
go run cmd/main.go
```

On first startup GORM automatically creates all database tables. You'll see:

```
[migrate] all tables migrated successfully
[roamify] server starting on :8080 (env: development)
```

### 4. Seed demo data (optional)

```bash
go run cmd/seed/main.go
```

Seeds a complete presentation dataset: 20 users, 40 trips, 20 travel posts, 40 travel questions, 30 post comments, 50 challenges, 15 challenge participations, 30 likes/reactions, 40 flights, 40 hotels, and 40 notifications.

The seeder uses the existing app models where they exist. Because Roamify does not currently have persistent models for travel questions, flights, or hotels, the seeder also creates `travel_questions`, `flight_entries`, and `hotel_entries` as seed-owned presentation tables.

All demo users use the password `Roamify2026!` and emails ending in `@roamify.demo`.

To reseed, run the same command again. It clears previous `@roamify.demo` records and seed-owned flight/hotel/question data first, then inserts a fresh dataset. Existing non-demo users are left alone.

To run migrations only, start the API once:

```bash
go run cmd/main.go
```

The seed command also calls `AutoMigrate` before inserting data, so it can prepare the database directly for a final demo.

---

## API reference

Base URL: `/api/v1`

All protected endpoints require:
```
Authorization: Bearer <jwt_token>
```

### Auth — `/auth`

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/auth/register` | — | Create account with email + password |
| POST | `/auth/login` | — | Log in, receive JWT |
| POST | `/auth/social` | — | OAuth login (Google, TikTok, Apple) |
| GET | `/auth/verify-email` | — | Verify email address via token link |
| POST | `/auth/forgot-password` | — | Request password reset code |
| POST | `/auth/verify-reset-code` | — | Validate reset code |
| POST | `/auth/reset-password` | — | Set new password |
| POST | `/auth/resend-verification` | — | Resend verification email |

### Users — `/users`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/me` | Get own profile |
| PATCH | `/users/me` | Update name, avatar |
| GET | `/users/me/vibe` | Get vibe profile |
| PUT | `/users/me/vibe` | Create or update vibe profile (pace, budget style, interests…) |
| GET | `/users/me/privacy` | Get privacy settings |
| PATCH | `/users/me/privacy` | Toggle ghost mode, data sharing, map visibility |
| POST | `/users/follow` | Follow a user |
| DELETE | `/users/follow/:userId` | Unfollow a user |
| GET | `/users/search?q=` | Search users by name |
| GET | `/users/:userId` | Public profile |
| GET | `/users/:userId/followers` | Followers list |
| GET | `/users/:userId/following` | Following list |
| GET | `/users/:userId/posts` | Posts by a specific user |

### Trips — `/trips`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/trips/` | Create a trip manually |
| POST | `/trips/plan-and-create` | Create trip + generate AI itinerary in one call |
| GET | `/trips/` | My trips |
| GET | `/trips/all` | All public trips |
| GET | `/trips/:tripId` | Trip detail |
| PATCH | `/trips/:tripId` | Update trip |
| DELETE | `/trips/:tripId` | Delete trip |
| POST | `/trips/:tripId/join` | Join an open trip |
| **Members** | | |
| POST | `/trips/:tripId/members` | Invite a member |
| GET | `/trips/:tripId/members` | List members |
| PATCH | `/trips/:tripId/members/status` | Accept / decline invitation |
| DELETE | `/trips/:tripId/members/:userId` | Remove member |
| **Itinerary** | | |
| POST | `/trips/:tripId/itinerary` | Add itinerary item |
| POST | `/trips/:tripId/itinerary/generate-ai` | Generate itinerary with AI |
| GET | `/trips/:tripId/itinerary` | Get full itinerary |
| PATCH | `/trips/:tripId/itinerary/:itemId` | Update itinerary item |
| DELETE | `/trips/:tripId/itinerary/:itemId` | Delete itinerary item |
| **Expenses** | | |
| POST | `/trips/:tripId/expenses` | Log an expense |
| GET | `/trips/:tripId/expenses` | Get expense breakdown |
| PATCH | `/trips/:tripId/expenses/:expenseId` | Edit expense |
| DELETE | `/trips/:tripId/expenses/:expenseId` | Delete expense |
| **Squad Chat** | | |
| GET | `/trips/:tripId/chat` | Get chat history |
| POST | `/trips/:tripId/chat` | Send a message |
| GET | `/trips/:tripId/map` | Get map pins for the trip |

### Posts — `/posts`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/posts/` | Create post (image, location, tags, visibility) |
| GET | `/posts/` | Social feed |
| GET | `/posts/:postId` | Single post |
| PATCH | `/posts/:postId` | Edit post |
| DELETE | `/posts/:postId` | Delete post |
| POST | `/posts/:postId/like` | Like a post |
| DELETE | `/posts/:postId/like` | Unlike a post |
| GET | `/posts/:postId/comments` | Get comments |
| POST | `/posts/:postId/comments` | Add comment |
| DELETE | `/posts/:postId/comments` | Delete comment |

### Discovery — `/discovery`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/discovery/home` | Personalised home dashboard |
| GET | `/discovery/vibe-search?query=` | AI-powered vibe-based destination search |
| GET | `/discovery/atlas` | Browse destinations (filtered) |
| GET | `/discovery/atlas/geojson` | Atlas as GeoJSON for map rendering |
| GET | `/discovery/atlas/:locationId` | Location detail |
| GET | `/discovery/price-drops` | Personalised flight price drops |
| GET | `/discovery/recommended` | AI-recommended destinations |
| POST | `/discovery/locations/generate` | Generate destinations from quiz answers |
| GET | `/search/global?q=` | Global search (users, trips, places) |
| POST | `/assistant/travel` | AI travel assistant chat |

### Challenges & Gamification — `/challenges`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/challenges/` | List all active challenges |
| POST | `/challenges/` | Create a challenge (admin) |
| POST | `/challenges/accept` | Accept a challenge |
| POST | `/challenges/complete` | Mark a challenge complete |
| GET | `/challenges/my-progress` | My challenge progress |
| GET | `/challenges/leaderboard` | Global leaderboard |
| GET | `/challenges/trivia` | List trivia questions |
| POST | `/challenges/trivia` | Create trivia question |
| POST | `/challenges/trivia/answer` | Answer a trivia question |

### Notifications — `/notifications`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/notifications` | In-app notification inbox |
| GET | `/notifications/unread-count` | Unread count badge |
| PATCH | `/notifications/read-all` | Mark all as read |
| PATCH | `/notifications/:notifId/read` | Mark one as read |
| GET | `/notifications/settings` | Get notification preferences |
| PATCH | `/notifications/settings` | Toggle push / email / in-app per type |

### Passport Vault — `/passport`

| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/passport/vault` | Store encrypted passport data |
| GET | `/passport/vault` | Retrieve vault (masked fields only) |
| DELETE | `/passport/vault` | Delete vault |
| POST | `/passport/stamps` | Add a country stamp |
| GET | `/passport/stamps` | List all stamps |
| DELETE | `/passport/stamps/:stampId` | Remove a stamp |

### Wishlist — `/wishlist`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/wishlist/items` | Save a place |
| GET | `/wishlist/items` | All saved places |
| PATCH | `/wishlist/items/:itemId` | Edit saved place |
| DELETE | `/wishlist/items/:itemId` | Remove saved place |
| POST | `/wishlist/collections` | Create a collection |
| GET | `/wishlist/collections` | All collections |
| GET | `/wishlist/collections/:collectionId` | Collection detail with items |
| PATCH | `/wishlist/collections/:collectionId` | Edit collection |
| DELETE | `/wishlist/collections/:collectionId` | Delete collection |
| POST | `/wishlist/collections/:collectionId/items` | Add item to collection |
| DELETE | `/wishlist/collections/:collectionId/items/:itemId` | Remove from collection |

### Upload — `/upload`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/upload/image` | Upload image to Cloudinary, returns secure URL |

### Admin — `/admin`

Admin endpoints require both a valid JWT **and** the `admin` role.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/users` | List all users |
| GET | `/admin/users/:id` | User detail |
| PATCH | `/admin/users/:id/role` | Change role |
| GET | `/admin/users/:id/activity` | User activity log |
| PUT | `/admin/users/:id/ban` | Ban user |
| PUT | `/admin/users/:id/unban` | Unban user |
| DELETE | `/admin/users/:id` | Delete user |
| GET | `/admin/posts` | All posts |
| PUT | `/admin/posts/:id/hide` | Hide post |
| PUT | `/admin/posts/:id/unhide` | Unhide post |
| DELETE | `/admin/posts/:id` | Delete post |
| GET | `/admin/comments` | All comments |
| DELETE | `/admin/comments/:id` | Delete comment |
| GET | `/admin/trips` | All trips |
| DELETE | `/admin/trips/:id` | Delete trip |
| GET | `/admin/reports` | Content reports queue |
| PUT | `/admin/reports/:id/resolve` | Resolve a report |
| GET | `/admin/stats` | Platform statistics |

---

## Data models

### User
Supports email/password and social OAuth. Every user has an optional **VibeProfile** that powers personalised discovery — travel pace (chill / balanced / adventure), budget style (backpacker / mid-range / luxury), preferred vibes, interests, and a gamification score (RoamifyPoints, ExplorerLevel, CountriesVisited).

### Trip
The core entity. A trip has an owner, optional members (`TripMember` with roles: owner / admin / member), a status lifecycle (`planning → ongoing → completed → archived`), vibe tags, budget, and relationships to itinerary items, expenses, and chat messages.

### Post
A social post with content, location, image, visibility (public / followers / private), tags, likes, and comments. Drives the main social feed.

### Passport Vault
Encrypted storage for sensitive passport data. Only masked fields (nationality, expiry) are ever returned from the API. Separate `PassportStamp` records track countries visited.

### Challenge / Trivia
Gamification layer. Users accept challenges, earn RoamifyPoints on completion, and appear on the global leaderboard. Trivia questions award points per correct answer.

---

## Key features

**AI itinerary generation** — `POST /trips/plan-and-create` sends destination, dates, budget and a natural-language prompt to Groq. The API returns a fully structured trip with day-by-day itinerary items in a single call.

**Vibe search** — `GET /discovery/vibe-search?query=dark+academia+city` uses AI to match user intent to curated destinations, bypassing keyword-only search.

**Squad chat** — Real-time-capable chat history is stored per trip, giving group trip members a shared communication thread alongside the itinerary and expenses.

**Passport Vault** — Encrypted client-side before upload. The server stores only the ciphertext and masked metadata, so raw passport data is never exposed in API responses.

**Ghost mode** — Users can enable ghost mode via privacy settings, hiding their location from the atlas and map features.

---

## Deployment

### Render (API)

1. Create a new **Web Service** pointed at this repo.
2. Set **Build Command**: `go build -o main ./cmd/main.go`
3. Set **Start Command**: `./main`
4. Add all environment variables from the section above.
5. Render sets `PORT` and `RENDER=true` automatically — the app respects both.

### Database

Any managed PostgreSQL service works: Render Postgres, Supabase, Railway, Neon. Set `DATABASE_URL` to the full connection string.

### Frontend

The frontend lives at [roamify-zeta.vercel.app](https://roamify-zeta.vercel.app) and is deployed separately to Vercel. Point its API base URL env var at the Render service URL.

---

## Environment variable reference

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅* | — | Full Postgres DSN. Takes precedence over individual params. |
| `DB_HOST` | ✅* | `localhost` | Postgres host |
| `DB_PORT` | — | `5432` | Postgres port |
| `DB_USER` | — | `postgres` | Postgres user |
| `DB_PASSWORD` | ✅ | — | Postgres password |
| `DB_NAME` | — | `roamify` | Database name |
| `DB_SSLMODE` | — | `disable` | `disable` locally, `require` on Render |
| `JWT_SECRET` | ✅ | — | HS256 signing key |
| `JWT_EXPIRY_HOURS` | — | `72` | Token lifetime |
| `GROQ_API_KEY` | ✅ | — | Powers AI itinerary, vibe search, assistant |
| `CLOUDINARY_URL` | ✅ | — | Image upload service |
| `APP_ENV` | — | `development` | `production` disables GORM query logs |
| `APP_BASE_URL` | ✅ (prod) | `http://localhost:8080` | Used in email verification links |
| `ADMIN_EMAILS` | — | — | Comma-separated emails promoted to admin on startup |
| `SMTP_HOST` | — | — | Email host |
| `SMTP_PORT` | — | `587` | Email port |
| `SMTP_USERNAME` | — | — | Email username |
| `SMTP_PASSWORD` | — | — | Email password / app password |
| `SMTP_FROM_EMAIL` | — | — | Sender address |
| `SMTP_FROM_NAME` | — | `Roamify` | Sender display name |

*Either `DATABASE_URL` or the individual `DB_*` params are required.

---

## Health check

```
GET /health
→ 200 { "status": "ok", "service": "roamify-api" }
```

---

## License

MIT
