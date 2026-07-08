# shrt — Design Modernization Spec

**Version:** 1.0 · **Date:** 2026-07-01 · Destination: `sukhera/shrt` → `docs/REDESIGN.md`
Part of the portfolio design system established in ping's `DESIGN.md` — read that first for principles and full token rationale.

---

## 1. Current state (review)

shrt works, but visually it's the shadcn default with violet swapped in (`--primary: 262 83% 58%`, radius 0.5rem, zinc neutrals, stock spacing). Specific issues:

1. **Template identity** — nothing distinguishes it from any shadcn scaffold.
2. **The product's nouns aren't typographic.** Slugs, destination URLs, and click counts are the content, yet they're set in the same UI sans as button labels. No tabular figures for counts.
3. **The dashboard is a flat table** — no glanceable summary, no trend information, although click data exists.
4. **The home page undersells the instant-shorten moment.** Paste → short link is the magic; the result state deserves a designed moment (large mono link, one-click copy, subtle success motion).
5. **Light-first default** with dark as an inversion, not a designed surface.
6. **Status badge is color-only** (WCAG 1.4.1 issue) and the violet accent also carries state-ish meanings in places, diluting both.

## 2. Target

Adopt the portfolio design system: **ping's neutral surface architecture + Geist pairing + mono-as-brand**, keeping shrt's **violet** as its accent (each product keeps one hue: shrt violet · scrt emerald · ping blue). Dark-first. One signature element: the **per-link activity sparkline**.

### 2.1 Tokens (drop-in for `globals.css`)

Same surfaces/text/borders as ping `DESIGN.md` §4 (`--bg #0B0D10`, `--surface #12151A`, `--surface-2 #1A1E25`, `--border #232830`, `--text #E6EAF0`, `--text-dim #8B94A3`, `--text-faint #5C6470`), with:

```css
--accent: #8B7CF6;            /* violet — refined, slightly desaturated from violet-600 */
--accent-soft: rgba(139,124,246,.12);
--ok: #2DD4A7; --warn: #F5B84B; --bad: #F4564E; --off: #5C6470;   /* link states: active/expiring/expired/disabled */
--radius: 0.375rem;
```

Rules: violet = actions/links/focus only. Link *status* uses the semantic trio + shape (dot=active, triangle=expiring soon, square=expired, ring=disabled) + label — never color alone. Typography: Geist Sans UI / Geist Mono for every slug, URL, count, and date (`tabular-nums`). Hairline borders, no default shadows.

## 3. Screen-by-screen

### 3.1 Home (anonymous shorten)

- Hero = the form. One oversized input (mono placeholder `https://…`), one violet button. Nothing else above the fold except the wordmark and a one-line pitch.
- **The result moment:** on success the form morphs into a result card — short link in large Geist Mono, copy button with ✓ feedback, QR icon (v2), "shorten another" ghost link, one expanding-ring pulse on the card border (600ms, `prefers-reduced-motion` aware). This is the screenshot moment.
- Sub-fold: three terse feature lines (mono glyphs, not icon-library icons), sign-in nudge for management.

### 3.2 Dashboard

- **Stat strip** above the table (ping pattern): Total links · Clicks 7d · Top link (name + count) — big mono numerals.
- Table rows: `[status shape] [slug · mono, violet on hover] [destination, truncated middle, dim] [clicks · mono tabular + 30-day sparkline] [created/expires · relative] [⋯]`.
- **Signature element — the sparkline:** 30-day click activity per link, hand-rolled SVG, 96×20px, violet line at 60% opacity, filled dot on today. Gives the table life and demonstrates the data shrt already collects. (Backend: needs a `clicks(link_id, day, count)` rollup — small migration + increment on redirect; see T-6.)
- Expired rows drop to `--off` treatment wholesale (like ping's paused). Search + status filter in URL params. Empty state with pulse-style glyph and inline example, never blank.

### 3.3 Link detail / edit dialog

Promote edit from a bare dialog to a two-pane sheet: left = fields (slug shown as `shrt.dev/` prefix + mono editable suffix), right = live stats (clicks total/7d/30d, sparkline large, last-clicked relative). Destructive delete stays behind a typed-confirm only when clicks > 100.

### 3.4 Auth pages

Match ping §7.6: centered card on `--bg`, wordmark above, zero marketing chrome.

## 4. Wordmark

`shrt` lowercase, Geist Mono, with the violet slash as the mark: **`s/`** (the URL-path glyph). Favicon: `s/` on `--bg`.

## 5. What does NOT change

Backend API, auth flows, data model (except the tiny clicks rollup), routing, shadcn as the component base, existing tests. This is a re-skin + two small features (sparkline data, result moment), not a rebuild.

## 6. Tickets

### SHRT-R1: Retoken to portfolio design system `size:M`
Swap `globals.css` tokens per §2.1, radius 6px, hairline borders, Geist Sans/Mono via `next/font`, dark-first default. AC: no raw hex outside `globals.css`; all existing screens render correctly in both themes; dark is default; contrast ≥ 4.5:1 for `--text-dim` on `--surface`.

### SHRT-R2: Status chip component (shape + color + label) `size:S`
Replace `status-badge.tsx` with the shared chip pattern; states active/expiring(<7d)/expired/disabled. AC: axe passes; state readable in grayscale; used everywhere status appears.

### SHRT-R3: Home hero + result moment `size:M`
§3.1 including morph animation and copy feedback. AC: paste→copy in ≤ 2 interactions; pulse respects reduced-motion; screenshot-ready.

### SHRT-R4: Dashboard stat strip + row redesign `size:M`
§3.2 minus sparkline. AC: mono tabular numerics; middle-truncated destinations; expired treatment; URL-param filters; matches updated screenshots.

### SHRT-R5: Click rollups (backend) `size:S`
`clicks_daily(link_id, day, count)` migration + upsert on redirect (async, never blocking the redirect) + `GET /api/v1/links/{id}/stats`. AC: redirect latency unchanged (bench before/after); rollup idempotent; down migration works.

### SHRT-R6: Sparkline + link detail sheet `size:M` · deps: R4, R5
§3.2 sparkline + §3.3 sheet. AC: sparkline renders 30 cells incl. zero-days; sheet keyboard-accessible; large sparkline matches small one's data.

### SHRT-R7: README screenshot refresh + wordmark `size:S` · deps: R1–R6
New dark screenshots, `s/` favicon/wordmark, README hero images swapped. AC: README renders with new shots; favicon correct in both themes.

Order: R1 → R2 → (R3 ∥ R4) → R5 → R6 → R7. Roughly 3–4 focused days total.
