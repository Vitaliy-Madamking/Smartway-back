const MATCH_REASON_LABELS: Record<string, string> = {
    'Similar addresses': 'Похожие адреса',
    'Similar geo': 'Близкие координаты',
    'Similar locations': 'Похожие локации',
    'Similar names': 'Похожие названия',
};

export const getMatchReasonLabel = (reason: string): string => MATCH_REASON_LABELS[reason] ?? reason;
