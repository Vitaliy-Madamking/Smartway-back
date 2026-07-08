import Typography from '@mui/material/Typography';

import type { ReasonContribution } from '@/entities/match-result/model';
import { ScoreBar } from '@/shared/ui/ScoreBar';

import styles from './ReasonContributionPanel.module.scss';
import * as React from "react";

const FEATURES = [
    { key: 'name', label: 'Название' },
    { key: 'address', label: 'Адрес' },
    { key: 'geo', label: 'Гео' },
    { key: 'city', label: 'Город' },
] as const;

interface ReasonContributionPanelProps {
    reasonContribution?: ReasonContribution;
}

const ReasonContributionPanel: React.FC<ReasonContributionPanelProps> = ({ reasonContribution }) => {
    if (!reasonContribution) {
        return null;
    }

    return (
        <div className={styles.diagnosticPanel}>
            <Typography variant="h6" component="h2" className={styles.diagnosticTitle}>
                Вклад признаков (в среднем)
            </Typography>
            <div className={styles.featureList}>
                {FEATURES.map(({ key, label }) => (
                    <ScoreBar key={key} label={label} value={reasonContribution[key] * 100} />
                ))}
            </div>
        </div>
    );
};

export default ReasonContributionPanel;
