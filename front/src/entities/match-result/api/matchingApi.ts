import type { MatchResult, MatchingRequest } from '@/entities/match-result/model/types';
import { mapUploadResponseToMatchResult } from './mapUploadResponse';

export type UploadErrorCode = 'bad-request' | 'missing-file' | 'network' | 'server' | 'unexpected-response' | 'unknown';

export type UploadRequestError = Error & {
    code: UploadErrorCode;
    status?: number;
    detail?: string;
};

const createUploadError = (
    code: UploadErrorCode,
    message: string,
    extra?: { status?: number; detail?: string }
): UploadRequestError => {
    const error = new Error(message) as UploadRequestError;

    error.name = 'UploadRequestError';
    error.code = code;
    error.status = extra?.status;
    error.detail = extra?.detail;

    return error;
};

export const isUploadRequestError = (error: unknown): error is UploadRequestError =>
    error instanceof Error && 'code' in error;

export const createMatchingJob = async (request: MatchingRequest): Promise<MatchResult> => {
    if (!request.file) {
        throw createUploadError('missing-file', 'CSV file is required');
    }

    const formData = new FormData();

    formData.append('file', request.file);
    if (request.algorithm) {
        formData.append('algorithm', request.algorithm);
    }
    formData.append('threshold', (request.threshold / 100).toString());
    formData.append('page', request.page.toString());
    formData.append('limit', request.limit.toString());
    formData.append('search', request.search);
    if (request.sortBy && request.sortDir) {
        formData.append('sortBy', request.sortBy);
        formData.append('sortDir', request.sortDir);
    }

    let response: Response;

    try {
        response = await fetch(`${import.meta.env.VITE_API_URL}/upload`, {
            method: 'POST',
            body: formData
        });
    } catch (error) {
        throw createUploadError('network', error instanceof Error ? error.message : 'Network request failed');
    }

    if (!response.ok) {
        const detail = (await response.text().catch(() => '')).trim() || undefined;
        const code = response.status === 400 ? 'bad-request' : response.status >= 500 ? 'server' : 'unknown';
        throw createUploadError(code, 'Upload failed', { status: response.status, detail });
    }

    try {
        const data = await response.json();
        return mapUploadResponseToMatchResult(data);
    } catch {
        throw createUploadError('unexpected-response', 'Upload response is invalid or has unexpected shape');
    }
};

