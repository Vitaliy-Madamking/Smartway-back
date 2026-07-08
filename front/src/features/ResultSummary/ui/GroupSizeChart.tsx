import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import { PieChart } from '@mui/x-charts';

import type { GroupSizeStatsDto } from '@/entities/match-result/api';

import styles from './ResultSummary.module.scss';
import React from "react";

type GroupSizeBarChartProps = {
    data: GroupSizeStatsDto[];
};

const GroupSizeChart: React.FC<GroupSizeBarChartProps> = ({ data }) => {
    const total = data.reduce((sum, item) => sum + item.groupsCount, 0);

    return (
        <Paper className={styles.card} variant="outlined">
            <Typography variant="h6">
                Размеры групп
            </Typography>

            <div className={styles.chart}>
                <PieChart
                    height={250}
                    series={[
                        {
                            data: data.map((item) => {
                                const percent = (item.groupsCount / total) * 100;

                                return {
                                    id: item.hotelsCount,
                                    value: item.groupsCount,
                                    label: `${item.hotelsCount} (${percent.toFixed(2)}%)`,
                                };
                            }),
                            innerRadius: 30,
                            outerRadius: 100,
                            paddingAngle: 2,
                            cornerRadius: 4,
                        },
                    ]}
                    slotProps={{
                        legend: {
                            direction: 'vertical',
                            position: {
                                vertical: 'middle',
                                horizontal: 'end',
                            },
                        },
                    }}
                />
            </div>
        </Paper>
    );
};

export default GroupSizeChart;