#!/bin/bash

# Скрипт для генерации документации
# Использование: ./scripts/generate-docs.sh

set -e

echo "📚 Генерация документации для Linka Type Backend..."

# Проверяем, что мы в корневой директории проекта
if [ ! -f "go.mod" ]; then
    echo "❌ Ошибка: Запустите скрипт из корневой директории проекта"
    exit 1
fi

# Создаем директорию docs если её нет
mkdir -p docs

# Генерируем документацию
echo "🔧 Генерация документации из кода..."
go run ./cmd/docs

# Проверяем, что файлы созданы
if [ -f "docs/generated.md" ]; then
    echo "✅ Markdown документация создана: docs/generated.md"
else
    echo "❌ Ошибка: Markdown документация не создана"
    exit 1
fi

if [ -f "docs/generated.html" ]; then
    echo "✅ HTML документация создана: docs/generated.html"
else
    echo "❌ Ошибка: HTML документация не создана"
    exit 1
fi

if [ -f "docs/generated.json" ]; then
    echo "✅ JSON документация создана: docs/generated.json"
else
    echo "❌ Ошибка: JSON документация не создана"
    exit 1
fi

echo ""
echo "🎉 Документация успешно сгенерирована!"
echo ""
echo "📁 Файлы документации:"
echo "   - docs/generated.md   (Markdown)"
echo "   - docs/generated.html (HTML)"
echo "   - docs/generated.json (JSON)"
echo ""
echo "🌐 Для просмотра HTML документации:"
echo "   make docs-serve"
echo ""
echo "📖 Для просмотра Markdown документации:"
echo "   cat docs/generated.md" 