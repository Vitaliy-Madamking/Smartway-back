export type MatchingRequestPatch = {
    page?: number;
    limit?: number;
    search?: string;
    sortBy?: string;
    sortDir?: 'asc' | 'desc';
};