# Smartway Frontend

Frontend-приложение сервиса сопоставления отелей.

## Варианты использования

- Загрузка CSV-файлов
- Выбор алгоритма сопоставления
- Просмотр результатов анализа
- Отображение групп совпадений
- Информация о группе

## Технологии

- React
- TypeScript
- Vite
- React Router
- Material UI
- SCSS Modules
- pnpm
- React Hook Form + Zod
- React Query

## Установка

```bash
git clone https://github.com/VladFomin17/smartway-front.git

cd smartway-front

pnpm install
```
## Настройка окружения

Создайте в корне проекта файл .env и добавьте в него адрес API:
```
VITE_API_URL=http://localhost:8080
```

## Запуск

### Development

```bash
pnpm dev
```

### Production build

```bash
pnpm build
```

### Preview

```bash
pnpm preview
```

### Lint

```bash
pnpm lint
```

## Структура проекта

```
src/
├── app/
├── pages/
├── widgets/
├── features/
├── entities/
├── shared/
```