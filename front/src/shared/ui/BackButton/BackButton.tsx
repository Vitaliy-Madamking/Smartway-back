import ArrowBackIosNew from '@mui/icons-material/ArrowBackIosNew';
import { useNavigate } from 'react-router-dom';

import styles from './BackButton.module.scss';

const BackButton = () => {
    const navigate = useNavigate();

    return (
        <button className={styles.backIconContainer} type="button" onClick={() => navigate(-1)}>
            <ArrowBackIosNew className={styles.backIcon} /> Назад
        </button>
    );
};

export default BackButton;
