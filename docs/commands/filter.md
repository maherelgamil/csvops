# 🔍 csvops filter

Filter rows in a CSV file based on a column's value using common conditions.

---

## 🧪 Examples

```bash
# Keep rows where country is Egypt
csvops filter --input data.csv --column country --eq Egypt --output egypt.csv

# Filter rows where age > 18
csvops filter --input people.csv --column age --gt 18 --enable-gt

# Filter names that contain "john"
csvops filter --input names.csv --column name --contains john
```

---

## 🔧 Available Flags

| Flag             | Description                                                | Default   |
|------------------|------------------------------------------------------------|-----------|
| `--input`        | Path to the input CSV file                                 | *(required)* |
| `--output`       | Path to the output CSV file                                | `stdout`  |
| `--column`       | Column to filter by                                        | *(required)* |
| `--eq`           | Keep rows where value equals this                          |           |
| `--contains`     | Keep rows where value contains this (case-insensitive)     |           |
| `--gt`           | Keep rows where value is greater than this (numeric only)  |           |
| `--lt`           | Keep rows where value is less than this (numeric only)     |           |
| `--enable-gt`    | Enable the --gt flag (must be set to apply it)             | `false`   |
| `--enable-lt`    | Enable the --lt flag (must be set to apply it)             | `false`   |
| `--with-header`  | Include the header row in the output                       | `true`    |

---

## 💡 Notes

- You can combine multiple filters (`--eq`, `--gt`, `--contains`) — any match passes.
- Numeric filters (`--gt`, `--lt`) only work if values can be parsed as floats.
- By default, output is printed to terminal unless `--output` is used.
- Case-insensitive matching is used for `--contains` by default.

