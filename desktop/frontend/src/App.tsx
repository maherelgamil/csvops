import { useEffect, useState } from "react";
import {
  OpenCSVFile,
  OpenDirectory,
  PreviewCSV,
  StatsCSV,
  FilterCSV,
  SaveCSVFile,
  SaveDBFile,
  SplitCSV,
  DedupeCSV,
  MergeCSV,
  ToSQLiteCSV,
} from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";
import "./App.css";

type Tab = "preview" | "stats" | "filter" | "split" | "dedupe" | "merge" | "sqlite";

type ProgressEvent = { op: string; done: number; total: number };

export default function App() {
  const [path, setPath] = useState<string>("");
  const [tab, setTab] = useState<Tab>("preview");
  const [progress, setProgress] = useState<ProgressEvent | null>(null);

  useEffect(() => {
    EventsOn("progress", (p: ProgressEvent) => setProgress(p));
    return () => EventsOff("progress");
  }, []);

  async function pickFile() {
    const p = await OpenCSVFile();
    if (p) setPath(p);
  }

  // For tabs that don't need a single input file (merge), still allow opening one for context.
  return (
    <div className="app">
      <header>
        <h1>csvops</h1>
        <p className="tagline">CSV operations toolkit.</p>
      </header>

      <section className="toolbar">
        <button onClick={pickFile}>📂 Open CSV…</button>
        {path && (
          <code className="path" title={path}>
            {path}
          </code>
        )}
      </section>

      <nav className="tabs">
        <TabButton id="preview" active={tab} onClick={setTab}>Preview</TabButton>
        <TabButton id="stats" active={tab} onClick={setTab}>Stats</TabButton>
        <TabButton id="filter" active={tab} onClick={setTab}>Filter</TabButton>
        <TabButton id="dedupe" active={tab} onClick={setTab}>Dedupe</TabButton>
        <TabButton id="split" active={tab} onClick={setTab}>Split</TabButton>
        <TabButton id="merge" active={tab} onClick={setTab}>Merge</TabButton>
        <TabButton id="sqlite" active={tab} onClick={setTab}>To SQLite</TabButton>
      </nav>

      <main className="tab-content">
        {tab === "preview" && <PreviewTab path={path} />}
        {tab === "stats" && <StatsTab path={path} />}
        {tab === "filter" && <FilterTab path={path} />}
        {tab === "dedupe" && <DedupeTab path={path} />}
        {tab === "split" && <SplitTab path={path} />}
        {tab === "merge" && <MergeTab />}
        {tab === "sqlite" && <SQLiteTab path={path} />}
      </main>

      {progress && progress.total > 0 && progress.done < progress.total && (
        <ProgressBar p={progress} />
      )}
    </div>
  );
}

function TabButton({ id, active, onClick, children }: {
  id: Tab; active: Tab; onClick: (id: Tab) => void; children: React.ReactNode;
}) {
  return (
    <button
      className={`tab ${active === id ? "tab-active" : ""}`}
      onClick={() => onClick(id)}
    >
      {children}
    </button>
  );
}

function ProgressBar({ p }: { p: ProgressEvent }) {
  const pct = p.total > 0 ? Math.round((p.done / p.total) * 100) : 0;
  return (
    <div className="progress-wrap">
      <div className="progress-label">
        {p.op}: {p.done}/{p.total} ({pct}%)
      </div>
      <div className="progress-bar">
        <div className="progress-fill" style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

function NeedFile() {
  return <div className="empty-state">Open a CSV file first.</div>;
}

function Error({ msg }: { msg: string }) {
  return <div className="error">⚠️ {msg}</div>;
}

function Success({ children }: { children: React.ReactNode }) {
  return <div className="success">✅ {children}</div>;
}

// ---------- Preview --------------------------------------------------------

function PreviewTab({ path }: { path: string }) {
  const [rows, setRows] = useState(20);
  const [data, setData] = useState<main.PreviewPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  if (!path) return <NeedFile />;

  async function run() {
    setLoading(true); setErr("");
    try { setData(await PreviewCSV(path, rows, false)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-row">
        <label>Rows:
          <input type="number" min={1} max={1000} value={rows}
            onChange={(e) => setRows(Number(e.target.value))} />
        </label>
        <button onClick={run} disabled={loading}>{loading ? "…" : "Preview"}</button>
      </div>
      {err && <Error msg={err} />}
      {data && (
        <div className="table-wrap">
          <table>
            <thead><tr>{data.headers.map((h, i) => <th key={i}>{h}</th>)}</tr></thead>
            <tbody>
              {data.rows.map((r, i) => (
                <tr key={i}>{r.map((c, j) => <td key={j}>{c}</td>)}</tr>
              ))}
            </tbody>
          </table>
          <div className="meta">{data.rows.length} row(s)</div>
        </div>
      )}
    </div>
  );
}

// ---------- Stats ----------------------------------------------------------

function StatsTab({ path }: { path: string }) {
  const [maxUnique, setMaxUnique] = useState(100000);
  const [data, setData] = useState<main.StatsPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  if (!path) return <NeedFile />;

  async function run() {
    setLoading(true); setErr("");
    try { setData(await StatsCSV(path, maxUnique)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-row">
        <label>Max unique tracked:
          <input type="number" min={0} value={maxUnique}
            onChange={(e) => setMaxUnique(Number(e.target.value))} />
        </label>
        <button onClick={run} disabled={loading}>{loading ? "Analyzing…" : "Analyze"}</button>
      </div>
      {err && <Error msg={err} />}
      {data && (
        <div className="table-wrap">
          <div className="stats-summary">
            Total rows: <strong>{data.totalRows}</strong> · Columns: <strong>{data.columns.length}</strong>
          </div>
          <table>
            <thead><tr><th>Column</th><th>Unique</th><th>Empty</th><th>Top 3</th></tr></thead>
            <tbody>
              {data.columns.map((c, i) => (
                <tr key={i}>
                  <td>{c.name}</td>
                  <td>{c.uniqueCapped ? `≥${c.unique} (capped)` : c.unique}</td>
                  <td>{c.empty}</td>
                  <td>{c.top.map((t) => `${t.value} (${t.count})`).join(", ")}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

// ---------- Filter ---------------------------------------------------------

function FilterTab({ path }: { path: string }) {
  const [column, setColumn] = useState("");
  const [eqSet, setEqSet] = useState(false);
  const [eq, setEq] = useState("");
  const [containsSet, setContainsSet] = useState(false);
  const [contains, setContains] = useState("");
  const [gtSet, setGtSet] = useState(false);
  const [gt, setGt] = useState(0);
  const [ltSet, setLtSet] = useState(false);
  const [lt, setLt] = useState(0);
  const [all, setAll] = useState(false);
  const [output, setOutput] = useState("");
  const [result, setResult] = useState<main.FilterPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  if (!path) return <NeedFile />;

  async function pickOutput() {
    const p = await SaveCSVFile("filtered.csv");
    if (p) setOutput(p);
  }

  async function run() {
    if (!output) { setErr("Pick an output file first."); return; }
    setLoading(true); setErr("");
    try {
      setResult(await FilterCSV({
        input: path, output, column,
        eq, eqSet, contains, containsSet, gt, gtSet, lt, ltSet,
        all, withHeader: true,
      } as any));
    } catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-grid">
        <label>Column
          <input type="text" value={column} onChange={(e) => setColumn(e.target.value)} placeholder="e.g. country" />
        </label>
        <label className="condition">
          <input type="checkbox" checked={eqSet} onChange={(e) => setEqSet(e.target.checked)} />
          equals <input type="text" value={eq} disabled={!eqSet} onChange={(e) => setEq(e.target.value)} />
        </label>
        <label className="condition">
          <input type="checkbox" checked={containsSet} onChange={(e) => setContainsSet(e.target.checked)} />
          contains <input type="text" value={contains} disabled={!containsSet} onChange={(e) => setContains(e.target.value)} />
        </label>
        <label className="condition">
          <input type="checkbox" checked={gtSet} onChange={(e) => setGtSet(e.target.checked)} />
          &gt; <input type="number" value={gt} disabled={!gtSet} onChange={(e) => setGt(Number(e.target.value))} />
        </label>
        <label className="condition">
          <input type="checkbox" checked={ltSet} onChange={(e) => setLtSet(e.target.checked)} />
          &lt; <input type="number" value={lt} disabled={!ltSet} onChange={(e) => setLt(Number(e.target.value))} />
        </label>
        <label>
          <input type="checkbox" checked={all} onChange={(e) => setAll(e.target.checked)} />
          Match ALL conditions (AND)
        </label>
      </div>
      <div className="form-row">
        <button onClick={pickOutput}>💾 Output file…</button>
        {output && <code className="path" title={output}>{output}</code>}
      </div>
      <div className="form-row">
        <button onClick={run} disabled={loading || !column}>{loading ? "Filtering…" : "Run filter"}</button>
      </div>
      {err && <Error msg={err} />}
      {result && (
        <Success>
          Matched <strong>{result.matched}</strong> of <strong>{result.totalRows}</strong> rows.
          Written to <code>{output}</code>.
        </Success>
      )}
    </div>
  );
}

// ---------- Dedupe --------------------------------------------------------

function DedupeTab({ path }: { path: string }) {
  const [keys, setKeys] = useState("");
  const [keepLast, setKeepLast] = useState(false);
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [output, setOutput] = useState("");
  const [result, setResult] = useState<main.DedupePayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  if (!path) return <NeedFile />;

  async function pickOutput() {
    const p = await SaveCSVFile("dedup.csv");
    if (p) setOutput(p);
  }

  async function run() {
    if (!output) { setErr("Pick an output file first."); return; }
    if (!keys.trim()) { setErr("Enter at least one key column."); return; }
    setLoading(true); setErr("");
    try {
      setResult(await DedupeCSV({
        input: path, output, keyColumns: keys,
        keepLast, caseSensitive,
      } as any));
    } catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-grid">
        <label>Key column(s) — comma-separated
          <input type="text" value={keys} onChange={(e) => setKeys(e.target.value)} placeholder="e.g. email or first,last" />
        </label>
        <label>
          <input type="checkbox" checked={keepLast} onChange={(e) => setKeepLast(e.target.checked)} />
          Keep last occurrence (default keeps first)
        </label>
        <label>
          <input type="checkbox" checked={caseSensitive} onChange={(e) => setCaseSensitive(e.target.checked)} />
          Case-sensitive comparison
        </label>
      </div>
      <div className="form-row">
        <button onClick={pickOutput}>💾 Output file…</button>
        {output && <code className="path" title={output}>{output}</code>}
      </div>
      <div className="form-row">
        <button onClick={run} disabled={loading}>{loading ? "Deduplicating…" : "Run dedupe"}</button>
      </div>
      {err && <Error msg={err} />}
      {result && (
        <Success>
          {result.uniqueRows} unique row(s) kept · {result.duplicates} duplicates removed
          (total: {result.totalRows}). Written to <code>{output}</code>.
        </Success>
      )}
    </div>
  );
}

// ---------- Split ---------------------------------------------------------

function SplitTab({ path }: { path: string }) {
  const [outDir, setOutDir] = useState("");
  const [rowsPerFile, setRowsPerFile] = useState(1000);
  const [withHeader, setWithHeader] = useState(true);
  const [result, setResult] = useState<main.SplitPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  if (!path) return <NeedFile />;

  async function pickOutDir() {
    const p = await OpenDirectory("Select output directory");
    if (p) setOutDir(p);
  }

  async function run() {
    if (!outDir) { setErr("Pick an output directory first."); return; }
    setLoading(true); setErr("");
    try {
      setResult(await SplitCSV({
        input: path, outputDir: outDir,
        rowsPerFile, withHeader,
      } as any));
    } catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-grid">
        <label>Rows per output file
          <input type="number" min={1} value={rowsPerFile} onChange={(e) => setRowsPerFile(Number(e.target.value))} />
        </label>
        <label>
          <input type="checkbox" checked={withHeader} onChange={(e) => setWithHeader(e.target.checked)} />
          Include header in each chunk
        </label>
      </div>
      <div className="form-row">
        <button onClick={pickOutDir}>📁 Output directory…</button>
        {outDir && <code className="path" title={outDir}>{outDir}</code>}
      </div>
      <div className="form-row">
        <button onClick={run} disabled={loading}>{loading ? "Splitting…" : "Run split"}</button>
      </div>
      {err && <Error msg={err} />}
      {result && (
        <Success>
          Wrote {result.filesCreated} file(s) totaling {result.rowsProcessed} row(s) into <code>{outDir}</code>.
        </Success>
      )}
    </div>
  );
}

// ---------- Merge ---------------------------------------------------------

function MergeTab() {
  const [inDir, setInDir] = useState("");
  const [output, setOutput] = useState("");
  const [withHeader, setWithHeader] = useState(true);
  const [result, setResult] = useState<main.MergePayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function pickInDir() {
    const p = await OpenDirectory("Select directory of CSV files to merge");
    if (p) setInDir(p);
  }
  async function pickOutput() {
    const p = await SaveCSVFile("merged.csv");
    if (p) setOutput(p);
  }

  async function run() {
    if (!inDir) { setErr("Pick an input directory."); return; }
    if (!output) { setErr("Pick an output file."); return; }
    setLoading(true); setErr("");
    try {
      setResult(await MergeCSV({ inputDir: inDir, output, withHeader } as any));
    } catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-row">
        <button onClick={pickInDir}>📁 Input directory…</button>
        {inDir && <code className="path" title={inDir}>{inDir}</code>}
      </div>
      <div className="form-row">
        <button onClick={pickOutput}>💾 Output file…</button>
        {output && <code className="path" title={output}>{output}</code>}
      </div>
      <div className="form-row">
        <label>
          <input type="checkbox" checked={withHeader} onChange={(e) => setWithHeader(e.target.checked)} />
          Use header from first file
        </label>
      </div>
      <div className="form-row">
        <button onClick={run} disabled={loading}>{loading ? "Merging…" : "Run merge"}</button>
      </div>
      {err && <Error msg={err} />}
      {result && (
        <Success>
          Merged {result.filesProcessed} file(s) — {result.rowsWritten} row(s) → <code>{output}</code>.
        </Success>
      )}
    </div>
  );
}

// ---------- ToSQLite ------------------------------------------------------

function SQLiteTab({ path }: { path: string }) {
  const [dbPath, setDbPath] = useState("");
  const [table, setTable] = useState("");
  const [ifExists, setIfExists] = useState("replace");
  const [result, setResult] = useState<main.ToSQLitePayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  if (!path) return <NeedFile />;

  async function pickDB() {
    const p = await SaveDBFile("data.db");
    if (p) setDbPath(p);
  }

  async function run() {
    if (!dbPath) { setErr("Pick an output .db file first."); return; }
    setLoading(true); setErr("");
    try {
      setResult(await ToSQLiteCSV({
        input: path, dbPath, table, ifExists,
      } as any));
    } catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="panel">
      <div className="form-grid">
        <label>Table name (default: filename)
          <input type="text" value={table} onChange={(e) => setTable(e.target.value)} placeholder="e.g. users" />
        </label>
        <label>If table exists
          <select value={ifExists} onChange={(e) => setIfExists(e.target.value)}>
            <option value="replace">replace</option>
            <option value="append">append</option>
            <option value="skip">skip</option>
            <option value="fail">fail</option>
          </select>
        </label>
      </div>
      <div className="form-row">
        <button onClick={pickDB}>💾 Output .db…</button>
        {dbPath && <code className="path" title={dbPath}>{dbPath}</code>}
      </div>
      <div className="form-row">
        <button onClick={run} disabled={loading}>{loading ? "Importing…" : "Run import"}</button>
      </div>
      {err && <Error msg={err} />}
      {result && result.skipped && (
        <Success>Table <code>{result.table}</code> already exists — skipped.</Success>
      )}
      {result && !result.skipped && (
        <Success>
          Imported {result.rowsImported} row(s) into table <code>{result.table}</code> at <code>{dbPath}</code>.
        </Success>
      )}
    </div>
  );
}
