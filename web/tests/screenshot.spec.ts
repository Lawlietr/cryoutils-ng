import { test, expect } from '@playwright/test'

test.describe('CryoUtils NG UI', () => {
  test('loads and shows status section', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1')).toContainText('CryoUtils NG')
    await expect(page.locator('h2', { hasText: 'System Status' })).toBeVisible()
  })

  test('shows swap settings section', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h2', { hasText: 'Swap Settings' })).toBeVisible()
    await expect(page.locator('label', { hasText: 'Swap Size' })).toBeVisible()
    await expect(page.locator('label', { hasText: 'Swappiness' })).toBeVisible()
  })

  test('shows memory settings section', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h2', { hasText: 'Memory Settings' })).toBeVisible()
    const toggles = page.locator('.memory-name')
    await expect(toggles).toHaveCount(5)
  })

  test('shows VRAM read-only section', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h2', { hasText: 'VRAM' })).toBeVisible()
  })

  test('shows preset buttons', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h2', { hasText: 'Presets' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Recommended' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Stock' })).toBeVisible()
  })

  test('screenshot at 1280x800 (Steam Deck native)', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.screenshot({ path: 'tests/screenshots/steam-deck-1280x800.png', fullPage: true })
    await expect(page.locator('h1')).toBeVisible()
  })

  test('screenshot at 3840x2160 (4K)', async ({ page }) => {
    await page.setViewportSize({ width: 3840, height: 2160 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.screenshot({ path: 'tests/screenshots/4k-3840x2160.png', fullPage: true })
    await expect(page.locator('h1')).toBeVisible()
  })
})
