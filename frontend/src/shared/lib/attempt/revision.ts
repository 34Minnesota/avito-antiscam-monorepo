export const parseRevision = (raw: string | number | null | undefined) => {
  const value = raw == null ? 0 : Number(raw);
  return Number.isFinite(value) && value >= 0 ? value : 0;
};
