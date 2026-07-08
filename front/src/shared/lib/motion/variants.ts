import type { Variants } from 'motion/react';

import { transitions } from './transitions';

export const fade: Variants = {
    hidden: {
        opacity: 0,
    },

    visible: {
        opacity: 1,
        transition: transitions.normal,
    },

    exit: {
        opacity: 0,
        transition: transitions.fast,
    },
};

export const slideUp: Variants = {
    hidden: {
        opacity: 0,
        y: 16,
    },

    visible: {
        opacity: 1,
        y: 0,
        transition: transitions.normal,
    },

    exit: {
        opacity: 0,
        y: -16,
        transition: transitions.fast,
    },
};

export const staggerContainer: Variants = {
    hidden: {},

    visible: {
        transition: {
            staggerChildren: 0.06,
        },
    },
};

export const staggerItem: Variants = {
    hidden: {
        opacity: 0,
        y: 12,
    },

    visible: {
        opacity: 1,
        y: 0,
        transition: transitions.normal,
    },
};