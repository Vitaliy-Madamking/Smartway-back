import Button from '@mui/material/Button';
import CloudUploadOutlined from '@mui/icons-material/CloudUploadOutlined';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import { type MouseEvent, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';

import { CSV_DROPZONE_ACCEPT, CSV_FILE_ERROR_MESSAGE, isCsvFile } from '@/features/UploadFileForm/lib';

import styles from './CsvFileDropzone.module.scss';

type JsonFileDropzoneProps = {
    value: File | null;
    error?: string;
    disabled?: boolean;
    onChange: (file: File | null) => void;
    onError: (message: string) => void;
};

const CsvFileDropzone = ({ value, error, disabled, onChange, onError }: JsonFileDropzoneProps) => {
    const selectedFileLabel = value
        ? `${value.name} · ${Math.max(1, Math.round(value.size / 1024))} KB`
        : 'Файл не выбран';

    const handleDropAccepted = useCallback(
        (acceptedFiles: File[]) => {
            const file = acceptedFiles[0];

            if (!file) {
                return;
            }

            if (!isCsvFile(file)) {
                onError(CSV_FILE_ERROR_MESSAGE);
                return;
            }

            onChange(file);
        },
        [onChange, onError],
    );

    const handleDropRejected = useCallback(
        () => {
            onError(CSV_FILE_ERROR_MESSAGE);
        },
        [onError],
    );

    const handleResetClick = useCallback(
        (event: MouseEvent<HTMLButtonElement>) => {
            event.stopPropagation();
            onChange(null);
        },
        [onChange],
    );

    const { getInputProps, getRootProps, isDragActive, isDragReject } = useDropzone({
        accept: CSV_DROPZONE_ACCEPT,
        disabled,
        maxFiles: 1,
        multiple: false,
        onDropAccepted: handleDropAccepted,
        onDropRejected: handleDropRejected,
        validator: (file) =>
            isCsvFile(file) ? null : { code: 'file-invalid-type', message: CSV_FILE_ERROR_MESSAGE },
    });

    const dropzoneClassName = [
        styles.dropzone,
        isDragActive ? styles.dropzoneActive : '',
        isDragReject || error ? styles.dropzoneError : '',
    ]
        .filter(Boolean)
        .join(' ');

    return (
        <FormControl error={Boolean(error)} fullWidth>
            <div {...getRootProps({ className: dropzoneClassName })}>
                <input {...getInputProps()} />
                <div className={styles.dropzoneHeader}>
                    <CloudUploadOutlined className={styles.uploadIcon} />
                    <div>
                        <div className={styles.dropzoneTitle}>Загрузите CSV с отелями поставщиков</div>
                        <div className={styles.dropzoneHint}>
                            Перетащите файл сюда или выберите его на устройстве
                        </div>
                    </div>
                </div>
                <div className={styles.fileMeta}>
                    <span className={styles.fileName}>{selectedFileLabel}</span>
                    {value ? (
                        <Button
                            type="button"
                            size="small"
                            variant="text"
                            disabled={disabled}
                            onClick={handleResetClick}
                        >
                            Удалить файл
                        </Button>
                    ) : null}
                </div>
            </div>
            {error ? <FormHelperText className={styles.helperText}>{error}</FormHelperText> : null}
        </FormControl>
    );
};

export default CsvFileDropzone;
