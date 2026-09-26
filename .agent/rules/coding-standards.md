# Coding Standards & Conventions

> **Directive**: This rule outlines conventions for Go, TypeScript, React, and SQL within this repository.

---

## 1. Go (Golang) Conventions

- **Go Version**: 1.26+ with multi-module workspace (`go.work`).
- **Error Handling**:
  - Always wrap errors with context: `fmt.Errorf("failed to fetch deployment %s: %w", id, err)`.
  - Use custom domain errors in `shared/errors` for business logic (e.g. `ErrNotFound`, `ErrUnauthorized`, `ErrConflict`).
- **Logging**:
  - Use structured JSON logging (`shared/logger` based on `log/slog`).
  - Never log raw Discord tokens, OAuth secrets, or customer credit card / payment details.
- **Context Propagation**:
  - Pass `ctx context.Context` as the first argument in all HTTP handlers, repository methods, and external RPC calls.
  - Respect request cancellation and timeouts.
- **HTTP Routing**:
  - Use standard Go `net/http` or standard mux with idiomatic REST paths: `/api/v1/...`.
  - Always respond with standardized JSON responses and proper HTTP status codes.

---

## 2. TypeScript & Node.js Conventions (Manager Bot)

- **TypeScript**: Strict mode enabled (`"strict": true`). No `any` types unless strictly necessary for dynamic Discord component payloads.
- **Discord.js v14**:
  - Slash commands defined with `SlashCommandBuilder`.
  - Interactions deferred with `await interaction.deferReply({ ephemeral: true })` if background REST calls take > 2.5 seconds.
  - Multi-bot selection components built using `StringSelectMenuBuilder` and `ActionRowBuilder<StringSelectMenuBuilder>`.
  - Component collectors use `ComponentType.StringSelect` with a 60-second timeout.
- **REST Client**:
  - Manager Bot communicates with microservices using typed helper classes (`src/api/client.ts`).
  - Always handle network timeouts, non-200 HTTP status codes, and JSON parse failures gracefully with user-friendly Discord error messages.

---

## 3. Next.js 16 & React 19 Conventions (Frontend)

- **Next.js App Router**: Routes live in `src/app/`.
- **Turbopack**: Ensure compatibility with Turbopack (`next build` / `next dev --turbo`).
- **Styling**: Tailwind CSS with dark theme aesthetic (slate/zinc dark palette, indigo/violet accents).
- **Icons**: Lucide React (`lucide-react`).
- **Component Design**:
  - Use `"use client"` directive on interactive dashboard/checkout components.
  - Extract reusable logic into `src/lib/api.ts`.
  - State management uses React hooks (`useState`, `useEffect`, `useCallback`).

---

## 4. SQL & Database Conventions

- All migrations and seed scripts reside in `scripts/init-databases.sql`.
- Table names use snake_case plural (e.g., `subscriptions`, `deployments`, `promo_codes`).
- Primary keys use `VARCHAR(64)` with descriptive prefix IDs (e.g., `sub_...`, `dep_...`, `usr_...`).
- Always add indexes on foreign keys, lookups (`guild_id`, `user_id`, `status`), and temporal queries (`checked_at DESC`).
