const ApiClient = require('../utils/api-client');

describe('Simple API Tests', () => {
  let apiClient;
  let testUser;
  let authToken;

  beforeEach(async () => {
    apiClient = new ApiClient();
    
    // Создаем тестового пользователя
    const email = `test-${Date.now()}@example.com`;
    const password = 'StrongPassword123!';
    
    try {
      const registerResponse = await apiClient.register(email, password);
      testUser = registerResponse.user;
      authToken = registerResponse.token;
      apiClient.setToken(authToken);
    } catch (error) {
      console.log('Registration failed, trying login:', error.message);
      // Если регистрация не удалась, попробуем логин
      const loginResponse = await apiClient.login(email, password);
      testUser = loginResponse.user;
      authToken = loginResponse.token;
      apiClient.setToken(authToken);
    }
  });

  afterEach(async () => {
    apiClient.clearToken();
  });

  test('should register and login user', async () => {
    expect(testUser).toBeDefined();
    expect(authToken).toBeDefined();
    expect(testUser.email).toContain('@example.com');
  });

  test('should create and get categories', async () => {
    // Создаем категорию
    const categoryTitle = 'Test Category';
    const categoryResponse = await apiClient.createCategory(categoryTitle);
    
    expect(categoryResponse).toBeDefined();
    expect(categoryResponse.title).toBe(categoryTitle);
    expect(categoryResponse.userId).toBe(testUser.id);

    // Получаем все категории
    const categoriesResponse = await apiClient.getCategories();
    expect(categoriesResponse.categories).toBeDefined();
    expect(categoriesResponse.categories.length).toBeGreaterThan(0);
    
    // Проверяем, что наша категория есть в списке
    const foundCategory = categoriesResponse.categories.find(c => c.id === categoryResponse.id);
    expect(foundCategory).toBeDefined();
    expect(foundCategory.title).toBe(categoryTitle);
  });

  test('should create and get statements', async () => {
    // Создаем категорию
    const categoryResponse = await apiClient.createCategory('Test Category');
    
    // Создаем statement
    const statementTitle = 'Test Statement';
    const statementResponse = await apiClient.createStatement(statementTitle, categoryResponse.id);
    
    expect(statementResponse).toBeDefined();
    expect(statementResponse.text).toBe(statementTitle);
    expect(statementResponse.userId).toBe(testUser.id);
    expect(statementResponse.categoryId).toBe(categoryResponse.id);

    // Получаем все statements
    const statementsResponse = await apiClient.getStatements();
    expect(statementsResponse.statements).toBeDefined();
    expect(statementsResponse.statements.length).toBeGreaterThan(0);
    
    // Проверяем, что наш statement есть в списке
    const foundStatement = statementsResponse.statements.find(s => s.id === statementResponse.id);
    expect(foundStatement).toBeDefined();
    expect(foundStatement.text).toBe(statementTitle);
  });

  test('should update category', async () => {
    // Создаем категорию
    const categoryResponse = await apiClient.createCategory('Original Title');
    
    // Обновляем категорию
    const updatedTitle = 'Updated Title';
    const updatedCategory = await apiClient.updateCategory(categoryResponse.id, updatedTitle);
    
    expect(updatedCategory).toBeDefined();
    expect(updatedCategory.title).toBe(updatedTitle);
    expect(updatedCategory.id).toBe(categoryResponse.id);
  });

  test('should update statement', async () => {
    // Создаем категорию
    const categoryResponse = await apiClient.createCategory('Test Category');
    
    // Создаем statement
    const statementResponse = await apiClient.createStatement('Original Statement', categoryResponse.id);
    
    // Обновляем statement
    const updatedTitle = 'Updated Statement';
    const updatedStatement = await apiClient.updateStatement(statementResponse.id, updatedTitle, categoryResponse.id);
    
    expect(updatedStatement).toBeDefined();
    expect(updatedStatement.text).toBe(updatedTitle);
    expect(updatedStatement.id).toBe(statementResponse.id);
  });

  test('should delete category', async () => {
    // Создаем категорию
    const categoryResponse = await apiClient.createCategory('Category to Delete');
    
    // Удаляем категорию
    const deleteResponse = await apiClient.deleteCategory(categoryResponse.id);
    expect(deleteResponse).toBeDefined();
    expect(deleteResponse.message).toContain('deleted');
  });

  test('should delete statement', async () => {
    // Создаем категорию
    const categoryResponse = await apiClient.createCategory('Test Category');
    
    // Создаем statement
    const statementResponse = await apiClient.createStatement('Statement to Delete', categoryResponse.id);
    
    // Удаляем statement
    const deleteResponse = await apiClient.deleteStatement(statementResponse.id);
    expect(deleteResponse).toBeDefined();
    expect(deleteResponse.message).toContain('deleted');
  });
});
