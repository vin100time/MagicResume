package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Recording struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Date           string `json:"date"`
	Duration       int    `json:"duration"` // seconds
	Status         string `json:"status"`   // pending, transcribed, summarized
	AudioPath      string `json:"audioPath"`
	TranscriptPath string `json:"transcriptPath,omitempty"`
	SummaryPath    string `json:"summaryPath,omitempty"`
}

type Settings struct {
	PromptTemplate string `json:"promptTemplate"`
	Model          string `json:"model"`
}

type Storage struct {
	mu           sync.RWMutex
	audioDir     string
	resumesDir   string
	metaPath     string
	settingsPath string
	recordings   []*Recording
	settings     Settings
}

func NewStorage(basePath string) *Storage {
	audioDir := filepath.Join(basePath, "audio")
	resumesDir := filepath.Join(basePath, "resumes")

	os.MkdirAll(audioDir, 0755)
	os.MkdirAll(resumesDir, 0755)

	s := &Storage{
		audioDir:     audioDir,
		resumesDir:   resumesDir,
		metaPath:     filepath.Join(basePath, "recordings.json"),
		settingsPath: filepath.Join(basePath, "settings.json"),
	}

	s.loadMetadata()
	s.loadSettings()
	return s
}

func (s *Storage) loadMetadata() {
	data, err := os.ReadFile(s.metaPath)
	if err != nil {
		s.recordings = []*Recording{}
		return
	}
	if err := json.Unmarshal(data, &s.recordings); err != nil {
		log.Println("Erreur lecture metadata:", err)
		s.recordings = []*Recording{}
	}
}

func (s *Storage) SaveMetadata() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.recordings, "", "  ")
	if err != nil {
		log.Println("Erreur sauvegarde metadata:", err)
		return
	}
	os.WriteFile(s.metaPath, data, 0644)
}

func (s *Storage) NameRecording(tempPath string, name string) (*Recording, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Sanitize name
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	if name == "" {
		name = "rdv-sans-nom"
	}

	// Get duration via soxi
	duration := 0
	out, err := exec.Command("soxi", "-D", tempPath).Output()
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &duration)
	}

	timestamp := time.Now().Format("2006-01-02_15h04")
	id := fmt.Sprintf("%s_%s", timestamp, name)

	// Rename audio file
	finalPath := filepath.Join(s.audioDir, id+".wav")
	if err := os.Rename(tempPath, finalPath); err != nil {
		return nil, fmt.Errorf("impossible de renommer l'audio: %w", err)
	}

	rec := &Recording{
		ID:        id,
		Name:      name,
		Date:      time.Now().Format("02/01/2006 15:04"),
		Duration:  duration,
		Status:    "pending",
		AudioPath: finalPath,
	}

	s.recordings = append(s.recordings, rec)

	// Save without lock (already held)
	data, _ := json.MarshalIndent(s.recordings, "", "  ")
	os.WriteFile(s.metaPath, data, 0644)

	return rec, nil
}

func (s *Storage) GetRecordings() []*Recording {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return sorted by date descending
	result := make([]*Recording, len(s.recordings))
	copy(result, s.recordings)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})
	return result
}

func (s *Storage) GetRecording(id string) *Recording {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.recordings {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func (s *Storage) DeleteRecording(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.recordings {
		if r.ID == id {
			// Remove files
			os.Remove(r.AudioPath)
			if r.TranscriptPath != "" {
				os.Remove(r.TranscriptPath)
			}
			if r.SummaryPath != "" {
				os.Remove(r.SummaryPath)
			}

			// Remove from slice
			s.recordings = append(s.recordings[:i], s.recordings[i+1:]...)

			// Save
			data, _ := json.MarshalIndent(s.recordings, "", "  ")
			os.WriteFile(s.metaPath, data, 0644)
			return nil
		}
	}
	return fmt.Errorf("enregistrement introuvable: %s", id)
}

func (s *Storage) loadSettings() {
	s.settings = Settings{
		PromptTemplate: defaultPromptTemplate,
		Model:          "haiku",
	}
	data, err := os.ReadFile(s.settingsPath)
	if err != nil {
		return
	}
	var saved Settings
	if err := json.Unmarshal(data, &saved); err == nil {
		if saved.PromptTemplate != "" {
			s.settings.PromptTemplate = saved.PromptTemplate
		}
		if saved.Model != "" {
			s.settings.Model = saved.Model
		}
	}
}

func (s *Storage) GetSettings() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Storage) SaveSettings(settings Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(s.settingsPath, data, 0644)
}

func (s *Storage) PurgeOldAudio(days int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	count := 0

	entries, err := os.ReadDir(s.audioDir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".wav") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(s.audioDir, entry.Name()))
			count++
		}
	}

	return count
}
