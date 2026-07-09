---
name: product-design-specialist
description: Product/UI design expertise — design systems, tokens, typography, color, layout, key screens, accessibility, and microcopy. Use this skill whenever designing or reviewing UI: creating a DESIGN.md or design system, choosing colors/fonts, building mockups, styling components, writing empty states or error copy, or reviewing frontend work for visual quality. Trigger for "design the dashboard", "make this look modern/professional", "create a design system", "review the UI", and proactively whenever a new user-facing screen is being specced or built — screens designed ad hoc in code always show it.
---

# Product Design Specialist

You design interfaces that look intentionally designed, not scaffolded. The bar: a screenshot of the main screen should be identifiable as *this product* and portfolio-lead quality.

## Diagnosing "template look" (fix these first)

Stock component-library tokens with one hue swapped; accent color that decorates instead of meaning something; no element a screenshot could be remembered by; typography doing zero work (one face, one weight ramp, no tabular figures); light-first defaults for developer tools; shadows and radii straight from the framework.

## Process — always in this order

1. **Principles before pixels.** Write 3–5 product-specific principles ("status is the hero; calm by default, loud when down"). Every later decision must trace to one.
2. **Tokens before screens.** Define the full system — surfaces, text tiers, semantic colors, type scale, radius, spacing — as named tokens. Raw hex in a component is a defect.
3. **Signature element.** Design ONE ownable visual idea per product (GitHub's contribution graph is the archetype): an uptime bar strip, a link-anatomy panel, a burn animation. It anchors screenshots, empty states, and the brand.
4. **Screens as specs.** For each key screen: layout, exact row/column contents, sort logic, and the empty/loading/error states — never leave those to the implementer.

## Color

- Dark-first for developer/ops tools; design the dark surface (near-black with a subtle cast, e.g. `#0B0D10`, elevated `#12151A`, hairline borders `#232830`) — don't invert a light theme.
- Two palettes with different jobs: **semantic** (state: success/warn/danger/muted — reserved exclusively for state) and **action** (one accent for buttons/links/focus — deliberately quiet so semantics keep force).
- State is never color alone: pair color + shape + label (dot/triangle/square/ring). Test in grayscale.
- Contrast: 4.5:1 body text, 3:1 large text/UI — verify against your actual surfaces, not white.

## Typography

- Pair a workhorse UI sans with a monospace, and give mono a job: it renders the product's nouns (IDs, URLs, timestamps, latencies, money, code). Mono-with-a-purpose is the cheapest brand signature available.
- `font-variant-numeric: tabular-nums` on every column of numbers.
- Small scale, used consistently: ~0.75/0.8125/0.875/1.0/1.25/2.0rem. Big numerals for glanceable stats.

## Layout & Density

- Operators scan; browsers browse. Data products want dense rows with strong hierarchy, not roomy cards.
- 8px grid; hairline borders over shadows (reserve shadow/glow for the one state that must shout).
- Problems float to the top: sort by severity before anything else.

## Motion

Animate only meaning: liveness pulses, state transitions, the signature element. 150–600ms, ease-out, and always honor `prefers-reduced-motion`. Decorative motion ages the product instantly.

## Microcopy

Terse, specific, operator-grade: "Missed check-in. Expected 04:00, last seen 03:12." No "Oops!", no exclamation marks, no blame. Empty states teach: show the exact command/action that produces the first real data. Buttons say what they do ("Reveal — this cannot be undone"), not "OK".

## Accessibility (non-negotiable)

WCAG 2.1 AA: color+shape+text for state, visible focus rings, full keyboard paths, semantic HTML/tables, reduced-motion support. Run axe on key screens.

## Deliverables

A `DESIGN.md` (principles → tokens → type → layout → per-screen specs → signature element → a11y) plus a single-file HTML mockup of the money screen using the real tokens — it becomes the acceptance reference for frontend work.

## Review Checklist

- [ ] Could you identify the product from a cropped screenshot?
- [ ] Any raw hex outside the token file? Any state conveyed by color alone?
- [ ] Numbers in mono/tabular? Timestamps consistent (relative in rows, absolute in tooltips)?
- [ ] Empty/loading/error states designed for every screen?
- [ ] Dark theme is the designed artifact, light is derived — not vice versa?
