(() => {
  const baseUrlInput = document.getElementById("baseUrl");
  const tokenInput = document.getElementById("token");
  const templateSelect = document.getElementById("template");
  const methodSelect = document.getElementById("method");
  const pathInput = document.getElementById("path");
  const bodyInput = document.getElementById("body");
  const sendButton = document.getElementById("send");
  const output = document.getElementById("output");
  const status = document.getElementById("status");

  const templates = [
    {
      label: "POST /v1/auth",
      method: "POST",
      path: "/v1/auth",
      body: JSON.stringify({ email: "user@example.com", password: "secret" }, null, 2)
    },
    {
      label: "GET /v1/categories",
      method: "GET",
      path: "/v1/categories",
      body: ""
    },
    {
      label: "POST /v1/categories",
      method: "POST",
      path: "/v1/categories",
      body: JSON.stringify({ label: "Мои фразы", created: Date.now() }, null, 2)
    },
    {
      label: "PATCH /v1/categories/{id}",
      method: "PATCH",
      path: "/v1/categories/<categoryId>",
      body: JSON.stringify({ label: "Обновленная категория" }, null, 2)
    },
    {
      label: "GET /v1/categories/{id}/statements",
      method: "GET",
      path: "/v1/categories/<categoryId>/statements",
      body: ""
    },
    {
      label: "POST /v1/statements",
      method: "POST",
      path: "/v1/statements",
      body: JSON.stringify({ categoryId: "<categoryId>", text: "Здравствуйте!", created: Date.now() }, null, 2)
    },
    {
      label: "PATCH /v1/statements/{id}",
      method: "PATCH",
      path: "/v1/statements/<statementId>",
      body: JSON.stringify({ text: "Новый текст" }, null, 2)
    },
    {
      label: "GET /v1/user/state",
      method: "GET",
      path: "/v1/user/state",
      body: ""
    },
    {
      label: "PUT /v1/quickes",
      method: "PUT",
      path: "/v1/quickes",
      body: JSON.stringify({ quickes: ["Да", "Нет", "Спасибо", "Помогите"] }, null, 2)
    },
    {
      label: "GET /v1/global/categories?include_statements=true",
      method: "GET",
      path: "/v1/global/categories?include_statements=true",
      body: ""
    },
    {
      label: "POST /v1/onboarding/phrases",
      method: "POST",
      path: "/v1/onboarding/phrases",
      body: JSON.stringify({
        questions: [
          { question_id: "greeting", value: "Здравствуйте" },
          { question_id: "needs", value: "Мне нужна помощь" }
        ]
      }, null, 2)
    }
  ];

  const setOutput = (text) => {
    output.textContent = text;
  };

  const setStatus = (value) => {
    status.textContent = value;
  };

  const applyTemplate = (index) => {
    const template = templates[index];
    if (!template) {
      return;
    }
    methodSelect.value = template.method;
    pathInput.value = template.path;
    bodyInput.value = template.body;
  };

  const populateTemplates = () => {
    templateSelect.innerHTML = "";
    templates.forEach((template, index) => {
      const option = document.createElement("option");
      option.value = String(index);
      option.textContent = template.label;
      templateSelect.appendChild(option);
    });
    applyTemplate(0);
  };

  const request = async () => {
    const baseUrl = (baseUrlInput.value || window.location.origin).trim();
    const path = pathInput.value.trim();
    const method = methodSelect.value;
    const token = tokenInput.value.trim();

    if (!path) {
      setStatus("Ошибка");
      setOutput("Укажите путь запроса.");
      return;
    }

    const url = new URL(path, baseUrl);
    const headers = {};

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    let body = bodyInput.value.trim();
    if (body && !["GET", "HEAD"].includes(method)) {
      try {
        const parsed = JSON.parse(body);
        body = JSON.stringify(parsed);
        headers["Content-Type"] = "application/json";
      } catch (error) {
        setStatus("Ошибка");
        setOutput("JSON тело невалидно. Проверьте синтаксис и попробуйте снова.");
        return;
      }
    } else {
      body = undefined;
    }

    setStatus("...");
    setOutput("Отправляем запрос...");

    try {
      const response = await fetch(url.toString(), {
        method,
        headers,
        body
      });

      const contentType = response.headers.get("content-type") || "";
      const raw = await response.text();

      setStatus(`${response.status} ${response.statusText}`);

      if (!raw) {
        setOutput("Пустой ответ.");
        return;
      }

      if (contentType.includes("application/json")) {
        try {
          const formatted = JSON.stringify(JSON.parse(raw), null, 2);
          setOutput(formatted);
        } catch (error) {
          setOutput(raw);
        }
        return;
      }

      setOutput(raw);
    } catch (error) {
      setStatus("Ошибка сети");
      setOutput(error.message || "Не удалось выполнить запрос.");
    }
  };

  baseUrlInput.value = window.location.origin;
  populateTemplates();

  templateSelect.addEventListener("change", (event) => {
    const index = Number(event.target.value);
    applyTemplate(index);
  });

  sendButton.addEventListener("click", (event) => {
    event.preventDefault();
    request();
  });
})();
