export type UploadResponseDto = {
    metrics: UploadMetricsDto;
    groups: UploadGroupDto[];
    pagination: UploadPaginationDto;
};

export type UploadPaginationDto = {
    page: number;
    limit: number;
    totalItems: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
};

export type UploadMetricsDto = {
    totalHotels: number;
    totalGroups: number;
    totalDuplicates: number;
    totalProviders: number;
    averageConfidence: number;
    groupSizeStats: GroupSizeStatsDto[];
};

export type GroupSizeStatsDto = {
    hotelsCount: number;
    groupsCount: number;
}

export type UploadGroupDto = {
    groupId: string;
    primaryName: string;
    matchScore: number;
    confidenceScore: number;
    providersCount: number;
    hotelsCount: number;
    matchReasons?: string[];
    hotels: UploadHotelDto[];
    pairwiseMatrix?: PairwiseMatrixItemDto[];
    featureContribution?: FeatureContributionDto;
};

export type UploadHotelDto = {
    id: string;
    source: string;
    name: string;
    address?: string;
    city?: string;
    country?: string;
    latitude?: number;
    longitude?: number;
};

export type PairwiseMatrixItemDto = {
    indexA: number;
    indexB: number;
    similarity: number;
};

export type FeatureContributionDto = {
    name: number;
    address: number;
    geo: number;
    city: number;
};
