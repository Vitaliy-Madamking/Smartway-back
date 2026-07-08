import Container from '@mui/material/Container';

import { ResultsOverview } from '@/widgets/ResultsOverview';

const ResultsPage = () => (
    <Container maxWidth="lg" sx={{ py: { xs: 3, md: 4 } }}>
        <ResultsOverview />
    </Container>
);

export default ResultsPage;
