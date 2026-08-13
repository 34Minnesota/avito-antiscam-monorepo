import { useEffect, useRef } from 'react';
import { Message } from '@/shared/api/contracts';
import { Avatar } from '@/shared/ui/Avatar';
import { classNames } from '@/shared/lib/classNames';
import cls from './TrainingChat.module.scss';

interface Props {
  messages: Message[];
  counterpartName: string;
  isCounterpartTyping?: boolean;
}

export const TrainingChat = ({ messages, counterpartName, isCounterpartTyping = false }: Props) => {
  const chatRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const chat = chatRef.current;
    if (chat) chat.scrollTop = chat.scrollHeight;
  }, [isCounterpartTyping, messages]);

  return (
    <div ref={chatRef} className={cls.Chat} role="log" aria-live="polite" aria-label="История чата">
      {messages.map((message, index) => {
        const mine = message.author === 'user';
        const system = message.author === 'system';
        return (
          <div
            key={
              message.id ??
              `${message.author}-${message.text}-${message.attachment?.title ?? ''}-${index}`
            }
            className={classNames(cls.row, mine && cls.mine, system && cls.system)}
          >
            {!mine && !system && <Avatar name={counterpartName} size="s" />}
            <div className={cls.message}>
              <div className={cls.author}>{system ? 'Система' : mine ? 'Вы' : counterpartName}</div>
              <div className={cls.bubble}>
                {message.text}
                {message.attachment && (
                  <div className={cls.attachment}>
                    <span>▧</span>
                    <div>
                      <b>{message.attachment.title}</b>
                      <small>Вложение в переписке</small>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        );
      })}
      {isCounterpartTyping && (
        <div className={cls.row} role="status" aria-label={`${counterpartName} печатает`}>
          <Avatar name={counterpartName} size="s" />
          <div className={cls.message}>
            <div className={cls.author}>{counterpartName}</div>
            <div className={classNames(cls.bubble, cls.typing)} aria-hidden="true">
              <span />
              <span />
              <span />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
