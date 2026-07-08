import Container from '@mui/material/Container';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

import { UploadFileForm } from '@/features/UploadFileForm';

const UploadPage = () => (
    <Container maxWidth="lg" sx={{ py: { xs: 3, md: 4 } }}>
        <Stack spacing={3}>
            <div>
                <Typography variant="h4" component="h1">
                    Загрузка CSV
                </Typography>
            </div>
            <UploadFileForm />
        </Stack>
    </Container>
);

export default UploadPage;
