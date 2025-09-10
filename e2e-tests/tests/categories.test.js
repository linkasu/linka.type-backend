const ApiClient = require('../utils/api-client');

describe('Categories Tests', () => {
  let apiClient;
  let testUser;
  let authToken;

  beforeEach(async () => {
    apiClient = new ApiClient();
    
    // Создаем тестового пользователя
    const email = generateTestEmail();
    const password = generateTestPassword();
    
    const registerResponse = await apiClient.register(email, password);
    testUser = registerResponse.user;
    authToken = registerResponse.token;
    apiClient.setToken(authToken);
  });

  afterEach(() => {
    apiClient.clearToken();
  });

  describe('CRUD Operations', () => {
    test('should create category successfully', async () => {
      const categoryTitle = 'Test Category';

      const response = await apiClient.createCategory(categoryTitle);

      expect(response).toHaveProperty('id');
      expect(response).toHaveProperty('title');
      expect(response).toHaveProperty('userId');
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
      expect(response.title).toBe(categoryTitle);
      expect(response.userId).toBe(testUser.id);
      expect(typeof response.id).toBe('string');
      expect(response.id.length).toBeGreaterThan(0);
    });

    test('should get all categories for user', async () => {
      const categoryTitle1 = 'Category 1';
      const categoryTitle2 = 'Category 2';

      // Создаем две категории
      await apiClient.createCategory(categoryTitle1);
      await apiClient.createCategory(categoryTitle2);

      const response = await apiClient.getCategories();

      expect(response).toHaveProperty('categories');
      expect(Array.isArray(response.categories)).toBe(true);
      expect(response.categories.length).toBe(2);

      const titles = response.categories.map(cat => cat.title);
      expect(titles).toContain(categoryTitle1);
      expect(titles).toContain(categoryTitle2);

      // Проверяем структуру каждой категории
      response.categories.forEach(category => {
        expect(category).toHaveProperty('id');
        expect(category).toHaveProperty('title');
        expect(category).toHaveProperty('userId');
        expect(category).toHaveProperty('createdAt');
        expect(category).toHaveProperty('updatedAt');
        expect(category.userId).toBe(testUser.id);
      });
    });

    test('should get specific category by id', async () => {
      const categoryTitle = 'Test Category';

      const createResponse = await apiClient.createCategory(categoryTitle);
      const categoryId = createResponse.id;

      const response = await apiClient.getCategory(categoryId);

      expect(response).toHaveProperty('id');
      expect(response).toHaveProperty('title');
      expect(response).toHaveProperty('userId');
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
      expect(response.id).toBe(categoryId);
      expect(response.title).toBe(categoryTitle);
      expect(response.userId).toBe(testUser.id);
    });

    test('should update category successfully', async () => {
      const originalTitle = 'Original Title';
      const updatedTitle = 'Updated Title';

      const createResponse = await apiClient.createCategory(originalTitle);
      const categoryId = createResponse.id;

      const response = await apiClient.updateCategory(categoryId, updatedTitle);

      expect(response).toHaveProperty('id');
      expect(response).toHaveProperty('title');
      expect(response).toHaveProperty('userId');
      expect(response).toHaveProperty('createdAt');
      expect(response).toHaveProperty('updatedAt');
      expect(response.id).toBe(categoryId);
      expect(response.title).toBe(updatedTitle);
      expect(response.userId).toBe(testUser.id);
    });

    test('should delete category successfully', async () => {
      const categoryTitle = 'Category to Delete';

      const createResponse = await apiClient.createCategory(categoryTitle);
      const categoryId = createResponse.id;

      const response = await apiClient.deleteCategory(categoryId);

      expect(response).toHaveProperty('message');
      expect(response.message).toContain('deleted successfully');

      // Проверяем, что категория действительно удалена
      await expect(apiClient.getCategory(categoryId)).rejects.toThrow();
    });
  });

  describe('Validation and Error Handling', () => {
    test('should reject creating category without title', async () => {
      await expect(apiClient.createCategory('')).rejects.toThrow();
    });

    test('should reject creating category with empty title', async () => {
      await expect(apiClient.createCategory('   ')).rejects.toThrow();
    });

    test('should reject updating non-existent category', async () => {
      const nonExistentId = 'category_nonexistent';

      await expect(apiClient.updateCategory(nonExistentId, 'New Title')).rejects.toThrow();
    });

    test('should reject deleting non-existent category', async () => {
      const nonExistentId = 'category_nonexistent';

      await expect(apiClient.deleteCategory(nonExistentId)).rejects.toThrow();
    });

    test('should reject getting non-existent category', async () => {
      const nonExistentId = 'category_nonexistent';

      await expect(apiClient.getCategory(nonExistentId)).rejects.toThrow();
    });
  });

  describe('Authorization', () => {
    test('should reject access without token', async () => {
      apiClient.clearToken();

      await expect(apiClient.createCategory('Test Category')).rejects.toThrow();
      await expect(apiClient.getCategories()).rejects.toThrow();
    });

    test('should reject access with invalid token', async () => {
      apiClient.setToken('invalid-token');

      await expect(apiClient.createCategory('Test Category')).rejects.toThrow();
      await expect(apiClient.getCategories()).rejects.toThrow();
    });
  });

  describe('Data Isolation', () => {
    test('should only return categories for authenticated user', async () => {
      // Создаем категорию для первого пользователя
      await apiClient.createCategory('User 1 Category');

      // Создаем второго пользователя
      const email2 = generateTestEmail();
      const password2 = generateTestPassword();
      const user2Response = await apiClient.register(email2, password2);
      
      // Переключаемся на второго пользователя
      apiClient.setToken(user2Response.token);
      
      // Создаем категорию для второго пользователя
      await apiClient.createCategory('User 2 Category');

      // Получаем категории второго пользователя
      const response = await apiClient.getCategories();

      expect(response.categories.length).toBe(1);
      expect(response.categories[0].title).toBe('User 2 Category');
      expect(response.categories[0].userId).toBe(user2Response.user.id);
    });
  });
});
