# shrt — Frontend Context

The react-expert and typescript-expert skills/agents cover all React/TS best practices.
This file covers project-specific decisions only — not general React or TypeScript knowledge.

## Structure

```
frontend/
├── app/                  # Next.js App Router pages and layouts
│   ├── (auth)/login/
│   ├── (auth)/register/
│   ├── dashboard/        # loading.tsx + error.tsx required
│   └── page.tsx          # home page
├── components/ui/        # shadcn generated — never hand-edit
├── components/app/       # project components (max ~300 lines each)
├── hooks/use-links.ts    # all TanStack Query hooks for link data
├── hooks/use-auth.ts     # auth state
├── lib/api.ts            # typed API client — the only place fetch is called
├── lib/auth.ts           # token storage (in-memory access, httpOnly refresh)
├── providers/            # QueryProvider, ThemeProvider wrappers
├── types/api.ts          # all shared API types — no inline types in components
└── mocks/                # MSW handlers for component tests
```

## Libraries

| Purpose | Library |
|---------|---------|
| UI components | `shadcn/ui` New York style, Zinc base |
| Fonts | `Geist Sans` (body), `Geist Mono` (slugs/URLs) |
| Icons | `Lucide React` only |
| Server state | `@tanstack/react-query` v5 |
| Forms | `react-hook-form` + `zod` |
| Toasts | `Sonner` |
| Dark mode | `next-themes` |
| E2E tests | `Playwright` |
| API mocking | `msw` (Mock Service Worker) |
| Pre-commit | `Husky` + `lint-staged` |

## Key decisions

- **TanStack Query for all remote data.** No `useState + useEffect + fetch` for server data. Query hooks live in `hooks/use-links.ts` with a `linkKeys` factory for cache key management.
- **Access token in memory only.** Never `localStorage`. Lost on page refresh — recovered silently via refresh token.
- **Refresh token in httpOnly cookie** set by Next.js API route at `/api/refresh`. Never accessible to JS.
- **All backend calls through `lib/api.ts`.** Never call fetch directly in a component or hook.
- **All shared types in `types/api.ts`.** No `any`. No inline type definitions in components.
- **Forms validate on blur/submit**, not on every keystroke.
- **shadcn/ui is the only component library.** No Mantine, Radix direct, Ant Design, or other libraries.
- **Semantic Tailwind only.** Use `bg-background`, `text-foreground`, `text-primary` — never `bg-white` or `text-violet-600` (breaks dark mode).

## Design system

- Primary accent: violet-600 (CSS variable `--primary`)
- Slugs and short URLs: `font-mono text-sm text-primary`
- Status badges: active → green, expires-soon → amber, expired → muted
- Max content width: `max-w-4xl mx-auto px-4`
- Destructive actions always use `AlertDialog`, never plain `Dialog`
- Icon-only buttons always have `aria-label`

## Running locally

```bash
npm run dev
npm run lint
npx tsc --noEmit
```

## CI must pass

```bash
npx tsc --noEmit
next lint
next build
```
