package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// metadataCategory maps a JSON key from 1C response to a human-readable title.
type metadataCategory struct {
	key   string // JSON key from 1C (e.g. "Справочники")
	title string // Display name (e.g. "Справочники")
}

// metadataCategories defines all known 1C metadata categories in display order.
var metadataCategories = []metadataCategory{
	{"Справочники", "Справочники"},
	{"Документы", "Документы"},
	{"Перечисления", "Перечисления"},
	{"Обработки", "Обработки"},
	{"Отчеты", "Отчёты"},
	{"РегистрыСведений", "Регистры сведений"},
	{"РегистрыНакопления", "Регистры накопления"},
	{"РегистрыБухгалтерии", "Регистры бухгалтерии"},
	{"РегистрыРасчета", "Регистры расчёта"},
	{"ПланыСчетов", "Планы счетов"},
	{"ПланыВидовХарактеристик", "Планы видов характеристик"},
	{"ПланыВидовРасчета", "Планы видов расчёта"},
	{"ПланыОбмена", "Планы обмена"},
	{"БизнесПроцессы", "Бизнес-процессы"},
	{"Задачи", "Задачи"},
	{"ЖурналыДокументов", "Журналы документов"},
	{"Константы", "Константы"},
	{"ОбщиеМодули", "Общие модули"},
	{"ОбщиеФормы", "Общие формы"},
	{"ОбщиеКоманды", "Общие команды"},
	{"ОбщиеМакеты", "Общие макеты"},
	{"Роли", "Роли"},
	{"Подсистемы", "Подсистемы"},
	{"РегулярныеЗадания", "Регулярные задания"},
	{"ВебСервисы", "Веб-сервисы"},
	{"HTTPСервисы", "HTTP-сервисы"},
}

// MetadataTool returns the MCP tool definition for get_metadata_tree.
func MetadataTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "get_metadata_tree",
		Description: "Получить дерево метаданных конфигурации 1С: справочники, документы, перечисления, " +
			"обработки, отчёты, регистры сведений/накопления/бухгалтерии/расчёта, планы счетов, " +
			"планы видов характеристик/расчёта, планы обмена, бизнес-процессы, задачи, " +
			"журналы документов, константы, общие модули/формы/команды/макеты, роли, подсистемы, " +
			"регулярные задания, веб-сервисы, HTTP-сервисы. " +
			"Используй когда нужно узнать структуру конфигурации, какие объекты есть в базе.",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}
}

// NewMetadataHandler returns a ToolHandler that fetches the metadata tree from 1C.
func NewMetadataHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var tree map[string][]string
		if err := client.Get(ctx, "/metadata", &tree); err != nil {
			return nil, fmt.Errorf("fetching metadata from 1C: %w", err)
		}

		return textResult(formatMetadataTree(tree)), nil
	}
}

// formatMetadataTree formats the metadata tree as markdown text.
// Known categories are rendered first in a stable order, then any unknown
// categories are appended at the end for forward compatibility.
func formatMetadataTree(tree map[string][]string) string {
	var b strings.Builder
	b.WriteString("# Метаданные конфигурации 1С\n\n")

	// Track which keys have been rendered.
	rendered := make(map[string]bool, len(metadataCategories))

	// Render known categories in defined order.
	for _, cat := range metadataCategories {
		items, ok := tree[cat.key]
		if !ok || len(items) == 0 {
			continue
		}
		rendered[cat.key] = true
		writeSection(&b, cat.title, items)
	}

	// Collect and render unknown categories (forward compatibility).
	unknown := make([]string, 0)
	for key := range tree {
		if !rendered[key] {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)

	for _, key := range unknown {
		items := tree[key]
		if len(items) == 0 {
			continue
		}
		writeSection(&b, key, items)
	}

	return b.String()
}

// writeSection writes a markdown section with the given title and items.
func writeSection(b *strings.Builder, title string, items []string) {
	fmt.Fprintf(b, "## %s\n", title)
	for _, name := range items {
		b.WriteString("- ")
		b.WriteString(name)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
}
