import { useState } from "react";
import {
  OpenCSVFile,
  PreviewCSV,
  StatsCSV,
  FilterCSV,
  SaveCSVFile,
} from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import "./App.css";

type Tab = "preview" | "stats" | "filter";

export default function App() {
  const [path, setPath] = useState<string>("");
  const [tab, setTab] = useState<Tab>("preview");

  async function pickFile() {
    const p = await OpenCSVFile();
    if (p) setPath(p);
  }

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

      {path && (
        <>
          <nav className="tabs">
            <TabButton id="preview" active={tab} onClick={setTab}>
              Preview
            </TabButton>
            <TabButton id="stats" active={tab} onClick={setTab}>
              Stats
            </TabButton>
            <TabButton id="filter" active={tab} onClick={setTab}>
              Filter
            </TabButton>
          </nav>

          <main className="tab-content">
            {tab === "preview" && <PreviewTab path={path} />}
            {tab === "stats" && <StatsTab path={path} />}
            {tab === "filter" && <FilterTab path={path} />}
          </main>
        </>
      )}

      {!path && (
        <div className="empty-state">
          Open a CSV file to get started.
        </div>
      )}
    </div>
  );
}

function TabButton({
  id,
  active,
  onClick,
  children,
}: {
  id: Tab;
  active: Tab;
  onClick: (id: Tab) => void;
  children: React.ReactNode;
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

// ---------- Preview tab ----------------------------------------------------

function PreviewTab({ path }: { path: string }) {
  const [rows, setRows] = useState(20);
  const [data, setData] = useState<main.PreviewPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function run() {
    setLoading(true);
    setErr("");
    try {
      setData(await PreviewCSV(path, rows, false));
    } catch (e: any) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="panel">
      <div className="form-row">
        <label>
          Rows:
          <input
            type="number"
            min={1}
            max={1000}
            value={rows}
            onChange={(e) => setRows(Number(e.target.value))}
          />
        </label>
        <button onClick={run} disabled={loading}>
          {loading ? "…" : "Preview"}
        </button>
      </div>
      {err && <Error msg={err} />}
      {data && (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                {data.headers.map((h, i) => (
                  <th key={i}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.rows.map((r, i) => (
                <tr key={i}>
                  {r.map((c, j) => (
                    <td key={j}>{c}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          <div className="meta">{data.rows.length} row(s)</div>
        </div>
      )}
    </div>
  );
}

// ---------- Stats tab ------------------------------------------------------

function StatsTab({ path }: { path: string }) {
  const [maxUnique, setMaxUnique] = useState(100000);
  const [data, setData] = useState<main.StatsPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function run() {
    setLoading(true);
    setErr("");
    try {
      setData(await StatsCSV(path, maxUnique));
    } catch (e: any) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="panel">
      <div className="form-row">
        <label>
          Max unique tracked:
          <input
            type="number"
            min={0}
            value={maxUnique}
            onChange={(e) => setMaxUnique(Number(e.target.value))}
          />
        </label>
        <button onClick={run} disabled={loading}>
          {loading ? "Analyzing…" : "Analyze"}
        </button>
      </div>
      {err && <Error msg={err} />}
      {data && (
        <div className="table-wrap">
          <div className="stats-summary">
            Total rows: <strong>{data.totalRows}</strong> · Columns:{" "}
            <strong>{data.columns.length}</strong>
          </div>
          <table>
            <thead>
              <tr>
                <th>Column</th>
                <th>Unique</th>
                <th>Empty</th>
                <th>Top 3</th>
              </tr>
            </thead>
            <tbody>
              {data.columns.map((c, i) => (
                <tr key={i}>
                  <td>{c.name}</td>
                  <td>
                    {c.uniqueCapped ? `≥${c.unique} (capped)` : c.unique}
                  </td>
                  <td>{c.empty}</td>
                  <td>
                    {c.top
                      .map((t) => `${t.value} (${t.count})`)
                      .join(", ")}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

// ---------- Filter tab -----------------------------------------------------

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

  async function pickOutput() {
    const p = await SaveCSVFile("filtered.csv");
    if (p) setOutput(p);
  }

  async function run() {
    if (!output) {
      setErr("Pick an output file first.");
      return;
    }
    setLoading(true);
    setErr("");
    try {
      const r = await FilterCSV({
        input: path,
        output,
        column,
        eq,
        eqSet,
        contains,
        containsSet,
        gt,
        gtSet,
        lt,
        ltSet,
        all,
        withHeader: true,
      } as any);
      setResult(r);
    } catch (e: any) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="panel">
      <div className="form-grid">
        <label>
          Column
          <input
            type="text"
            value={column}
            onChange={(e) => setColumn(e.target.value)}
            placeholder="e.g. country"
          />
        </label>
        <label className="condition">
          <input
            type="checkbox"
            checked={eqSet}
            onChange={(e) => setEqSet(e.target.checked)}
          />
          equals
          <input
            type="text"
            value={eq}
            disabled={!eqSet}
            onChange={(e) => setEq(e.target.value)}
          />
        </label>
        <label className="condition">
          <input
            type="checkbox"
            checked={containsSet}
            onChange={(e) => setContainsSet(e.target.checked)}
          />
          contains
          <input
            type="text"
            value={contains}
            disabled={!containsSet}
            onChange={(e) => setContains(e.target.value)}
          />
        </label>
        <label className="condition">
          <input
            type="checkbox"
            checked={gtSet}
            onChange={(e) => setGtSet(e.target.checked)}
          />
          &gt;
          <input
            type="number"
            value={gt}
            disabled={!gtSet}
            onChange={(e) => setGt(Number(e.target.value))}
          />
        </label>
        <label className="condition">
          <input
            type="checkbox"
            checked={ltSet}
            onChange={(e) => setLtSet(e.target.checked)}
          />
          &lt;
          <input
            type="number"
            value={lt}
            disabled={!ltSet}
            onChange={(e) => setLt(Number(e.target.value))}
          />
        </label>
        <label>
          <input
            type="checkbox"
            checked={all}
            onChange={(e) => setAll(e.target.checked)}
          />
          Match ALL conditions (AND)
        </label>
      </div>
      <div className="form-row">
        <button onClick={pickOutput}>💾 Output file…</button>
        {output && (
          <code className="path" title={output}>
            {output}
          </code>
        )}
      </div>
      <div className="form-row">
        <button onClick={run} disabled={loading || !column}>
          {loading ? "Filtering…" : "Run filter"}
        </button>
      </div>
      {err && <Error msg={err} />}
      {result && (
        <div className="success">
          ✅ Matched <strong>{result.matched}</strong> of{" "}
          <strong>{result.totalRows}</strong> rows. Written to{" "}
          <code>{output}</code>.
        </div>
      )}
    </div>
  );
}

function Error({ msg }: { msg: string }) {
  return <div className="error">⚠️ {msg}</div>;
}
