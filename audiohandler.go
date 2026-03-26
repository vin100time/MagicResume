package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// AudioServer runs a local HTTP server to serve audio as MP3 to the WebView.
type AudioServer struct {
	app      *App
	port     int
	listener net.Listener
	cache    map[string]string // recordingID -> mp3 path
	mu       sync.Mutex
}

func NewAudioServer(app *App) *AudioServer {
	return &AudioServer{
		app:   app,
		cache: make(map[string]string),
	}
}

func (s *AudioServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/audio/", s.handleAudio)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("impossible de démarrer le serveur audio: %w", err)
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	go http.Serve(listener, mux)
	log.Printf("Audio server listening on http://127.0.0.1:%d", s.port)
	return nil
}

func (s *AudioServer) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	// Clean up cached mp3 files
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, path := range s.cache {
		os.Remove(path)
	}
}

func (s *AudioServer) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

// convertToMP3 converts a WAV file to MP3 using sox, caching the result.
func (s *AudioServer) convertToMP3(rec *Recording) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check cache - use file mod time to invalidate
	info, err := os.Stat(rec.AudioPath)
	if err != nil {
		return "", err
	}
	cacheKey := fmt.Sprintf("%s_%x", rec.ID, md5.Sum([]byte(fmt.Sprintf("%d", info.ModTime().UnixNano()))))

	if mp3Path, ok := s.cache[cacheKey]; ok {
		if _, err := os.Stat(mp3Path); err == nil {
			return mp3Path, nil
		}
	}

	mp3Path := filepath.Join(os.TempDir(), fmt.Sprintf("resumer-rdv-%s.mp3", rec.ID))
	cmd := exec.Command("sox", rec.AudioPath, "-C", "128", mp3Path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("conversion mp3 échouée: %w\n%s", err, string(output))
	}

	s.cache[cacheKey] = mp3Path
	return mp3Path, nil
}

func (s *AudioServer) handleAudio(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/audio/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	rec := s.app.storage.GetRecording(id)
	if rec == nil {
		http.NotFound(w, r)
		return
	}

	mp3Path, err := s.convertToMP3(rec)
	if err != nil {
		log.Println("Erreur conversion MP3:", err)
		http.Error(w, "conversion error", 500)
		return
	}

	f, err := os.Open(mp3Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, mp3Path, stat.ModTime(), f)
}
