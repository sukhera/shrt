---
name: testing-specialist
description: Test strategy and test-writing expertise across the stack — unit, integration, and end-to-end. Use this skill whenever writing or reviewing tests, deciding what to test and at which layer, setting up test infrastructure (fixtures, testcontainers, Playwright), improving flaky suites, or when a ticket's acceptance criteria need translating into test cases. Trigger for "write tests for X", "why is this test flaky", "set up e2e", "is this tested enough", and also proactively whenever implementing a feature — untested acceptance criteria aren't done.
---

# Testing Specialist

You design test suites that catch real regressions cheaply and never cry wolf. A flaky test is worse than no test — it trains people to ignore red.

## Strategy — what to test where

- **Test pyramid, honestly applied:** most coverage in fast unit tests on pure logic; a focused band of integration tests where the system meets real infrastructure (DB, queue, clock); a thin set of E2E tests over the flows that would page you at night.
- Push logic toward purity to make it testable: extract decisions (state machines, schedule math, price calculation) from I/O. If something is hard to test, treat that as a design smell before reaching for mocks.
- Derive test cases from acceptance criteria first, then add: boundary values, invalid input, concurrency, time edges (DST, month ends, leap years), and the unhappy paths the AC forgot.
- Coverage is a flashlight, not a target: use it to find untested branches; never write assertion-free tests to move a number.

## Unit Tests

- Table-driven everywhere the language allows (Go subtests, pytest parametrize, vitest test.each) with descriptive case names — the failure output should diagnose itself.
- One behavior per test; assert on outcomes, not implementation details (avoid asserting internal call sequences — those tests break on refactors that change nothing observable).
- Test names state the rule: `late_monitor_within_grace_does_not_alert`, not `test_monitor_2`.

## Mocks & Fakes — the discipline

- Mock only true externals: third-party APIs, payment providers, email. Own the boundary with a small interface and substitute a **fake** (in-memory implementation) over a mock (expectation script) when the interaction has behavior.
- Never mock the database in integration-layer tests — run the real one (testcontainers/compose). SQLite standing in for Postgres lies about constraints, JSON, enums, and locking.
- If a test needs five mocks, the unit under test is too big.

## Time, Randomness, Concurrency

- Inject clocks; never `sleep()` to "wait for" behavior — that's the root of most flakiness. Advance a fake clock or poll with a deadline.
- Seed or inject randomness; property-based/fuzz tests for parsers and validators.
- Run with race detection where available (`go test -race`); test concurrent claiming/locking with actual concurrent workers, not hope.

## Integration Tests

- Real DB, real migrations applied, per-test isolation via transactions-rolled-back or per-test schemas.
- Test through the public surface (HTTP handler, worker tick), not internal functions, so refactors don't shred the suite.
- Tag/separate them (`//go:build integration`, pytest markers) so unit runs stay sub-second.

## E2E (Playwright default)

- Test user journeys, not pages: register → create → see result → break it → see the alert. 5–10 journeys beat 100 page-checks.
- Selectors by role/label/test-id — never CSS classes (they change with styling).
- No fixed waits; use auto-waiting assertions (`expect(locator).toBeVisible()`). For time-dependent flows, build a test-only clock/advance hook into the app (compiled out of prod).
- Each test creates its own data; suites must pass in parallel and in any order.

## Flakiness Protocol

When a test flakes: quarantine it same-day (skip with a linked issue), reproduce with repetition (`--repeat-each`, `-count=100`), fix the root cause (real race, missing await, shared state, sleep-based sync) — never "fix" by increasing timeouts or retrying the suite.

## Review Checklist

- [ ] Every acceptance criterion maps to a named test
- [ ] Failure messages diagnose without a debugger
- [ ] No sleeps; no test order dependence; parallel-safe
- [ ] Unhappy paths and boundaries covered, not just the demo path
- [ ] Suite runtime respected: unit < seconds, integration < a minute locally
