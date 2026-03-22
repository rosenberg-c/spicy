package params

// Base builds a normalized parameter map for history entries.
func Base(model string, verbose bool, history bool, save bool, output string) map[string]interface{} {
	return map[string]interface{}{
		"model":   model,
		"verbose": verbose,
		"history": history,
		"save":    save,
		"output":  output,
	}
}
