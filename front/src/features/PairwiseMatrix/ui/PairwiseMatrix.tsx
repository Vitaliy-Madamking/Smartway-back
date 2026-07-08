import Typography from '@mui/material/Typography';
import { Fragment } from 'react';

import type { PairwiseMatrixItem } from '@/entities/match-result/model';

import styles from './PairwiseMatrix.module.scss';
import {useTheme} from "@mui/material/styles";
import * as React from "react";

const getMatrixCellColor = (baseColor: string, value: number): string => {
    const opacity = Math.min(Math.max(value * 100, 0), 100);

    return `color-mix(in srgb, ${baseColor} ${opacity}%, transparent)`;
};

const getMatrixRows = (pairwiseMatrix: PairwiseMatrixItem[], size: number): Array<Array<number | null>> => {
    const rows = Array.from({ length: size }, () => Array<number | null>(size).fill(null));

    for (let index = 0; index < size; index += 1) {
        rows[index][index] = 1;
    }

    pairwiseMatrix.forEach(({ indexA, indexB, similarity }) => {
        if (
            Number.isInteger(indexA) &&
            Number.isInteger(indexB) &&
            indexA >= 0 &&
            indexB >= 0 &&
            indexA < size &&
            indexB < size
        ) {
            rows[indexA][indexB] = similarity;
            rows[indexB][indexA] = similarity;
        }
    });

    return rows;
};

interface PairwiseMatrixProps {
    pairwiseMatrix: PairwiseMatrixItem[];
    size: number
}

const PairwiseMatrix: React.FC<PairwiseMatrixProps> = ({ pairwiseMatrix, size }) => {
    const theme = useTheme();
    const rows = getMatrixRows(pairwiseMatrix, size);
    const gridTemplateColumns = `28px repeat(${size}, minmax(44px, 1fr))`;

    const getCellClassName = (value: number | null) => {
        const cellClassName = [
            styles.matrixCell,
            value ? '' : styles.matrixCellEmpty,
            value !== null && value >= 0.72
                ? styles.matrixCellContrast
                : '',
        ]
            .filter(Boolean)
            .join(' ');

        return cellClassName
    }

    return (
        <div className={styles.diagnosticPanel}>
            <Typography variant="h6" component="h2" className={styles.diagnosticTitle}>
                Попарная матрица (сходство)
            </Typography>
            <div className={styles.matrixWrap}>
                <div className={styles.matrixGrid} style={{ gridTemplateColumns }}>
                    <span/>
                    {rows.map((_, index) => (
                        <span key={`column-${index}`} className={styles.matrixAxisLabel}>
                            {index + 1}
                        </span>
                    ))}
                    {rows.map((row, rowIndex) => (
                        <Fragment key={`row-${rowIndex}`}>
                            <span className={styles.matrixAxisLabel}>{rowIndex + 1}</span>
                            {row.map((value, columnIndex) => {
                                const cellClassName = getCellClassName(value);

                                return (
                                    <span
                                        key={`${rowIndex}-${columnIndex}`}
                                        className={cellClassName}
                                        style={value ? { backgroundColor: getMatrixCellColor(theme.palette.primary.main, value) } : undefined}
                                    >
                                        {value ? value?.toFixed(2) : '-'}
                                    </span>
                                );
                            })}
                        </Fragment>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default PairwiseMatrix;
