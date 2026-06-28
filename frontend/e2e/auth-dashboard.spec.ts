import { test, expect } from "@playwright/test"

import { TEST_PASSWORD, uniqueEmail } from "./helpers"

// Critical path: a user registers, lands authenticated, creates a link, edits
// it, and deletes it from the dashboard.
test("register, create, edit, and delete a link", async ({ page }) => {
  const email = uniqueEmail()

  // Register (auto-logs in and redirects to the dashboard).
  await page.goto("/register")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill(TEST_PASSWORD)
  await page.getByRole("button", { name: "Create account" }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByRole("heading", { name: "Your links" })).toBeVisible()

  // New account starts with an empty table.
  await expect(page.getByText("You haven't created any links yet.")).toBeVisible()

  // Create a link from the home page while authenticated.
  const dest = "https://www.example.com/dashboard-e2e"
  await page.goto("/")
  await page.getByLabel("URL to shorten").fill(dest)
  await page.getByRole("button", { name: "Shorten" }).click()
  await expect(page.getByRole("link", { name: /https?:\/\/.+\// })).toBeVisible()

  // Back on the dashboard the link appears.
  await page.goto("/dashboard")
  const row = page.getByRole("row").filter({ hasText: "example.com/dashboard-e2e" })
  await expect(row).toBeVisible()

  // Edit the destination via the row actions menu.
  await row.getByRole("button", { name: /Actions for/ }).click()
  await page.getByRole("menuitem", { name: "Edit" }).click()
  const newDest = "https://www.example.com/edited-e2e"
  await page.getByLabel("Destination URL").fill(newDest)
  await page.getByRole("button", { name: "Save changes" }).click()
  await expect(page.getByText("Link updated")).toBeVisible()
  await expect(
    page.getByRole("row").filter({ hasText: "example.com/edited-e2e" }),
  ).toBeVisible()

  // Delete it (with the AlertDialog confirmation).
  const editedRow = page.getByRole("row").filter({ hasText: "example.com/edited-e2e" })
  await editedRow.getByRole("button", { name: /Actions for/ }).click()
  await page.getByRole("menuitem", { name: "Delete" }).click()
  await page.getByRole("button", { name: "Delete", exact: true }).click()
  await expect(page.getByText("Link deleted")).toBeVisible()
  await expect(page.getByText("You haven't created any links yet.")).toBeVisible()
})

// Security: an unauthenticated visit to the dashboard redirects to login.
test("unauthenticated dashboard redirects to login", async ({ page }) => {
  await page.goto("/dashboard")
  await expect(page).toHaveURL(/\/login$/)
})

// A registered user can log out and back in.
test("login with existing credentials", async ({ page, request }) => {
  const email = uniqueEmail("login")

  // Seed the account via the same UI register flow, then log out.
  await page.goto("/register")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill(TEST_PASSWORD)
  await page.getByRole("button", { name: "Create account" }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  await page.getByRole("button", { name: "Log out" }).click()

  // Log back in.
  await page.goto("/login")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill(TEST_PASSWORD)
  await page.getByRole("button", { name: "Log in" }).click()
  await expect(page).toHaveURL(/\/dashboard$/)

  // Wrong password is rejected with an inline error.
  await page.goto("/login")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill("totally-wrong-pw")
  await page.getByRole("button", { name: "Log in" }).click()
  await expect(page.getByRole("alert")).toBeVisible()
  void request // request fixture reserved for future API assertions
})
