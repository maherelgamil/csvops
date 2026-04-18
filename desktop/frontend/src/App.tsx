import { useEffect, useMemo, useState } from "react";
import {
  OpenCSVFile,
  OpenDirectory,
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
  ReadPage,
} from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Progress } from "@/components/ui/progress";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  FileText,
  Folder,
  Save,
  Filter as FilterIcon,
  Copy as CopyIcon,
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
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  MoreHorizontal,
  BarChart3,
  Wrench,
} from "lucide-react";

// ---------- helpers --------------------------------------------------------

type ProgressEvent = { op: string; done: number; total: number };

function formatBytes(b: number) {
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
  if (b < 1024 ** 3) return `${(b / 1024 / 1024).toFixed(1)} MB`;
  return `${(b / 1024 ** 3).toFixed(2)} GB`;
}

function suggestOutput(input: string, suffix: string, ext?: string) {
  if (!input) return "";
  const sep = input.includes("\\") ? "\\" : "/";
  const i = input.lastIndexOf(sep);
  const dir = i >= 0 ? input.slice(0, i + 1) : "";
  const name = i >= 0 ? input.slice(i + 1) : input;
  const dot = name.lastIndexOf(".");
  const stem = dot > 0 ? name.slice(0, dot) : name;
  const finalExt = ext ?? (dot > 0 ? name.slice(dot) : ".csv");
  return `${dir}${stem}.${suffix}${finalExt}`;
}

const PAGE_SIZE = 100;

// ---------- root -----------------------------------------------------------

type ActionKind = "filter" | "dedupe" | "split" | "sqlite" | "merge" | null;

export default function App() {
  const [info, setInfo] = useState<main.FileInfo | null>(null);
  const [page, setPage] = useState<main.PagePayload | null>(null);
  const [offset, setOffset] = useState(0);
  const [loadingFile, setLoadingFile] = useState(false);
  const [loadingPage, setLoadingPage] = useState(false);
  const [stats, setStats] = useState<main.StatsPayload | null>(null);
  const [progress, setProgress] = useState<ProgressEvent | null>(null);
  const [dragging, setDragging] = useState(false);
  const [action, setAction] = useState<ActionKind>(null);

  useEffect(() => {
    EventsOn("progress", (p: ProgressEvent) => setProgress(p));
    EventsOn("file-dropped", (path: string) => loadFile(path));
    return () => {
      EventsOff("progress");
      EventsOff("file-dropped");
    };
  }, []);

  useEffect(() => {
    const onDragOver = (e: DragEvent) => { e.preventDefault(); setDragging(true); };
    const onDragLeave = (e: DragEvent) => { if (e.target === document.documentElement) setDragging(false); };
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
    setStats(null);
    try {
      const i = await FileInfoCSV(path);
      setInfo(i);
      setOffset(0);
      const p = await ReadPage(path, 0, PAGE_SIZE);
      setPage(p);
    } finally {
      setLoadingFile(false);
    }
  }

  async function pickFile() {
    const p = await OpenCSVFile();
    if (p) await loadFile(p);
  }

  async function gotoOffset(newOffset: number) {
    if (!info) return;
    setLoadingPage(true);
    try {
      const p = await ReadPage(info.path, Math.max(0, newOffset), PAGE_SIZE);
      setPage(p);
      setOffset(p.offset);
    } finally {
      setLoadingPage(false);
    }
  }

  async function loadStatsIfNeeded(): Promise<main.StatsPayload | null> {
    if (stats) return stats;
    if (!info) return null;
    const s = await StatsCSV(info.path, 100000);
    setStats(s);
    return s;
  }

  return (
    <div className="relative flex h-screen flex-col bg-[hsl(210_40%_98%)]">
      <Header info={info} onOpen={pickFile} loading={loadingFile} />

      {!info ? (
        <div className="flex-1 overflow-auto p-6">
          <EmptyState onOpen={pickFile} />
        </div>
      ) : (
        <>
          <FileInfoBar info={info} />
          <Toolbar
            info={info}
            offset={offset}
            page={page}
            onPrev={() => gotoOffset(offset - PAGE_SIZE)}
            onNext={() => gotoOffset(offset + PAGE_SIZE)}
            onJump={(o) => gotoOffset(o)}
            onAction={setAction}
          />
          <div className="flex-1 overflow-auto px-4 pb-4">
            <DataTable
              page={page}
              loading={loadingPage}
              offset={offset}
              loadStats={loadStatsIfNeeded}
              stats={stats}
            />
          </div>
        </>
      )}

      {progress && progress.total > 0 && progress.done < progress.total && (
        <ProgressBar p={progress} />
      )}

      {dragging && <DropOverlay />}

      {info && (
        <Sheet open={!!action} onOpenChange={(o) => !o && setAction(null)}>
          <SheetContent
            title={actionTitle(action)}
            description={actionDescription(action)}
          >
            {action === "filter" && <FilterAction info={info} onDone={(out) => loadFile(out)} />}
            {action === "dedupe" && <DedupeAction info={info} onDone={(out) => loadFile(out)} />}
            {action === "split" && <SplitAction info={info} />}
            {action === "sqlite" && <SQLiteAction info={info} />}
            {action === "merge" && <MergeAction />}
          </SheetContent>
        </Sheet>
      )}
    </div>
  );
}

function actionTitle(a: ActionKind) {
  switch (a) {
    case "filter": return "Filter";
    case "dedupe": return "Dedupe";
    case "split": return "Split";
    case "sqlite": return "Export to SQLite";
    case "merge": return "Merge directory of CSVs";
    default: return "";
  }
}
function actionDescription(a: ActionKind) {
  switch (a) {
    case "filter": return "Keep only rows matching one or more conditions.";
    case "dedupe": return "Remove duplicate rows by one or more key columns.";
    case "split": return "Break the file into chunks of N rows each.";
    case "sqlite": return "Import the file into a SQLite table.";
    case "merge": return "Combine all CSVs in a directory into one file.";
    default: return "";
  }
}

// ---------- chrome --------------------------------------------------------

function Header({ info, onOpen, loading }: { info: main.FileInfo | null; onOpen: () => void; loading: boolean }) {
  return (
    <header className="flex shrink-0 items-center justify-between border-b border-border bg-card px-6 py-3">
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
    <div className="flex shrink-0 items-center gap-5 border-b border-border bg-secondary/40 px-6 py-2 text-xs">
      <code className="max-w-[420px] truncate font-mono text-muted-foreground" title={info.path}>{info.path}</code>
      <span className="flex items-center gap-1.5 text-muted-foreground">
        <Hash className="h-3 w-3" />
        <strong className="font-semibold text-foreground">{info.rows.toLocaleString()}</strong> rows
      </span>
      <span className="flex items-center gap-1.5 text-muted-foreground">
        <BarChart3 className="h-3 w-3" />
        <strong className="font-semibold text-foreground">{(info.headers || []).length}</strong> columns
      </span>
      <span className="flex items-center gap-1.5 text-muted-foreground">
        <HardDrive className="h-3 w-3" />
        <strong className="font-semibold text-foreground">{formatBytes(info.size)}</strong>
      </span>
      <button onClick={() => RevealFile(info.path)} className="ml-auto flex items-center gap-1 text-muted-foreground transition-colors hover:text-foreground">
        <ExternalLink className="h-3 w-3" />Reveal
      </button>
    </div>
  );
}

function Toolbar({ info, offset, page, onPrev, onNext, onJump, onAction }: {
  info: main.FileInfo;
  offset: number;
  page: main.PagePayload | null;
  onPrev: () => void;
  onNext: () => void;
  onJump: (o: number) => void;
  onAction: (a: ActionKind) => void;
}) {
  const total = page?.totalRows ?? info.rows;
  const showingFrom = total === 0 ? 0 : offset + 1;
  const showingTo = Math.min(offset + PAGE_SIZE, total);
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const canPrev = offset > 0;
  const canNext = offset + PAGE_SIZE < total;

  const [pageInput, setPageInput] = useState(String(currentPage));
  useEffect(() => setPageInput(String(currentPage)), [currentPage]);

  function jumpFromInput() {
    const n = Math.max(1, Math.min(totalPages, Number(pageInput) || 1));
    onJump((n - 1) * PAGE_SIZE);
  }

  return (
    <div className="flex shrink-0 items-center gap-2 border-b border-border bg-card px-4 py-2">
      <Button variant="outline" size="sm" onClick={onPrev} disabled={!canPrev}>
        <ChevronLeft className="h-3.5 w-3.5" />
      </Button>
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        Page
        <Input
          className="h-7 w-14 text-center text-xs"
          value={pageInput}
          onChange={(e) => setPageInput(e.target.value)}
          onBlur={jumpFromInput}
          onKeyDown={(e) => e.key === "Enter" && jumpFromInput()}
        />
        of <strong className="text-foreground">{totalPages.toLocaleString()}</strong>
      </div>
      <Button variant="outline" size="sm" onClick={onNext} disabled={!canNext}>
        <ChevronRight className="h-3.5 w-3.5" />
      </Button>
      <span className="ml-3 text-xs text-muted-foreground">
        Rows <strong className="text-foreground">{showingFrom.toLocaleString()}</strong>–
        <strong className="text-foreground">{showingTo.toLocaleString()}</strong>{" "}
        of <strong className="text-foreground">{total.toLocaleString()}</strong>
      </span>

      <div className="ml-auto flex items-center gap-2">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="sm">
              <Wrench className="h-3.5 w-3.5" />
              Actions
              <ChevronDown className="h-3.5 w-3.5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onSelect={() => onAction("filter")}>
              <FilterIcon className="h-4 w-4" />Filter rows…
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => onAction("dedupe")}>
              <CopyIcon className="h-4 w-4" />Dedupe…
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => onAction("split")}>
              <Scissors className="h-4 w-4" />Split into chunks…
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => onAction("sqlite")}>
              <Database className="h-4 w-4" />Export to SQLite…
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => onAction("merge")}>
              <Combine className="h-4 w-4" />Merge directory…
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}

function EmptyState({ onOpen }: { onOpen: () => void }) {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="w-[460px] rounded-lg border border-border bg-card p-8 text-center shadow-sm">
        <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-secondary">
          <Upload className="h-6 w-6 text-muted-foreground" />
        </div>
        <h2 className="text-base font-semibold">Drop a CSV here</h2>
        <p className="mt-1 text-sm text-muted-foreground">Or pick one — preview, analyze, and transform locally.</p>
        <div className="mt-5">
          <Button onClick={onOpen}>
            <FolderOpen className="h-4 w-4" />
            Choose CSV file…
          </Button>
        </div>
      </div>
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

// ---------- table --------------------------------------------------------

function DataTable({ page, loading, offset, loadStats, stats }: {
  page: main.PagePayload | null;
  loading: boolean;
  offset: number;
  loadStats: () => Promise<main.StatsPayload | null>;
  stats: main.StatsPayload | null;
}) {
  if (!page) return <div className="p-8 text-sm text-muted-foreground">Loading…</div>;

  return (
    <div className={`relative h-full overflow-auto rounded-lg border border-border bg-card ${loading ? "opacity-60" : ""}`}>
      <table className="w-full text-sm">
        <thead className="sticky top-0 z-10 bg-secondary">
          <tr>
            <th className="w-14 border-b border-border px-3 py-2 text-right font-mono text-xs text-muted-foreground">#</th>
            {page.headers.map((h, i) => (
              <ColumnHeader key={i} name={h} index={i} loadStats={loadStats} stats={stats} />
            ))}
          </tr>
        </thead>
        <tbody>
          {page.rows.map((r, i) => (
            <tr key={i} className="border-b border-border last:border-0 hover:bg-secondary/60">
              <td className="px-3 py-1.5 text-right font-mono text-xs text-muted-foreground">
                {(offset + i + 1).toLocaleString()}
              </td>
              {r.map((c, j) => (
                <td key={j} className="max-w-[280px] truncate px-3 py-1.5 font-mono text-xs text-foreground" title={c}>
                  {c}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {page.rows.length === 0 && (
        <div className="p-8 text-center text-sm text-muted-foreground">No rows on this page.</div>
      )}
    </div>
  );
}

function ColumnHeader({ name, index, loadStats, stats }: {
  name: string;
  index: number;
  loadStats: () => Promise<main.StatsPayload | null>;
  stats: main.StatsPayload | null;
}) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const colStats = stats?.columns?.[index];

  async function maybeLoad() {
    if (!colStats && !loading) {
      setLoading(true);
      try { await loadStats(); } finally { setLoading(false); }
    }
  }

  return (
    <th className="border-b border-border px-3 py-2 text-left font-semibold text-foreground">
      <Popover open={open} onOpenChange={(o) => { setOpen(o); if (o) maybeLoad(); }}>
        <PopoverTrigger asChild>
          <button className="-mx-2 flex w-full items-center justify-between gap-2 rounded px-2 py-0.5 text-left transition-colors hover:bg-card">
            <span className="truncate" title={name}>{name}</span>
            <ChevronDown className="h-3 w-3 shrink-0 opacity-50" />
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-80">
          <div className="mb-3 flex items-center justify-between">
            <h4 className="text-sm font-semibold">{name}</h4>
            <span className="text-xs text-muted-foreground">column {index + 1}</span>
          </div>
          {loading && <div className="text-xs text-muted-foreground">Analyzing…</div>}
          {!loading && !colStats && <div className="text-xs text-muted-foreground">No stats yet.</div>}
          {colStats && <ColumnStatsView c={colStats} />}
        </PopoverContent>
      </Popover>
    </th>
  );
}

function ColumnStatsView({ c }: { c: main.StatsColumn }) {
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3 text-sm">
        <div>
          <div className="text-xs uppercase tracking-wide text-muted-foreground">Unique</div>
          <div className="font-semibold">{c.uniqueCapped ? `≥${c.unique}` : c.unique.toLocaleString()}</div>
        </div>
        <div>
          <div className="text-xs uppercase tracking-wide text-muted-foreground">Empty</div>
          <div className="font-semibold">{c.empty.toLocaleString()}</div>
        </div>
      </div>
      {c.top.length > 0 && (
        <div>
          <div className="mb-1 text-xs uppercase tracking-wide text-muted-foreground">Top values</div>
          <div className="space-y-1">
            {c.top.map((t, i) => (
              <div key={i} className="flex items-center justify-between gap-2 text-xs">
                <span className="truncate font-mono" title={t.value}>{t.value || <em className="text-muted-foreground">(empty)</em>}</span>
                <span className="rounded bg-secondary px-2 py-0.5 font-semibold">{t.count.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ---------- shared form widgets ------------------------------------------

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function PathPicker({ label, value, onPick, icon: Icon = Folder }: {
  label: string; value: string; onPick: () => void; icon?: any;
}) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onPick}>
          <Icon className="h-3.5 w-3.5" />Choose…
        </Button>
        {value ? (
          <code className="flex-1 truncate rounded-md bg-secondary px-3 py-1.5 font-mono text-xs text-muted-foreground" title={value}>{value}</code>
        ) : (
          <span className="text-xs italic text-muted-foreground">None</span>
        )}
      </div>
    </div>
  );
}

function Banner({ kind, output, children }: { kind: "success" | "error"; output?: string; children: React.ReactNode }) {
  const Icon = kind === "success" ? CheckCircle2 : AlertCircle;
  const bg = kind === "success" ? "bg-success/10 border-success/30" : "bg-destructive/10 border-destructive/30";
  const iconColor = kind === "success" ? "text-success" : "text-destructive";
  return (
    <div className={`flex items-start gap-3 rounded-lg border p-3 text-sm ${bg}`}>
      <Icon className={`mt-0.5 h-4 w-4 shrink-0 ${iconColor}`} />
      <div className="flex-1 text-foreground">{children}</div>
      {output && (
        <Button size="sm" variant="outline" onClick={() => RevealFile(output)}>
          <ExternalLink className="h-3.5 w-3.5" />Reveal
        </Button>
      )}
    </div>
  );
}

function GoButton({ onClick, loading, disabled, children }: {
  onClick: () => void; loading: boolean; disabled?: boolean; children: React.ReactNode;
}) {
  return (
    <Button onClick={onClick} disabled={loading || disabled} className="w-full">
      {loading && <Loader2 className="h-4 w-4 animate-spin" />}
      {children}
    </Button>
  );
}

// ---------- actions (slide-over forms) -----------------------------------

function FilterAction({ info, onDone }: { info: main.FileInfo; onDone: (output: string) => void }) {
  const headers = info.headers || [];
  const [column, setColumn] = useState(headers[0] || "");
  const [eqSet, setEqSet] = useState(false); const [eq, setEq] = useState("");
  const [containsSet, setContainsSet] = useState(false); const [contains, setContains] = useState("");
  const [gtSet, setGtSet] = useState(false); const [gt, setGt] = useState(0);
  const [ltSet, setLtSet] = useState(false); const [lt, setLt] = useState(0);
  const [all, setAll] = useState(false);
  const [output, setOutput] = useState(suggestOutput(info.path, "filtered"));
  const [result, setResult] = useState<main.FilterPayload | null>(null);
  const [err, setErr] = useState(""); const [loading, setLoading] = useState(false);

  async function pickOutput() { const p = await SaveCSVFile(output || "filtered.csv"); if (p) setOutput(p); }
  async function run() {
    if (!output) { setErr("Choose an output file."); return; }
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
    <div className="space-y-4">
      <Field label="Column">
        <Select value={column} onValueChange={setColumn}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            {headers.map((h) => <SelectItem key={h} value={h}>{h}</SelectItem>)}
          </SelectContent>
        </Select>
      </Field>

      <div className="space-y-2">
        <Label>Conditions</Label>
        <Cond checked={eqSet} onCheck={setEqSet} label="equals">
          <Input value={eq} disabled={!eqSet} onChange={(e) => setEq(e.target.value)} placeholder="value" />
        </Cond>
        <Cond checked={containsSet} onCheck={setContainsSet} label="contains">
          <Input value={contains} disabled={!containsSet} onChange={(e) => setContains(e.target.value)} placeholder="substring" />
        </Cond>
        <Cond checked={gtSet} onCheck={setGtSet} label="greater than">
          <Input type="number" value={gt} disabled={!gtSet} onChange={(e) => setGt(Number(e.target.value))} />
        </Cond>
        <Cond checked={ltSet} onCheck={setLtSet} label="less than">
          <Input type="number" value={lt} disabled={!ltSet} onChange={(e) => setLt(Number(e.target.value))} />
        </Cond>
        <label className="flex cursor-pointer items-center gap-2 pt-2 text-sm">
          <Checkbox checked={all} onCheckedChange={(v) => setAll(!!v)} />
          Match <strong>ALL</strong> conditions (AND)
        </label>
      </div>

      <PathPicker label="Output file" value={output} onPick={pickOutput} icon={Save} />

      <GoButton onClick={run} loading={loading} disabled={!column}>Run filter</GoButton>

      {err && <Banner kind="error">{err}</Banner>}
      {result && (
        <Banner kind="success" output={output}>
          Matched <strong>{result.matched.toLocaleString()}</strong> of{" "}
          <strong>{result.totalRows.toLocaleString()}</strong> rows.{" "}
          <button className="underline" onClick={() => onDone(output)}>Open result</button>
        </Banner>
      )}
    </div>
  );
}

function Cond({ checked, onCheck, label, children }: {
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

function DedupeAction({ info, onDone }: { info: main.FileInfo; onDone: (output: string) => void }) {
  const headers = info.headers || [];
  const [picked, setPicked] = useState<string[]>([]);
  const keysCSV = useMemo(() => picked.join(","), [picked]);
  const [keepLast, setKeepLast] = useState(false);
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [output, setOutput] = useState(suggestOutput(info.path, "dedup"));
  const [result, setResult] = useState<main.DedupePayload | null>(null);
  const [err, setErr] = useState(""); const [loading, setLoading] = useState(false);

  function toggle(h: string) { setPicked((p) => p.includes(h) ? p.filter((x) => x !== h) : [...p, h]); }
  async function pickOutput() { const p = await SaveCSVFile(output || "dedup.csv"); if (p) setOutput(p); }
  async function run() {
    if (!output) { setErr("Choose an output file."); return; }
    if (picked.length === 0) { setErr("Pick at least one key column."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await DedupeCSV({ input: info.path, output, keyColumns: keysCSV, keepLast, caseSensitive } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="space-y-4">
      <Field label={`Key columns (${picked.length} selected)`}>
        <div className="flex flex-wrap gap-2">
          {headers.map((h) => {
            const on = picked.includes(h);
            return (
              <button key={h} onClick={() => toggle(h)}
                className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${on ? "border-primary bg-primary text-primary-foreground" : "border-border bg-card text-foreground hover:bg-secondary"}`}>
                {h}
              </button>
            );
          })}
        </div>
      </Field>
      <div className="flex flex-wrap gap-5">
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={keepLast} onCheckedChange={(v) => setKeepLast(!!v)} />
          Keep last occurrence
        </label>
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={caseSensitive} onCheckedChange={(v) => setCaseSensitive(!!v)} />
          Case-sensitive
        </label>
      </div>
      <PathPicker label="Output file" value={output} onPick={pickOutput} icon={Save} />
      <GoButton onClick={run} loading={loading}>Run dedupe</GoButton>
      {err && <Banner kind="error">{err}</Banner>}
      {result && (
        <Banner kind="success" output={output}>
          Kept <strong>{result.uniqueRows.toLocaleString()}</strong>, removed{" "}
          <strong>{result.duplicates.toLocaleString()}</strong> duplicates.{" "}
          <button className="underline" onClick={() => onDone(output)}>Open result</button>
        </Banner>
      )}
    </div>
  );
}

function SplitAction({ info }: { info: main.FileInfo }) {
  const [outDir, setOutDir] = useState("");
  const [rowsPerFile, setRowsPerFile] = useState(1000);
  const [withHeader, setWithHeader] = useState(true);
  const [result, setResult] = useState<main.SplitPayload | null>(null);
  const [err, setErr] = useState(""); const [loading, setLoading] = useState(false);

  async function pickOutDir() { const p = await OpenDirectory("Select output directory"); if (p) setOutDir(p); }
  async function run() {
    if (!outDir) { setErr("Choose an output directory."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await SplitCSV({ input: info.path, outputDir: outDir, rowsPerFile, withHeader } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="space-y-4">
      <Field label="Rows per output file">
        <Input type="number" min={1} value={rowsPerFile} onChange={(e) => setRowsPerFile(Number(e.target.value))} />
      </Field>
      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <Checkbox checked={withHeader} onCheckedChange={(v) => setWithHeader(!!v)} />
        Include header in each chunk
      </label>
      <PathPicker label="Output directory" value={outDir} onPick={pickOutDir} />
      <GoButton onClick={run} loading={loading}>Run split</GoButton>
      {err && <Banner kind="error">{err}</Banner>}
      {result && (
        <Banner kind="success" output={outDir}>
          Wrote <strong>{result.filesCreated}</strong> file(s),{" "}
          <strong>{result.rowsProcessed.toLocaleString()}</strong> row(s) total.
        </Banner>
      )}
    </div>
  );
}

function MergeAction() {
  const [inDir, setInDir] = useState("");
  const [output, setOutput] = useState("");
  const [withHeader, setWithHeader] = useState(true);
  const [result, setResult] = useState<main.MergePayload | null>(null);
  const [err, setErr] = useState(""); const [loading, setLoading] = useState(false);

  async function pickInDir() { const p = await OpenDirectory("Select directory of CSV files"); if (p) setInDir(p); }
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
    <div className="space-y-4">
      <PathPicker label="Input directory" value={inDir} onPick={pickInDir} />
      <PathPicker label="Output file" value={output} onPick={pickOutput} icon={Save} />
      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <Checkbox checked={withHeader} onCheckedChange={(v) => setWithHeader(!!v)} />
        Use header from first file
      </label>
      <GoButton onClick={run} loading={loading}>Run merge</GoButton>
      {err && <Banner kind="error">{err}</Banner>}
      {result && (
        <Banner kind="success" output={output}>
          Merged <strong>{result.filesProcessed}</strong> file(s),{" "}
          <strong>{result.rowsWritten.toLocaleString()}</strong> row(s).
        </Banner>
      )}
    </div>
  );
}

function SQLiteAction({ info }: { info: main.FileInfo }) {
  const [dbPath, setDbPath] = useState(suggestOutput(info.path, "data", ".db"));
  const [table, setTable] = useState("");
  const [ifExists, setIfExists] = useState("replace");
  const [result, setResult] = useState<main.ToSQLitePayload | null>(null);
  const [err, setErr] = useState(""); const [loading, setLoading] = useState(false);

  async function pickDB() { const p = await SaveDBFile(dbPath || "data.db"); if (p) setDbPath(p); }
  async function run() {
    if (!dbPath) { setErr("Choose an output .db file."); return; }
    setLoading(true); setErr(""); setResult(null);
    try { setResult(await ToSQLiteCSV({ input: info.path, dbPath, table, ifExists } as any)); }
    catch (e: any) { setErr(String(e)); }
    finally { setLoading(false); }
  }

  return (
    <div className="space-y-4">
      <Field label="Table name" hint="Defaults to the input filename.">
        <Input value={table} onChange={(e) => setTable(e.target.value)} placeholder="users" />
      </Field>
      <Field label="If table exists">
        <Select value={ifExists} onValueChange={setIfExists}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="replace">replace — drop and re-create</SelectItem>
            <SelectItem value="append">append — add rows to existing</SelectItem>
            <SelectItem value="skip">skip — do nothing if exists</SelectItem>
            <SelectItem value="fail">fail — error if exists</SelectItem>
          </SelectContent>
        </Select>
      </Field>
      <PathPicker label="Output database (.db)" value={dbPath} onPick={pickDB} icon={Database} />
      <GoButton onClick={run} loading={loading}>Run import</GoButton>
      {err && <Banner kind="error">{err}</Banner>}
      {result?.skipped && <Banner kind="success" output={dbPath}>Table <code>{result.table}</code> already exists — skipped.</Banner>}
      {result && !result.skipped && (
        <Banner kind="success" output={dbPath}>
          Imported <strong>{result.rowsImported.toLocaleString()}</strong> row(s) into{" "}
          <code>{result.table}</code>.
        </Banner>
      )}
    </div>
  );
}
