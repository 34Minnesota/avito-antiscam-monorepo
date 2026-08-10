const REPORT_URL = import.meta.env.VITE_ERROR_REPORT_URL?.trim();

export const reportError = (error: unknown, context: Record<string, unknown> = {}) => {
  const payload = {
    message: error instanceof Error ? error.message : String(error),
    stack: error instanceof Error ? error.stack : undefined,
    context,
    url: window.location.href,
    timestamp: new Date().toISOString(),
  };

  console.error('[AntiScam]', payload);

  if (!REPORT_URL) return;
  try {
    void fetch(REPORT_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      keepalive: true,
    });
  } catch {
    // Observability must never break the application.
  }
};

export const installGlobalErrorReporting = () => {
  const onError = (event: ErrorEvent) =>
    reportError(event.error ?? event.message, { source: 'window' });
  const onRejection = (event: PromiseRejectionEvent) =>
    reportError(event.reason, { source: 'unhandledrejection' });
  window.addEventListener('error', onError);
  window.addEventListener('unhandledrejection', onRejection);
  return () => {
    window.removeEventListener('error', onError);
    window.removeEventListener('unhandledrejection', onRejection);
  };
};
