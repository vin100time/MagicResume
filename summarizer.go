package main

import (
	"fmt"
	"os/exec"
	"strings"
)

const defaultPromptTemplate = `Tu es un assistant qui résume des réunions professionnelles.
Voici la transcription d'un RDV nommé '{{NOM}}'.
Produis un résumé structuré en français avec:

## Contexte
Brève description du RDV

## Points clés discutés
- Les sujets principaux abordés

## Décisions prises
- Ce qui a été décidé

## Actions à faire
- [ ] Action — Responsable — Deadline (si mentionnée)

## Notes importantes
- Tout détail important à retenir

---
Transcription:

{{TRANSCRIPTION}}`

func SummarizeTranscript(transcript string, rdvName string, promptTemplate string, model string) (string, error) {
	if promptTemplate == "" {
		promptTemplate = defaultPromptTemplate
	}
	if model == "" {
		model = "haiku"
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{NOM}}", rdvName)
	prompt = strings.ReplaceAll(prompt, "{{TRANSCRIPTION}}", transcript)

	cmd := exec.Command("claude", "--print", "--model", model, prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude a échoué: %w\n%s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("résumé vide")
	}

	return result, nil
}
