const WebSocket = require('ws');

class WebSocketClient {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.ws = null;
    this.messages = [];
    this.connected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 3;
  }

  connect() {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url, {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        });

        this.ws.on('open', () => {
          console.log('WebSocket connected');
          this.connected = true;
          this.reconnectAttempts = 0;
          resolve();
        });

        this.ws.on('message', (data) => {
          try {
            const message = JSON.parse(data.toString());
            console.log('WebSocket message received:', message);
            this.messages.push(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        });

        this.ws.on('close', (code, reason) => {
          console.log(`WebSocket closed: ${code} - ${reason}`);
          this.connected = false;
        });

        this.ws.on('error', (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        });

        // Timeout для подключения
        setTimeout(() => {
          if (!this.connected) {
            reject(new Error('WebSocket connection timeout'));
          }
        }, 5000);

      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.connected = false;
    }
  }

  send(message) {
    if (this.ws && this.connected) {
      this.ws.send(JSON.stringify(message));
    } else {
      throw new Error('WebSocket not connected');
    }
  }

  waitForMessage(type, timeout = 10000) {
    return new Promise((resolve, reject) => {
      const startTime = Date.now();
      
      const checkMessages = () => {
        const message = this.messages.find(msg => msg.type === type);
        if (message) {
          resolve(message);
          return;
        }

        if (Date.now() - startTime > timeout) {
          reject(new Error(`Timeout waiting for message type: ${type}`));
          return;
        }

        setTimeout(checkMessages, 100);
      };

      checkMessages();
    });
  }

  waitForMessageWithPayload(type, payloadCheck, timeout = 10000) {
    return new Promise((resolve, reject) => {
      const startTime = Date.now();
      
      const checkMessages = () => {
        const message = this.messages.find(msg => {
          if (msg.type !== type) return false;
          return payloadCheck(msg.payload);
        });
        
        if (message) {
          resolve(message);
          return;
        }

        if (Date.now() - startTime > timeout) {
          reject(new Error(`Timeout waiting for message type: ${type} with specific payload`));
          return;
        }

        setTimeout(checkMessages, 100);
      };

      checkMessages();
    });
  }

  getMessages() {
    return [...this.messages];
  }

  clearMessages() {
    this.messages = [];
  }

  isConnected() {
    return this.connected && this.ws && this.ws.readyState === WebSocket.OPEN;
  }
}

module.exports = WebSocketClient;
