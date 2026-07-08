import { GroupDetailsPanel } from '@/widgets/GroupDetailsPanel';
import Container from "@mui/material/Container";

const GroupDetailsPage = () => (
    <Container maxWidth="lg" sx={{ py: { xs: 3, md: 4 } }}>
        <GroupDetailsPanel />
    </Container>
);

export default GroupDetailsPage;
