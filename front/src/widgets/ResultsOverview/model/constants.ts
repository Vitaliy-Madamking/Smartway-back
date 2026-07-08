export const LIMIT_OPTIONS = [20, 30, 50, 100];

export const SEARCH_DEBOUNCE_MS = 500;

export const SORT_OPTIONS = [
    {
        value: 'hotelsCount-desc',
        label: 'Количество отелей ↓',
        sortBy: 'hotelsCount',
        sortDir: 'desc' as const,
    },
    {
        value: 'hotelsCount-asc',
        label: 'Количество отелей ↑',
        sortBy: 'hotelsCount',
        sortDir: 'asc' as const,
    },
    {
        value: 'confidence-desc',
        label: 'Уверенность ↓',
        sortBy: 'confidence',
        sortDir: 'desc' as const,
    },
    {
        value: 'confidence-asc',
        label: 'Уверенность ↑',
        sortBy: 'confidence',
        sortDir: 'asc' as const,
    },
    {
        value: 'name-asc',
        label: 'Название А–Я',
        sortBy: 'name',
        sortDir: 'asc' as const,
    },
    {
        value: 'name-desc',
        label: 'Название Я–А',
        sortBy: 'name',
        sortDir: 'desc' as const,
    },
];