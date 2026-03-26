interface Props {
  message: string;
  percent: number; // -1 = no bar, 0-100 = show progress
}

function StatusBar({ message, percent }: Props) {
  if (!message) return null;

  return (
    <div className="status-bar">
      <span className="status-dot" />
      <span className="status-text">{message}</span>
      {percent >= 0 && (
        <div className="progress-bar">
          <div className="progress-fill" style={{ width: `${percent}%` }} />
        </div>
      )}
    </div>
  );
}

export default StatusBar;
