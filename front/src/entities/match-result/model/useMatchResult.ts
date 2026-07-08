import { useQuery } from '@tanstack/react-query';

import { MATCH_RESULT_QUERY_KEY } from '@/entities/match-result/model/storage';
import type { MatchResult } from '@/entities/match-result/model/types';

export const useMatchResult = () =>
  useQuery({
    queryKey: MATCH_RESULT_QUERY_KEY,
    queryFn: async (): Promise<MatchResult | null> => null,
    staleTime: Number.POSITIVE_INFINITY,
  });
