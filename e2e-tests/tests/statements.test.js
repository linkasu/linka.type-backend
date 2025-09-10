const ApiClient = require('../utils/api-client');

describe('Statements Tests', () => {
  let apiClient;
  let testUser;
  let authToken;
  let testCategory;

  beforeEach(async () => {
    apiClient = new ApiClient();
    
    // Создаем тестового пользователя
    const email = generateTestEmail();
    const password = generateTestPassword();
    
    const registerResponse = await apiClient.register(email, password);
    testUser = registerResponse.user;
    authToken = registerResponse.token;
    apiClient.setToken(authToken);

    // Создаем тестовую категорию
    const categoryResponse = await apiClient.createCategory('Test Category');
    testCategory = categoryResponse;
  });

  afterEach(() => {
    apiClient.clearToken();
  });

  describe('CRUD Operations', () => {
    test('should create statement successfully', async () => {
      const statementTitle = 'Test Statement';

      const response = await apiClient.createStatement(statementTitle, testCategory.id);

      expect(response).toHaveProperty('id');
      expect(response).toHaveProperty('text');
      expect(response).toHaveProperty('userId');
      expect(response).toHaveProperty('categoryId');
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
      expect(response.text).toBe(statementTitle);
      expect(response.userId).toBe(testUser.id);
      expect(response.categoryId).toBe(testCategory.id);
      expect(typeof response.id).toBe('string');
      expect(response.id.length).toBeGreaterThan(0);
    });

    test('should get all statements for user', async () => {
      const statementTitle1 = 'Statement 1';
      const statementTitle2 = 'Statement 2';

      // Создаем два statement
      await apiClient.createStatement(statementTitle1, testCategory.id);
      await apiClient.createStatement(statementTitle2, testCategory.id);

      const response = await apiClient.getStatements();

      expect(response).toHaveProperty('statements');
      expect(Array.isArray(response.statements)).toBe(true);
      expect(response.statements.length).toBe(2);

      const titles = response.statements.map(stmt => stmt.text);
      expect(titles).toContain(statementTitle1);
      expect(titles).toContain(statementTitle2);

      // Проверяем структуру каждого statement
      response.statements.forEach(statement => {
        expect(statement).toHaveProperty('id');
        expect(statement).toHaveProperty('text');
        expect(statement).toHaveProperty('userId');
        expect(statement).toHaveProperty('categoryId');
        expect(statement).toHaveProperty('createdAt');
        expect(statement).toHaveProperty('updatedAt');
        expect(statement.userId).toBe(testUser.id);
        expect(statement.categoryId).toBe(testCategory.id);
      });
    });

    test('should get specific statement by id', async () => {
      const statementTitle = 'Test Statement';

      const createResponse = await apiClient.createStatement(statementTitle, testCategory.id);
      const statementId = createResponse.id;

      const response = await apiClient.getStatement(statementId);

      expect(response).toHaveProperty('id');
      expect(response).toHaveProperty('text');
      expect(response).toHaveProperty('userId');
      expect(response).toHaveProperty('categoryId');
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
      expect(response.id).toBe(statementId);
      expect(response.text).toBe(statementTitle);
      expect(response.userId).toBe(testUser.id);
      expect(response.categoryId).toBe(testCategory.id);
    });

    test('should update statement successfully', async () => {
      const originalTitle = 'Original Statement';
      const updatedTitle = 'Updated Statement';

      const createResponse = await apiClient.createStatement(originalTitle, testCategory.id);
      const statementId = createResponse.id;

      const response = await apiClient.updateStatement(statementId, updatedTitle, testCategory.id);

      expect(response).toHaveProperty('id');
      expect(response).toHaveProperty('text');
      expect(response).toHaveProperty('userId');
      expect(response).toHaveProperty('categoryId');
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
      expect(response.id).toBe(statementId);
      expect(response.text).toBe(updatedTitle);
      expect(response.userId).toBe(testUser.id);
      expect(response.categoryId).toBe(testCategory.id);
    });

    test('should delete statement successfully', async () => {
      const statementTitle = 'Statement to Delete';

      const createResponse = await apiClient.createStatement(statementTitle, testCategory.id);
      const statementId = createResponse.id;

      const response = await apiClient.deleteStatement(statementId);

      expect(response).toHaveProperty('message');
      expect(response.message).toContain('deleted successfully');

      // Проверяем, что statement действительно удален
      await expect(apiClient.getStatement(statementId)).rejects.toThrow();
    });

    test('should update statement with different category', async () => {
      const statementTitle = 'Test Statement';
      const newCategoryTitle = 'New Category';

      // Создаем statement в первой категории
      const createResponse = await apiClient.createStatement(statementTitle, testCategory.id);
      const statementId = createResponse.id;

      // Создаем новую категорию
      const newCategoryResponse = await apiClient.createCategory(newCategoryTitle);

      // Перемещаем statement в новую категорию
      const response = await apiClient.updateStatement(statementId, statementTitle, newCategoryResponse.id);

      expect(response.categoryId).toBe(newCategoryResponse.id);
      expect(response.text).toBe(statementTitle);
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
    });
  });

  describe('Validation and Error Handling', () => {
    test('should reject creating statement without title', async () => {
      await expect(apiClient.createStatement('', testCategory.id)).rejects.toThrow();
    });

    test('should reject creating statement with empty title', async () => {
      await expect(apiClient.createStatement('   ', testCategory.id)).rejects.toThrow();
    });

    test('should reject creating statement without categoryId', async () => {
      await expect(apiClient.createStatement('Test Statement', '')).rejects.toThrow();
    });

    test('should reject creating statement with non-existent categoryId', async () => {
      const nonExistentCategoryId = 'category_nonexistent';

      await expect(apiClient.createStatement('Test Statement', nonExistentCategoryId)).rejects.toThrow();
    });

    test('should reject updating non-existent statement', async () => {
      const nonExistentId = 'statement_nonexistent';

      await expect(apiClient.updateStatement(nonExistentId, 'New Title', testCategory.id)).rejects.toThrow();
    });

    test('should reject deleting non-existent statement', async () => {
      const nonExistentId = 'statement_nonexistent';

      await expect(apiClient.deleteStatement(nonExistentId)).rejects.toThrow();
    });

    test('should reject getting non-existent statement', async () => {
      const nonExistentId = 'statement_nonexistent';

      await expect(apiClient.getStatement(nonExistentId)).rejects.toThrow();
    });
  });

  describe('Authorization', () => {
    test('should reject access without token', async () => {
      apiClient.clearToken();

      await expect(apiClient.createStatement('Test Statement', testCategory.id)).rejects.toThrow();
      await expect(apiClient.getStatements()).rejects.toThrow();
    });

    test('should reject access with invalid token', async () => {
      apiClient.setToken('invalid-token');

      await expect(apiClient.createStatement('Test Statement', testCategory.id)).rejects.toThrow();
      await expect(apiClient.getStatements()).rejects.toThrow();
    });
  });

  describe('Data Isolation', () => {
    test('should only return statements for authenticated user', async () => {
      // Создаем statement для первого пользователя
      await apiClient.createStatement('User 1 Statement', testCategory.id);

      // Создаем второго пользователя
      const email2 = generateTestEmail();
      const password2 = generateTestPassword();
      const user2Response = await apiClient.register(email2, password2);
      
      // Переключаемся на второго пользователя
      apiClient.setToken(user2Response.token);
      
      // Создаем категорию для второго пользователя
      const category2Response = await apiClient.createCategory('User 2 Category');
      
      // Создаем statement для второго пользователя
      await apiClient.createStatement('User 2 Statement', category2Response.id);

      // Получаем statements второго пользователя
      const response = await apiClient.getStatements();

      expect(response.statements.length).toBe(1);
      expect(response.statements[0].text).toBe('User 2 Statement');
      expect(response.statements[0].userId).toBe(user2Response.user.id);
    });

    test('should not allow access to other user statements', async () => {
      // Создаем statement для первого пользователя
      const statementResponse = await apiClient.createStatement('User 1 Statement', testCategory.id);

      // Создаем второго пользователя
      const email2 = generateTestEmail();
      const password2 = generateTestPassword();
      const user2Response = await apiClient.register(email2, password2);
      
      // Переключаемся на второго пользователя
      apiClient.setToken(user2Response.token);
      
      // Пытаемся получить statement первого пользователя
      await expect(apiClient.getStatement(statementResponse.id)).rejects.toThrow();
    });
  });

  describe('Category Relationships', () => {
    test('should maintain category relationship after updates', async () => {
      const statementTitle = 'Test Statement';

      // Создаем statement
      const createResponse = await apiClient.createStatement(statementTitle, testCategory.id);
      const statementId = createResponse.id;

      // Проверяем, что statement связан с правильной категорией
      let response = await apiClient.getStatement(statementId);
      expect(response.categoryId).toBe(testCategory.id);

      // Создаем новую категорию
      const newCategoryResponse = await apiClient.createCategory('New Category');

      // Обновляем statement с новой категорией
      response = await apiClient.updateStatement(statementId, statementTitle, newCategoryResponse.id);
      expect(response.categoryId).toBe(newCategoryResponse.id);

      // Проверяем, что statement теперь связан с новой категорией
      response = await apiClient.getStatement(statementId);
      expect(response.categoryId).toBe(newCategoryResponse.id);
    });
  });
});
