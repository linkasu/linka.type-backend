# Инструкция по развертыванию на сервер linka.su

## Подготовка

1. Убедитесь, что у вас есть SSH доступ к серверу (пользователь: `aacidov`)
2. Проверьте, что Docker и Docker Compose установлены на сервере
3. Настройте переменные окружения в файле `.env` на сервере

## Автоматическое развертывание

### Через Docker Hub (рекомендуется)

Полное развертывание с загрузкой в Docker Hub:

```bash
./deploy-via-dockerhub.sh
```

### Быстрое обновление на сервере

Если образ уже загружен в Docker Hub:

```bash
./update-server-only.sh
```

### Прямая передача файлов

```bash
./deploy-to-server.sh
```

## Ручное развертывание

### 1. Создание образа локально

```bash
docker compose build
docker save linka.type-backend:latest -o linka-type-backend-latest.tar
```

### 2. Загрузка файлов на сервер

```bash
scp linka-type-backend-latest.tar root@linka.su:/home/
scp docker-compose.yml root@linka.su:/home/
scp env.server.example root@linka.su:/home/
```

### 3. Настройка на сервере

Подключитесь к серверу:

```bash
ssh root@linka.su
cd /home
```

Настройте переменные окружения:

```bash
cp env.server.example .env
nano .env  # Отредактируйте значения
```

### 4. Развертывание

```bash
# Остановите существующие контейнеры
docker compose down

# Удалите старые образы
docker images | grep linka | awk '{print $3}' | xargs -r docker rmi -f

# Загрузите новый образ
docker load -i linka-type-backend-latest.tar

# Запустите сервисы
docker compose up -d
```

### 5. Проверка

```bash
# Статус контейнеров
docker ps

# Логи сервера
docker compose logs server

# Проверка здоровья
docker compose ps
```

## Docker Hub

Образ автоматически загружается в Docker Hub: `aacidov/linka-type-backend:latest`

Для загрузки в Docker Hub:
```bash
docker login
docker push aacidov/linka-type-backend:latest
```

## Структура файлов на сервере

```
/home/linka.type-backend/
├── docker-compose.yml             # Конфигурация сервисов
├── .env                          # Переменные окружения
├── firebase.json                 # Конфигурация Firebase
└── deploy.sh                     # Скрипт развертывания
```

## Полезные команды

```bash
# Просмотр логов
docker compose logs -f server

# Перезапуск сервиса
docker compose restart server

# Обновление без простоя
docker compose up -d --no-deps server

# Очистка неиспользуемых образов
docker image prune -f
```

## Устранение неполадок

1. **Порт занят**: Проверьте `netstat -tlnp | grep 8080`
2. **Проблемы с БД**: Проверьте логи `docker compose logs db`
3. **Проблемы с правами**: Убедитесь, что Docker может читать файлы
4. **Проблемы с сетью**: Проверьте firewall и настройки Docker

## Мониторинг

```bash
# Статистика использования ресурсов
docker stats

# Проверка здоровья сервисов
curl -f http://localhost:8080/api/health
```
