require('dotenv').config();

// Увеличиваем timeout для e2e тестов
jest.setTimeout(30000);

// Глобальные переменные для тестов
global.BASE_URL = process.env.BASE_URL || 'http://localhost:8081';
global.API_BASE = `${global.BASE_URL}/api`;

// Функция для ожидания
global.wait = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Функция для генерации случайного email
global.generateTestEmail = () => `test-${Date.now()}-${Math.random().toString(36).substr(2, 9)}@example.com`;

// Функция для генерации случайного пароля
global.generateTestPassword = () => `Password123!-${Math.random().toString(36).substr(2, 9)}`;

console.log('E2E Test Setup Complete');
console.log(`Base URL: ${global.BASE_URL}`);
console.log(`API Base: ${global.API_BASE}`);
