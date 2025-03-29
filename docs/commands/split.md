# 📂 csvops split

Split a large CSV file into smaller chunks based on a fixed number of rows per file.

---

## 🧪 Example

```bash
csvops split \
  --input big.csv \
  --rows 1000 \
  --output-dir ./chunks \
  --with-header
```

---

## 🔧 Available Flags

| Flag           | Description                                         | Default       |
|----------------|-----------------------------------------------------|---------------|
| `--input`      | Path to the input CSV file                         | *(required)*  |
| `--rows`       | Max rows per output file                           | `1000`        |
| `--output-dir` | Directory to write the output files                | `./output`    |
| `--with-header`| Include the header row in every output chunk       | `true`        |
| `--delimiter`  | Delimiter character used in CSV (e.g., `;`)        | `,`           |

---

## 💡 Notes

- The tool automatically creates the `output-dir` if it doesn't exist.
- File names will follow the pattern: `part_1.csv`, `part_2.csv`, etc.
- If `--with-header=false`, the header row will only appear in the first file (or none).

