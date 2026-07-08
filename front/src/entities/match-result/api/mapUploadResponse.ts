import type { UploadHotelDto, UploadResponseDto } from './types';
import type { HotelRecord } from '@/entities/hotel/model';
import type { MatchResult } from '@/entities/match-result/model/types';
import { getMatchReasonLabel } from '@/entities/match-result/model/matchReasons';

const toPercent = (value: number): number => Math.round(value * 100);

const mapHotel = (hotel: UploadHotelDto): HotelRecord => ({
    id: `${hotel.source}:${hotel.id}`,
    providerId: hotel.source,
    externalId: hotel.id,
    name: hotel.name,
    address: hotel.address,
    city: hotel.city,
    country: hotel.country,
    coordinates:
        hotel.latitude != null && hotel.longitude != null
            ? {
                latitude: hotel.latitude,
                longitude: hotel.longitude,
            }
            : undefined,
});

export const mapUploadResponseToMatchResult = (data: UploadResponseDto): MatchResult => {
    const groups = data.groups.map((group) => ({
        id: group.groupId,
        primaryName: group.primaryName,
        confidence: toPercent(group.confidenceScore),
        matchScore: toPercent(group.matchScore),
        providersCount: group.providersCount,
        hotelsCount: group.hotelsCount,
        hotels: group.hotels.map(mapHotel),
        reasons: group.matchReasons ? group.matchReasons.map((reason) => getMatchReasonLabel(reason)) : [],
        pairwiseMatrix: group.pairwiseMatrix,
        reasonContribution: group.featureContribution,
    }));

    return {
        metrics: {
            totalHotels: data.metrics.totalHotels,
            totalGroups: data.metrics.totalGroups,
            totalDuplicates: data.metrics.totalDuplicates,
            totalProviders: data.metrics.totalProviders,
            averageConfidence: toPercent(data.metrics.averageConfidence),
            groupSizeStats: data.metrics.groupSizeStats
        },
        pagination: {
            page: data.pagination.page,
            limit: data.pagination.limit,
            totalItems: data.pagination.totalItems,
            totalPages: data.pagination.totalPages,
            hasNext: data.pagination.hasNext,
            hasPrev: data.pagination.hasPrev,
        },
        groups,
    };
};
