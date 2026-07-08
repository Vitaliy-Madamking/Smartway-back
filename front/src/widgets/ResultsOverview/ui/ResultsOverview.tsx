import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {useCallback, useEffect, useState} from 'react';
import { Navigate } from 'react-router-dom';

import { createMatchingJob } from '@/entities/match-result/api';
import { MATCH_REQUEST_QUERY_KEY, MATCH_RESULT_QUERY_KEY } from '@/entities/match-result/model';
import type { MatchingRequest } from '@/entities/match-result/model';
import { useMatchResult } from '@/entities/match-result/model';
import { MatchGroupList } from '@/features/MatchGroupList';
import { ResultSummary } from '@/features/ResultSummary';
import { ROUTES } from '@/shared/config';
import { BackButton } from '@/shared/ui/BackButton';
import { EmptyState } from '@/shared/ui/EmptyState';
import ResultsControls from "@/widgets/ResultsOverview/ui/ResultsControls.tsx";
import {SEARCH_DEBOUNCE_MS, SORT_OPTIONS} from "@/widgets/ResultsOverview/model/constants.ts";
import type {MatchingRequestPatch} from "@/widgets/ResultsOverview/model/types.ts";

const ResultsOverview = () => {
    const queryClient = useQueryClient();
    const { data: result } = useMatchResult();
    const { data: request } = useQuery({
        queryKey: MATCH_REQUEST_QUERY_KEY,
        queryFn: async (): Promise<MatchingRequest | null> => null,
        staleTime: Number.POSITIVE_INFINITY,
    });
    const [searchValue, setSearchValue] = useState(request?.search ?? '');

    const { isPending: isLoading, mutate } = useMutation({
        mutationFn: createMatchingJob,
        onMutate: (nextRequest) => {
            const previousRequest = queryClient.getQueryData<MatchingRequest | null>(MATCH_REQUEST_QUERY_KEY);

            queryClient.setQueryData(MATCH_REQUEST_QUERY_KEY, nextRequest);

            return { previousRequest };
        },
        onSuccess: (nextResult, nextRequest) => {
            queryClient.setQueryData(MATCH_REQUEST_QUERY_KEY, nextRequest);
            queryClient.setQueryData(MATCH_RESULT_QUERY_KEY, nextResult);
        },
        onError: (_error, _nextRequest, context) => {
            if (context?.previousRequest) {
                queryClient.setQueryData(MATCH_REQUEST_QUERY_KEY, context.previousRequest);
            }
        },
    });

    const selectedSort =
        SORT_OPTIONS.find(
            (option) =>
                option.sortBy === request?.sortBy &&
                option.sortDir === request?.sortDir,
        )?.value ?? SORT_OPTIONS[0].value;

    const updateRequest = useCallback(
        (patch: MatchingRequestPatch) => {
            if (!request) {
                return;
            }

            mutate({
                ...request,
                ...patch,
            });
        },
        [mutate, request],
    );

    useEffect(() => {
        if (!request || isLoading) {
            return;
        }

        const nextSearch = searchValue.trim();

        if (nextSearch === request.search) {
            return;
        }

        const timeoutId = window.setTimeout(() => {
            updateRequest({
                page: 1,
                search: nextSearch,
            });
        }, SEARCH_DEBOUNCE_MS);

        return () => {
            window.clearTimeout(timeoutId);
        };
    }, [isLoading, request, searchValue, updateRequest]);

    const onSortChange = (value: string) => {
        const option = SORT_OPTIONS.find(
            (item) => item.value === value,
        );

        if (!option) {
            return;
        }

        updateRequest({
            sortBy: option.sortBy,
            sortDir: option.sortDir,
            page: 1,
        });
    }
    const onLimitChange = (limit: number) => {
        updateRequest({
            limit: limit,
            page: 1,
        });
    }

    if (!result || !request) {
        return <Navigate to={ROUTES.UPLOAD} replace />;
    }

    return (
        <Stack spacing={3}>
            <BackButton />
            <div>
                <Typography variant="h4" component="h1">
                    Результаты сопоставления
                </Typography>
            </div>
            <ResultSummary result={result.metrics} />
            <ResultsControls
                searchValue={searchValue}
                selectedSort={selectedSort}
                limit={request.limit}
                disabled={isLoading}
                onSearchChange={setSearchValue}
                onSortChange={onSortChange}
                onLimitChange={onLimitChange}
            />

            {result.groups.length > 0 ? (
                <MatchGroupList
                    groups={result.groups}
                    page={request.page}
                    totalPages={result.pagination.totalPages}
                    disabled={isLoading}
                    onPageChange={(page) => {
                        updateRequest({page});
                    }}
                />
            ) : (
                <EmptyState
                    title="Групп нет"
                    description="В загруженном файле нет отелей, соответствующих параметрам"
                />
            )}

        </Stack>
    );
};

export default ResultsOverview;
