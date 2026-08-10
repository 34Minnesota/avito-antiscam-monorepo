import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig({
  plugins: [react()],
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  test: {
    environment: 'jsdom',
    environmentOptions: { jsdom: { url: 'http://localhost/' } },
    include: ['src/**/*.test.{ts,tsx}'],
    exclude: ['e2e/**', 'node_modules/**'],
    setupFiles: ['./src/test/setup.ts'],
    css: true,
    coverage: {
      provider: 'v8',
      // Critical-domain coverage: deterministic training state, persistence, revision and progress logic.
      include: [
        'src/features/training/model/**/*.ts',
        'src/shared/lib/progress/**/*.ts',
        'src/shared/lib/attempt/**/*.ts',
      ],
      reporter: ['text', 'html', 'lcov'],
      thresholds: { lines: 85, functions: 85, statements: 85, branches: 70 },
      exclude: ['src/test/**', '**/*.d.ts', '**/index.ts'],
    },
  },
});
