import { test, expect } from "@playwright/test"

import { BACKEND_URL } from "./helpers"

// Critical path: an anonymous visitor shortens a URL on the home page, copies
// the result, and the short link redirects to the original destination.
test.describe("anonymous shorten flow", () => {
  test.use({ permissions: ["clipboard-read", "clipboard-write"] })

  test("shorten, copy, and redirect", async ({ page, context }) => {
    const target = "https://www.example.com/some/long/path?ref=e2e"

    await page.goto("/")

    await page.getByLabel("URL to shorten").fill(target)
    await page.getByRole("button", { name: "Shorten" }).click()

    // The result card shows a short URL link.
    const shortLink = page.getByRole("link", { name: /https?:\/\/.+\// })
    await expect(shortLink).toBeVisible()
    const shortUrl = await shortLink.getAttribute("href")
    expect(shortUrl).toBeTruthy()

    // Copy button copies the short URL and fires a toast.
    await page.getByRole("button", { name: "Copy short URL" }).click()
    await expect(page.getByText("Link copied")).toBeVisible()

    const clipboard = await page.evaluate(() => navigator.clipboard.readText())
    expect(clipboard).toBe(shortUrl)

    // Visiting the short URL redirects to the original destination. Use an API
    // request without following redirects so we can assert the 30x + Location.
    const res = await context.request.get(shortUrl!, { maxRedirects: 0 })
    expect([301, 302]).toContain(res.status())
    expect(res.headers()["location"]).toBe(target)
  })

  test("rejects an invalid URL", async ({ page }) => {
    await page.goto("/")
    // The native URL input blocks obviously malformed input; assert the
    // backend-backed validation path too via a value the input accepts but the
    // backend rejects (unsupported scheme).
    const input = page.getByLabel("URL to shorten")
    await input.fill("ftp://example.com")
    await page.getByRole("button", { name: "Shorten" }).click()
    await expect(page.getByRole("alert")).toBeVisible()
  })
})

// The backend redirect contract, exercised directly: unknown slug → 404.
test("unknown slug returns 404", async ({ request }) => {
  const res = await request.get(`${BACKEND_URL}/nonexistentslug`, { maxRedirects: 0 })
  expect(res.status()).toBe(404)
})
