package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const whisperBin = "/Users/vince/Library/Python/3.9/bin/whisper"
const whisperModel = "small"

func TranscribeAudio(audioPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "whisper-*")
	if err != nil {
		return "", fmt.Errorf("impossible de créer le dossier temporaire: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command(whisperBin, audioPath,
		"--language", "fr",
		"--model", whisperModel,
		"--output_format", "txt",
		"--output_dir", tmpDir,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("whisper a échoué: %w\n%s", err, string(output))
	}

	// Find the .txt output file
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".txt") {
			data, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
			if err != nil {
				return "", err
			}
			text := strings.TrimSpace(string(data))
			if text == "" {
				return "", fmt.Errorf("transcription vide")
			}
			return text, nil
		}
	}

	return "", fmt.Errorf("aucun fichier de transcription généré")
}
