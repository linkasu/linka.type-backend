const ApiClient = require('../utils/api-client');
const WebSocketClient = require('../utils/websocket-client');

describe('Integration Tests', () => {
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

  describe('Complete User Workflow', () => {
    test('should complete full user workflow with WebSocket notifications', async () => {
      // Подключаемся к WebSocket
      wsClient = new WebSocketClient(wsUrl, authToken);
      await wsClient.connect();
      wsClient.clearMessages();

      // 1. Создаем категории
      const category1 = await apiClient.createCategory('Работа');
      const category2 = await apiClient.createCategory('Личное');

      // Ждем уведомления о создании категорий
      const category1Message = await wsClient.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.category && payload.category.id === category1.id,
        5000
      );
      const category2Message = await wsClient.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.category && payload.category.id === category2.id,
        5000
      );

      expect(category1Message.payload.category.title).toBe('Работа');
      expect(category2Message.payload.category.title).toBe('Личное');

      // 2. Создаем statements в разных категориях
      const statement1 = await apiClient.createStatement('Выполнить проект', category1.id);
      const statement2 = await apiClient.createStatement('Купить продукты', category2.id);
      const statement3 = await apiClient.createStatement('Позвонить маме', category2.id);

      // Ждем уведомления о создании statements
      await wsClient.waitForMessageWithPayload(
        'statement_update',
        (payload) => payload.statement && payload.statement.id === statement1.id,
        5000
      );
      await wsClient.waitForMessageWithPayload(
        'statement_update',
        (payload) => payload.statement && payload.statement.id === statement2.id,
        5000
      );
      await wsClient.waitForMessageWithPayload(
        'statement_update',
        (payload) => payload.statement && payload.statement.id === statement3.id,
        5000
      );

      // 3. Получаем все данные и проверяем
      const categoriesResponse = await apiClient.getCategories();
      const statementsResponse = await apiClient.getStatements();

      expect(categoriesResponse.categories).toHaveLength(2);
      expect(statementsResponse.statements).toHaveLength(3);

      // 4. Обновляем данные
      const updatedCategoryTitle = 'Работа и проекты';
      await apiClient.updateCategory(category1.id, updatedCategoryTitle);

      const updatedStatementTitle = 'Выполнить проект до пятницы';
      await apiClient.updateStatement(statement1.id, updatedStatementTitle, category1.id);

      // Ждем уведомления об обновлениях
      await wsClient.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.category && payload.category.id === category1.id && payload.action === 'updated',
        5000
      );
      await wsClient.waitForMessageWithPayload(
        'statement_update',
        (payload) => payload.statement && payload.statement.id === statement1.id && payload.action === 'updated',
        5000
      );

      // 5. Проверяем обновленные данные
      const updatedCategory = await apiClient.getCategory(category1.id);
      const updatedStatement = await apiClient.getStatement(statement1.id);

      expect(updatedCategory.title).toBe(updatedCategoryTitle);
      expect(updatedStatement.text).toBe(updatedStatementTitle);

      // 6. Удаляем данные
      await apiClient.deleteStatement(statement3.id);
      await apiClient.deleteCategory(category2.id);

      // Ждем уведомления об удалениях
      await wsClient.waitForMessageWithPayload(
        'statement_update',
        (payload) => payload.statementId === statement3.id && payload.action === 'deleted',
        5000
      );
      await wsClient.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.categoryId === category2.id && payload.action === 'deleted',
        5000
      );

      // 7. Проверяем финальное состояние
      const finalCategories = await apiClient.getCategories();
      const finalStatements = await apiClient.getStatements();

      expect(finalCategories.categories).toHaveLength(1);
      expect(finalStatements.statements).toHaveLength(1);
      expect(finalCategories.categories[0].title).toBe(updatedCategoryTitle);
      expect(finalStatements.statements[0].text).toBe(updatedStatementTitle);
    });
  });

  describe('Multi-User Scenario', () => {
    test('should handle multiple users with data isolation', async () => {
      // Создаем второго пользователя
      const email2 = generateTestEmail();
      const password2 = generateTestPassword();
      const user2Response = await apiClient.register(email2, password2);
      const apiClient2 = new ApiClient();
      apiClient2.setToken(user2Response.token);

      // Создаем WebSocket подключения для обоих пользователей
      const wsClient1 = new WebSocketClient(wsUrl, authToken);
      const wsClient2 = new WebSocketClient(wsUrl, user2Response.token);
      
      await wsClient1.connect();
      await wsClient2.connect();
      
      wsClient1.clearMessages();
      wsClient2.clearMessages();

      // Пользователь 1 создает данные
      const user1Category = await apiClient.createCategory('User 1 Category');
      const user1Statement = await apiClient.createStatement('User 1 Statement', user1Category.id);

      // Пользователь 2 создает данные
      const user2Category = await apiClient2.createCategory('User 2 Category');
      const user2Statement = await apiClient2.createStatement('User 2 Statement', user2Category.id);

      // Ждем уведомления
      await wsClient1.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.category && payload.category.id === user1Category.id,
        5000
      );
      await wsClient2.waitForMessageWithPayload(
        'category_update',
        (payload) => payload.category && payload.category.id === user2Category.id,
        5000
      );

      // Проверяем изоляцию данных
      const user1Categories = await apiClient.getCategories();
      const user1Statements = await apiClient.getStatements();
      const user2Categories = await apiClient2.getCategories();
      const user2Statements = await apiClient2.getStatements();

      expect(user1Categories.categories).toHaveLength(1);
      expect(user1Statements.statements).toHaveLength(1);
      expect(user2Categories.categories).toHaveLength(1);
      expect(user2Statements.statements).toHaveLength(1);

      expect(user1Categories.categories[0].title).toBe('User 1 Category');
      expect(user2Categories.categories[0].title).toBe('User 2 Category');

      // Проверяем, что пользователи не видят данные друг друга
      await expect(apiClient.getCategory(user2Category.id)).rejects.toThrow();
      await expect(apiClient2.getCategory(user1Category.id)).rejects.toThrow();

      // Проверяем, что WebSocket уведомления изолированы
      const user1Messages = wsClient1.getMessages();
      const user2Messages = wsClient2.getMessages();

      const user1CategoryMessages = user1Messages.filter(
        msg => msg.type === 'category_update' && msg.payload.category && msg.payload.category.id === user1Category.id
      );
      const user2CategoryMessages = user2Messages.filter(
        msg => msg.type === 'category_update' && msg.payload.category && msg.payload.category.id === user2Category.id
      );

      expect(user1CategoryMessages.length).toBeGreaterThan(0);
      expect(user2CategoryMessages.length).toBeGreaterThan(0);

      // Пользователь 1 не должен получать уведомления о действиях пользователя 2
      const user1User2Messages = user1Messages.filter(
        msg => msg.type === 'category_update' && msg.payload.category && msg.payload.category.id === user2Category.id
      );
      expect(user1User2Messages.length).toBe(0);

      wsClient1.disconnect();
      wsClient2.disconnect();
    });
  });

  describe('Error Recovery', () => {
    test('should handle errors gracefully and maintain data consistency', async () => {
      // Создаем валидные данные
      const category = await apiClient.createCategory('Test Category');
      const statement = await apiClient.createStatement('Test Statement', category.id);

      // Пытаемся создать statement с несуществующей категорией
      await expect(apiClient.createStatement('Invalid Statement', 'non-existent-category')).rejects.toThrow();

      // Проверяем, что валидные данные остались нетронутыми
      const categories = await apiClient.getCategories();
      const statements = await apiClient.getStatements();

      expect(categories.categories).toHaveLength(1);
      expect(statements.statements).toHaveLength(1);
      expect(categories.categories[0].id).toBe(category.id);
      expect(statements.statements[0].id).toBe(statement.id);

      // Пытаемся обновить несуществующий statement
      await expect(apiClient.updateStatement('non-existent-statement', 'New Title', category.id)).rejects.toThrow();

      // Проверяем, что данные все еще валидны
      const updatedStatement = await apiClient.getStatement(statement.id);
      expect(updatedStatement.text).toBe('Test Statement');

      // Пытаемся удалить несуществующий statement
      await expect(apiClient.deleteStatement('non-existent-statement')).rejects.toThrow();

      // Проверяем, что валидный statement все еще существует
      const finalStatements = await apiClient.getStatements();
      expect(finalStatements.statements).toHaveLength(1);
    });
  });

  describe('Performance and Load', () => {
    test('should handle multiple rapid operations', async () => {
      const operations = [];
      const numOperations = 10;

      // Создаем много категорий быстро
      for (let i = 0; i < numOperations; i++) {
        operations.push(apiClient.createCategory(`Category ${i}`));
      }

      // Ждем завершения всех операций
      const categories = await Promise.all(operations);

      expect(categories).toHaveLength(numOperations);

      // Проверяем, что все категории созданы
      const allCategories = await apiClient.getCategories();
      expect(allCategories.categories).toHaveLength(numOperations);

      // Создаем statements для каждой категории
      const statementOperations = [];
      for (let i = 0; i < numOperations; i++) {
        statementOperations.push(apiClient.createStatement(`Statement ${i}`, categories[i].id));
      }

      const statements = await Promise.all(statementOperations);
      expect(statements).toHaveLength(numOperations);

      // Проверяем, что все statements созданы
      const allStatements = await apiClient.getStatements();
      expect(allStatements.statements).toHaveLength(numOperations);
    });
  });

  describe('Data Integrity', () => {
    test('should maintain referential integrity', async () => {
      // Создаем категорию и statement
      const category = await apiClient.createCategory('Test Category');
      const statement = await apiClient.createStatement('Test Statement', category.id);

      // Проверяем, что statement связан с правильной категорией
      const retrievedStatement = await apiClient.getStatement(statement.id);
      expect(retrievedStatement.categoryId).toBe(category.id);

      // Создаем новую категорию
      const newCategory = await apiClient.createCategory('New Category');

      // Перемещаем statement в новую категорию
      await apiClient.updateStatement(statement.id, 'Updated Statement', newCategory.id);

      // Проверяем, что statement теперь связан с новой категорией
      const updatedStatement = await apiClient.getStatement(statement.id);
      expect(updatedStatement.categoryId).toBe(newCategory.id);
      expect(updatedStatement.text).toBe('Updated Statement');

      // Удаляем старую категорию
      await apiClient.deleteCategory(category.id);

      // Проверяем, что statement все еще существует и связан с новой категорией
      const finalStatement = await apiClient.getStatement(statement.id);
      expect(finalStatement.categoryId).toBe(newCategory.id);
    });
  });
});
