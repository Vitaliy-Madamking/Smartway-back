import React from 'react';
import TextField from "@mui/material/TextField";
import InputAdornment from "@mui/material/InputAdornment";
import CircularProgress from "@mui/material/CircularProgress";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import {LIMIT_OPTIONS, SORT_OPTIONS} from "@/widgets/ResultsOverview/model/constants.ts";

type ResultsControlsProps = {
    searchValue: string;
    selectedSort: string;
    limit: number;
    disabled: boolean;
    onSearchChange: (value: string) => void;
    onSortChange: (value: string) => void;
    onLimitChange: (limit: number) => void;
};

const ResultsControls: React.FC<ResultsControlsProps> = ({searchValue, selectedSort, limit, disabled, onSearchChange, onSortChange, onLimitChange}) => {
    return (
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
            <TextField
                fullWidth
                label="Поиск"
                value={searchValue}
                disabled={disabled}
                slotProps={{
                    input: {
                        endAdornment: disabled ? (
                            <InputAdornment position="end">
                                <CircularProgress color="inherit" size={20} />
                            </InputAdornment>
                        ) : null,
                    },
                }}
                onChange={(event) => onSearchChange(event.target.value)}
            />
            <FormControl sx={{ minWidth: 210 }} disabled={disabled}>
                <InputLabel id="match-sort-label">Сортировка</InputLabel>
                <Select
                    labelId="match-sort-label"
                    label="Сортировка"
                    value={selectedSort}
                    onChange={(event) => onSortChange(event.target.value)}
                >
                    {SORT_OPTIONS.map((option) => (
                        <MenuItem key={option.value} value={option.value}>
                            {option.label}
                        </MenuItem>
                    ))}
                </Select>
            </FormControl>
            <FormControl sx={{ minWidth: 160 }} disabled={disabled}>
                <InputLabel id="match-limit-label">Лимит</InputLabel>
                <Select
                    labelId="match-limit-label"
                    label="Лимит"
                    value={limit}
                    onChange={(event) => onLimitChange(Number(event.target.value))}
                >
                    {LIMIT_OPTIONS.map((limit) => (
                        <MenuItem key={limit} value={limit}>
                            {limit}
                        </MenuItem>
                    ))}
                </Select>
            </FormControl>
        </Stack>
    );
};

export default ResultsControls;