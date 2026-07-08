import CircularProgress from '@mui/material/CircularProgress';
import {lazy, Suspense, useEffect, type ReactNode} from 'react';
import {Navigate, Route, Routes, useLocation} from 'react-router-dom';

import { ROUTES } from '@/shared/config';

import styles from './AppRouter.module.scss';
import {RouteTransition} from "@/shared/ui/motion";
import { AnimatePresence } from 'motion/react';

const UploadPage = lazy(() => import('@/pages/UploadPage'));
const ResultsPage = lazy(() => import('@/pages/ResultsPage'));
const GroupDetailsPage = lazy(() => import('@/pages/GroupDetailsPage'));

const withSuspense = (page: ReactNode) => (
    <Suspense
        fallback={
            <div className={styles.fallback}>
                <CircularProgress />
            </div>
        }
    >
        <RouteTransition>
            {page}
        </RouteTransition>
    </Suspense>
);

const ScrollToTop = () => {
    const { pathname } = useLocation();

    useEffect(() => {
        window.scrollTo(0, 0);
    }, [pathname]);

    return null;
};

const AppRouter = () => {
    const location = useLocation();

    return (
        <AnimatePresence mode="wait">
            <ScrollToTop />
            <Routes location={location} key={location.pathname}>
                <Route index element={<Navigate to={ROUTES.UPLOAD} replace/>}/>
                <Route path={ROUTES.UPLOAD} element={withSuspense(<UploadPage/>)}/>
                <Route path={ROUTES.RESULTS} element={withSuspense(<ResultsPage/>)}/>
                <Route path={`${ROUTES.GROUPS}/:groupId`} element={withSuspense(<GroupDetailsPage/>)}/>
                <Route path="*" element={<Navigate to={ROUTES.UPLOAD} replace/>}/>
            </Routes>
        </AnimatePresence>
    )
};

export default AppRouter;
