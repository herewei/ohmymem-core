package huh

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// ErrCancelled is returned when user cancels the prompt (Ctrl+C)
var ErrCancelled = errors.New("user cancelled")

// Confirm asks for yes/no confirmation
// Returns error if user interrupts (never silently falls back)
func Confirm(message string, defaultYes bool) (bool, error) {
	confirmed := defaultYes

	err := huh.NewConfirm().
		Title(message).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, ErrCancelled
		}
		return false, fmt.Errorf("prompt failed: %w", err)
	}

	return confirmed, nil
}

// SelectOne shows single-select prompt with arrow key navigation
func SelectOne(label string, options []string) (int, string, error) {
	if len(options) == 0 {
		return -1, "", nil
	}

	var selected string
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewSelect[string]().
		Title(label).
		Options(huhOptions...).
		Value(&selected).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return -1, "", ErrCancelled
		}
		return -1, "", fmt.Errorf("prompt failed: %w", err)
	}

	// Find index
	for i, opt := range options {
		if opt == selected {
			return i, selected, nil
		}
	}

	return -1, selected, nil
}

// ContextOption represents a selectable context for multi-select
type ContextOption struct {
	ID          string
	Name        string
	Description string
	Selected    bool // Pre-selected based on detection
}

// SelectContexts shows interactive multi-select for context selection
// Uses huh.MultiSelect with native checkbox support
// Returns slice of selected context IDs
func SelectContexts(label string, options []ContextOption) ([]string, error) {
	if len(options) == 0 {
		return nil, nil
	}

	// Build huh options with pre-selection
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		displayText := opt.Name
		if opt.Description != "" {
			desc := opt.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			displayText = fmt.Sprintf("%s - %s", opt.Name, desc)
		}
		huhOpt := huh.NewOption(displayText, opt.ID)
		if opt.Selected {
			huhOpt = huhOpt.Selected(true)
		}
		huhOptions[i] = huhOpt
	}

	var selected []string
	err := huh.NewMultiSelect[string]().
		Title(label).
		Options(huhOptions...).
		Value(&selected).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrCancelled
		}
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	return selected, nil
}

// PromptInput prompts for text input with optional default value
// Returns error if user interrupts (never silently falls back)
func PromptInput(label string, defaultValue string) (string, error) {
	value := defaultValue

	err := huh.NewInput().
		Title(label).
		Value(&value).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	return value, nil
}
