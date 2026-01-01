package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates and configures the pomodoro MCP server
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"pomodoro",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s)

	return s
}

// registerTools adds all pomodoro tools to the server
func registerTools(s *server.MCPServer) {
	// start_pomodoro - Start a new Pomodoro
	s.AddTool(
		mcp.NewTool("start_pomodoro",
			mcp.WithDescription("Start a new Pomodoro timer"),
			mcp.WithString("description",
				mcp.Description("Description of what you're working on"),
			),
			mcp.WithNumber("duration",
				mcp.Description("Duration in minutes (default: 25)"),
			),
			mcp.WithArray("tags",
				mcp.Description("Tags for this Pomodoro"),
			),
		),
		handleStartPomodoro,
	)

	// get_status - Get current Pomodoro status
	s.AddTool(
		mcp.NewTool("get_status",
			mcp.WithDescription("Get the status of the current Pomodoro"),
		),
		handleGetStatus,
	)

	// finish_pomodoro - Finish early
	s.AddTool(
		mcp.NewTool("finish_pomodoro",
			mcp.WithDescription("Finish the current Pomodoro early"),
		),
		handleFinishPomodoro,
	)

	// cancel_pomodoro - Cancel active
	s.AddTool(
		mcp.NewTool("cancel_pomodoro",
			mcp.WithDescription("Cancel the current active Pomodoro"),
		),
		handleCancelPomodoro,
	)

	// clear_pomodoro - Clear finished
	s.AddTool(
		mcp.NewTool("clear_pomodoro",
			mcp.WithDescription("Clear a finished Pomodoro"),
		),
		handleClearPomodoro,
	)

	// start_break - Take a break
	s.AddTool(
		mcp.NewTool("start_break",
			mcp.WithDescription("Start a break timer"),
			mcp.WithNumber("duration",
				mcp.Description("Break duration in minutes (default: 5)"),
			),
		),
		handleStartBreak,
	)

	// repeat_pomodoro - Repeat last
	s.AddTool(
		mcp.NewTool("repeat_pomodoro",
			mcp.WithDescription("Repeat the last Pomodoro with the same description and tags"),
		),
		handleRepeatPomodoro,
	)

	// amend_pomodoro - Amend current
	s.AddTool(
		mcp.NewTool("amend_pomodoro",
			mcp.WithDescription("Amend the current Pomodoro's description, duration, or tags"),
			mcp.WithString("description",
				mcp.Description("New description"),
			),
			mcp.WithNumber("duration",
				mcp.Description("New duration in minutes"),
			),
			mcp.WithArray("tags",
				mcp.Description("New tags"),
			),
		),
		handleAmendPomodoro,
	)

	// get_history - Get history
	s.AddTool(
		mcp.NewTool("get_history",
			mcp.WithDescription("Get Pomodoro history"),
			mcp.WithNumber("limit",
				mcp.Description("Limit number of entries returned (0 = all)"),
			),
		),
		handleGetHistory,
	)

	// get_settings - Get settings
	s.AddTool(
		mcp.NewTool("get_settings",
			mcp.WithDescription("Get current Pomodoro settings"),
		),
		handleGetSettings,
	)
}
