import type {
  BatchAnswerItem,
  Group,
  Member,
  QuizStatisticsResponse,
  Song,
  TargetStat,
} from '@/types/generated';

/**
 * Calculates QuizStatisticsResponse locally from IndexedDB answer history items.
 * Mirroring the backend SQLite logic for offline / local-first usage (ADR-0007).
 */
export function calculateLocalStatistics(
  items: BatchAnswerItem[],
  groups: Group[],
  members: Member[],
  songs: Song[],
  limit = 5,
): QuizStatisticsResponse {
  if (items.length === 0) {
    return {
      total_answers: 0,
      total_correct: 0,
      accuracy_rate: 0,
      average_response_time_ms: 0,
      groups: [],
      weak_targets: [],
      extra: {},
    };
  }

  // 1. Overall summary
  const totalAnswers = items.length;
  let totalCorrect = 0;
  let totalTime = 0;

  for (const item of items) {
    if (item.is_correct) totalCorrect++;
    totalTime += item.response_time_ms;
  }

  const accuracyRate = totalAnswers > 0 ? totalCorrect / totalAnswers : 0;
  const averageResponseTimeMs =
    totalAnswers > 0 ? Math.round(totalTime / totalAnswers) : 0;

  // 2. Group statistics
  const groupMap = new Map<string, Group>(groups.map((g) => [g.id, g]));
  const groupStatsMap = new Map<
    string,
    { total: number; correct: number; time: number }
  >();

  for (const item of items) {
    const stat = groupStatsMap.get(item.group_id) || {
      total: 0,
      correct: 0,
      time: 0,
    };
    stat.total++;
    if (item.is_correct) stat.correct++;
    stat.time += item.response_time_ms;
    groupStatsMap.set(item.group_id, stat);
  }

  const groupStats = Array.from(groupStatsMap.entries())
    .map(([gid, s]) => {
      const g = groupMap.get(gid);
      return {
        group_id: gid,
        group_name: g?.name || '不明グループ',
        total_answers: s.total,
        correct_answers: s.correct,
        accuracy_rate: s.total > 0 ? s.correct / s.total : 0,
        average_response_time_ms:
          s.total > 0 ? Math.round(s.time / s.total) : 0,
      };
    })
    .sort((a, b) => b.total_answers - a.total_answers);

  // 3. Weakest targets (polymorphic: members and songs)
  const memberMap = new Map<string, Member>(members.map((m) => [m.id, m]));
  const songMap = new Map<string, Song>(songs.map((s) => [s.id, s]));

  const targetStatsMap = new Map<
    string,
    {
      targetId: string;
      type: 'member' | 'song';
      groupId: string;
      total: number;
      correct: number;
      time: number;
    }
  >();

  for (const item of items) {
    let targetKey: string | null = null;
    let targetType: 'member' | 'song' = 'member';

    if (item.target_member_id) {
      targetKey = item.target_member_id;
      targetType = 'member';
    } else if (item.target_song_id) {
      targetKey = item.target_song_id;
      targetType = 'song';
    }

    if (!targetKey) continue;

    const stat = targetStatsMap.get(targetKey) || {
      targetId: targetKey,
      type: targetType,
      groupId: item.group_id,
      total: 0,
      correct: 0,
      time: 0,
    };

    stat.total++;
    if (item.is_correct) stat.correct++;
    stat.time += item.response_time_ms;
    targetStatsMap.set(targetKey, stat);
  }

  const weakTargets: TargetStat[] = Array.from(targetStatsMap.values())
    .map((s) => {
      let name = s.targetId;
      if (s.type === 'member') {
        const m = memberMap.get(s.targetId);
        if (m) name = `${m.family_name} ${m.given_name}`;
      } else {
        const song = songMap.get(s.targetId);
        if (song) name = song.title;
      }

      return {
        target_id: s.targetId,
        target_type: s.type,
        group_id: s.groupId,
        name,
        total_answers: s.total,
        correct_answers: s.correct,
        accuracy_rate: s.total > 0 ? s.correct / s.total : 0,
        average_response_time_ms:
          s.total > 0 ? Math.round(s.time / s.total) : 0,
      };
    })
    .sort((a, b) => {
      if (a.accuracy_rate !== b.accuracy_rate) {
        return a.accuracy_rate - b.accuracy_rate; // worst accuracy first
      }
      return b.total_answers - a.total_answers; // then most answered first
    })
    .slice(0, limit);

  return {
    total_answers: totalAnswers,
    total_correct: totalCorrect,
    accuracy_rate: accuracyRate,
    average_response_time_ms: averageResponseTimeMs,
    groups: groupStats,
    weak_targets: weakTargets,
    extra: {
      is_local_computed: true,
    },
  };
}
