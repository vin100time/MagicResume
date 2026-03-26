# MagicResume

App macOS native pour enregistrer des rendez-vous, transcrire l'audio via Whisper et generer un resume structure via Claude.

## Fonctionnalites

- **Enregistrement audio** : capture micro avec pause/reprise, timer en temps reel
- **Transcription** : Whisper (modele small, francais)
- **Resume IA** : Claude CLI avec prompt personnalisable et choix du modele (Haiku/Sonnet/Opus)
- **Lecteur audio** : play/pause, seek, timestamps
- **Interface native** : app macOS sombre, rapide, zero SaaS

## Prerequis

- macOS (Apple Silicon ou Intel)
- [Go](https://go.dev/) >= 1.21
- [Wails](https://wails.io/) v2 : `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [Node.js](https://nodejs.org/) >= 18
- [SoX](https://sox.sourceforge.net/) : `brew install sox`
- [Whisper](https://github.com/openai/whisper) : `pip install openai-whisper`
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) : `npm install -g @anthropic-ai/claude-code`

## Build

```bash
git clone https://github.com/vin100time/MagicResume.git
cd MagicResume
wails build
```

L'app se trouve dans `build/bin/MagicResume.app`. Copier dans `/Applications` pour l'utiliser.

## Dev

```bash
wails dev
```

## Utilisation

1. **Enregistrer** : cliquer sur le bouton rouge, parler, stop, nommer le RDV
2. **Resumer** : selectionner l'enregistrement, cliquer "Resumer" — transcription + resume automatique
3. **Reglages** : icone roue crantee pour modifier le prompt et le modele Claude

## Structure

```
├── main.go              # Point d'entree Wails
├── app.go               # Logique principale + bindings
├── recorder.go          # Enregistrement audio (sox)
├── transcriber.go       # Transcription (whisper)
├── summarizer.go        # Resume (claude CLI)
├── storage.go           # Persistance JSON
├── audiohandler.go      # Serveur audio local (WAV -> MP3)
├── frontend/src/        # React TypeScript
│   ├── App.tsx
│   └── components/
│       ├── Recorder.tsx
│       ├── RecordingList.tsx
│       ├── ResumeView.tsx
│       ├── AudioPlayer.tsx
│       ├── SettingsModal.tsx
│       └── StatusBar.tsx
└── build/darwin/        # Config macOS (Info.plist)
```

## License

MIT
