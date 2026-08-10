# AntiScam Frontend

Production-style React frontend для хакатонного тренажёра безопасных сделок.

## Команды

```bash
npm ci
npm run typecheck
npm run lint
npm run format:check
npm test
npm run test:coverage
npm run build
npm run test:e2e
```

Для разработки:

```bash
npm run start
```

## Архитектура

- `app` — routing, providers, global styles;
- `pages` — route-level composition;
- `widgets` — крупные UI-секции;
- `features` — пользовательские действия и training state;
- `entities` — User, Scenario, Progress;
- `shared` — UI-kit, API base, storage и pure helpers.

`TrainingPage` намеренно оставлен тонким: серверный training flow и его восстановление находятся в `features/training/model/useTrainingSession.ts`.

## Надёжность training flow

Backend является source of truth для активной попытки. Frontend дополнительно сохраняет transcript, номер шага и revision одного активного attempt в `sessionStorage`, чтобы F5 не превращал активную попытку в новую. Backend остаётся source of truth.

При `409 Conflict` frontend заново запрашивает активную попытку через `POST /v1/attempts`, после чего восстанавливает текущую серверную сцену и локальный transcript.

## Security trade-off

Текущий backend использует `X-Session-ID`, поэтому frontend хранит session identifier в `localStorage`. Для production с изменяемым backend предпочтительнее HttpOnly/Secure/SameSite cookie. В рамках хакатона API-контракт backend не меняется.

## Testing

Unit/component tests находятся рядом с критической логикой и UI. Playwright покрывает основной путь:

`dashboard → scenario → decision → feedback → result`.

Тесты не требуют реального backend: E2E stub-ит API ответы, а production Docker использует реальный backend без его изменений.

## Reproducible install

`package-lock.json` входит в репозиторий. Для CI и production используйте только `npm ci`; зависимости не плавают между сборками. Версия Node зафиксирована в `.nvmrc`.

## Docker

Из корня репозитория:

```bash
cp .env.example .env
docker compose up --build
```

Frontend: `http://localhost:3000`
Backend API: `http://localhost:8080`

## CI

GitHub Actions проверяет typecheck, lint, форматирование, тесты и production build. Playwright E2E запускается отдельно локально или в CI job с установленными браузерами.

### Архитектурные границы

Frontend использует FSD-слои `shared → entities → features → widgets → pages → app`. ESLint дополнительно запрещает обратные зависимости между слоями, чтобы нарушение границ было ошибкой quality gate.

### История прогресса

Dashboard показывает последние завершённые попытки выбранной роли: результат, сценарий, исход и дату. На ResultPage дополнительно показывается изменение относительно предыдущего результата, если backend его предоставляет.

## Progress model

Frontend renders the progress returned by `GET /v1/progress` without reimplementing the
server's XP rules.

The dashboard exposes:

- server-calculated XP, level and level progress;
- earned and locked achievements;
- personalized recommendations, including active attempts;
- recent attempt history and the first safe attempt;
- initial vs latest score and trend;
- best score and attempt count per scenario;
- role progress and comparison;
- category-based skill summaries for the selected role.

Scenario cards also expose server-provided category, difficulty and personal statistics.

## Observability

React ErrorBoundary и глобальные `error`/`unhandledrejection` события отправляются в единый `reportError`. По умолчанию ошибки остаются в console; для внешнего collector можно задать `VITE_ERROR_REPORT_URL`.
