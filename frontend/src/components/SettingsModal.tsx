import { useState, useEffect } from "react";
import { GetSettings, SaveSettings } from "../../wailsjs/go/main/App";

interface Props {
  open: boolean;
  onClose: () => void;
}

const models = [
  { value: "haiku", label: "Haiku (rapide)" },
  { value: "sonnet", label: "Sonnet (équilibré)" },
  { value: "opus", label: "Opus (puissant)" },
];

function SettingsModal({ open, onClose }: Props) {
  const [promptTemplate, setPromptTemplate] = useState("");
  const [model, setModel] = useState("haiku");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (open) {
      GetSettings().then((s) => {
        setPromptTemplate(s.promptTemplate || "");
        setModel(s.model || "haiku");
        setSaved(false);
      });
    }
  }, [open]);

  const handleSave = async () => {
    await SaveSettings({ promptTemplate, model });
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  if (!open) return null;

  return (
    <div className="confirm-overlay" onClick={onClose}>
      <div className="settings-modal" onClick={(e) => e.stopPropagation()}>
        <div className="settings-header">
          <h2>Réglages</h2>
          <button className="btn-close" onClick={onClose}>×</button>
        </div>

        <div className="settings-field">
          <label>Modèle Claude</label>
          <select value={model} onChange={(e) => setModel(e.target.value)}>
            {models.map((m) => (
              <option key={m.value} value={m.value}>{m.label}</option>
            ))}
          </select>
        </div>

        <div className="settings-field">
          <label>
            Prompt de résumé
            <span className="hint">{" ({{NOM}} et {{TRANSCRIPTION}} seront remplacés)"}</span>
          </label>
          <textarea
            value={promptTemplate}
            onChange={(e) => setPromptTemplate(e.target.value)}
            rows={14}
          />
        </div>

        <div className="settings-footer">
          {saved && <div className="toast">Paramètres sauvegardés</div>}
          <button className="btn-save" onClick={handleSave}>
            Sauvegarder
          </button>
        </div>
      </div>
    </div>
  );
}

export default SettingsModal;
