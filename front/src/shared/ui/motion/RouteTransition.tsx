import type { PropsWithChildren } from 'react';
import { motion } from 'motion/react';

import { slideUp } from '@/shared/lib/motion';

const RouteTransition = ({ children }: PropsWithChildren) => (
    <motion.div
        variants={slideUp}
        initial="hidden"
        animate="visible"
        exit="exit"
        style={{ height: '100%' }}
    >
        {children}
    </motion.div>
);

export default RouteTransition;