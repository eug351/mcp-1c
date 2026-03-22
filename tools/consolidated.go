package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/feenlace/mcp-1c/dump"
	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// actionArgs is the common envelope for all consolidated tool requests.
// Only the action field is decoded here; individual handlers unmarshal
// the remaining fields from the original raw arguments.
type actionArgs struct {
	Action string `json:"action"`
}

// extractAction unmarshals the action field from a CallToolRequest.
func extractAction(req *mcp.CallToolRequest) (string, error) {
	if req.Params.Arguments == nil {
		return "", fmt.Errorf("missing arguments")
	}
	var args actionArgs
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return "", fmt.Errorf("parsing action: %w", err)
	}
	return args.Action, nil
}

// ---------- code_read ----------

// CodeReadTool returns the consolidated tool definition for reading 1C metadata and code.
func CodeReadTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "code_read",
		Title:       "Чтение метаданных и кода 1С",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Чтение метаданных и кода конфигурации 1С. " +
			"action=metadata_tree: дерево объектов конфигурации (filter для фильтрации по категории). " +
			"action=object_structure: реквизиты, табличные части и типы полей объекта (object_type, object_name). " +
			"action=form_structure: элементы формы, команды и обработчики (object_type, object_name, form_name). " +
			"action=config_info: название, версия и режим работы базы.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"action": {
					"type": "string",
					"enum": ["metadata_tree", "object_structure", "form_structure", "config_info"],
					"description": "Действие"
				},
				"filter": {
					"type": "string",
					"description": "Категория метаданных для фильтрации (для action=metadata_tree): Справочники, Документы, РегистрыСведений и др."
				},
				"object_type": {
					"type": "string",
					"description": "Тип объекта метаданных (для action=object_structure, form_structure): Document, Catalog, InformationRegister и др."
				},
				"object_name": {
					"type": "string",
					"description": "Имя объекта метаданных (для action=object_structure, form_structure)"
				},
				"form_name": {
					"type": "string",
					"description": "Имя формы (для action=form_structure, необязательно)"
				}
			},
			"required": ["action"]
		}`),
	}
}

// NewCodeReadHandler returns a ToolHandler that dispatches code_read actions
// to the appropriate existing handler.
func NewCodeReadHandler(client *onec.Client, dumpDir string) mcp.ToolHandler {
	metadataHandler := NewMetadataHandler(client)
	objectStructureHandler := NewObjectStructureHandler(client)
	formStructureHandler := NewFormStructureHandler(client, dumpDir)
	configInfoHandler := NewConfigurationInfoHandler(client)

	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := extractAction(req)
		if err != nil {
			return errResult(err.Error()), nil
		}
		switch action {
		case "metadata_tree":
			return metadataHandler(ctx, req)
		case "object_structure":
			return objectStructureHandler(ctx, req)
		case "form_structure":
			return formStructureHandler(ctx, req)
		case "config_info":
			return configInfoHandler(ctx, req)
		default:
			return errResult("Неизвестное действие: " + action), nil
		}
	}
}

// ---------- code_search ----------

// CodeSearchTool returns the consolidated tool definition for searching code and BSL help.
func CodeSearchTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "code_search",
		Title:       "Поиск по коду и справочник BSL",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Поиск по коду конфигурации 1С и справочник встроенных функций. " +
			"action=text: полнотекстовый поиск по модулям (query, mode, category, module, limit). " +
			"action=syntax_help: справка по встроенным функциям платформы 1С (query).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"action": {
					"type": "string",
					"enum": ["text", "syntax_help"],
					"description": "Действие"
				},
				"query": {
					"type": "string",
					"description": "Поисковый запрос или название функции"
				},
				"limit": {
					"type": "integer",
					"description": "Макс. количество результатов (для action=text, по умолчанию 50, макс. 500)"
				},
				"category": {
					"type": "string",
					"description": "Фильтр по типу метаданных (для action=text): Документ, Справочник, ОбщийМодуль и др."
				},
				"module": {
					"type": "string",
					"description": "Фильтр по типу модуля (для action=text): МодульОбъекта, МодульМенеджера, МодульФормы и др."
				},
				"mode": {
					"type": "string",
					"enum": ["smart", "regex", "exact"],
					"description": "Режим поиска (для action=text): smart (по умолчанию), regex, exact"
				}
			},
			"required": ["action", "query"]
		}`),
	}
}

// NewCodeSearchHandler returns a ToolHandler that dispatches code_search actions.
// If dumpIndex is nil, the text search action will return an error.
func NewCodeSearchHandler(dumpIndex *dump.Index) mcp.ToolHandler {
	var searchHandler mcp.ToolHandler
	if dumpIndex != nil {
		searchHandler = NewSearchCodeHandler(dumpIndex)
	}

	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := extractAction(req)
		if err != nil {
			return errResult(err.Error()), nil
		}
		switch action {
		case "text":
			if searchHandler == nil {
				return errResult("Поиск по коду недоступен: не указан путь к выгрузке конфигурации (--dump)."), nil
			}
			return searchHandler(ctx, req)
		case "syntax_help":
			return handleBSLHelpCompat(ctx, req)
		default:
			return errResult("Неизвестное действие: " + action), nil
		}
	}
}

// bslHelpArgs extracts the query field from raw arguments for BSL help.
type bslHelpArgs struct {
	Query string `json:"query"`
}

// handleBSLHelpCompat wraps HandleBSLHelp to match the ToolHandler signature.
func handleBSLHelpCompat(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args bslHelpArgs
	if req.Params.Arguments != nil {
		json.Unmarshal(req.Params.Arguments, &args) //nolint:errcheck
	}
	if args.Query == "" {
		return errResult("Параметр query обязателен для action=syntax_help."), nil
	}
	input := BSLHelpInput{Query: args.Query}
	result, _, err := HandleBSLHelp(ctx, req, input)
	return result, err
}

// ---------- code_execute ----------

// CodeExecuteTool returns the consolidated tool definition for query operations.
func CodeExecuteTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "code_execute",
		Title:       "Выполнение и проверка запросов",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Выполнение и проверка запросов 1С. " +
			"action=query: выполнить SELECT/ВЫБРАТЬ запрос к базе (query, limit, parameters). " +
			"action=validate: проверить синтаксис запроса без выполнения (query).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"action": {
					"type": "string",
					"enum": ["query", "validate"],
					"description": "Действие"
				},
				"query": {
					"type": "string",
					"description": "Текст запроса на языке 1С"
				},
				"limit": {
					"type": "integer",
					"description": "Макс. количество строк результата (для action=query, по умолчанию 100, макс. 1000)"
				},
				"parameters": {
					"type": "object",
					"description": "Параметры запроса (для action=query): {\"Контрагент\": \"ООО Ромашка\"}"
				}
			},
			"required": ["action", "query"]
		}`),
	}
}

// NewCodeExecuteHandler returns a ToolHandler that dispatches code_execute actions.
func NewCodeExecuteHandler(client *onec.Client) mcp.ToolHandler {
	queryHandler := NewQueryHandler(client)
	validateHandler := NewValidateQueryHandler(client)

	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := extractAction(req)
		if err != nil {
			return errResult(err.Error()), nil
		}
		switch action {
		case "query":
			return queryHandler(ctx, req)
		case "validate":
			return validateHandler(ctx, req)
		default:
			return errResult("Неизвестное действие: " + action), nil
		}
	}
}

// ---------- system ----------

// SystemTool returns the consolidated tool definition for system operations.
func SystemTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "system",
		Title:       "Системные операции 1С",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Системные операции 1С. " +
			"action=event_log: журнал регистрации -- ошибки, действия пользователей, системные события (start_date, end_date, level, user, limit).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"action": {
					"type": "string",
					"enum": ["event_log"],
					"description": "Действие"
				},
				"start_date": {
					"type": "string",
					"description": "Начало периода ISO 8601 (для action=event_log)"
				},
				"end_date": {
					"type": "string",
					"description": "Конец периода ISO 8601 (для action=event_log)"
				},
				"level": {
					"type": "string",
					"enum": ["Ошибка", "Предупреждение", "Информация", "Примечание"],
					"description": "Уровень важности (для action=event_log)"
				},
				"user": {
					"type": "string",
					"description": "Имя пользователя 1С для фильтрации (для action=event_log)"
				},
				"limit": {
					"type": "integer",
					"description": "Макс. количество записей (для action=event_log, по умолчанию 50, макс. 500)"
				}
			},
			"required": ["action"]
		}`),
	}
}

// NewSystemHandler returns a ToolHandler that dispatches system actions.
func NewSystemHandler(client *onec.Client) mcp.ToolHandler {
	eventLogHandler := NewEventLogHandler(client)

	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		action, err := extractAction(req)
		if err != nil {
			return errResult(err.Error()), nil
		}
		switch action {
		case "event_log":
			return eventLogHandler(ctx, req)
		default:
			return errResult("Неизвестное действие: " + action), nil
		}
	}
}

// ---------- helpers ----------

// errResult returns a tool result with IsError=true and the given message.
func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Ошибка: %s", msg)},
		},
		IsError: true,
	}
}
