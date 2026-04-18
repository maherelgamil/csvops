import { useState } from "react";
import { OpenCSVFile, PreviewCSV } from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import "./App.css";

export default function App() {
  const [path, setPath] = useState<string>("");
  const [rows, setRows] = useState<number>(20);
  const [preview, setPreview] = useState<main.PreviewPayload | null>(null);
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState(false);

  async function pickFile() {
    setError("");
    try {
      const p = await OpenCSVFile();
      if (p) {
        setPath(p);
        await loadPreview(p, rows);
      }
    } catch (e: any) {
      setError(String(e));
    }
  }

  async function loadPreview(p: string, n: number) {
    setLoading(true);
    setError("");
    try {
      const r = await PreviewCSV(p, n, false);
      setPreview(r);
    } catch (e: any) {
      setError(String(e));
      setPreview(null);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="app">
      <header>
        <h1>csvops</h1>
        <p className="tagline">Open a CSV to preview it.</p>
      </header>

      <section className="toolbar">
        <button onClick={pickFile} disabled={loading}>
          📂 Open CSV…
        </button>
        {path && (
          <>
            <code className="path" title={path}>
              {path}
            </code>
            <label className="rows">
              Rows:
              <input
                type="number"
                min={1}
                max={1000}
                value={rows}
                onChange={(e) => setRows(Number(e.target.value))}
              />
            </label>
            <button onClick={() => loadPreview(path, rows)} disabled={loading}>
              🔄 Reload
            </button>
          </>
        )}
      </section>

      {error && <div className="error">⚠️ {error}</div>}
      {loading && <div className="loading">Loading…</div>}

      {preview && preview.rows.length > 0 && (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                {preview.headers.map((h, i) => (
                  <th key={i}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {preview.rows.map((row, i) => (
                <tr key={i}>
                  {row.map((cell, j) => (
                    <td key={j}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          <div className="meta">Showing {preview.rows.length} row(s)</div>
        </div>
      )}

      {preview && preview.rows.length === 0 && (
        <div className="empty">No rows in file.</div>
      )}
    </div>
  );
}
