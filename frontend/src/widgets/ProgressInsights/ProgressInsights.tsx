import { Progress, Role, ScenarioCard } from '@/shared/api/contracts';
import {
  getExperience,
  getProgressDynamics,
  getSkills,
  getTrendLabel,
} from '@/shared/lib/progress/insights';
import { Card } from '@/shared/ui/Card';
import cls from './ProgressInsights.module.scss';

interface Props {
  progress: Progress;
  role: Role;
  scenarios: ScenarioCard[];
}

export const ProgressInsights = ({ progress, role, scenarios }: Props) => {
  const experience = getExperience(progress, role);
  const skills = getSkills(progress, role, scenarios);
  const dynamics = getProgressDynamics(progress, role);
  const achievements = progress.experience.achievements ?? [];
  const earnedAchievements = achievements.filter((item) => item.earned);

  return (
    <section className={cls.section} aria-label="Опыт, навыки и динамика">
      <div className={cls.heading}>
        <div>
          <span>РОСТ НАВЫКА</span>
          <h2>Ваш прогресс в обучении</h2>
        </div>
      </div>

      <div className={cls.grid}>
        <Card className={cls.levelCard}>
          <div className={cls.cardTop}>
            <span>УРОВЕНЬ</span>
            <strong>{experience.level}</strong>
          </div>
          <div className={cls.xpLine}>
            <b>{experience.xp} XP</b>
            <small>
              {experience.currentLevelXp} / {experience.nextLevelXp} XP до следующего уровня
            </small>
          </div>
          <div
            className={cls.track}
            role="progressbar"
            aria-label="Прогресс до следующего уровня"
            aria-valuenow={experience.progressPercent}
            aria-valuemin={0}
            aria-valuemax={100}
          >
            <span style={{ width: `${experience.progressPercent}%` }} />
          </div>
          <p>
            {experience.completedScenarios} завершённых сценария · {experience.safeScenarios}{' '}
            безопасных в роли {role === 'seller' ? 'продавца' : 'покупателя'}
          </p>
        </Card>

        <Card className={cls.dynamicCard}>
          <span>ДИНАМИКА РЕЗУЛЬТАТА</span>
          {dynamics.trend === 'none' ? (
            <>
              <strong>—</strong>
              <p>
                Пройдите сценарий повторно, чтобы увидеть динамику первого и последнего результата.
              </p>
            </>
          ) : (
            <>
              <div className={cls.delta}>
                <strong>
                  {dynamics.delta > 0 ? '+' : ''}
                  {dynamics.delta}
                </strong>
                <small>баллов</small>
              </div>
              <p>
                {getTrendLabel(dynamics.trend)} · {dynamics.initial}% → {dynamics.latest}% ·{' '}
                {dynamics.scenariosTracked}{' '}
                {dynamics.scenariosTracked === 1 ? 'сценарий' : 'сценария'}
              </p>
            </>
          )}
        </Card>

        <Card className={cls.skillsCard}>
          <div className={cls.cardTop}>
            <span>НАВЫКИ</span>
            <small>{skills.length ? `${skills.length} направлений` : 'нет данных'}</small>
          </div>
          {skills.length ? (
            <div className={cls.skills}>
              {skills.slice(0, 4).map((skill) => (
                <div className={cls.skill} key={skill.name}>
                  <div>
                    <span>{skill.name}</span>
                    <b>{skill.score}%</b>
                  </div>
                  <div className={cls.skillTrack}>
                    <span style={{ width: `${skill.score}%` }} />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className={cls.empty}>Навыки появятся после первого результата.</p>
          )}
        </Card>
      </div>

      <Card className={cls.achievements}>
        <div className={cls.achievementHead}>
          <div>
            <span>ДОСТИЖЕНИЯ</span>
            <h3>Ваши шаги вперёд</h3>
          </div>
          <b>
            {earnedAchievements.length} / {achievements.length}
          </b>
        </div>
        <div className={cls.achievementGrid}>
          {achievements.map((achievement) => (
            <div
              className={`${cls.achievement} ${achievement.earned ? cls.earned : ''}`}
              key={achievement.code}
            >
              <span aria-hidden="true">{achievement.earned ? '✓' : '○'}</span>
              <div>
                <strong>{achievement.title}</strong>
                <p>{achievement.description}</p>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </section>
  );
};
