import { test, expect } from "@playwright/test"

import { BACKEND_URL, createLinkViaApi } from "./helpers"

// Critical path: an expired short link returns 410 Gone rather than redirecting.
test("expired link returns 410", async ({ request }) => {
  // Seed a link whose expiry is already in the past.
  const pastIso = new Date(Date.now() - 60 * 60 * 1000).toISOString()
  const { slug } = await createLinkViaApi(request, "https://example.com/expired", {
    expiresAt: pastIso,
  })

  const res = await request.get(`${BACKEND_URL}/${slug}`, { maxRedirects: 0 })
  expect(res.status()).toBe(410)
})

// The branded 410 page renders for users routed to it.
test("branded 410 page renders", async ({ page }) => {
  await page.goto("/gone")
  await expect(page.getByText("410")).toBeVisible()
  await expect(page.getByRole("heading", { name: /expired/i })).toBeVisible()
})

// The branded 404 page renders for unknown app routes.
test("branded 404 page renders", async ({ page }) => {
  const res = await page.goto("/this-route-does-not-exist")
  expect(res?.status()).toBe(404)
  await expect(page.getByText("404")).toBeVisible()
})
