# 📊 csvops stats

Get a quick summary of the structure and content of a CSV file.

---

## 🧪 Example

```bash
csvops stats --input data.csv
```

---

## 🔧 Available Flags

| Flag         | Description                      | Default     |
|--------------|----------------------------------|-------------|
| `--input`    | Path to the input CSV file       | *(required)*|

---

## 📋 What It Shows

- Total row count (excluding header)
- Number of columns
- Column names
- Count of empty values per column
- Count of unique values per column
- Top 3 most frequent values per column

---

## 💡 Notes

- Output is displayed in a formatted table.
- Helps diagnose missing data, repetitive fields, or unexpected values.
- Use this before cleaning, filtering, or exporting your data.

