import Chip from '@mui/material/Chip';

type ConfidenceBadgeProps = {
    value: number;
};

const ConfidenceBadge = ({ value }: ConfidenceBadgeProps) => {
    const color = value >= 85 ? 'success' : value >= 70 ? 'warning' : 'error';

    return <Chip color={color} label={`${Math.round(value)}% уверенность`} size="small" variant="outlined" />;
};

export default ConfidenceBadge;
