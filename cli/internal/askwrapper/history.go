package askwrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	historyDirName  = ".askwrapper"
	historyFileName = "history.json"
)

func HistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, historyDirName, historyFileName), nil
}

func LoadHistory() ([]HistoryEntry, error) {
	path, err := HistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []HistoryEntry{}, nil
		}
		return nil, err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return []HistoryEntry{}, nil
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse history %s: %w", path, err)
	}

	return entries, nil
}

func AppendHistory(question, answer string) error {
	q := strings.TrimSpace(question)
	if q == "" {
		return nil
	}

	entries, err := LoadHistory()
	if err != nil {
		return err
	}

	entry := HistoryEntry{
		Question: q,
		Answer:   strings.TrimSpace(answer),
		At:       time.Now().Unix(),
	}
	entries = append([]HistoryEntry{entry}, entries...)
	return saveAll(entries)
}

func saveAll(entries []HistoryEntry) error {

	path, err := HistoryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), "history-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func DeleteHistoryAt(index int) error {
	entries, err := LoadHistory()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(entries) {
		return nil
	}

	entries = append(entries[:index], entries[index+1:]...)
	return saveAll(entries)
}
