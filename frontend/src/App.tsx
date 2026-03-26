import { useState, useEffect } from "react";
import { EventsOn } from "../wailsjs/runtime/runtime";
import Recorder from "./components/Recorder";
import RecordingList from "./components/RecordingList";
import ResumeView from "./components/ResumeView";
import StatusBar from "./components/StatusBar";
import SettingsModal from "./components/SettingsModal";
import "./styles/app.css";

function App() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  const [statusMessage, setStatusMessage] = useState("");
  const [statusPercent, setStatusPercent] = useState(-1);
  const [settingsOpen, setSettingsOpen] = useState(false);

  useEffect(() => {
    EventsOn("recordings_updated", () => {
      setRefreshKey((k) => k + 1);
    });
    EventsOn("notification", (msg: string) => {
      setStatusMessage(msg);
      setStatusPercent(-1);
      setTimeout(() => setStatusMessage(""), 5000);
    });
    EventsOn("status", (data: { id: string; step: string; message: string; percent?: number }) => {
      setStatusMessage(data.message);
      setStatusPercent(typeof data.percent === "number" ? data.percent : -1);
      if (data.step.endsWith("_done") || data.step === "error") {
        setTimeout(() => {
          setStatusMessage("");
          setStatusPercent(-1);
        }, 4000);
      }
    });
  }, []);

  return (
    <div className="app">
      <div className="titlebar-drag" />
      <header className="app-header">
        <h1>Résumer RDV</h1>
        <button className="btn-settings" onClick={() => setSettingsOpen(true)}>
          ⚙
        </button>
      </header>
      <main className="app-main">
        <div className="left-panel">
          <Recorder onRecordingSaved={() => setRefreshKey((k) => k + 1)} />
          <RecordingList
            key={refreshKey}
            selectedId={selectedId}
            onSelect={setSelectedId}
            onDeleted={() => {
              setSelectedId(null);
              setRefreshKey((k) => k + 1);
            }}
          />
        </div>
        <div className="right-panel">
          {selectedId ? (
            <ResumeView id={selectedId} key={selectedId + refreshKey} />
          ) : (
            <div className="placeholder">
              Sélectionnez un enregistrement pour voir le résumé
            </div>
          )}
        </div>
      </main>
      <StatusBar message={statusMessage} percent={statusPercent} />
      <SettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </div>
  );
}

export default App;
