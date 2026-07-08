import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { Link } from 'react-router-dom';

import styles from './EmptyState.module.scss';

type EmptyStateProps = {
    title: string;
    description: string;
    actionLabel?: string;
    actionTo?: string;
};

const EmptyState = ({ title, description, actionLabel, actionTo }: EmptyStateProps) => (
    <Paper className={styles.root} variant="outlined">
        <Stack className={styles.content} spacing={2}>
            <div>
                <Typography variant="h6" component="h2">
                    {title}
                </Typography>
                <Typography color="text.secondary">{description}</Typography>
            </div>
            {actionLabel && actionTo ? (
                <Button component={Link} to={actionTo} variant="contained">
                    {actionLabel}
                </Button>
            ) : null}
        </Stack>
    </Paper>
);

export default EmptyState;
