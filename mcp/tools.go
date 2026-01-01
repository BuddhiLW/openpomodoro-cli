package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/open-pomodoro/go-openpomodoro"
	"github.com/open-pomodoro/openpomodoro-cli/hook"
)

// StatusResponse represents the current Pomodoro status
type StatusResponse struct {
	Active       bool     `json:"active"`
	Done         bool     `json:"done"`
	Remaining    string   `json:"remaining,omitempty"`
	Duration     string   `json:"duration,omitempty"`
	Description  string   `json:"description,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	GoalComplete int      `json:"goal_complete"`
	GoalTotal    int      `json:"goal_total"`
}

// SettingsResponse represents the current settings
type SettingsResponse struct {
	DataDirectory           string   `json:"data_directory"`
	DailyGoal               int      `json:"daily_goal"`
	DefaultPomodoroDuration int      `json:"default_pomodoro_duration"`
	DefaultBreakDuration    int      `json:"default_break_duration"`
	DefaultTags             []string `json:"default_tags"`
}

// getClient creates a new openpomodoro client with default directory
func getClient() (*openpomodoro.Client, *openpomodoro.Settings, error) {
	client, err := openpomodoro.NewClient("")
	if err != nil {
		return nil, nil, err
	}
	settings, err := client.Settings()
	if err != nil {
		return nil, nil, err
	}
	return client, settings, nil
}

// durationAsTime formats a duration as mm:ss
func durationAsTime(d time.Duration) string {
	s := int(d.Seconds())
	if s < 0 {
		s = 0
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// handleStartPomodoro starts a new Pomodoro
func handleStartPomodoro(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, settings, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	p := openpomodoro.NewPomodoro()

	// Get description
	p.Description = request.GetString("description", "")

	// Get duration (in minutes)
	dur := request.GetFloat("duration", 0)
	if dur > 0 {
		p.Duration = time.Duration(dur) * time.Minute
	} else {
		p.Duration = settings.DefaultPomodoroDuration
	}

	// Get tags
	p.Tags = request.GetStringSlice("tags", nil)

	p.StartTime = time.Now()

	if err := client.Start(p); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to start pomodoro: %v", err)), nil
	}

	if err := hook.Run(client, "start"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hook failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pomodoro started: %s (%s)", p.Description, durationAsTime(p.Duration))), nil
}

// handleGetStatus returns the current Pomodoro status
func handleGetStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, settings, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	state, err := client.CurrentState()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get status: %v", err)), nil
	}

	p := state.Pomodoro
	response := StatusResponse{
		Active:       p.IsActive(),
		Done:         p.IsDone(),
		GoalComplete: 0,
		GoalTotal:    settings.DailyGoal,
	}

	if state.History != nil {
		response.GoalComplete = state.History.Date(time.Now()).Count()
	}

	if !p.IsInactive() {
		response.Remaining = durationAsTime(p.Remaining())
		response.Duration = durationAsTime(p.Duration)
		response.Description = p.Description
		response.Tags = p.Tags
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal status: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// handleFinishPomodoro finishes the current Pomodoro early
func handleFinishPomodoro(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, _, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	p, err := client.Pomodoro()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get pomodoro: %v", err)), nil
	}

	if p.IsInactive() {
		return mcp.NewToolResultError("no active pomodoro to finish"), nil
	}

	elapsed := time.Since(p.StartTime)

	if err := hook.Run(client, "stop"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hook failed: %v", err)), nil
	}

	if err := client.Finish(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to finish pomodoro: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pomodoro finished after %s", durationAsTime(elapsed))), nil
}

// handleCancelPomodoro cancels the current Pomodoro
func handleCancelPomodoro(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, _, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	p, err := client.Pomodoro()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get pomodoro: %v", err)), nil
	}

	if p.IsInactive() {
		return mcp.NewToolResultError("no active pomodoro to cancel"), nil
	}

	if err := hook.Run(client, "stop"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hook failed: %v", err)), nil
	}

	if err := client.Cancel(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to cancel pomodoro: %v", err)), nil
	}

	return mcp.NewToolResultText("Pomodoro cancelled"), nil
}

// handleClearPomodoro clears a finished Pomodoro
func handleClearPomodoro(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, _, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	if err := hook.Run(client, "stop"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hook failed: %v", err)), nil
	}

	if err := client.Clear(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to clear pomodoro: %v", err)), nil
	}

	return mcp.NewToolResultText("Pomodoro cleared"), nil
}

// handleStartBreak starts a break timer
func handleStartBreak(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, settings, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	duration := settings.DefaultBreakDuration

	dur := request.GetFloat("duration", 0)
	if dur > 0 {
		duration = time.Duration(dur) * time.Minute
	}

	if err := hook.Run(client, "break"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hook failed: %v", err)), nil
	}

	// Note: break is a blocking timer in CLI, but for MCP we just trigger the hook
	// and return immediately. The hook can handle notifications.

	return mcp.NewToolResultText(fmt.Sprintf("Break started (%s)", durationAsTime(duration))), nil
}

// handleRepeatPomodoro repeats the last Pomodoro
func handleRepeatPomodoro(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, settings, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	h, err := client.History()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get history: %v", err)), nil
	}

	p := h.Latest()
	if p == nil {
		return mcp.NewToolResultError("no previous pomodoro to repeat"), nil
	}

	if p.IsActive() {
		return mcp.NewToolResultError("cannot repeat an active pomodoro"), nil
	}

	p.StartTime = time.Now()
	p.Duration = settings.DefaultPomodoroDuration

	if err := client.Start(p); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to start pomodoro: %v", err)), nil
	}

	if err := hook.Run(client, "start"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hook failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pomodoro repeated: %s", p.Description)), nil
}

// handleAmendPomodoro amends the current Pomodoro
func handleAmendPomodoro(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, _, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	h, err := client.History()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get history: %v", err)), nil
	}

	p := h.Latest()
	if p == nil {
		return mcp.NewToolResultError("no pomodoro to amend"), nil
	}

	// Update description if provided
	desc := request.GetString("description", "")
	if desc != "" {
		p.Description = desc
	}

	// Update duration if provided
	dur := request.GetFloat("duration", 0)
	if dur > 0 {
		p.Duration = time.Duration(dur) * time.Minute
	}

	// Update tags if provided
	tags := request.GetStringSlice("tags", nil)
	if tags != nil {
		p.Tags = tags
	}

	if err := client.Start(p); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to amend pomodoro: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Pomodoro amended: %s", p.Description)), nil
}

// handleGetHistory returns Pomodoro history
func handleGetHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, _, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	h, err := client.History()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get history: %v", err)), nil
	}

	// Apply limit if specified
	limit := int(request.GetFloat("limit", 0))

	if limit > 0 && len(h.Pomodoros) > limit {
		start := len(h.Pomodoros) - limit
		h = &openpomodoro.History{
			Pomodoros: h.Pomodoros[start:],
		}
	}

	jsonBytes, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal history: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// handleGetSettings returns current settings
func handleGetSettings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, settings, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create client: %v", err)), nil
	}

	response := SettingsResponse{
		DataDirectory:           client.Directory,
		DailyGoal:               settings.DailyGoal,
		DefaultPomodoroDuration: int(settings.DefaultPomodoroDuration.Minutes()),
		DefaultBreakDuration:    int(settings.DefaultBreakDuration.Minutes()),
		DefaultTags:             settings.DefaultTags,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal settings: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
