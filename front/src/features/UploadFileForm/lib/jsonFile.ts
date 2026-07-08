import type { Accept } from 'react-dropzone';

export const CSV_FILE_ERROR_MESSAGE = 'Файл должен быть в формате .csv';

export const CSV_DROPZONE_ACCEPT: Accept = {
    'text/csv': ['.csv'],
    'application/vnd.ms-excel': ['.csv'],
};

export const isCsvFile = (file: File): boolean => {
    if (!file.name) return false;

    const fileName = file.name.trim().toLowerCase();

    return fileName.endsWith('.csv');
};
