# 🔗 csvops merge

Merge multiple CSV files into one. Files must have the same column structure.

---

## 🧪 Example

```bash
csvops merge \
  --input file1.csv,file2.csv,file3.csv \
  --output merged.csv
```

---

## 🔧 Available Flags

| Flag           | Description                                        | Default     |
|----------------|----------------------------------------------------|-------------|
| `--input`      | Comma-separated list of CSV files to merge         | *(required)*|
| `--output`     | Path to save the merged CSV                        | *(required)*|
| `--with-header`| Include header row once (from the first file)      | `true`      |

---

## 💡 Notes

- All files must have the **same column structure**.
- Header will be taken from the first file if `--with-header` is true.
- Order of merging follows the order of the files in `--input`.
- Supports large files — merging is done line-by-line.

---

## 🧼 Coming Soon

- Option to validate matching headers
- Support for inner/outer joins by key column (planned)

