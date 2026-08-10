export const getApiStatus = (error: unknown): number | undefined => {
  if (!error || typeof error !== 'object' || !('status' in error)) return undefined;
  const status = (error as { status?: unknown }).status;
  return typeof status === 'number' ? status : undefined;
};

export const isUnauthorized = (error: unknown) => getApiStatus(error) === 401;
export const isConflict = (error: unknown) => getApiStatus(error) === 409;

export const getApiMessage = (error: unknown, fallback: string) => {
  if (error && typeof error === 'object' && 'data' in error) {
    const data = (error as { data?: { message?: string } }).data;
    if (typeof data?.message === 'string' && data.message.trim()) return data.message;
  }
  return fallback;
};
