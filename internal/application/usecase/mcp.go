package usecase

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/herewei/ohmymem-core/internal/domain"
	"github.com/herewei/ohmymem-core/internal/infrastructure/adapters"
	"github.com/herewei/ohmymem-core/internal/infrastructure/persistence"
	"github.com/herewei/ohmymem-core/internal/version"
)

// McpUseCase handles MCP tool requests
type McpUseCase struct {
	memoryService *domain.MemoryService
	uuidGen       domain.UUIDGenerator
	timeProvider  domain.TimeProvider
}

// NewMcpUseCase creates a new MCP McpUseCase
func NewMcpUseCase(
	memoryService *domain.MemoryService,
	uuidGen domain.UUIDGenerator,
	timeProvider domain.TimeProvider,
) *McpUseCase {
	return &McpUseCase{
		memoryService: memoryService,
		uuidGen:       uuidGen,
		timeProvider:  timeProvider,
	}
}

// RegisterTools registers the MCP tools with the server
func (h *McpUseCase) RegisterTools(s *server.MCPServer) {
	// Register ohmymem_read tool
	readTool := mcp.NewTool("ohmymem_read",
		mcp.WithDescription("Read the working memory file (.ohmymem/memory.md). Returns the raw Markdown content containing constraints, decisions, patterns, and anti-patterns."),
	)

	s.AddTool(readTool, h.handleReadMemory)

	// Register ohmymem_capture tool
	captureTool := mcp.NewTool("ohmymem_capture",
		mcp.WithDescription("When you find some valueable to memory.use this tool to capture a new entry to the working memory file under a specific category."),
		mcp.WithString("category",
			mcp.Description("Category: 'constraints', 'decisions', 'patterns', 'anti-patterns' or 'note'. Defaults to 'note' if not specified."),
			mcp.Enum("constraints", "decisions", "patterns", "anti-patterns", "note"),
		),
		mcp.WithString("tag",
			mcp.Required(),
			mcp.Description("Tag for the entry (max 50 chars, auto-wrapped in brackets)"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to remember (max 2000 chars, no newlines or markdown headers)"),
		),
		mcp.WithString("rationale",
			mcp.Description("Optional reason/justification (max 500 chars)"),
		),
	)

	s.AddTool(captureTool, h.handleCaptureMemory)
}

// handleReadMemory handles the ohmymem_read tool request
func (h *McpUseCase) handleReadMemory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := h.memoryService.ReadMemory(ctx)
	if err != nil {
		slog.Error("failed to read memory", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read memory: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

// handleCaptureMemory handles the ohmymem_capture tool request
func (h *McpUseCase) handleCaptureMemory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	category := request.GetString("category", "")
	tag := request.GetString("tag", "")
	content := request.GetString("content", "")
	rationale := request.GetString("rationale", "")

	// Validate input
	input := domain.AppendInput{
		Category:  category,
		Tag:       tag,
		Content:   content,
		Rationale: rationale,
	}

	if err := h.memoryService.ValidateInput(input); err != nil {
		slog.Warn("validation failed", "error", err, "category", category, "tag", tag)
		return mcp.NewToolResultError(fmt.Sprintf("Validation failed: %v", err)), nil
	}

	// Generate ID and timestamp
	id, err := h.uuidGen.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate ID: %v", err)), nil
	}
	now := h.timeProvider.Now()

	// Append to memory
	if err := h.memoryService.AppendMemory(ctx, input, id, now); err != nil {
		slog.Error("failed to capture memory", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to capture to memory: %v", err)), nil
	}

	slog.Debug("memory entry added",
		"category", category,
		"tag", tag,
		"id", id)

	return mcp.NewToolResultText(fmt.Sprintf("Successfully captured entry to '%s' category.", category)), nil
}

// NewServer creates and configures a new MCP server
func NewServer(
	basePath string,
) (*server.MCPServer, *persistence.MarkdownMemoryRepository, error) {
	// Initialize infrastructure
	uuidGen := adapters.NewGoogleUUIDGenerator()
	timeProvider := adapters.NewSystemClock()
	repo := persistence.NewMemoryRepository(basePath, uuidGen, timeProvider)

	// Initialize domain service
	memoryService := domain.NewMemoryService(repo)

	// Create MCP server
	s := server.NewMCPServer(
		"OhMyMem MCP Server",
		version.Version,
		server.WithToolCapabilities(true),
	)

	// Create McpUseCase and register tools
	McpUseCase := NewMcpUseCase(memoryService, uuidGen, timeProvider)
	McpUseCase.RegisterTools(s)

	return s, repo, nil
}
