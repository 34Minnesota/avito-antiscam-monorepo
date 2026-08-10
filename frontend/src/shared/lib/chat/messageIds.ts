import { Message } from '@/shared/api/contracts';

export const withStableMessageIds = (messages: Message[], scope: string): Message[] => {
  const seen = new Map<string, number>();
  return messages.map((message) => {
    if (message.id) return message;
    const signature = [
      message.author,
      message.text,
      message.attachment?.type ?? '',
      message.attachment?.title ?? '',
    ].join('|');
    const occurrence = seen.get(signature) ?? 0;
    seen.set(signature, occurrence + 1);
    return { ...message, id: `${scope}:${occurrence}:${signature}` };
  });
};
