import { useRef, useState, useEffect } from "react";
import { GetAudioURL } from "../../wailsjs/go/main/App";

interface Props {
  recordingId: string;
}

function AudioPlayer({ recordingId }: Props) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);
  const [currentTime, setCurrent] = useState(0);
  const [duration, setDuration] = useState(0);
  const [audioSrc, setAudioSrc] = useState("");

  useEffect(() => {
    setPlaying(false);
    setCurrent(0);
    setDuration(0);
    setAudioSrc("");
    GetAudioURL(recordingId).then((url) => {
      if (url) setAudioSrc(url);
    });
  }, [recordingId]);

  const toggle = () => {
    const a = audioRef.current;
    if (!a) return;
    if (playing) {
      a.pause();
    } else {
      a.play();
    }
    setPlaying(!playing);
  };

  const onTimeUpdate = () => {
    if (audioRef.current) setCurrent(audioRef.current.currentTime);
  };

  const onLoaded = () => {
    if (audioRef.current) setDuration(audioRef.current.duration);
  };

  const onEnded = () => setPlaying(false);

  const seek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const t = parseFloat(e.target.value);
    if (audioRef.current) {
      audioRef.current.currentTime = t;
      setCurrent(t);
    }
  };

  const fmt = (s: number) => {
    if (!s || !isFinite(s)) return "00:00";
    const m = Math.floor(s / 60);
    const sec = Math.floor(s % 60);
    return `${m.toString().padStart(2, "0")}:${sec.toString().padStart(2, "0")}`;
  };

  if (!audioSrc) return null;

  return (
    <div className="audio-player">
      <audio
        ref={audioRef}
        src={audioSrc}
        onTimeUpdate={onTimeUpdate}
        onLoadedMetadata={onLoaded}
        onEnded={onEnded}
        preload="metadata"
      />
      <button className="btn-play" onClick={toggle}>
        {playing ? "⏸" : "▶"}
      </button>
      <span className="audio-time">{fmt(currentTime)}</span>
      <input
        type="range"
        className="audio-seek"
        min={0}
        max={duration || 0}
        step={0.1}
        value={currentTime}
        onChange={seek}
      />
      <span className="audio-time">{fmt(duration)}</span>
    </div>
  );
}

export default AudioPlayer;
