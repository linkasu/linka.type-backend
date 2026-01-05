(() => {
  const tokenKey = "adminToken";
  const loginPanel = document.getElementById("login-panel");
  const adminPanel = document.getElementById("admin-panel");
  const loginMessage = document.getElementById("login-message");
  const adminMessage = document.getElementById("admin-message");

  const emailInput = document.getElementById("login-email");
  const passwordInput = document.getElementById("login-password");
  const loginBtn = document.getElementById("login-btn");
  const logoutBtn = document.getElementById("logout");

  const statUsers = document.getElementById("stat-users");
  const statCategories = document.getElementById("stat-categories");
  const statStatements = document.getElementById("stat-statements");
  const statWindowSelect = document.getElementById("stat-window-select");

  const globalCategoriesTable = document.getElementById("global-categories-table");
  const refreshGlobalCategoriesBtn = document.getElementById("refresh-global-categories");
  const createGlobalCategoryBtn = document.getElementById("create-global-category");

  const questionsTable = document.getElementById("questions-table");
  const refreshQuestionsBtn = document.getElementById("refresh-questions");
  const createQuestionBtn = document.getElementById("create-question");

  const adminsTable = document.getElementById("admins-table");
  const refreshAdminsBtn = document.getElementById("refresh-admins");
  const addAdminBtn = document.getElementById("add-admin");
  const newAdminId = document.getElementById("new-admin-id");

  const keysTable = document.getElementById("keys-table");
  const refreshKeysBtn = document.getElementById("refresh-keys");
  const createKeyBtn = document.getElementById("create-key");
  const createKeyMessage = document.getElementById("create-key-message");
  const newClientId = document.getElementById("new-client-id");
  const newStatus = document.getElementById("new-status");

  function setMessage(target, text, tone = "muted") {
    if (!target) return;
    target.textContent = text || "";
    target.className = tone === "error" ? "text-danger" : tone === "success" ? "text-success" : "text-muted";
  }

  function getApiBase() {
    return window.location.origin.replace(/\/$/, "");
  }

  async function apiFetch(path, options = {}) {
    const url = `${getApiBase()}${path}`;
    const headers = Object.assign({}, options.headers || {});
    const token = localStorage.getItem(tokenKey);
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    if (options.body && !headers["Content-Type"]) {
      headers["Content-Type"] = "application/json";
    }
    const response = await fetch(url, Object.assign({}, options, { headers }));
    const text = await response.text();
    let data = null;
    try {
      data = text ? JSON.parse(text) : null;
    } catch (err) {
      data = null;
    }
    if (!response.ok) {
      const message = data && data.message ? data.message : `Request failed (${response.status})`;
      throw new Error(message);
    }
    return data;
  }

  async function login() {
    const email = emailInput.value.trim();
    const password = passwordInput.value.trim();
    if (!email || !password) {
      setMessage(loginMessage, "Введите email и пароль", "error");
      return;
    }
    setMessage(loginMessage, "Вход...");
    try {
      const data = await apiFetch("/v1/auth", {
        method: "POST",
        body: JSON.stringify({ email, password })
      });
      if (!data.token) {
        throw new Error("Токен не получен");
      }
      localStorage.setItem(tokenKey, data.token);
      setMessage(loginMessage, "Вход выполнен", "success");
      await showAdmin();
    } catch (err) {
      setMessage(loginMessage, err.message, "error");
    }
  }

  async function loadStats() {
    const windowValue = statWindowSelect.value || "24h";
    const data = await apiFetch(`/v1/admin/stats?window=${encodeURIComponent(windowValue)}`);
    statUsers.textContent = data.total_users ?? "—";
    statCategories.textContent = data.total_categories ?? "—";
    statStatements.textContent = data.total_statements ?? "—";
  }

  async function loadGlobalCategories() {
    const data = await apiFetch("/v1/admin/global/categories");
    globalCategoriesTable.innerHTML = "";
    if (!data || data.length === 0) {
      globalCategoriesTable.innerHTML = "<tr><td colspan=\"4\" class=\"text-muted\">Нет категорий</td></tr>";
      return;
    }
    data.forEach((item) => {
      const row = document.createElement("tr");
      row.innerHTML = `
        <td><small>${item.id || item.ID || "—"}</small></td>
        <td>${item.label || item.Label || "—"}</td>
        <td>${item.default || item.Default ? "Да" : "Нет"}</td>
        <td>
          <button class="btn btn-ghost btn-sm" data-edit-category="${item.id || item.ID}">Редактировать</button>
          <button class="btn btn-ghost btn-sm" data-delete-category="${item.id || item.ID}">Удалить</button>
        </td>
      `;
      globalCategoriesTable.appendChild(row);
    });
    globalCategoriesTable.querySelectorAll("button[data-delete-category]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const id = btn.getAttribute("data-delete-category");
        if (!confirm("Удалить категорию?")) return;
        try {
          await apiFetch(`/v1/admin/global/categories/${id}`, { method: "DELETE" });
          await loadGlobalCategories();
        } catch (err) {
          setMessage(adminMessage, err.message, "error");
        }
      });
    });
  }

  async function loadQuestions() {
    const data = await apiFetch("/v1/admin/factory/questions");
    questionsTable.innerHTML = "";
    if (!data || data.length === 0) {
      questionsTable.innerHTML = "<tr><td colspan=\"5\" class=\"text-muted\">Нет вопросов</td></tr>";
      return;
    }
    data.forEach((item) => {
      const row = document.createElement("tr");
      row.innerHTML = `
        <td><small>${item.id || item.ID || item.question_id || "—"}</small></td>
        <td>${item.label || item.Label || "—"}</td>
        <td>${item.category || item.Category || "—"}</td>
        <td>${item.type || item.Type || "—"}</td>
        <td>
          <button class="btn btn-ghost btn-sm" data-edit-question="${item.id || item.ID || item.question_id}">Редактировать</button>
          <button class="btn btn-ghost btn-sm" data-delete-question="${item.id || item.ID || item.question_id}">Удалить</button>
        </td>
      `;
      questionsTable.appendChild(row);
    });
    questionsTable.querySelectorAll("button[data-delete-question]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const id = btn.getAttribute("data-delete-question");
        if (!confirm("Удалить вопрос?")) return;
        try {
          await apiFetch(`/v1/admin/factory/questions/${id}`, { method: "DELETE" });
          await loadQuestions();
        } catch (err) {
          setMessage(adminMessage, err.message, "error");
        }
      });
    });
  }

  async function loadAdmins() {
    const data = await apiFetch("/v1/admin/admins");
    adminsTable.innerHTML = "";
    if (!data.items || data.items.length === 0) {
      adminsTable.innerHTML = "<tr><td colspan=\"2\" class=\"text-muted\">Нет админов</td></tr>";
      return;
    }
    data.items.forEach((userID) => {
      const row = document.createElement("tr");
      row.innerHTML = `
        <td>${userID}</td>
        <td><button class="btn btn-ghost btn-sm" data-remove-admin="${userID}">Удалить</button></td>
      `;
      adminsTable.appendChild(row);
    });
    adminsTable.querySelectorAll("button[data-remove-admin]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const userID = btn.getAttribute("data-remove-admin");
        if (!confirm("Удалить админа?")) return;
        try {
          await apiFetch(`/v1/admin/admins/${userID}`, { method: "DELETE" });
          await loadAdmins();
        } catch (err) {
          setMessage(adminMessage, err.message, "error");
        }
      });
    });
  }

  async function addAdmin() {
    const userID = newAdminId.value.trim();
    if (!userID) {
      setMessage(adminMessage, "Введите User ID", "error");
      return;
    }
    try {
      await apiFetch("/v1/admin/admins", {
        method: "POST",
        body: JSON.stringify({ user_id: userID })
      });
      newAdminId.value = "";
      await loadAdmins();
      setMessage(adminMessage, "Админ добавлен", "success");
    } catch (err) {
      setMessage(adminMessage, err.message, "error");
    }
  }

  async function loadKeys() {
    const data = await apiFetch("/v1/admin/client-keys");
    keysTable.innerHTML = "";
    if (!data.items || data.items.length === 0) {
      keysTable.innerHTML = "<tr><td colspan=\"5\" class=\"text-muted\">Нет ключей</td></tr>";
      return;
    }
    data.items.forEach((item) => {
      const row = document.createElement("tr");
      const createdAt = item.created_at ? new Date(item.created_at).toLocaleString("ru-RU") : (item.CreatedAt ? new Date(item.CreatedAt).toLocaleString("ru-RU") : "—");
      const isRevoked = (item.status || item.Status) === "revoked";
      row.innerHTML = `
        <td><small>${item.key_hash || item.KeyHash || "—"}</small></td>
        <td>${item.client_id || item.ClientID || "—"}</td>
        <td>${item.status || item.Status || "—"}</td>
        <td>${createdAt}</td>
        <td>${isRevoked ? "" : `<button class="btn btn-ghost btn-sm" data-revoke="${item.key_hash || item.KeyHash}">Отозвать</button>`}</td>
      `;
      keysTable.appendChild(row);
    });
    keysTable.querySelectorAll("button[data-revoke]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const hash = btn.getAttribute("data-revoke");
        if (!hash) return;
        if (!confirm("Отозвать ключ?")) return;
        try {
          await apiFetch(`/v1/admin/client-keys/${hash}`, { method: "DELETE" });
          await loadKeys();
        } catch (err) {
          setMessage(adminMessage, err.message, "error");
        }
      });
    });
  }

  async function createKey() {
    setMessage(createKeyMessage, "Создаем ключ...");
    const payload = {
      client_id: newClientId.value.trim(),
      status: newStatus.value
    };
    try {
      const data = await apiFetch("/v1/admin/client-keys", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      setMessage(createKeyMessage, `Новый ключ: ${data.api_key}`, "success");
      newClientId.value = "";
      await loadKeys();
    } catch (err) {
      setMessage(createKeyMessage, err.message, "error");
    }
  }

  async function showAdmin() {
    loginPanel.classList.add("d-none");
    adminPanel.classList.remove("d-none");
    await refreshAll();
  }

  function showLogin() {
    adminPanel.classList.add("d-none");
    loginPanel.classList.remove("d-none");
  }

  async function refreshAll() {
    setMessage(adminMessage, "Загружаем данные...");
    try {
      await Promise.all([loadStats(), loadGlobalCategories(), loadQuestions(), loadAdmins(), loadKeys()]);
      setMessage(adminMessage, "Данные обновлены", "success");
    } catch (err) {
      setMessage(adminMessage, err.message, "error");
    }
  }

  function logout() {
    localStorage.removeItem(tokenKey);
    showLogin();
  }

  loginBtn.addEventListener("click", login);
  logoutBtn.addEventListener("click", logout);
  refreshGlobalCategoriesBtn.addEventListener("click", loadGlobalCategories);
  refreshQuestionsBtn.addEventListener("click", loadQuestions);
  refreshAdminsBtn.addEventListener("click", loadAdmins);
  refreshKeysBtn.addEventListener("click", loadKeys);
  createKeyBtn.addEventListener("click", createKey);
  addAdminBtn.addEventListener("click", addAdmin);
  statWindowSelect.addEventListener("change", loadStats);

  createGlobalCategoryBtn.addEventListener("click", () => {
    const label = prompt("Название категории:");
    if (!label) return;
    apiFetch("/v1/admin/global/categories", {
      method: "POST",
      body: JSON.stringify({ label })
    }).then(() => loadGlobalCategories()).catch(err => setMessage(adminMessage, err.message, "error"));
  });

  createQuestionBtn.addEventListener("click", () => {
    const label = prompt("Название вопроса:");
    if (!label) return;
    const category = prompt("Категория:");
    if (!category) return;
    apiFetch("/v1/admin/factory/questions", {
      method: "POST",
      body: JSON.stringify({ label, category, type: "text", phrases: [], order_index: 0 })
    }).then(() => loadQuestions()).catch(err => setMessage(adminMessage, err.message, "error"));
  });

  if (localStorage.getItem(tokenKey)) {
    showAdmin().catch(() => logout());
  }
})();

