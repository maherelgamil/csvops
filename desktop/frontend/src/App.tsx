import { useEffect, useMemo, useState } from "react";
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
  FileInfoCSV,
  RevealFile,
} from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Progress } from "@/components/ui/progress";
import {
  FileText,
  Folder,
  Save,
  Eye,
  BarChart3,
  Filter as FilterIcon,
  Copy,
  Scissors,
  Combine,
  Database,
  CheckCircle2,
  AlertCircle,
  Loader2,
  FolderOpen,
  ExternalLink,
  Hash,
  HardDrive,
  Upload,
} from "lucide-react";

// ---------- helpers --------------------------------------------------------

type ProgressEvent = { op: string; done: number; total: number };

function formatBytes(b: number): string {
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
  if (b < 1024 * 1024 * 1024) return `${(b / 1024 / 1024).toFixed(1)} MB`;
  return `${(b / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

/** Suggest a sibling output path by inserting `.suffix` before the extension. */
function suggestOutput(input: string, suffix: string, newExt?: string): string {
  if (!input) return "";
  const sep = input.includes("\\") ? "\\" : "/";
  const i = input.lastIndexOf(sep);
  const dir = i >= 0 ? input.slice(0, i + 1) : "";
  const name = i >= 0 ? input.slice(i + 1) : input;
  const dot = name.lastIndexOf(".");
  const stem = dot > 0 ? name.slice(0, dot) : name;
  const ext = newExt ?? (dot > 0 ? name.slice(dot) : ".csv");
  return `${dir}${stem}.${suffix}${ext}`;
}

// ---------- root -----------------------------------------------------------

export default function App() {
  const [info, setInfo] = useState<main.FileInfo | null>(null);
  const [loadingFile, setLoadingFile] = useState(false);
  const [progress, setProgress] = useState<ProgressEvent | null>(null);
  const [dragging, setDragging] = useState(false);

  useEffect(() => {
    EventsOn("progress", (p: ProgressEvent) => setProgress(p));
    EventsOn("file-dropped", (path: string) => loadFile(path));
    return () => {
      EventsOff("progress");
      EventsOff("file-dropped");
    };
  }, []);

  // Native HTML drag/drop also works because Wails enables file drop.
  useEffect(() => {
    const onDragOver = (e: DragEvent) => {
      e.preventDefault();
      setDragging(true);
    };
    const onDragLeave = (e: DragEvent) => {
      if (e.target === document.documentElement) setDragging(false);
    };
    const onDrop = () => setDragging(false);
    window.addEventListener("dragover", onDragOver);
    window.addEventListener("dragleave", onDragLeave);
    window.addEventListener("drop", onDrop);
    return () => {
      window.removeEventListener("dragover", onDragOver);
      window.removeEventListener("dragleave", onDragLeave);
      window.removeEventListener("drop", onDrop);
    };
  }, []);

  async function loadFile(path: string) {
    if (!path) return;
    setLoadingFile(true);
    try {
      const i = await FileInfoCSV(path);
      setInfo(i);
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingFile(false);
    }
  }

  async function pickFile() {
    const p = await OpenCSVFile();
    if (p) await loadFile(p);
  }

  return (
    <div className="relative flex h-screen flex-col bg-[hsl(210_40%_98%)]">
      <Header info={info} onOpen={pickFile} loading={loadingFile} />
      {info && <FileInfoBar info={info} />}

      <div className="flex-1 overflow-auto px-6 py-5">
        {!info ? (
          <EmptyState onOpen={pickFile} />
        ) : (
          <Tabs defaultValue="preview" className="w-full">
            <TabsList className="mb-4">
              <TabsTrigger value="preview"><Eye className="h-3.5 w-3.5" />Preview</TabsTrigger>
              <TabsTrigger value="stats"><BarChart3 className="h-3.5 w-3.5" />Stats</TabsTrigger>
              <TabsTrigger value="filter"><FilterIcon className="h-3.5 w-3.5" />Filter</TabsTrigger>
              <TabsTrigger value="dedupe"><Copy className="h-3.5 w-3.5" />Dedupe</TabsTrigger>
              <TabsTrigger value="split"><Scissors className="h-3.5 w-3.5" />Split</TabsTrigger>
              <TabsTrigger value="merge"><Combine className="h-3.5 w-3.5" />Merge</TabsTrigger>
              <TabsTrigger value="sqlite"><Database className="h-3.5 w-3.5" />SQLite</TabsTrigger>
            </TabsList>

            <TabsContent value="preview"><PreviewTab info={info} /></TabsContent>
            <TabsContent value="stats"><StatsTab path={info.path} /></TabsContent>
            <TabsContent value="filter"><FilterTab info={info} /></TabsContent>
            <TabsContent value="dedupe"><DedupeTab info={info} /></TabsContent>
            <TabsContent value="split"><SplitTab path={info.path} /></TabsContent>
            <TabsContent value="merge"><MergeTab /></TabsContent>
            <TabsContent value="sqlite"><SQLiteTab info={info} /></TabsContent>
          </Tabs>
        )}
      </div>

      {progress && progress.total > 0 && progress.done < progress.total && (
        <ProgressBar p={progress} />
      )}

      {dragging && <DropOverlay />}
    </div>
  );
}

// ---------- Chrome --------------------------------------------------------

function Header({ info, onOpen, loading }: { info: main.FileInfo | null; onOpen: () => void; loading: boolean }) {
  return (
    <header className="flex items-center justify-between border-b border-border bg-card px-6 py-3">
      <div className="flex items-center gap-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <FileText className="h-4 w-4" />
        </div>
        <div>
          <h1 className="text-sm font-semibold leading-none">csvops</h1>
          <p className="mt-0.5 text-xs text-muted-foreground">CSV operations toolkit</p>
        </div>
      </div>
      <Button onClick={onOpen} size="sm" disabled={loading}>
        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <FolderOpen className="h-4 w-4" />}
        {info ? "Change file…" : "Open CSV…"}
      </Button>
    </header>
  );
}

function FileInfoBar({ info }: { info: main.FileInfo }) {
  return (
    <div className="flex items-center gap-5 border-b border-border bg-secondary/40 px-6 py-2 text-xs">
      <code
        className="max-w-[420px] truncate font-mono text-muted-foreground"
        title={info.path}
      >
        {info.path}
      </code>
      <span className="flex items-center gap-1.5 text-muted-foreground">
        <Hash className="h-3 w-3" />
        <strong className="font-semibold text-foreground">
          {info.rows.toLocaleString()}
        </strong>{" "}
        rows
      </span>
      <span className="flex items-center gap-1.5 text-muted-foreground">
        <HardDrive className="h-3 w-3" />
        <strong className="font-semibold text-foreground">
          {formatBytes(info.size)}
        </strong>
      </span>
      <span className="flex items-center gap-1.5 text-muted-foreground">
        <BarChart3 className="h-3 w-3" />
        <strong className="font-semibold text-foreground">
          {(info.headers || []).length}
        </strong>{" "}
        columns
      </span>
      <button
        onClick={() => RevealFile(info.path)}
        className="ml-auto flex items-center gap-1 text-muted-foreground transition-colors hover:text-foreground"
      >
        <ExternalLink className="h-3 w-3" />
        Reveal
      </button>
    </div>
  );
}

function EmptyState({ onOpen }: { onOpen: () => void }) {
  return (
    <div className="flex h-full items-center justify-center">
      <Card className="w-[460px]">
        <CardHeader className="items-center text-center">
          <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-secondary">
            <Upload className="h-6 w-6 text-muted-foreground" />
          </div>
          <CardTitle>Drop a CSV here</CardTitle>
          <CardDescription>
            Or pick one to preview, analyze, and transform — all locally.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center">
          <Button onClick={onOpen}>
            <FolderOpen className="h-4 w-4" />
            Choose CSV file…
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

function DropOverlay() {
  return (
    <div className="pointer-events-none absolute inset-0 z-50 flex items-center justify-center bg-primary/10 backdrop-blur-sm">
      <div className="rounded-2xl border-4 border-dashed border-primary bg-card px-12 py-10 text-center shadow-lg">
        <Upload className="mx-auto mb-3 h-10 w-10 text-primary" />
        <div className="text-lg font-semibold text-foreground">Drop your CSV</div>
        <div className="text-sm text-muted-foreground">to open it in csvops</div>
      </div>
    </div>
  );
}

function ProgressBar({ p }: { p: ProgressEvent }) {
  const pct = p.total > 0 ? Math.round((p.done / p.total) * 100) : 0;
  return (
    <div className="border-t border-border bg-card px-6 py-2.5">
      <div className="mb-1.5 flex items-center justify-between text-xs">
        <span className="font-medium capitalize text-foreground">{p.op}</span>
        <span className="text-muted-foreground">
          {p.done.toLocaleString()} / {p.total.toLocaleString()} ({pct}%)
        </span>
      </div>
      <Progress value={pct} />
    </div>
  );
}

// ---------- Common UI ----------------------------------------------------

function FormField({ label, children, hint }: {
  label: string; children: React.ReactNode; hint?: string;
}) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function FilePicker({ label, value, onPick, icon: Icon }: {
  label: string; value: string; onPick: () => void; icon?: any;
}) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <div className="flex items-center gap-2">
        <Button variant="outline" onClick={onPick}>
          {Icon ? <Icon className="h-4 w-4" /> : <Folder className="h-4 w-4" />}
          Choose…
        </Button>
        {value ? (
          <code className="flex-1 truncate rounded-md bg-secondary px-3 py-1.5 font-mono text-xs text-muted-foreground" title={value}>
            {value}
          </code>
        ) : (
          <span className="text-xs italic text-muted-foreground">No file selected</span>
        )}
      </div>
    </div>
  );
}

function ResultBanner({ kind, children, output }: {
  kind: "success" | "error";
  children: React.ReactNode;
  output?: string;
}) {
  const Icon = kind === "success" ? CheckCircle2 : AlertCircle;
  const bg =
    kind === "success"
      ? "bg-success/10 border-success/30"
      : "bg-destructive/10 border-destructive/30";
  return (
    <div className={`flex items-start gap-3 rounded-lg border p-3 text-sm ${bg}`}>
      <Icon className={`mt-0.5 h-4 w-4 shrink-0 ${kind === "success" ? "text-success" : "text-destructive"}`} />
      <div className="flex-1 text-foreground">{children}</div>
      {output && (
        <Button size="sm" variant="outline" onClick={() => RevealFile(output)}>
          <ExternalLink className="h-3.5 w-3.5" />
          Reveal
        </Button>
      )}
    </div>
  );
}

function RunButton({ onClick, loading, disabled, children }: {
  onClick: () => void; loading: boolean; disabled?: boolean; children: React.ReactNode;
}) {
  return (
    <Button onClick={onClick} disabled={loading || disabled}>
      {loading && <Loader2 className="h-4 w-4 animate-spin" />}
      {children}
    </Button>
  );
}

function OpCard({ title, description, children }: {
  title: string; description: string; children: React.ReactNode;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">{children}</CardContent>
    </Card>
  );
}

// ---------- Preview --------------------------------------------------------

function PreviewTab({ info }: { info: main.FileInfo }) {
  const [rows, setRows] = useState(20);
  const [data, setData] = useState<main.PreviewPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function run() {
    setLoading(true); setErr("");
    try { setData(await PreviewCSV(info.path, rows, false)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  useEffect(() => { run(); /* eslint-disable-next-line */ }, [info.path]);

  return (
    <OpCard title="Preview" description="Pretty-print the first N rows of the file.">
      <div className="flex items-end gap-3">
        <div className="w-32">
          <FormField label="Rows">
            <Input type="number" min={1} max={1000} value={rows} onChange={(e) => setRows(Number(e.target.value))} />
          </FormField>
        </div>
        <RunButton onClick={run} loading={loading}>Preview</RunButton>
      </div>
      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {data && data.rows.length > 0 && (
        <DataTable headers={data.headers} rows={data.rows} caption={`${data.rows.length} row(s)`} />
      )}
    </OpCard>
  );
}

// ---------- Stats ---------------------------------------------------------

function StatsTab({ path }: { path: string }) {
  const [maxUnique, setMaxUnique] = useState(100000);
  const [data, setData] = useState<main.StatsPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function run() {
    setLoading(true); setErr("");
    try { setData(await StatsCSV(path, maxUnique)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <OpCard title="Stats" description="Per-column unique values, empty cells, and top values.">
      <div className="flex items-end gap-3">
        <div className="w-48">
          <FormField label="Max unique tracked" hint="Caps memory on high-cardinality columns.">
            <Input type="number" min={0} value={maxUnique} onChange={(e) => setMaxUnique(Number(e.target.value))} />
          </FormField>
        </div>
        <RunButton onClick={run} loading={loading}>Analyze</RunButton>
      </div>
      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {data && (
        <>
          <div className="flex gap-6 text-sm text-muted-foreground">
            <span>Total rows: <strong className="text-foreground">{data.totalRows.toLocaleString()}</strong></span>
            <span>Columns: <strong className="text-foreground">{data.columns.length}</strong></span>
          </div>
          <DataTable
            headers={["Column", "Unique", "Empty", "Top 3"]}
            rows={data.columns.map((c) => [
              c.name,
              c.uniqueCapped ? `≥${c.unique} (capped)` : String(c.unique),
              String(c.empty),
              c.top.map((t) => `${t.value} (${t.count})`).join(", "),
            ])}
          />
        </>
      )}
    </OpCard>
  );
}

// ---------- Filter --------------------------------------------------------

function FilterTab({ info }: { info: main.FileInfo }) {
  const headers = info.headers || [];
  const [column, setColumn] = useState(headers[0] || "");
  useEffect(() => { setColumn(headers[0] || ""); }, [info.path]);

  const [eqSet, setEqSet] = useState(false);
  const [eq, setEq] = useState("");
  const [containsSet, setContainsSet] = useState(false);
  const [contains, setContains] = useState("");
  const [gtSet, setGtSet] = useState(false);
  const [gt, setGt] = useState(0);
  const [ltSet, setLtSet] = useState(false);
  const [lt, setLt] = useState(0);
  const [all, setAll] = useState(false);
  const [output, setOutput] = useState(suggestOutput(info.path, "filtered"));
  useEffect(() => { setOutput(suggestOutput(info.path, "filtered")); }, [info.path]);

  const [result, setResult] = useState<main.FilterPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function pickOutput() {
    const p = await SaveCSVFile(output || "filtered.csv");
    if (p) setOutput(p);
  }
  async function run() {
    if (!output) { setErr("Choose an output file first."); return; }
    setLoading(true); setErr(""); setResult(null);
    try {
      setResult(await FilterCSV({
        input: info.path, output, column,
        eq, eqSet, contains, containsSet, gt, gtSet, lt, ltSet,
        all, withHeader: true,
      } as any));
    } catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <OpCard title="Filter" description="Keep only rows matching one or more conditions.">
      <FormField label="Column">
        <Select value={column} onValueChange={setColumn}>
          <SelectTrigger><SelectValue placeholder="Pick a column…" /></SelectTrigger>
          <SelectContent>
            {headers.map((h) => <SelectItem key={h} value={h}>{h}</SelectItem>)}
          </SelectContent>
        </Select>
      </FormField>

      <div className="space-y-2">
        <Label>Conditions</Label>
        <ConditionRow checked={eqSet} onCheck={setEqSet} label="equals">
          <Input value={eq} disabled={!eqSet} onChange={(e) => setEq(e.target.value)} placeholder="value" />
        </ConditionRow>
        <ConditionRow checked={containsSet} onCheck={setContainsSet} label="contains">
          <Input value={contains} disabled={!containsSet} onChange={(e) => setContains(e.target.value)} placeholder="substring" />
        </ConditionRow>
        <ConditionRow checked={gtSet} onCheck={setGtSet} label="greater than">
          <Input type="number" value={gt} disabled={!gtSet} onChange={(e) => setGt(Number(e.target.value))} />
        </ConditionRow>
        <ConditionRow checked={ltSet} onCheck={setLtSet} label="less than">
          <Input type="number" value={lt} disabled={!ltSet} onChange={(e) => setLt(Number(e.target.value))} />
        </ConditionRow>

        <label className="flex cursor-pointer items-center gap-2 pt-2 text-sm">
          <Checkbox checked={all} onCheckedChange={(v) => setAll(!!v)} />
          Match <strong>ALL</strong> conditions (AND), not ANY (OR)
        </label>
      </div>

      <FilePicker label="Output file" value={output} onPick={pickOutput} icon={Save} />

      <RunButton onClick={run} loading={loading} disabled={!column}>Run filter</RunButton>

      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {result && (
        <ResultBanner kind="success" output={output}>
          Matched <strong>{result.matched.toLocaleString()}</strong> of{" "}
          <strong>{result.totalRows.toLocaleString()}</strong> rows.
        </ResultBanner>
      )}
    </OpCard>
  );
}

function ConditionRow({ checked, onCheck, label, children }: {
  checked: boolean; onCheck: (v: boolean) => void; label: string; children: React.ReactNode;
}) {
  return (
    <div className="flex items-center gap-3 rounded-md border border-border bg-card p-2.5">
      <Checkbox checked={checked} onCheckedChange={(v) => onCheck(!!v)} />
      <span className="w-28 text-sm text-foreground">{label}</span>
      <div className="flex-1">{children}</div>
    </div>
  );
}

// ---------- Dedupe --------------------------------------------------------

function DedupeTab({ info }: { info: main.FileInfo }) {
  const headers = info.headers || [];
  const [picked, setPicked] = useState<string[]>([]);
  useEffect(() => { setPicked([]); }, [info.path]);
  const keysCSV = useMemo(() => picked.join(","), [picked]);

  const [keepLast, setKeepLast] = useState(false);
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [output, setOutput] = useState(suggestOutput(info.path, "dedup"));
  useEffect(() => { setOutput(suggestOutput(info.path, "dedup")); }, [info.path]);

  const [result, setResult] = useState<main.DedupePayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  function toggle(h: string) {
    setPicked((p) => p.includes(h) ? p.filter((x) => x !== h) : [...p, h]);
  }

  async function pickOutput() { const p = await SaveCSVFile(output || "dedup.csv"); if (p) setOutput(p); }
  async function run() {
    if (!output) { setErr("Choose an output file first."); return; }
    if (picked.length === 0) { setErr("Pick at least one key column."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await DedupeCSV({ input: info.path, output, keyColumns: keysCSV, keepLast, caseSensitive } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <OpCard title="Dedupe" description="Remove duplicate rows by one or more key columns. Output preserves original file order.">
      <div className="space-y-1.5">
        <Label>Key columns ({picked.length} selected)</Label>
        <div className="flex flex-wrap gap-2">
          {headers.map((h) => {
            const on = picked.includes(h);
            return (
              <button
                key={h}
                onClick={() => toggle(h)}
                className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
                  on
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-border bg-card text-foreground hover:bg-secondary"
                }`}
              >
                {h}
              </button>
            );
          })}
        </div>
      </div>

      <div className="flex flex-wrap gap-5">
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={keepLast} onCheckedChange={(v) => setKeepLast(!!v)} />
          Keep last occurrence (default keeps first)
        </label>
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={caseSensitive} onCheckedChange={(v) => setCaseSensitive(!!v)} />
          Case-sensitive
        </label>
      </div>

      <FilePicker label="Output file" value={output} onPick={pickOutput} icon={Save} />
      <RunButton onClick={run} loading={loading}>Run dedupe</RunButton>

      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {result && (
        <ResultBanner kind="success" output={output}>
          Kept <strong>{result.uniqueRows.toLocaleString()}</strong> unique row(s),
          removed <strong>{result.duplicates.toLocaleString()}</strong> duplicates from{" "}
          <strong>{result.totalRows.toLocaleString()}</strong> total.
        </ResultBanner>
      )}
    </OpCard>
  );
}

// ---------- Split --------------------------------------------------------

function SplitTab({ path }: { path: string }) {
  const [outDir, setOutDir] = useState("");
  const [rowsPerFile, setRowsPerFile] = useState(1000);
  const [withHeader, setWithHeader] = useState(true);
  const [result, setResult] = useState<main.SplitPayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function pickOutDir() { const p = await OpenDirectory("Select output directory"); if (p) setOutDir(p); }
  async function run() {
    if (!outDir) { setErr("Choose an output directory first."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await SplitCSV({ input: path, outputDir: outDir, rowsPerFile, withHeader } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <OpCard title="Split" description="Break a large CSV into chunks of N rows each.">
      <div className="grid grid-cols-2 gap-4">
        <FormField label="Rows per output file">
          <Input type="number" min={1} value={rowsPerFile} onChange={(e) => setRowsPerFile(Number(e.target.value))} />
        </FormField>
        <div className="flex items-end">
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <Checkbox checked={withHeader} onCheckedChange={(v) => setWithHeader(!!v)} />
            Include header in each chunk
          </label>
        </div>
      </div>
      <FilePicker label="Output directory" value={outDir} onPick={pickOutDir} icon={Folder} />
      <RunButton onClick={run} loading={loading}>Run split</RunButton>
      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {result && (
        <ResultBanner kind="success" output={outDir}>
          Wrote <strong>{result.filesCreated}</strong> file(s),{" "}
          <strong>{result.rowsProcessed.toLocaleString()}</strong> row(s) total.
        </ResultBanner>
      )}
    </OpCard>
  );
}

// ---------- Merge --------------------------------------------------------

function MergeTab() {
  const [inDir, setInDir] = useState("");
  const [output, setOutput] = useState("");
  const [withHeader, setWithHeader] = useState(true);
  const [result, setResult] = useState<main.MergePayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function pickInDir() {
    const p = await OpenDirectory("Select directory of CSV files");
    if (p) {
      setInDir(p);
      if (!output) setOutput(suggestOutput(p + "/merged", "csv", ""));
    }
  }
  async function pickOutput() { const p = await SaveCSVFile(output || "merged.csv"); if (p) setOutput(p); }
  async function run() {
    if (!inDir) { setErr("Choose an input directory."); return; }
    if (!output) { setErr("Choose an output file."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await MergeCSV({ inputDir: inDir, output, withHeader } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <OpCard title="Merge" description="Combine all CSV files in a directory into one output file.">
      <FilePicker label="Input directory" value={inDir} onPick={pickInDir} icon={Folder} />
      <FilePicker label="Output file" value={output} onPick={pickOutput} icon={Save} />
      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <Checkbox checked={withHeader} onCheckedChange={(v) => setWithHeader(!!v)} />
        Use header from the first file
      </label>
      <RunButton onClick={run} loading={loading}>Run merge</RunButton>
      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {result && (
        <ResultBanner kind="success" output={output}>
          Merged <strong>{result.filesProcessed}</strong> file(s),{" "}
          <strong>{result.rowsWritten.toLocaleString()}</strong> row(s).
        </ResultBanner>
      )}
    </OpCard>
  );
}

// ---------- ToSQLite ----------------------------------------------------

function SQLiteTab({ info }: { info: main.FileInfo }) {
  const [dbPath, setDbPath] = useState(suggestOutput(info.path, "data", ".db"));
  useEffect(() => { setDbPath(suggestOutput(info.path, "data", ".db")); }, [info.path]);

  const [table, setTable] = useState("");
  const [ifExists, setIfExists] = useState("replace");
  const [result, setResult] = useState<main.ToSQLitePayload | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function pickDB() { const p = await SaveDBFile(dbPath || "data.db"); if (p) setDbPath(p); }
  async function run() {
    if (!dbPath) { setErr("Choose an output .db file first."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await ToSQLiteCSV({ input: info.path, dbPath, table, ifExists } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <OpCard title="To SQLite" description="Import the CSV into a SQLite database table (all columns as TEXT).">
      <div className="grid grid-cols-2 gap-4">
        <FormField label="Table name" hint="Defaults to the input filename.">
          <Input value={table} onChange={(e) => setTable(e.target.value)} placeholder="users" />
        </FormField>
        <FormField label="If table exists">
          <Select value={ifExists} onValueChange={setIfExists}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="replace">replace — drop and re-create</SelectItem>
              <SelectItem value="append">append — add rows to existing</SelectItem>
              <SelectItem value="skip">skip — do nothing if exists</SelectItem>
              <SelectItem value="fail">fail — error if exists</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
      </div>
      <FilePicker label="Output database (.db)" value={dbPath} onPick={pickDB} icon={Database} />
      <RunButton onClick={run} loading={loading}>Run import</RunButton>
      {err && <ResultBanner kind="error">{err}</ResultBanner>}
      {result?.skipped && (
        <ResultBanner kind="success" output={dbPath}>
          Table <code>{result.table}</code> already exists — skipped.
        </ResultBanner>
      )}
      {result && !result.skipped && (
        <ResultBanner kind="success" output={dbPath}>
          Imported <strong>{result.rowsImported.toLocaleString()}</strong> row(s) into table{" "}
          <code>{result.table}</code>.
        </ResultBanner>
      )}
    </OpCard>
  );
}

// ---------- Data table ---------------------------------------------------

function DataTable({ headers, rows, caption }: {
  headers: string[]; rows: string[][]; caption?: string;
}) {
  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <div className="max-h-[420px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="sticky top-0 z-10 bg-secondary">
            <tr>
              {headers.map((h, i) => (
                <th key={i} className="border-b border-border px-3 py-2 text-left font-semibold text-foreground">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} className="border-b border-border last:border-0 hover:bg-secondary/60">
                {r.map((c, j) => (
                  <td key={j} className="px-3 py-2 font-mono text-xs text-foreground" title={c}>{c}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {caption && (
        <div className="border-t border-border bg-secondary/50 px-3 py-1.5 text-xs text-muted-foreground">
          {caption}
        </div>
      )}
    </div>
  );
}
