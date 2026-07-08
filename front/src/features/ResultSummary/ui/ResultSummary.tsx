import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';

import { formatPercent } from '@/shared/lib';

import styles from './ResultSummary.module.scss';
import type { MetricsModel } from '@/entities/match-result/model';
import GroupSizeChart from './GroupSizeChart';

type ResultSummaryProps = {
    result: MetricsModel;
};

const metricsItems = [
    { key: 'totalHotels', label: 'Отелей', percent: false },
    { key: 'averageConfidence', label: 'Средняя уверенность', percent: true },
] as const;

const sideMetrics = [
    { key: 'totalGroups', label: 'Групп совпадений', percent: false },
    { key: 'totalDuplicates', label: 'Дубликатов', percent: false },
    { key: 'totalProviders', label: 'Поставщиков', percent: false },
] as const;

const ResultSummary = ({ result }: ResultSummaryProps) => (
    <div className={styles.summary}>
        <div className={styles.chartMetrics}>
            <div className={styles.metricsGrid}>
                {metricsItems.map((item) => {
                    const value = result[item.key];

                    return (
                        <Paper key={item.key} className={styles.item} variant="outlined">
                            <Typography color="text.secondary" variant="body2">
                                {item.label}
                            </Typography>
                            <Typography className={styles.value} variant="h4" component="p">
                                {item.percent ? formatPercent(value) : value}
                            </Typography>
                        </Paper>
                    );
                })}
            </div>
            <GroupSizeChart data={result.groupSizeStats}/>
        </div>
        <div className={styles.sideMetrics}>
            {sideMetrics.map((item) => {
                const value = result[item.key];

                return (
                    <Paper key={item.key} className={styles.item} variant="outlined">
                        <Typography color="text.secondary" variant="body2">
                            {item.label}
                        </Typography>
                        <Typography className={styles.value} variant="h4" component="p">
                            {item.percent ? formatPercent(value) : value}
                        </Typography>
                    </Paper>
                );
            })}
        </div>
    </div>
);

export default ResultSummary;
