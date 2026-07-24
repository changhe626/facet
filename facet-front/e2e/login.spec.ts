import { test, expect } from '@playwright/test';

test.describe('Login', () => {
  test('shows login page with all form fields', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByText('Facet')).toBeVisible();
    await expect(page.getByText('Feature Flag Management Platform')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();
    await expect(page.getByLabel('Captcha')).toBeVisible();
  });

  test('redirects authenticated user to overview', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel('Email').fill('admin@facet.com');
    await page.getByLabel('Password').fill('admin123');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/overview', { timeout: 10000 });
    await expect(page.getByText('Overview')).toBeVisible();
  });
});
