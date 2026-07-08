import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import { type ChangeEvent, useState } from 'react';

import { formatCoordinates } from '@/shared/lib';
import { EmptyState } from '@/shared/ui/EmptyState';
import type { HotelRecord } from '@/entities/hotel/model';

import styles from './HotelTable.module.scss';

const pageSizeOptions = [10, 25, 50] as const;

type HotelTableProps = {
    hotels: HotelRecord[];
};

const HotelTable = ({ hotels }: HotelTableProps) => {
    const [page, setPage] = useState(0);
    const [pageSize, setPageSize] = useState(10);
    const maxPage = Math.max(0, Math.ceil(hotels.length / pageSize) - 1);
    const currentPage = Math.min(page, maxPage);
    const firstRowIndex = currentPage * pageSize;
    const visibleHotels = hotels.slice(firstRowIndex, firstRowIndex + pageSize);

    if (hotels.length === 0) {
        return (
            <EmptyState
                title="Отели не найдены"
                description="Измените поисковый запрос или загрузите файл с записями."
            />
        );
    }

    const handleRowsPerPageChange = (event: ChangeEvent<HTMLInputElement>) => {
        setPageSize(Number(event.target.value));
        setPage(0);
    };

    return (
        <Paper variant="outlined">
            <TableContainer>
                <Table className={styles.table}>
                    <TableHead>
                        <TableRow>
                            <TableCell>Поставщик</TableCell>
                            <TableCell className={styles.nameCell}>Название</TableCell>
                            <TableCell>Адрес</TableCell>
                            <TableCell>Город</TableCell>
                            <TableCell>Страна</TableCell>
                            <TableCell>Координаты</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {visibleHotels.map((hotel) => (
                            <TableRow key={hotel.id} hover>
                                <TableCell>{hotel.providerId}</TableCell>
                                <TableCell className={styles.nameCell}>
                                    <Typography className={styles.name} title={hotel.name}>
                                        {hotel.name}
                                    </Typography>
                                    <Typography color="text.secondary" variant="body2">
                                        {hotel.externalId}
                                    </Typography>
                                </TableCell>
                                <TableCell>{hotel.address ?? '-'}</TableCell>
                                <TableCell>{hotel.city ?? '-'}</TableCell>
                                <TableCell>{hotel.country ?? '-'}</TableCell>
                                <TableCell>
                                    {hotel.coordinates
                                        ? formatCoordinates(hotel.coordinates.latitude, hotel.coordinates.longitude)
                                        : '-'}
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>
            <TablePagination
                component="div"
                count={hotels.length}
                page={currentPage}
                rowsPerPage={pageSize}
                rowsPerPageOptions={pageSizeOptions}
                labelRowsPerPage="Строк на странице"
                labelDisplayedRows={({ from, to, count }) => `${from}-${to} из ${count}`}
                onPageChange={(_, nextPage) => setPage(nextPage)}
                onRowsPerPageChange={handleRowsPerPageChange}
            />
        </Paper>
    );
};

export default HotelTable;
