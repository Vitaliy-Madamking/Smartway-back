import CssBaseline from '@mui/material/CssBaseline';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { PropsWithChildren } from 'react';
import { BrowserRouter } from 'react-router-dom';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: false,
            staleTime: 5 * 60 * 1000,
        },
    },
});

const theme = createTheme({
    palette: {
        primary: {
            main: '#d24233',
        },
        background: {
            default: '#f8fafc',
            paper: '#ffffff',
        },
        text: {
            primary: '#111827',
            secondary: '#64748b',
        },
    },
    shape: {
        borderRadius: 8,
    },
    typography: {
        fontFamily: 'Inter, Arial, sans-serif',
    },
});

const AppProviders = ({ children }: PropsWithChildren) => (
    <QueryClientProvider client={queryClient}>
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <BrowserRouter>{children}</BrowserRouter>
        </ThemeProvider>
    </QueryClientProvider>
);

export default AppProviders;
