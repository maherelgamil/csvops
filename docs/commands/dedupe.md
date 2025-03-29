# ♻️ csvops dedupe

Remove duplicate rows from a CSV file based on one or more key columns.

---

## 🧪 Example

```bash
csvops dedupe \
  --input users.csv \
  --output unique.csv \
  --key email
```

---

## 🔧 Available Flags

| Flag               | Description                                    | Default      |              |
| ------------------ | ---------------------------------------------- | ------------ | ------------ |
| `--input`          | Path to the input CSV file                     |              | *(required)* |
| `--output`         | Path to write the output file                  | *(required)* |              |
| `--key`            | Comma-separated column(s) to use as unique key | *(required)* |              |
| `--keep-last`      | Keep the last occurrence instead of the first  | `false`      |              |
| `--case-sensitive` | Treat key values as case-sensitive             | `false`      |              |

---

## 💡 Notes

- By default, only the first occurrence of each key is kept.
- Use `--keep-last` to reverse this behavior.
- Rows with missing key columns are skipped.
- Use multiple keys like: `--key email,phone`

