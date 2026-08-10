import { expect, Page, test } from '@playwright/test';

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

const startResult = (attemptId: string) => ({
  attempt_id: attemptId,
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
});

const summary = {
  score: 100,
  outcome: 'safe',
  ending: { outcome: 'safe', title: 'Отлично', text: 'Вы сохранили контроль.' },
  missed_flags: [],
  takeaway: 'Проверяйте факт оплаты.',
  steps_total: 1,
};

async function mockApp(page: Page) {
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
        body: JSON.stringify({
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
              total_scenarios: 0,
              completed_scenarios: 0,
              passed_scenarios: 0,
              completion_percent: 0,
              passed_percent: 0,
              scenarios: [],
            },
          ],
          role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
          recommendations: [],
        }),
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
        body: JSON.stringify(startResult('attempt-' + Date.now())),
      });
    if (url.pathname.includes('/choice'))
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          feedback: { verdict: 'safe', text: 'Вы проверили оплату до действия.' },
          reaction: [],
          next_scene: null,
          finished: true,
          summary,
          revision: 1,
        }),
      });
    if (url.pathname.includes('/summary'))
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(summary),
      });
    return route.fulfill({ status: 404, body: '{}' });
  });
}

test('retry starts a fresh training attempt instead of returning to the old result', async ({
  page,
}) => {
  await mockApp(page);
  await page.goto('/');
  await page.getByRole('button', { name: 'Начать' }).click();
  await page.getByRole('button', { name: /Проверить оплату/ }).click();
  await page.getByRole('button', { name: 'Посмотреть итог' }).click();
  await expect(page.getByText('Отлично')).toBeVisible();
  await page.getByRole('button', { name: 'Пройти ещё раз' }).click();
  await expect(page.getByText('Что сделать?')).toBeVisible();
});

test('revision conflict reloads the active attempt', async ({ page }) => {
  let choices = 0;
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
        body: JSON.stringify({
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
          ],
          role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
          recommendations: [],
        }),
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
        body: JSON.stringify(startResult('attempt-recovered')),
      });
    if (url.pathname.includes('/choice')) {
      choices += 1;
      if (choices === 1)
        return route.fulfill({
          status: 409,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'revision conflict' }),
        });
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          feedback: { verdict: 'safe', text: 'Вы проверили оплату до действия.' },
          reaction: [],
          next_scene: null,
          finished: true,
          summary,
          revision: 1,
        }),
      });
    }
    if (url.pathname.includes('/summary'))
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(summary),
      });
    return route.fulfill({ status: 404, body: '{}' });
  });

  await page.goto('/training/scenario-1');
  await expect(page.getByText('Что сделать?')).toBeVisible();
  await page.getByRole('button', { name: /Проверить оплату/ }).click();
  await expect(page.getByText('Что сделать?')).toBeVisible();
  await page.getByRole('button', { name: /Проверить оплату/ }).click();
  await expect(page.getByText('Безопасное решение')).toBeVisible();
});

test('refresh resumes the active attempt with the same scene and revision', async ({ page }) => {
  let startCalls = 0;
  let choiceCalls = 0;
  const scene1 = {
    scene_id: 'scene-1',
    intro: [{ author: 'counterpart', text: 'Первый шаг.' }],
    prompt: 'Что сделать?',
    options: [{ id: 'safe', text: 'Проверить оплату' }],
  };
  const scene2 = {
    scene_id: 'scene-2',
    intro: [{ author: 'counterpart', text: 'Второй шаг.' }],
    prompt: 'Что делать дальше?',
    options: [{ id: 'done', text: 'Закончить' }],
  };

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
        body: JSON.stringify({
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
          ],
          role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
          recommendations: [],
        }),
      });
    if (url.pathname === '/v1/scenarios')
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ scenarios: [scenario] }),
      });
    if (url.pathname === '/v1/attempts' && route.request().method() === 'POST') {
      startCalls += 1;
      return route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          ...startResult('attempt-resume'),
          scene: startCalls === 1 ? scene1 : scene2,
          scenes_total: 2,
        }),
      });
    }
    if (url.pathname.includes('/choice')) {
      choiceCalls += 1;
      if (choiceCalls === 1) {
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            feedback: { verdict: 'safe', text: 'Первый шаг сохранён.' },
            reaction: [],
            next_scene: scene2,
            finished: false,
            summary: null,
            revision: 1,
          }),
        });
      }
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          feedback: { verdict: 'safe', text: 'Готово.' },
          reaction: [],
          next_scene: null,
          finished: true,
          summary,
          revision: 2,
        }),
      });
    }
    if (url.pathname.includes('/summary'))
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(summary),
      });
    return route.fulfill({ status: 404, body: '{}' });
  });

  await page.goto('/training/scenario-1');
  await expect(page.getByText('Что сделать?')).toBeVisible();
  await page.getByRole('button', { name: /Проверить оплату/ }).click();
  await page.getByRole('button', { name: 'Продолжить →' }).click();
  await expect(page.getByText('Что делать дальше?')).toBeVisible();

  await page.reload();

  await expect(page.getByText('Что делать дальше?')).toBeVisible();
});
