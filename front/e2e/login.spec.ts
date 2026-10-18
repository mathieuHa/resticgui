import { expect, test } from '@playwright/test';

import { PASSWORD } from '../playwright.config';

test('log in and out', async ({ page }) => {
  await page.goto('/ui/repositories/new');
  await expect(page).toHaveURL(/\/ui\/login$/);

  await page.getByPlaceholder('Username').fill('admin');
  await page.getByPlaceholder('Password').fill('wrong password');
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByText('wrong username or password')).toBeVisible();

  await page.getByPlaceholder('Password').fill(PASSWORD);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByRole('heading', { name: 'No repository yet' })).toBeVisible();

  // the API refuses requests without the session
  const res = await page.request.get('/api/v2/repositories', { headers: { cookie: '' } });
  expect(res.status()).toBe(401);

  await page.getByRole('button', { name: 'Log out admin' }).click();
  await expect(page).toHaveURL(/\/ui\/login$/);
});
