import { Message } from '@/shared/api/contracts';
import { Avatar } from '@/shared/ui/Avatar';
import { classNames } from '@/shared/lib/classNames';
import cls from './TrainingChat.module.scss';

interface Props {
  messages: Message[];
  counterpartName: string;
}

export const TrainingChat = ({ messages, counterpartName }: Props) => (
  <div className={cls.Chat} role="log" aria-live="polite" aria-label="История чата">
    {messages.map((message, index) => {
      const mine = message.author === 'user';
      const system = message.author === 'system';
      return (
        <div
          key={`${message.author}-${index}-${message.text}`}
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
  </div>
);
