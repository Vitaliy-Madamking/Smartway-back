import { getPluralForm } from '@/shared/lib';

const HOTEL_RECORD_FORMS = {
    one: "запись",
    few: "записи",
    many: "записей"
}

export const getPluralHotel = (count: number) => {
    const word = getPluralForm(count, HOTEL_RECORD_FORMS);
    return `${count} ${word}`;
}