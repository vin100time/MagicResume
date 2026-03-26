import { useState, useEffect } from "react";
import {
  GetRecordings,
  DeleteRecording,
  TranscribeAndSummarize,
} from "../../wailsjs/go/main/App";

interface Recording {
  id: string;
  name: string;
  date: string;
  duration: number;
  status: string;
}

interface Props {
  selectedId: string | null;
  onSelect: (id: string) => void;
  onDeleted: () => void;
}

function RecordingList({ selectedId, onSelect, onDeleted }: Props) {
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [processing, setProcessing] = useState<Set<string>>(new Set());
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  useEffect(() => {
    loadRecordings();
  }, []);

  const loadRecordings = async () => {
    try {
      const recs = await GetRecordings();
      setRecordings(recs || []);
    } catch {
      setRecordings([]);
    }
  };

  const handleDeleteClick = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setConfirmDeleteId(id);
  };

  const handleDeleteConfirm = async () => {
    if (!confirmDeleteId) return;
    try {
      await DeleteRecording(confirmDeleteId);
      setConfirmDeleteId(null);
      onDeleted();
    } catch (err) {
      console.error(err);
    }
  };

  const handleResume = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setProcessing((prev) => new Set(prev).add(id));
    try {
      await TranscribeAndSummarize(id);
      await loadRecordings();
    } catch (err) {
      console.error(err);
    } finally {
      setProcessing((prev) => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  };

  const formatDuration = (s: number) => {
    if (!s) return "—";
    const m = Math.floor(s / 60);
    const sec = s % 60;
    return `${m}m${sec.toString().padStart(2, "0")}s`;
  };

  const statusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return <span className="badge badge-pending">Audio</span>;
      case "transcribed":
        return <span className="badge badge-transcribed">Transcrit</span>;
      case "summarized":
        return <span className="badge badge-summarized">Résumé</span>;
      default:
        return null;
    }
  };

  return (
    <div className="recording-list">
      <h2>Enregistrements</h2>
      {recordings.length === 0 ? (
        <p className="empty">Aucun enregistrement</p>
      ) : (
        <ul>
          {recordings.map((rec) => (
            <li
              key={rec.id}
              className={`recording-item ${selectedId === rec.id ? "selected" : ""}`}
              onClick={() => onSelect(rec.id)}
            >
              <div className="rec-info">
                <span className="rec-name">{rec.name}</span>
                <span className="rec-meta">
                  {rec.date} · {formatDuration(rec.duration)}
                </span>
              </div>
              <div className="rec-actions">
                {statusBadge(rec.status)}
                {rec.status !== "summarized" && (
                  <button
                    className="btn-sm btn-process"
                    onClick={(e) => handleResume(rec.id, e)}
                    disabled={processing.has(rec.id)}
                  >
                    {processing.has(rec.id) ? "..." : "Résumer"}
                  </button>
                )}
                <button
                  className="btn-sm btn-delete"
                  onClick={(e) => handleDeleteClick(rec.id, e)}
                >
                  ×
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
      {confirmDeleteId && (
        <div className="confirm-overlay" onClick={() => setConfirmDeleteId(null)}>
          <div className="confirm-dialog" onClick={(e) => e.stopPropagation()}>
            <p>Supprimer cet enregistrement ?</p>
            <div className="confirm-buttons">
              <button className="btn-sm" onClick={() => setConfirmDeleteId(null)}>
                Annuler
              </button>
              <button className="btn-sm btn-delete" onClick={handleDeleteConfirm}>
                Supprimer
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default RecordingList;
