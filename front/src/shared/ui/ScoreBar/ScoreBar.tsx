import LinearProgress from '@mui/material/LinearProgress';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

import { formatPercent } from '@/shared/lib';

import styles from './ScoreBar.module.scss';

type ScoreBarProps = {
    label: string;
    value: number;
};

const ScoreBar = ({ label, value }: ScoreBarProps) => (
    <Stack spacing={0.75}>
        <div className={styles.header}>
            <Typography variant="body2" color="text.secondary">
                {label}
            </Typography>
            <Typography className={styles.value} variant="body2">
                {formatPercent(value)}
            </Typography>
        </div>
        <LinearProgress variant="determinate" value={Math.min(100, Math.max(0, value))} />
    </Stack>
);

export default ScoreBar;
