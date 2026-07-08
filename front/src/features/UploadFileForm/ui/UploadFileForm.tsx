import { zodResolver } from '@hookform/resolvers/zod';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { Controller, useController, useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { z } from 'zod';

import { createMatchingJob, isUploadRequestError } from '@/entities/match-result/api';
import {
    MATCH_REQUEST_QUERY_KEY,
    MATCH_RESULT_QUERY_KEY,
    MATCHING_ALGORITHMS,
    type MatchingAlgorithm,
    type MatchingRequest,
} from '@/entities/match-result/model';
import { ROUTES } from '@/shared/config';
import { CSV_FILE_ERROR_MESSAGE, isCsvFile } from '@/features/UploadFileForm/lib';
import { CsvFileDropzone } from '@/features/UploadFileForm/ui/CsvFileDropzone';

import styles from './UploadFileForm.module.scss';

const uploadJsonFormSchema = z.object({
    file: z.custom<File | null>().superRefine((value, context) => {
        if (typeof File === 'undefined' || !(value instanceof File)) {
            context.addIssue({
                code: 'custom',
                message: 'Выберите CSV-файл',
            });
            return;
        }

        if (!isCsvFile(value)) {
            context.addIssue({
                code: 'custom',
                message: CSV_FILE_ERROR_MESSAGE,
            });
        }
    }),
    threshold: z
        .number({ error: 'Введите порог уверенности' })
        .min(0, { error: 'Порог должен быть не меньше 0' })
        .max(100, { error: 'Порог должен быть не больше 100' }),
    algorithm: z.enum(MATCHING_ALGORITHMS, {
        error: 'Выберите алгоритм',
    }),
});

const INITIAL_PAGE = 1;
const INITIAL_LIMIT = 30;

const DEFAULT_UPLOAD_ERROR_MESSAGE = 'Не удалось отправить файл. Попробуйте еще раз.';

const withServerDetail = (message: string, detail?: string): string => {
    if (!detail) {
        return message;
    }

    try {
        const parsed = JSON.parse(detail) as { message?: string };

        return parsed.message
            ? `${message} ${parsed.message}`
            : `${message} ${detail}`;
    } catch {
        return `${message} ${detail}`;
    }
};

const getUploadErrorMessage = (error: unknown): string => {
    if (!isUploadRequestError(error)) {
        return DEFAULT_UPLOAD_ERROR_MESSAGE;
    }

    switch (error.code) {
        case 'missing-file':
            return 'Выберите CSV-файл перед отправкой.';
        case 'bad-request':
            return withServerDetail(
                'Сервер не принял файл или параметры загрузки.',
                error.detail,
            );
        case 'server':
            return withServerDetail(
                'На сервере произошла ошибка при обработке файла.',
                error.detail,
            );
        case 'network':
            return 'Не удалось подключиться к серверу. Проверьте интернет или доступность API.';
        case 'unexpected-response':
            return 'Сервер вернул неожиданный ответ. Обновите страницу и попробуйте еще раз.';
        case 'unknown':
            return error.status
                ? `Сервер вернул ошибку ${error.status}. Попробуйте еще раз или проверьте файл.`
                : DEFAULT_UPLOAD_ERROR_MESSAGE;
        default:
            return DEFAULT_UPLOAD_ERROR_MESSAGE;
    }
};

type UploadJsonFormValues = {
    file: File | null;
    threshold: number;
    algorithm: MatchingAlgorithm;
};

const UploadFileForm = () => {
    const queryClient = useQueryClient();
    const navigate = useNavigate();

    const {
        clearErrors,
        control,
        formState: { errors, isSubmitting },
        handleSubmit,
        register,
        setError,
    } = useForm<UploadJsonFormValues>({
        resolver: zodResolver(uploadJsonFormSchema),
        defaultValues: {
            algorithm: 'universal',
            file: null,
            threshold: 85,
        },
    });

    const {
        field: { onChange: setFileValue, value: selectedFile },
        fieldState: { error: fileError },
    } = useController<UploadJsonFormValues, 'file'>({
        control,
        name: 'file',
    });

    const mutation = useMutation({
        mutationFn: createMatchingJob,
        onSuccess: (result, request) => {
            queryClient.setQueryData(MATCH_REQUEST_QUERY_KEY, request);
            queryClient.setQueryData(MATCH_RESULT_QUERY_KEY, result);
            navigate(ROUTES.RESULTS);
        },
    });
    const isLoading = isSubmitting || mutation.isPending;
    const uploadErrorMessage = mutation.error ? getUploadErrorMessage(mutation.error) : null;

    const handleFileChange = useCallback(
        (file: File | null) => {
            setFileValue(file);
            clearErrors('file');
        },
        [clearErrors, setFileValue],
    );

    const handleFileError = useCallback(
        (message: string) => {
            setFileValue(null);
            setError('file', {
                message,
                type: 'manual',
            });
        },
        [setError, setFileValue],
    );

    const submitForm = handleSubmit(async ({ file, threshold, algorithm }) => {
        if (!file) {
            return;
        }

        const request: MatchingRequest = {
            algorithm: algorithm,
            file,
            limit: INITIAL_LIMIT,
            page: INITIAL_PAGE,
            search: '',
            threshold,
        };

        await mutation.mutateAsync(request);
    });

    return (
        <Paper className={styles.root} variant="outlined">
            <form className={styles.form} onSubmit={submitForm}>
                <Stack spacing={3}>
                    <CsvFileDropzone
                        value={selectedFile}
                        error={fileError?.message}
                        disabled={isLoading}
                        onChange={handleFileChange}
                        onError={handleFileError}
                    />

                    <TextField
                        label="Порог уверенности"
                        type="number"
                        slotProps={{ htmlInput: { min: 0, max: 100, step: 1 } }}
                        error={Boolean(errors.threshold)}
                        helperText={errors.threshold?.message ?? 'Значение от 0 до 100'}
                        disabled={isLoading}
                        {...register('threshold', { valueAsNumber: true })}
                    />

                    <Controller
                        name="algorithm"
                        control={control}
                        render={({ field, fieldState }) => (
                            <FormControl error={Boolean(fieldState.error)} disabled={isLoading}>
                                <InputLabel id="matching-algorithm-label">Алгоритм</InputLabel>
                                <Select
                                    labelId="matching-algorithm-label"
                                    label="Алгоритм"
                                    {...field}
                                >
                                    {MATCHING_ALGORITHMS.map((algorithmValue) => (
                                        <MenuItem key={algorithmValue} value={algorithmValue}>
                                            {algorithmValue}
                                        </MenuItem>
                                    ))}
                                </Select>
                                {fieldState.error?.message ? (
                                    <FormHelperText>{fieldState.error.message}</FormHelperText>
                                ) : null}
                            </FormControl>
                        )}
                    />

                    {uploadErrorMessage ? <Alert severity="error">{uploadErrorMessage}</Alert> : null}

                    <div className={styles.actions}>
                        <Button type="submit" variant="contained" disabled={isLoading}>
                            {isLoading ? <CircularProgress color="inherit" size={20} /> : null}
                            Отправить
                        </Button>
                    </div>
                </Stack>
            </form>
        </Paper>
    );
};

export default UploadFileForm;
