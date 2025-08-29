const ApiClient = require('../utils/api-client');
const WebSocketClient = require('../utils/websocket-client');

describe('WebSocket Tests', () => {
  let apiClient;
  let testUser;
  let authToken;
  let wsClient;
  let wsUrl;

  beforeEach(async () => {
    apiClient = new ApiClient();
    
    // Создаем тестового пользователя
    const email = generateTestEmail();
    const password = generateTestPassword();
    
    const registerResponse = await apiClient.register(email, password);
    testUser = registerResponse.user;
    authToken = registerResponse.token;
    apiClient.setToken(authToken);

    // Настраиваем WebSocket URL
    wsUrl = global.BASE_URL.replace('http', 'ws') + '/api/ws';
  });

  afterEach(async () => {
    if (wsClient) {
      wsClient.disconnect();
    }
    apiClient.clearToken();
  });

  describe('Connection', () => {
    test('should connect to WebSocket with valid token', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);

      await expect(wsClient.connect()).resolves.not.toThrow();
      expect(wsClient.isConnected()).toBe(true);
    });

    test('should reject connection without token', async () => {
      wsClient = new WebSocketClient(wsUrl, '');

      await expect(wsClient.connect()).rejects.toThrow();
    });

    test('should reject connection with invalid token', async () => {
      wsClient = new WebSocketClient(wsUrl, 'invalid-token');

      await expect(wsClient.connect()).rejects.toThrow();
    });

    test('should handle connection close gracefully', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      expect(wsClient.isConnected()).toBe(true);

      wsClient.disconnect();
      expect(wsClient.isConnected()).toBe(false);
    });
  });

  describe('Category Notifications', () => {
    test('should receive notification when creating category', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Создаем категорию
      const categoryTitle = 'WebSocket Test Category';
      const categoryResponse = await apiClient.createCategory(categoryTitle);

      // Ждем уведомление о создании категории
      const message = await wsClient.waitForMessage('category_update', 5000);

      expect(message).toHaveProperty('type', 'category_update');
      expect(message).toHaveProperty('payload');
      expect(message.payload).toHaveProperty('action', 'created');
      expect(message.payload).toHaveProperty('category');
      expect(message.payload.category).toHaveProperty('id', categoryResponse.id);
      expect(message.payload.category).toHaveProperty('title', categoryTitle);
      expect(message.payload.category).toHaveProperty('userId', testUser.id);
    });

    test('should receive notification when updating category', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Создаем категорию
      const categoryResponse = await apiClient.createCategory('Original Title');

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Обновляем категорию
      const updatedTitle = 'Updated Title';
      await apiClient.updateCategory(categoryResponse.id, updatedTitle);

      // Ждем уведомление об обновлении категории
      const message = await wsClient.waitForMessage('category_update', 5000);

      expect(message).toHaveProperty('type', 'category_update');
      expect(message).toHaveProperty('payload');
      expect(message.payload).toHaveProperty('action', 'updated');
      expect(message.payload).toHaveProperty('category');
      expect(message.payload.category).toHaveProperty('id', categoryResponse.id);
      expect(message.payload.category).toHaveProperty('title', updatedTitle);
      expect(message.payload.category).toHaveProperty('userId', testUser.id);
    });

    test('should receive notification when deleting category', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Создаем категорию
      const categoryResponse = await apiClient.createCategory('Category to Delete');

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Удаляем категорию
      await apiClient.deleteCategory(categoryResponse.id);

      // Ждем уведомление об удалении категории
      const message = await wsClient.waitForMessage('category_update', 5000);

      expect(message).toHaveProperty('type', 'category_update');
      expect(message).toHaveProperty('payload');
      expect(message.payload).toHaveProperty('action', 'deleted');
      expect(message.payload).toHaveProperty('categoryId', categoryResponse.id);
    });
  });

  describe('Statement Notifications', () => {
    let testCategory;

    beforeEach(async () => {
      // Создаем тестовую категорию для statements
      const categoryResponse = await apiClient.createCategory('Test Category');
      testCategory = categoryResponse;
    });

    test('should receive notification when creating statement', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Создаем statement
      const statementTitle = 'WebSocket Test Statement';
      const statementResponse = await apiClient.createStatement(statementTitle, testCategory.id);

      // Ждем уведомление о создании statement
      const message = await wsClient.waitForMessage('statement_update', 5000);

      expect(message).toHaveProperty('type', 'statement_update');
      expect(message).toHaveProperty('payload');
      expect(message.payload).toHaveProperty('action', 'created');
      expect(message.payload).toHaveProperty('statement');
      expect(message.payload.statement).toHaveProperty('id', statementResponse.id);
      expect(message.payload.statement).toHaveProperty('text', statementTitle);
      expect(message.payload.statement).toHaveProperty('userId', testUser.id);
      expect(message.payload.statement).toHaveProperty('categoryId', testCategory.id);
    });

    test('should receive notification when updating statement', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Создаем statement
      const statementResponse = await apiClient.createStatement('Original Statement', testCategory.id);

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Обновляем statement
      const updatedTitle = 'Updated Statement';
      await apiClient.updateStatement(statementResponse.id, updatedTitle, testCategory.id);

      // Ждем уведомление об обновлении statement
      const message = await wsClient.waitForMessage('statement_update', 5000);

      expect(message).toHaveProperty('type', 'statement_update');
      expect(message).toHaveProperty('payload');
      expect(message.payload).toHaveProperty('action', 'updated');
      expect(message.payload).toHaveProperty('statement');
      expect(message.payload.statement).toHaveProperty('id', statementResponse.id);
      expect(message.payload.statement).toHaveProperty('text', updatedTitle);
      expect(message.payload.statement).toHaveProperty('userId', testUser.id);
      expect(message.payload.statement).toHaveProperty('categoryId', testCategory.id);
    });

    test('should receive notification when deleting statement', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Создаем statement
      const statementResponse = await apiClient.createStatement('Statement to Delete', testCategory.id);

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Удаляем statement
      await apiClient.deleteStatement(statementResponse.id);

      // Ждем уведомление об удалении statement
      const message = await wsClient.waitForMessage('statement_update', 5000);

      expect(message).toHaveProperty('type', 'statement_update');
      expect(message).toHaveProperty('payload');
      expect(message.payload).toHaveProperty('action', 'deleted');
      expect(message.payload).toHaveProperty('statementId', statementResponse.id);
    });
  });

  describe('Multiple Users', () => {
    test('should only receive notifications for own actions', async () => {
      // Создаем первого пользователя и подключаем WebSocket
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Создаем второго пользователя
      const email2 = generateTestEmail();
      const password2 = generateTestPassword();
      const user2Response = await apiClient.register(email2, password2);
      const apiClient2 = new ApiClient();
      apiClient2.setToken(user2Response.token);

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Второй пользователь создает категорию
      await apiClient2.createCategory('User 2 Category');

      // Первый пользователь не должен получить уведомление
      await expect(wsClient.waitForMessage('category_update', 3000)).rejects.toThrow();

      // Первый пользователь создает категорию
      await apiClient.createCategory('User 1 Category');

      // Теперь должен получить уведомление
      const message = await wsClient.waitForMessage('category_update', 5000);
      expect(message.payload.category.title).toBe('User 1 Category');
    });
  });

  describe('Message Handling', () => {
    test('should handle multiple messages correctly', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      // Создаем несколько категорий быстро
      await apiClient.createCategory('Category 1');
      await apiClient.createCategory('Category 2');
      await apiClient.createCategory('Category 3');

      // Ждем немного для получения всех сообщений
      await wait(1000);

      const messages = wsClient.getMessages();
      const categoryMessages = messages.filter(msg => msg.type === 'category_update');

      expect(categoryMessages.length).toBeGreaterThanOrEqual(3);

      const titles = categoryMessages
        .filter(msg => msg.payload.action === 'created')
        .map(msg => msg.payload.category.title);

      expect(titles).toContain('Category 1');
      expect(titles).toContain('Category 2');
      expect(titles).toContain('Category 3');
    });

    test('should handle message with specific payload', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      // Очищаем предыдущие сообщения
      wsClient.clearMessages();

      const categoryTitle = 'Specific Category';
      await apiClient.createCategory(categoryTitle);

      // Ждем сообщение с конкретным заголовком
      const message = await wsClient.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.category && payload.category.title === categoryTitle,
        5000
      );

      expect(message.payload.category.title).toBe(categoryTitle);
    });
  });

  describe('Reconnection', () => {
    test('should handle reconnection gracefully', async () => {
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();

      expect(wsClient.isConnected()).toBe(true);

      // Отключаемся
      wsClient.disconnect();
      expect(wsClient.isConnected()).toBe(false);

      // Подключаемся снова
      await wsClient.connect();
      expect(wsClient.isConnected()).toBe(true);

      // Проверяем, что можем получать сообщения
      wsClient.clearMessages();
      await apiClient.createCategory('Reconnection Test');

      const message = await wsClient.waitForMessage('category_update', 5000);
      expect(message.payload.category.title).toBe('Reconnection Test');
    });
  });
});
