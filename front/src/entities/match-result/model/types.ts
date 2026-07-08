import type { HotelRecord } from '@/entities/hotel/model';
import type { GroupSizeStatsDto } from '@/entities/match-result/api/types';

export const MATCHING_ALGORITHMS = [
    'universal',
    'jaro-winkler',
    'jaro',
    'levenshtein',
    'damerau-levenshtein',
    'soundex',
    'ngram',
] as const;

export type MatchingAlgorithm = (typeof MATCHING_ALGORITHMS)[number];

export type SortDirection = 'asc' | 'desc';

export type MatchingRequest = {
    file: File;
    threshold: number;
    algorithm?: MatchingAlgorithm;
    page: number;
    limit: number;
    search: string;
    sortBy?: string;
    sortDir?: SortDirection;
};

export type MatchGroup = {
    id: string;
    primaryName: string;
    confidence: number;
    matchScore: number;
    hotelsCount: number;
    providersCount: number;
    hotels: HotelRecord[];
    reasons?: string[];
    pairwiseMatrix?: PairwiseMatrixItem[];
    reasonContribution?: ReasonContribution;
};

export type PairwiseMatrixItem = {
    indexA: number;
    indexB: number;
    similarity: number;
};

export type ReasonContribution = {
    name: number;
    address: number;
    geo: number;
    city: number;
};

export type MetricsModel = {
    totalHotels: number,
    totalGroups: number,
    totalDuplicates: number,
    totalProviders: number,
    averageConfidence: number
    groupSizeStats: GroupSizeStatsDto[];
}

export type MatchPaginationModel = {
    page: number;
    limit: number;
    totalItems: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
};

export type MatchResult = {
    metrics: MetricsModel;
    pagination: MatchPaginationModel;
    groups: MatchGroup[];
};
