import Button from '@mui/material/Button';
import Pagination from '@mui/material/Pagination';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { forwardRef, useCallback, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';

import { getPluralHotel } from '@/entities/hotel/lib';
import type { MatchGroup } from '@/entities/match-result/model';
import { ROUTES } from '@/shared/config';
import { ConfidenceBadge } from '@/shared/ui/ConfidenceBadge';
import { ScoreBar } from '@/shared/ui/ScoreBar';

import styles from './MatchGroupList.module.scss';

type MatchGroupListProps = {
    groups: MatchGroup[];
    page: number;
    totalPages: number;
    disabled?: boolean;
    onPageChange: (page: number) => void;
};

type MatchGroupPaginationProps = {
    page: number;
    pagesCount: number;
    disabled?: boolean;
    onPageChange: (page: number) => void;
};

const MatchGroupPagination = forwardRef<HTMLDivElement, MatchGroupPaginationProps>(
    ({ page, pagesCount, disabled, onPageChange }, ref) => (
        <div ref={ref} className={styles.pagination}>
            <Pagination
                color="primary"
                count={pagesCount}
                disabled={disabled}
                page={page}
                showFirstButton
                showLastButton
                siblingCount={1}
                boundaryCount={1}
                onChange={(_, nextPage) => {
                    onPageChange(nextPage);
                }}
            />
        </div>
    ),
);

MatchGroupPagination.displayName = 'MatchGroupPagination';

const MatchGroupList = ({
    groups,
    page,
    totalPages,
    disabled,
    onPageChange,
}: MatchGroupListProps) => {
    const pagesCount = Math.max(1, totalPages);

    const topPaginationRef = useRef<HTMLDivElement | null>(null);
    const pendingScrollPageRef = useRef<number | null>(null);

    const scrollToTopPagination = useCallback(() => {
        topPaginationRef.current?.scrollIntoView({
            behavior: 'smooth',
            block: 'start',
        });
    }, []);

    useEffect(() => {
        const pendingScrollPage = pendingScrollPageRef.current;

        if (disabled || pendingScrollPage === null) {
            return;
        }

        pendingScrollPageRef.current = null;

        if (pendingScrollPage !== page) {
            return;
        }

        const animationFrameId = window.requestAnimationFrame(scrollToTopPagination);

        return () => {
            window.cancelAnimationFrame(animationFrameId);
        };
    }, [disabled, page, scrollToTopPagination]);

    const handleBottomPageChange = (nextPage: number) => {
        pendingScrollPageRef.current = nextPage;
        onPageChange(nextPage);
    };

    return (
        <div className={styles.root}>
            <MatchGroupPagination
                ref={topPaginationRef}
                disabled={disabled}
                page={page}
                pagesCount={pagesCount}
                onPageChange={onPageChange}
            />

            <div className={styles.wrapper}>
                <div className={styles.groupsGrid}>
                    {groups.map((group) => (
                        <Paper key={group.id} className={styles.card} variant="outlined">
                            <Stack spacing={2}>
                                <div className={styles.header}>
                                    <div>
                                        <Typography variant="h6" component="h2">
                                            {group.primaryName}
                                        </Typography>
                                        <Typography color="text.secondary">
                                            {getPluralHotel(group.hotels.length)} от поставщиков
                                        </Typography>
                                    </div>
                                    <ConfidenceBadge value={group.confidence}/>
                                </div>
                                {group.hotelsCount > 1 &&
                                    <ScoreBar label="Степень совпадения" value={group.matchScore}/>
                                }
                                {group.reasons?.length ? (
                                    <div className={styles.section}>
                                        <Typography color="text.secondary" variant="body2">
                                            Причины:
                                        </Typography>
                                        <div className={styles.reasons}>
                                            {group.reasons.map((reason) => (
                                                <span key={reason}>{reason}</span>
                                            ))}
                                        </div>
                                    </div>
                                ) : null}
                                <div className={styles.section}>
                                    <Typography color="text.secondary" variant="body2">
                                        Отели:
                                    </Typography>
                                    <div className={styles.hotels}>
                                        {group.hotels.map((hotel) => (
                                            <span key={hotel.id}>{hotel.providerId}: {hotel.name}</span>
                                        ))}
                                    </div>
                                </div>
                            </Stack>
                            <Button component={Link} to={`${ROUTES.GROUPS}/${group.id}`} variant="outlined">
                                Открыть группу
                            </Button>
                        </Paper>
                    ))}
                </div>

                {disabled && (
                    <div className={styles.overlay}>
                    </div>
                )}
            </div>

            <MatchGroupPagination
                disabled={disabled}
                page={page}
                pagesCount={pagesCount}
                onPageChange={handleBottomPageChange}
            />
        </div>
    );
};

export default MatchGroupList;
