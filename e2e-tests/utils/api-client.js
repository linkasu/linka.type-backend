const axios = require('axios');

class ApiClient {
  constructor(baseURL = global.API_BASE) {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    // Добавляем interceptor для логирования
    this.client.interceptors.request.use(
      (config) => {
        console.log(`${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('Request error:', error);
        return Promise.reject(error);
      }
    );

    this.client.interceptors.response.use(
      (response) => {
        console.log(`Response: ${response.status} ${response.statusText}`);
        return response;
      },
      (error) => {
        console.error('Response error:', error.response?.status, error.response?.data);
        // Создаем Error объект для правильной обработки в тестах
        const errorMessage = error.response?.data?.error || error.message || 'Request failed';
        const httpError = new Error(errorMessage);
        httpError.status = error.response?.status;
        httpError.statusText = error.response?.statusText;
        httpError.data = error.response?.data;
        return Promise.reject(httpError);
      }
    );
  }

  setToken(token) {
    this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }

  clearToken() {
    delete this.client.defaults.headers.common['Authorization'];
  }

  // Auth endpoints
  async register(email, password) {
    const response = await this.client.post('/register', { email, password });
    return response.data;
  }

  async login(email, password) {
    const response = await this.client.post('/login', { email, password });
    return response.data;
  }

  async getProfile() {
    const response = await this.client.get('/profile');
    return response.data;
  }

  // OTP endpoints
  async registerWithOTP(email, password) {
    const response = await this.client.post('/auth/register', { email, password });
    return response.data;
  }

  async verifyEmail(email, code) {
    const response = await this.client.post('/auth/verify-email', { email, code });
    return response.data;
  }

  async requestPasswordReset(email) {
    const response = await this.client.post('/auth/reset-password', { email });
    return response.data;
  }

  async verifyPasswordResetOTP(email, code) {
    const response = await this.client.post('/auth/reset-password/verify', { email, code });
    return response.data;
  }

  async confirmPasswordReset(email, code, password) {
    const response = await this.client.post('/auth/reset-password/confirm', { email, code, password });
    return response.data;
  }

  // Categories endpoints
  async getCategories() {
    const response = await this.client.get('/categories');
    return response.data;
  }

  async getCategory(id) {
    const response = await this.client.get(`/categories/${id}`);
    return response.data;
  }

  async createCategory(title) {
    const response = await this.client.post('/categories', { title });
    return response.data;
  }

  async updateCategory(id, title) {
    const response = await this.client.put(`/categories/${id}`, { title });
    return response.data;
  }

  async deleteCategory(id) {
    const response = await this.client.delete(`/categories/${id}`);
    return response.data;
  }

  // Statements endpoints
  async getStatements() {
    const response = await this.client.get('/statements');
    return response.data;
  }

  async getStatement(id) {
    const response = await this.client.get(`/statements/${id}`);
    return response.data;
  }

  async createStatement(title, categoryId) {
    const response = await this.client.post('/statements', { title, categoryId });
    return response.data;
  }

  async updateStatement(id, title, categoryId) {
    const response = await this.client.put(`/statements/${id}`, { title, categoryId });
    return response.data;
  }

  async deleteStatement(id) {
    const response = await this.client.delete(`/statements/${id}`);
    return response.data;
  }

  // Health check
  async healthCheck() {
    const response = await this.client.get('/health');
    return response.data;
  }
}

module.exports = ApiClient;
