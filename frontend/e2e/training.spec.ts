import { test, expect } from '@playwright/test';

const scenario = {
  id: 'scenario-1',
  slug: 'safe-deal',
  role: 'seller',
  category: 'Оплата',
  difficulty: 3,
  title: 'Срочная предоплата',
  description: 'Проверьте безопасную оплату.',
  stats: null,
};

const progress = {
  total_scenarios: 1,
  completed_scenarios: 0,
  passed_scenarios: 0,
  completion_percent: 0,
  passed_percent: 0,
  roles: [
    {
      role: 'seller',
      total_scenarios: 1,
      completed_scenarios: 0,
      passed_scenarios: 0,
      completion_percent: 0,
      passed_percent: 0,
      scenarios: [
        {
          scenario_slug: 'safe-deal',
          title: scenario.title,
          completed: false,
          passed: false,
          attempts_count: 0,
          active_attempt_id: null,
          recent_attempts: [],
        },
      ],
    },
    {
      role: 'buyer',
      total_scenarios: 1,
      completed_scenarios: 0,
      passed_scenarios: 0,
      completion_percent: 0,
      passed_percent: 0,
      scenarios: [],
    },
  ],
  role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
  recommendations: [],
};

test('user can complete a training and reach the result', async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('antiscam_session_id', 'test-session'));
  await page.route('**/v1/**', async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname === '/v1/users/me')
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'u1',
          nickname: 'Tester',
          email: 'test@example.com',
          created_at: '2026-01-01',
        }),
      });
    if (url.pathname === '/v1/progress')
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(progress),
      });
    if (url.pathname === '/v1/scenarios')
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ scenarios: [scenario] }),
      });
    if (url.pathname === '/v1/attempts' && route.request().method() === 'POST')
      return route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          attempt_id: 'attempt-1',
          listing: { title: 'Велосипед', price: 100000, location: 'Москва', image: '' },
          counterpart: { name: 'Алексей', rating: 4.9, reviews: 18, registered: '2024' },
          role: 'seller',
          title: scenario.title,
          scene: {
            scene_id: 'scene-1',
            intro: [{ author: 'counterpart', text: 'Давайте быстро решим вопрос.' }],
            prompt: 'Что сделать?',
            options: [{ id: 'safe', text: 'Проверить оплату' }],
          },
          scenes_total: 1,
        }),
      });
    if (url.pathname === '/v1/attempts/attempt-1/choice')
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          feedback: { verdict: 'safe', text: 'Вы проверили оплату до действия.' },
          reaction: [
            { author: 'user', text: 'Проверю оплату перед действием.' },
            { author: 'counterpart', text: 'Хорошо.' },
          ],
          next_scene: null,
          finished: true,
          summary: {
            score: 100,
            outcome: 'safe',
            ending: { outcome: 'safe', title: 'Отлично', text: 'Вы сохранили контроль.' },
            missed_flags: [],
            takeaway: 'Проверяйте факт оплаты.',
            steps_total: 1,
          },
          revision: 1,
        }),
      });
    if (url.pathname === '/v1/attempts/attempt-1/summary')
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          score: 100,
          outcome: 'safe',
          ending: { outcome: 'safe', title: 'Отлично', text: 'Вы сохранили контроль.' },
          missed_flags: [],
          takeaway: 'Проверяйте факт оплаты.',
          steps_total: 1,
        }),
      });
    return route.fulfill({ status: 404, body: '{}' });
  });

  await page.goto('/');
  await expect(page.getByText('Срочная предоплата')).toBeVisible();
  await page.getByRole('button', { name: 'Начать' }).click();
  await expect(page.getByText('Что сделать?')).toBeVisible();
  await page.getByRole('button', { name: /Проверить оплату/ }).click();
  await expect(page.getByRole('status', { name: /печатает/ })).toBeVisible();
  await expect(page.getByText('Проверю оплату перед действием.')).toBeVisible();
  await expect(page.getByRole('button', { name: /Проверить оплату/ })).not.toBeVisible();
  await expect(page.getByText('Безопасное решение')).toBeVisible();
  await expect(page.getByRole('status', { name: /печатает/ })).not.toBeVisible();
  await expect(page.getByText('Хорошо.')).toBeVisible();
  await page.getByRole('button', { name: 'Посмотреть итог' }).click();
  await expect(page.getByText('Отлично')).toBeVisible();
});
