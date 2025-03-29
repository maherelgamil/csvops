# 🗃️ csvops to-sqlite

Convert a CSV file into a SQLite database with a single table.

---

## 🧪 Example

```bash
csvops to-sqlite \
  --input data.csv \
  --output data.db \
  --table users
```

---

## 🔧 Available Flags

| Flag           | Description                                                   | Default       |
|----------------|---------------------------------------------------------------|---------------|
| `--input`      | Path to the input CSV file                                    | *(required)*  |
| `--output`     | Path to the output `.db` SQLite database file                 | *(required)*  |
| `--table`      | Name of the table to create (defaults to CSV filename)        | *(auto)*      |
| `--delimiter`  | CSV delimiter character                                       | `,`           |
| `--if-exists`  | What to do if the DB/table exists: `skip` or `replace`        | `replace`     |

---

## 💡 Notes

- If no `--table` is provided, the table name is inferred from the input file name.
- Column types are inferred as `TEXT` by default.
- If `--if-exists=replace`, the table will be dropped and recreated.
- Includes a real-time progress bar for inserting rows.
- Use SQLite tools like `sqlite3` to inspect or query the database.

---

## 📂 Example Workflow

```bash
csvops to-sqlite --input users.csv --output users.db
sqlite3 users.db "SELECT COUNT(*) FROM users;"
```
