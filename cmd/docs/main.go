package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DocItem представляет элемент документации
type DocItem struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	File        string            `json:"file"`
	Line        int               `json:"line"`
	Params      []Param           `json:"params,omitempty"`
	Returns     []string          `json:"returns,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Param представляет параметр функции
type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// PackageInfo представляет информацию о пакете
type PackageInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Functions   []DocItem `json:"functions"`
	Structs     []DocItem `json:"structs"`
	Constants   []DocItem `json:"constants"`
	Variables   []DocItem `json:"variables"`
}

// ProjectDocs представляет всю документацию проекта
type ProjectDocs struct {
	GeneratedAt time.Time               `json:"generated_at"`
	Packages    map[string]*PackageInfo `json:"packages"`
	API         *APIInfo                `json:"api"`
}

// APIInfo представляет информацию об API
type APIInfo struct {
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint представляет API endpoint
type Endpoint struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Description string            `json:"description"`
	Handler     string            `json:"handler"`
	Auth        bool              `json:"auth"`
	Params      []Param           `json:"params"`
	Response    map[string]string `json:"response"`
}

func main() {
	log.Println("Starting documentation generator...")

	// Создаем структуру для документации
	docs := &ProjectDocs{
		GeneratedAt: time.Now(),
		Packages:    make(map[string]*PackageInfo),
		API:         &APIInfo{},
	}

	// Сканируем пакеты
	packages := []string{
		"auth",
		"db",
		"handlers",
		"utils",
		"websocket",
		"bl",
		"fb",
	}

	for _, pkg := range packages {
		if info := scanPackage(pkg); info != nil {
			docs.Packages[pkg] = info
		}
	}

	// Сканируем API endpoints
	docs.API.Endpoints = scanAPIEndpoints()

	// Генерируем документацию
	generateMarkdown(docs)
	generateHTML(docs)
	generateJSON(docs)

	log.Println("Documentation generated successfully!")
}

// scanPackage сканирует пакет и извлекает документацию
func scanPackage(pkgName string) *PackageInfo {
	info := &PackageInfo{
		Name:        pkgName,
		Description: "",
		Functions:   []DocItem{},
		Structs:     []DocItem{},
		Constants:   []DocItem{},
		Variables:   []DocItem{},
	}

	// Сканируем файлы в пакете
	files, err := filepath.Glob(fmt.Sprintf("%s/*.go", pkgName))
	if err != nil {
		log.Printf("Error scanning package %s: %v", pkgName, err)
		return nil
	}

	for _, file := range files {
		items := scanFile(file)
		for _, item := range items {
			switch item.Type {
			case "function":
				info.Functions = append(info.Functions, item)
			case "struct":
				info.Structs = append(info.Structs, item)
			case "constant":
				info.Constants = append(info.Constants, item)
			case "variable":
				info.Variables = append(info.Variables, item)
			}
		}
	}

	return info
}

// scanFile сканирует Go файл и извлекает документацию
func scanFile(filePath string) []DocItem {
	var items []DocItem

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s: %v", filePath, err)
		return items
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			item := extractFunctionDoc(x, fset, filePath)
			if item.Name != "" {
				items = append(items, item)
			}
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							item := extractStructDoc(typeSpec, structType, x.Doc, fset, filePath)
							items = append(items, item)
						}
					}
				}
			}
		}
		return true
	})

	return items
}

// extractFunctionDoc извлекает документацию функции
func extractFunctionDoc(fn *ast.FuncDecl, fset *token.FileSet, filePath string) DocItem {
	item := DocItem{
		Name:        fn.Name.Name,
		Type:        "function",
		File:        filePath,
		Line:        fset.Position(fn.Pos()).Line,
		Description: extractComment(fn.Doc),
		Params:      []Param{},
		Returns:     []string{},
		Tags:        extractTags(fn.Doc),
	}

	// Извлекаем параметры
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			paramType := formatType(field.Type)
			for _, name := range field.Names {
				item.Params = append(item.Params, Param{
					Name:        name.Name,
					Type:        paramType,
					Description: "",
					Required:    true,
				})
			}
		}
	}

	// Извлекаем возвращаемые значения
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			item.Returns = append(item.Returns, formatType(field.Type))
		}
	}

	return item
}

// extractStructDoc извлекает документацию структуры
func extractStructDoc(typeSpec *ast.TypeSpec, structType *ast.StructType, doc *ast.CommentGroup, fset *token.FileSet, filePath string) DocItem {
	item := DocItem{
		Name:        typeSpec.Name.Name,
		Type:        "struct",
		File:        filePath,
		Line:        fset.Position(typeSpec.Pos()).Line,
		Description: extractComment(doc),
		Params:      []Param{},
		Tags:        extractTags(doc),
	}

	// Извлекаем поля структуры
	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			fieldType := formatType(field.Type)
			for _, name := range field.Names {
				item.Params = append(item.Params, Param{
					Name:        name.Name,
					Type:        fieldType,
					Description: extractComment(field.Doc),
					Required:    true,
				})
			}
		}
	}

	return item
}

// extractComment извлекает комментарий из CommentGroup
func extractComment(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var comments []string
	for _, comment := range doc.List {
		text := strings.TrimSpace(comment.Text)
		if strings.HasPrefix(text, "//") {
			text = strings.TrimSpace(text[2:])
		} else if strings.HasPrefix(text, "/*") {
			text = strings.TrimSpace(text[2 : len(text)-2])
		}
		if text != "" {
			comments = append(comments, text)
		}
	}

	return strings.Join(comments, " ")
}

// extractTags извлекает теги из комментариев
func extractTags(doc *ast.CommentGroup) map[string]string {
	tags := make(map[string]string)
	if doc == nil {
		return tags
	}

	tagRegex := regexp.MustCompile(`@(\w+)\s+(.+)`)
	for _, comment := range doc.List {
		matches := tagRegex.FindAllStringSubmatch(comment.Text, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				tags[match[1]] = strings.TrimSpace(match[2])
			}
		}
	}

	return tags
}

// formatType форматирует тип
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// scanAPIEndpoints сканирует API endpoints
func scanAPIEndpoints() []Endpoint {
	return []Endpoint{
		{
			Method:      "POST",
			Path:        "/api/register",
			Description: "Регистрация нового пользователя",
			Handler:     "handlers.Register",
			Auth:        false,
			Params: []Param{
				{Name: "email", Type: "string", Description: "Email пользователя", Required: true},
				{Name: "password", Type: "string", Description: "Пароль пользователя", Required: true},
			},
			Response: map[string]string{
				"200": "Успешная регистрация",
				"400": "Ошибка валидации",
				"500": "Внутренняя ошибка сервера",
			},
		},
		{
			Method:      "POST",
			Path:        "/api/login",
			Description: "Авторизация пользователя",
			Handler:     "handlers.Login",
			Auth:        false,
			Params: []Param{
				{Name: "email", Type: "string", Description: "Email пользователя", Required: true},
				{Name: "password", Type: "string", Description: "Пароль пользователя", Required: true},
			},
			Response: map[string]string{
				"200": "Успешная авторизация",
				"401": "Неверные учетные данные",
				"500": "Внутренняя ошибка сервера",
			},
		},
		{
			Method:      "GET",
			Path:        "/api/statements",
			Description: "Получение списка statements",
			Handler:     "handlers.GetStatements",
			Auth:        true,
			Params:      []Param{},
			Response: map[string]string{
				"200": "Список statements",
				"401": "Не авторизован",
				"500": "Внутренняя ошибка сервера",
			},
		},
		{
			Method:      "POST",
			Path:        "/api/statements",
			Description: "Создание нового statement",
			Handler:     "handlers.CreateStatement",
			Auth:        true,
			Params: []Param{
				{Name: "title", Type: "string", Description: "Заголовок statement", Required: true},
				{Name: "content", Type: "string", Description: "Содержание statement", Required: true},
				{Name: "category_id", Type: "string", Description: "ID категории", Required: true},
			},
			Response: map[string]string{
				"201": "Statement создан",
				"400": "Ошибка валидации",
				"401": "Не авторизован",
				"500": "Внутренняя ошибка сервера",
			},
		},
		{
			Method:      "GET",
			Path:        "/api/categories",
			Description: "Получение списка категорий",
			Handler:     "handlers.GetCategories",
			Auth:        true,
			Params:      []Param{},
			Response: map[string]string{
				"200": "Список категорий",
				"401": "Не авторизован",
				"500": "Внутренняя ошибка сервера",
			},
		},
		{
			Method:      "POST",
			Path:        "/api/categories",
			Description: "Создание новой категории",
			Handler:     "handlers.CreateCategory",
			Auth:        true,
			Params: []Param{
				{Name: "name", Type: "string", Description: "Название категории", Required: true},
				{Name: "description", Type: "string", Description: "Описание категории", Required: false},
			},
			Response: map[string]string{
				"201": "Категория создана",
				"400": "Ошибка валидации",
				"401": "Не авторизован",
				"500": "Внутренняя ошибка сервера",
			},
		},
		{
			Method:      "GET",
			Path:        "/api/ws",
			Description: "WebSocket подключение для real-time уведомлений",
			Handler:     "handlers.HandleWebSocket",
			Auth:        true,
			Params:      []Param{},
			Response: map[string]string{
				"101": "WebSocket upgrade успешен",
				"401": "Не авторизован",
			},
		},
	}
}

// generateMarkdown генерирует Markdown документацию
func generateMarkdown(docs *ProjectDocs) {
	content := fmt.Sprintf(`# Документация проекта Linka Type Backend

**Сгенерировано:** %s

## Содержание

- [API Endpoints](#api-endpoints)
`, docs.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Добавляем ссылки на пакеты
	for pkgName := range docs.Packages {
		content += fmt.Sprintf("- [%s](#%s)\n", pkgName, pkgName)
	}

	content += "\n## API Endpoints\n\n"

	// Добавляем API endpoints
	for _, endpoint := range docs.API.Endpoints {
		content += fmt.Sprintf("### %s %s\n\n", endpoint.Method, endpoint.Path)
		content += fmt.Sprintf("%s\n\n", endpoint.Description)
		content += fmt.Sprintf("**Handler:** %s\n", endpoint.Handler)
		authText := "Нет"
		if endpoint.Auth {
			authText = "Да"
		}
		content += fmt.Sprintf("**Auth:** %s\n\n", authText)

		if len(endpoint.Params) > 0 {
			content += "**Параметры:**\n"
			for _, param := range endpoint.Params {
				required := ""
				if param.Required {
					required = "*обязательный*"
				}
				content += fmt.Sprintf("- `%s` (%s) %s - %s\n", param.Name, param.Type, required, param.Description)
			}
			content += "\n"
		}

		content += "**Ответы:**\n"
		for code, desc := range endpoint.Response {
			content += fmt.Sprintf("- `%s` - %s\n", code, desc)
		}
		content += "\n---\n\n"
	}

	// Добавляем пакеты
	for pkgName, pkgInfo := range docs.Packages {
		content += fmt.Sprintf("## %s\n\n", pkgName)
		if pkgInfo.Description != "" {
			content += fmt.Sprintf("%s\n\n", pkgInfo.Description)
		}

		// Функции
		if len(pkgInfo.Functions) > 0 {
			content += "### Функции\n\n"
			for _, fn := range pkgInfo.Functions {
				content += fmt.Sprintf("#### %s\n\n", fn.Name)
				if fn.Description != "" {
					content += fmt.Sprintf("%s\n\n", fn.Description)
				}

				if len(fn.Params) > 0 {
					content += "**Параметры:**\n"
					for _, param := range fn.Params {
						required := ""
						if param.Required {
							required = "*обязательный*"
						}
						content += fmt.Sprintf("- `%s` (%s) %s - %s\n", param.Name, param.Type, required, param.Description)
					}
					content += "\n"
				}

				if len(fn.Returns) > 0 {
					content += "**Возвращает:**\n"
					for _, ret := range fn.Returns {
						content += fmt.Sprintf("- %s\n", ret)
					}
					content += "\n"
				}

				if len(fn.Tags) > 0 {
					content += "**Теги:**\n"
					for key, value := range fn.Tags {
						content += fmt.Sprintf("- `%s`: %s\n", key, value)
					}
					content += "\n"
				}

				content += fmt.Sprintf("**Файл:** %s:%d\n\n---\n\n", fn.File, fn.Line)
			}
		}

		// Структуры
		if len(pkgInfo.Structs) > 0 {
			content += "### Структуры\n\n"
			for _, st := range pkgInfo.Structs {
				content += fmt.Sprintf("#### %s\n\n", st.Name)
				if st.Description != "" {
					content += fmt.Sprintf("%s\n\n", st.Description)
				}

				if len(st.Params) > 0 {
					content += "**Поля:**\n"
					for _, field := range st.Params {
						content += fmt.Sprintf("- `%s` (%s) - %s\n", field.Name, field.Type, field.Description)
					}
					content += "\n"
				}

				if len(st.Tags) > 0 {
					content += "**Теги:**\n"
					for key, value := range st.Tags {
						content += fmt.Sprintf("- `%s`: %s\n", key, value)
					}
					content += "\n"
				}

				content += fmt.Sprintf("**Файл:** %s:%d\n\n---\n\n", st.File, st.Line)
			}
		}

		// Константы
		if len(pkgInfo.Constants) > 0 {
			content += "### Константы\n\n"
			for _, constItem := range pkgInfo.Constants {
				content += fmt.Sprintf("#### %s\n\n", constItem.Name)
				if constItem.Description != "" {
					content += fmt.Sprintf("%s\n\n", constItem.Description)
				}
				content += fmt.Sprintf("**Файл:** %s:%d\n\n---\n\n", constItem.File, constItem.Line)
			}
		}

		// Переменные
		if len(pkgInfo.Variables) > 0 {
			content += "### Переменные\n\n"
			for _, varItem := range pkgInfo.Variables {
				content += fmt.Sprintf("#### %s\n\n", varItem.Name)
				if varItem.Description != "" {
					content += fmt.Sprintf("%s\n\n", varItem.Description)
				}
				content += fmt.Sprintf("**Файл:** %s:%d\n\n---\n\n", varItem.File, varItem.Line)
			}
		}
	}

	file, err := os.Create("docs/generated.md")
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		log.Printf("Error writing file: %v", err)
		return
	}

	log.Println("Markdown documentation generated: docs/generated.md")
}

// generateHTML генерирует HTML документацию
func generateHTML(docs *ProjectDocs) {
	tmpl := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Linka Type Backend - Документация</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f8f9fa;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
            border-bottom: 2px solid #ecf0f1;
            padding-bottom: 5px;
        }
        h3 {
            color: #2c3e50;
            margin-top: 25px;
        }
        h4 {
            color: #34495e;
            margin-top: 20px;
        }
        .endpoint {
            background: #f8f9fa;
            border-left: 4px solid #3498db;
            padding: 15px;
            margin: 15px 0;
            border-radius: 4px;
        }
        .method {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 12px;
            text-transform: uppercase;
        }
        .method.get { background: #28a745; color: white; }
        .method.post { background: #007bff; color: white; }
        .method.put { background: #ffc107; color: black; }
        .method.delete { background: #dc3545; color: white; }
        .auth {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 11px;
            margin-left: 10px;
        }
        .auth.required { background: #dc3545; color: white; }
        .auth.optional { background: #6c757d; color: white; }
        .param {
            background: #e9ecef;
            padding: 8px;
            margin: 5px 0;
            border-radius: 4px;
            font-family: monospace;
        }
        .required { color: #dc3545; font-weight: bold; }
        .optional { color: #6c757d; }
        .function, .struct {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            padding: 15px;
            margin: 15px 0;
            border-radius: 4px;
        }
        .file-info {
            font-size: 12px;
            color: #6c757d;
            font-style: italic;
        }
        .toc {
            background: #e9ecef;
            padding: 20px;
            border-radius: 4px;
            margin: 20px 0;
        }
        .toc ul {
            list-style-type: none;
            padding-left: 0;
        }
        .toc li {
            margin: 5px 0;
        }
        .toc a {
            text-decoration: none;
            color: #007bff;
        }
        .toc a:hover {
            text-decoration: underline;
        }
        pre {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
        }
        code {
            background: #f1f3f4;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📚 Linka Type Backend - Документация</h1>
        <p><strong>Сгенерировано:</strong> {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>

        <div class="toc">
            <h3>📋 Содержание</h3>
            <ul>
                <li><a href="#api-endpoints">🚀 API Endpoints</a></li>
                {{range $pkg, $info := .Packages}}
                <li><a href="#{{$info.Name}}">📦 {{$info.Name}}</a></li>
                {{end}}
            </ul>
        </div>

        <h2 id="api-endpoints">🚀 API Endpoints</h2>

        {{range .API.Endpoints}}
        <div class="endpoint">
            <h3>
                <span class="method {{.Method | lower}}">{{.Method}}</span>
                <code>{{.Path}}</code>
                {{if .Auth}}
                <span class="auth required">Auth Required</span>
                {{else}}
                <span class="auth optional">No Auth</span>
                {{end}}
            </h3>
            <p>{{.Description}}</p>
            <p><strong>Handler:</strong> <code>{{.Handler}}</code></p>

            {{if .Params}}
            <h4>📝 Параметры:</h4>
            {{range .Params}}
            <div class="param">
                <code>{{.Name}}</code> ({{.Type}}) 
                {{if .Required}}<span class="required">*обязательный*</span>{{else}}<span class="optional">опциональный</span>{{end}}
                - {{.Description}}
            </div>
            {{end}}
            {{end}}

            <h4>📤 Ответы:</h4>
            {{range $code, $desc := .Response}}
            <div class="param">
                <code>{{$code}}</code> - {{$desc}}
            </div>
            {{end}}
        </div>
        {{end}}

        {{range $pkg, $info := .Packages}}
        <h2 id="{{$info.Name}}">📦 {{$info.Name}}</h2>
        {{if $info.Description}}<p>{{$info.Description}}</p>{{end}}

        {{if $info.Functions}}
        <h3>🔧 Функции</h3>
        {{range $info.Functions}}
        <div class="function">
            <h4>{{.Name}}</h4>
            <p>{{.Description}}</p>

            {{if .Params}}
            <h5>📝 Параметры:</h5>
            {{range .Params}}
            <div class="param">
                <code>{{.Name}}</code> ({{.Type}}) 
                {{if .Required}}<span class="required">*обязательный*</span>{{else}}<span class="optional">опциональный</span>{{end}}
                - {{.Description}}
            </div>
            {{end}}
            {{end}}

            {{if .Returns}}
            <h5>📤 Возвращает:</h5>
            {{range .Returns}}
            <div class="param"><code>{{.}}</code></div>
            {{end}}
            {{end}}

            {{if .Tags}}
            <h5>🏷️ Теги:</h5>
            {{range $key, $value := .Tags}}
            <div class="param"><code>{{$key}}</code>: {{$value}}</div>
            {{end}}
            {{end}}

            <div class="file-info">📁 {{.File}}:{{.Line}}</div>
        </div>
        {{end}}
        {{end}}

        {{if $info.Structs}}
        <h3>🏗️ Структуры</h3>
        {{range $info.Structs}}
        <div class="struct">
            <h4>{{.Name}}</h4>
            <p>{{.Description}}</p>

            {{if .Params}}
            <h5>📝 Поля:</h5>
            {{range .Params}}
            <div class="param">
                <code>{{.Name}}</code> ({{.Type}}) - {{.Description}}
            </div>
            {{end}}
            {{end}}

            {{if .Tags}}
            <h5>🏷️ Теги:</h5>
            {{range $key, $value := .Tags}}
            <div class="param"><code>{{$key}}</code>: {{$value}}</div>
            {{end}}
            {{end}}

            <div class="file-info">📁 {{.File}}:{{.Line}}</div>
        </div>
        {{end}}
        {{end}}

        {{if $info.Constants}}
        <h3>🔢 Константы</h3>
        {{range $info.Constants}}
        <div class="function">
            <h4>{{.Name}}</h4>
            <p>{{.Description}}</p>
            <div class="file-info">📁 {{.File}}:{{.Line}}</div>
        </div>
        {{end}}
        {{end}}

        {{if $info.Variables}}
        <h3>📊 Переменные</h3>
        {{range $info.Variables}}
        <div class="function">
            <h4>{{.Name}}</h4>
            <p>{{.Description}}</p>
            <div class="file-info">📁 {{.File}}:{{.Line}}</div>
        </div>
        {{end}}
        {{end}}

        {{end}}
    </div>
</body>
</html>`

	t, err := template.New("html").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(tmpl)
	if err != nil {
		log.Printf("Error parsing HTML template: %v", err)
		return
	}

	file, err := os.Create("docs/generated.html")
	if err != nil {
		log.Printf("Error creating HTML file: %v", err)
		return
	}
	defer file.Close()

	if err := t.Execute(file, docs); err != nil {
		log.Printf("Error executing HTML template: %v", err)
		return
	}

	log.Println("HTML documentation generated: docs/generated.html")
}

// generateJSON генерирует JSON документацию
func generateJSON(docs *ProjectDocs) {
	file, err := os.Create("docs/generated.json")
	if err != nil {
		log.Printf("Error creating JSON file: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(docs); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		return
	}

	log.Println("JSON documentation generated: docs/generated.json")
}
