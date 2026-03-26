package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	recorder    *Recorder
	storage     *Storage
	audioServer *AudioServer
	basePath    string
}

func NewApp() *App {
	home, _ := os.UserHomeDir()
	basePath := filepath.Join(home, "Resumer-RDV-App", "data")

	storage := NewStorage(basePath)
	return &App{
		storage:  storage,
		basePath: basePath,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.recorder = NewRecorder(a.storage, ctx)

	// Start local audio server
	a.audioServer = NewAudioServer(a)
	if err := a.audioServer.Start(); err != nil {
		log.Println("Erreur serveur audio:", err)
	}

	count := a.storage.PurgeOldAudio(30)
	if count > 0 {
		runtime.EventsEmit(ctx, "notification", fmt.Sprintf("%d ancien(s) audio(s) supprimé(s)", count))
	}

	log.Println("Résumer RDV started, data dir:", a.basePath)
}

func (a *App) shutdown(ctx context.Context) {
	if a.recorder != nil && a.recorder.IsRecording() {
		a.recorder.Stop()
	}
	if a.audioServer != nil {
		a.audioServer.Stop()
	}
}

// GetAudioURL returns the base URL for streaming audio files
func (a *App) GetAudioURL(id string) string {
	if a.audioServer == nil {
		return ""
	}
	return fmt.Sprintf("%s/audio/%s", a.audioServer.URL(), id)
}

// --- Recording ---

func (a *App) StartRecording() error {
	return a.recorder.Start()
}

func (a *App) PauseRecording() error {
	return a.recorder.Pause()
}

func (a *App) ResumeRecording() error {
	return a.recorder.Resume()
}

func (a *App) StopRecording() (string, error) {
	return a.recorder.Stop()
}

func (a *App) NameRecording(tempPath string, name string) (*Recording, error) {
	return a.storage.NameRecording(tempPath, name)
}

// --- Recordings list ---

func (a *App) GetRecordings() []*Recording {
	return a.storage.GetRecordings()
}

func (a *App) DeleteRecording(id string) error {
	return a.storage.DeleteRecording(id)
}

// --- Settings (prompt + model) ---

func (a *App) GetSettings() Settings {
	return a.storage.GetSettings()
}

func (a *App) SaveSettings(settings Settings) {
	a.storage.SaveSettings(settings)
}

// --- Transcribe + Summarize pipelined ---

func (a *App) TranscribeAndSummarize(id string) error {
	rec := a.storage.GetRecording(id)
	if rec == nil {
		return fmt.Errorf("enregistrement introuvable: %s", id)
	}

	// Step 1: Transcription
	runtime.EventsEmit(a.ctx, "status", map[string]string{
		"id":      id,
		"step":    "transcription",
		"message": "Transcription en cours...",
	})

	transcript, err := TranscribeAudio(rec.AudioPath)
	if err != nil {
		runtime.EventsEmit(a.ctx, "status", map[string]string{
			"id":      id,
			"step":    "error",
			"message": fmt.Sprintf("Erreur transcription: %v", err),
		})
		return err
	}

	// Step 2: Save transcript in background + launch Claude immediately
	runtime.EventsEmit(a.ctx, "status", map[string]string{
		"id":      id,
		"step":    "summary",
		"message": "Transcription OK — Résumé en cours...",
	})

	rec.TranscriptPath = filepath.Join(a.storage.resumesDir, rec.ID+"_transcription.txt")
	go func() {
		os.WriteFile(rec.TranscriptPath, []byte(transcript), 0644)
		rec.Status = "transcribed"
		a.storage.SaveMetadata()
	}()

	settings := a.storage.GetSettings()
	summary, err := SummarizeTranscript(transcript, rec.Name, settings.PromptTemplate, settings.Model)
	if err != nil {
		rec.Status = "transcribed"
		a.storage.SaveMetadata()
		runtime.EventsEmit(a.ctx, "recordings_updated", "")
		runtime.EventsEmit(a.ctx, "status", map[string]string{
			"id":      id,
			"step":    "error",
			"message": fmt.Sprintf("Erreur résumé: %v", err),
		})
		return err
	}

	rec.SummaryPath = filepath.Join(a.storage.resumesDir, rec.ID+"_resume.md")
	if err := os.WriteFile(rec.SummaryPath, []byte(summary), 0644); err != nil {
		return err
	}
	rec.Status = "summarized"
	a.storage.SaveMetadata()

	runtime.EventsEmit(a.ctx, "status", map[string]string{
		"id":      id,
		"step":    "summary_done",
		"message": "Résumé terminé",
	})
	runtime.EventsEmit(a.ctx, "recordings_updated", "")

	return nil
}

// --- Read content ---

func (a *App) GetTranscript(id string) (string, error) {
	rec := a.storage.GetRecording(id)
	if rec == nil {
		return "", fmt.Errorf("enregistrement introuvable")
	}
	if rec.TranscriptPath == "" {
		return "", fmt.Errorf("pas encore transcrit")
	}
	data, err := os.ReadFile(rec.TranscriptPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) GetSummary(id string) (string, error) {
	rec := a.storage.GetRecording(id)
	if rec == nil {
		return "", fmt.Errorf("enregistrement introuvable")
	}
	if rec.SummaryPath == "" {
		return "", fmt.Errorf("pas encore résumé")
	}
	data, err := os.ReadFile(rec.SummaryPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- Clipboard ---

func (a *App) CopyToClipboard(text string) error {
	runtime.ClipboardSetText(a.ctx, text)
	return nil
}
