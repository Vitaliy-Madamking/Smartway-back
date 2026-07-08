import type { Transition } from 'motion/react';

export const transitions = {
    fast: {
        duration: 0.15,
        ease: 'easeOut',
    },

    normal: {
        duration: 0.25,
        ease: 'easeOut',
    },

    slow: {
        duration: 0.4,
        ease: 'easeOut',
    },

    spring: {
        type: 'spring',
        stiffness: 260,
        damping: 24,
    },
} satisfies Record<string, Transition>;