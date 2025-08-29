FROM node:18-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем package.json и package-lock.json
COPY package*.json ./

# Устанавливаем зависимости
RUN npm install

# Копируем исходный код тестов
COPY . .

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nodejs -u 1001

# Создаем директорию для результатов тестов
RUN mkdir -p /app/test-results

# Меняем владельца файлов
RUN chown -R nodejs:nodejs /app
RUN chmod 777 /app/test-results
USER nodejs

# Открываем порт (если нужно)
EXPOSE 3000

# Устанавливаем переменные окружения
ENV NODE_ENV=test
ENV BASE_URL=http://server:8080

# Команда по умолчанию
CMD ["npm", "test"]
