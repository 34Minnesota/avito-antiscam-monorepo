import React from 'react';
import { createRoot } from 'react-dom/client';
import { StoreProvider } from '@/app/providers/StoreProvider';
import { ErrorBoundary } from '@/app/providers/ErrorBoundary';
import { App } from '@/app/App';
import '@/app/styles/index.scss';
import { installGlobalErrorReporting } from '@/shared/observability';

installGlobalErrorReporting();

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <StoreProvider>
      <ErrorBoundary>
        <App />
      </ErrorBoundary>
    </StoreProvider>
  </React.StrictMode>,
);
