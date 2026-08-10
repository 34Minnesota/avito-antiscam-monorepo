import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeAll } from 'vitest';

const createMemoryStorage = (): Storage => {
  const data = new Map<string, string>();
  return {
    get length() {
      return data.size;
    },
    clear() {
      data.clear();
    },
    getItem(key) {
      return data.get(String(key)) ?? null;
    },
    key(index) {
      return Array.from(data.keys())[index] ?? null;
    },
    removeItem(key) {
      data.delete(String(key));
    },
    setItem(key, value) {
      data.set(String(key), String(value));
    },
  };
};

beforeAll(() => {
  const installStorage = (name: 'localStorage' | 'sessionStorage') => {
    try {
      if (typeof globalThis[name] === 'undefined') {
        Object.defineProperty(globalThis, name, {
          configurable: true,
          value: createMemoryStorage(),
        });
      }
    } catch {
      Object.defineProperty(globalThis, name, { configurable: true, value: createMemoryStorage() });
    }
  };

  installStorage('localStorage');
  installStorage('sessionStorage');
});

afterEach(() => cleanup());
