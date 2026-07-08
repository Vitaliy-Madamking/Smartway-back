const pluralRules = new Intl.PluralRules('ru-RU');

interface PluralForms {
    one: string;
    few: string;
    many: string;
}

export const getPluralForm = (count: number, forms: PluralForms): string => {
    const rule = pluralRules.select(count) as keyof PluralForms;
    return forms[rule];
};