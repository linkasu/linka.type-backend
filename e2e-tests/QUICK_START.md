# Быстрый старт E2E тестов

## 1. Запуск сервера
```bash
# Из корневой директории проекта
docker-compose up -d
```

## 2. Установка зависимостей
```bash
cd e2e-tests
npm install
```

## 3. Настройка окружения
```bash
cp env.example .env
```

## 4. Запуск тестов
```bash
# Все тесты
./scripts/run-tests.sh

# Или напрямую
npm test

# Watch режим
npm run test:watch

# С покрытием
npm run test:coverage
```

## 5. Проверка готовности сервера
```bash
curl http://localhost:8081/api/health
```

## Полезные команды

### Запуск конкретного теста
```bash
npm test -- --testNamePattern="should register new user"
```

### Запуск конкретного файла
```bash
npm test -- auth.test.js
```

### Отладка
```bash
npm run test:debug
```

### Просмотр логов сервера
```bash
docker-compose logs server
```

### Остановка сервера
```bash
docker-compose down
```
