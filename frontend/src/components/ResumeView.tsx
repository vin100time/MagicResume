import { useState, useEffect } from "react";
import ReactMarkdown from "react-markdown";
import { GetSummary, GetTranscript, CopyToClipboard } from "../../wailsjs/go/main/App";
import AudioPlayer from "./AudioPlayer";

interface Props {
  id: string;
}

function ResumeView({ id }: Props) {
  const [summary, setSummary] = useState("");
  const [transcript, setTranscript] = useState("");
  const [showTranscript, setShowTranscript] = useState(false);
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadContent();
  }, [id]);

  const loadContent = async () => {
    setLoading(true);
    try {
      const s = await GetSummary(id);
      setSummary(s);
    } catch {
      setSummary("");
    }
    try {
      const t = await GetTranscript(id);
      setTranscript(t);
    } catch {
      setTranscript("");
    }
    setLoading(false);
  };

  const handleCopy = async () => {
    const text = summary || transcript;
    if (!text) return;
    try {
      await CopyToClipboard(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (e) {
      console.error(e);
    }
  };

  if (loading) {
    return <div className="resume-view loading">Chargement...</div>;
  }

  return (
    <div className="resume-view">
      <AudioPlayer recordingId={id} />

      <div className="resume-toolbar">
        {(summary || transcript) && (
          <button className="btn-sm" onClick={handleCopy}>
            {copied ? "Copié !" : "Copier"}
          </button>
        )}
        {transcript && (
          <button className="btn-sm" onClick={() => setShowTranscript(!showTranscript)}>
            {showTranscript ? "Masquer transcription" : "Transcription"}
          </button>
        )}
      </div>

      {summary && (
        <div className="markdown-content">
          <ReactMarkdown>{summary}</ReactMarkdown>
        </div>
      )}

      {!summary && !transcript && (
        <div className="empty-view">
          <p>Cliquez sur "Résumer" pour transcrire et générer le résumé.</p>
        </div>
      )}

      {showTranscript && transcript && (
        <div className="transcript-content">
          <h3>Transcription brute</h3>
          <pre>{transcript}</pre>
        </div>
      )}
    </div>
  );
}

export default ResumeView;
