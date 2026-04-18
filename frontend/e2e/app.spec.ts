import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test('should redirect to login when not authenticated', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should show login form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('input[placeholder*="用户名"], input[type="text"]').first()).toBeVisible()
    await expect(page.locator('input[type="password"]').first()).toBeVisible()
  })
})

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authenticated state
    await page.goto('/login')
    await page.evaluate(() => {
      localStorage.setItem('token', 'fake-jwt-token-for-e2e')
    })
    await page.goto('/')
  })

  test('should navigate to memories page', async ({ page }) => {
    await page.click('a[href="/memories"], [data-nav="memories"]')
    await expect(page).toHaveURL(/\/memories/)
  })

  test('should navigate to knowledge page', async ({ page }) => {
    await page.click('a[href="/knowledge"], [data-nav="knowledge"]')
    await expect(page).toHaveURL(/\/knowledge/)
  })

  test('should navigate to settings page', async ({ page }) => {
    await page.click('a[href="/settings"], [data-nav="settings"]')
    await expect(page).toHaveURL(/\/settings/)
  })
})

test.describe('Chat Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.evaluate(() => {
      localStorage.setItem('token', 'fake-jwt-token-for-e2e')
    })
    await page.goto('/')
  })

  test('should display chat interface', async ({ page }) => {
    // Chat view should have a message input
    const chatInput = page.locator('textarea, input[type="text"]').last()
    await expect(chatInput).toBeVisible()
  })
})

test.describe('Memories Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.evaluate(() => {
      localStorage.setItem('token', 'fake-jwt-token-for-e2e')
    })
    await page.goto('/memories')
  })

  test('should display memories page title', async ({ page }) => {
    await expect(page.locator('text=记忆, text=Memories').first()).toBeVisible()
  })
})

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.evaluate(() => {
      localStorage.setItem('token', 'fake-jwt-token-for-e2e')
    })
    await page.goto('/settings')
  })

  test('should display settings tabs', async ({ page }) => {
    // Should have profile and password tabs
    await expect(page.locator('text=密码, text=Password, text=修改密码').first()).toBeVisible()
  })
})
