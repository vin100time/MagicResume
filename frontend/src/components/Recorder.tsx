import { useState, useEffect } from "react";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import {
  StartRecording,
  StopRecording,
  PauseRecording,
  ResumeRecording,
  NameRecording,
} from "../../wailsjs/go/main/App";

interface Props {
  onRecordingSaved: () => void;
}

function Recorder({ onRecordingSaved }: Props) {
  const [state, setState] = useState<"idle" | "recording" | "paused" | "naming">("idle");
  const [seconds, setSeconds] = useState(0);
  const [tempPath, setTempPath] = useState("");
  const [rdvName, setRdvName] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    EventsOn("recording_timer", (s: number) => {
      setSeconds(s);
    });
  }, []);

  const formatTime = (s: number) => {
    const m = Math.floor(s / 60);
    const sec = s % 60;
    return `${m.toString().padStart(2, "0")}:${sec.toString().padStart(2, "0")}`;
  };

  const handleStart = async () => {
    setError("");
    try {
      await StartRecording();
      setState("recording");
      setSeconds(0);
    } catch (e: any) {
      setError(e);
    }
  };

  const handlePause = async () => {
    try {
      await PauseRecording();
      setState("paused");
    } catch (e: any) {
      setError(e);
    }
  };

  const handleResume = async () => {
    try {
      await ResumeRecording();
      setState("recording");
    } catch (e: any) {
      setError(e);
    }
  };

  const handleStop = async () => {
    try {
      const path = await StopRecording();
      setTempPath(path);
      setState("naming");
    } catch (e: any) {
      setError(e);
      setState("idle");
    }
  };

  const handleSave = async () => {
    if (!rdvName.trim()) return;
    try {
      await NameRecording(tempPath, rdvName.trim());
      setState("idle");
      setRdvName("");
      setTempPath("");
      setSeconds(0);
      onRecordingSaved();
    } catch (e: any) {
      setError(e);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSave();
  };

  return (
    <div className="recorder">
      {state === "idle" && (
        <button className="btn-record" onClick={handleStart}>
          Enregistrer
        </button>
      )}

      {(state === "recording" || state === "paused") && (
        <div className="recording-controls">
          <div className={`timer ${state === "paused" ? "paused" : ""}`}>
            {formatTime(seconds)}
          </div>
          <div className={`recording-indicator ${state === "recording" ? "active" : ""}`} />
          <div className="btn-group">
            {state === "recording" ? (
              <button className="btn-pause" onClick={handlePause}>
                Pause
              </button>
            ) : (
              <button className="btn-resume" onClick={handleResume}>
                Reprendre
              </button>
            )}
            <button className="btn-stop" onClick={handleStop}>
              Stop
            </button>
          </div>
        </div>
      )}

      {state === "naming" && (
        <div className="naming-dialog">
          <label>Nom du RDV :</label>
          <input
            type="text"
            value={rdvName}
            onChange={(e) => setRdvName(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="ex: rdv-client-dupont"
            autoFocus
          />
          <button className="btn-save" onClick={handleSave} disabled={!rdvName.trim()}>
            Sauvegarder
          </button>
        </div>
      )}

      {error && <div className="error">{String(error)}</div>}
    </div>
  );
}

export default Recorder;
