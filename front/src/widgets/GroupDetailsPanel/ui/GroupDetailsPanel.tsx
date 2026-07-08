import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { useMemo } from 'react';
import { Navigate, useParams } from 'react-router-dom';

import { HotelTable } from '@/entities/hotel/ui';
import { useMatchResult } from '@/entities/match-result/model';
import { ROUTES } from '@/shared/config';
import { EmptyState } from '@/shared/ui/EmptyState';
import { ScoreBar } from '@/shared/ui/ScoreBar';

import styles from './GroupDetailsPanel.module.scss';
import { BackButton } from '@/shared/ui/BackButton';
import { getPluralHotel } from '@/entities/hotel/lib';
import { PairwiseMatrix } from '@/features/PairwiseMatrix';
import { ReasonContributionPanel } from '@/features/ReasonContributionPanel';

const GroupDetailsPanel = () => {
    const { groupId } = useParams();
    const { data: result } = useMatchResult();
    const group = useMemo(
        () => result?.groups.find((currentGroup) => currentGroup.id === groupId) ?? null,
        [groupId, result],
    );

    if (!result) {
        return <Navigate to={ROUTES.UPLOAD} replace />;
    }

    if (!group) {
        return (
            <EmptyState
                title="Группа не найдена"
                description="Такой группы нет в текущем результате сопоставления."
                actionLabel="Вернуться к результатам"
                actionTo={ROUTES.RESULTS}
            />
        );
    }

    const hasPairwiseMatrix = group.pairwiseMatrix?.length && group.hotels.length > 1;

    return (
        <Stack spacing={3}>
            <BackButton/>
            <div className={styles.header}>
                <div>
                    <Typography variant="h4" component="h1">
                        {group.primaryName}
                    </Typography>
                    <Typography color="text.secondary">
                        {getPluralHotel(group.hotelsCount)} в группе
                    </Typography>
                </div>
            </div>

            {group.hotelsCount > 1 &&
                <Paper className={styles.metrics} variant="outlined">
                    <ScoreBar label="Степень совпадения" value={group.matchScore}/>
                    <ScoreBar label="Оценка уверенности" value={group.confidence}/>
                </Paper>
            }

            {group.hotelsCount > 1 &&
                <Paper className={styles.diagnostics} variant="outlined">
                    {hasPairwiseMatrix && group.pairwiseMatrix ? (
                        <PairwiseMatrix pairwiseMatrix={group.pairwiseMatrix} size={group.hotelsCount} />
                    ) : null}
                    <ReasonContributionPanel reasonContribution={group.reasonContribution} />
                </Paper>
            }

            <HotelTable hotels={group.hotels} />
        </Stack>
    );
};

export default GroupDetailsPanel;
