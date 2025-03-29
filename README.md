# 📊 csvops [![GitHub stars](https://img.shields.io/github/stars/maherelgamil/csvops?style=social)](https://github.com/maherelgamil/csvops/stargazers)

A blazing fast, modular CLI toolkit for working with CSV files.

**Built with Go. Designed for humans.**

---

## 🚀 Quick Install

```bash
brew tap maherelgamil/tap
brew install csvops
```

Or download the binary manually from [Releases](https://github.com/maherelgamil/csvops/releases).

---

## 🛠 Supported Commands

| Command     | Description                                                  |
|-------------|--------------------------------------------------------------|
| `split`     | Split a large CSV file into smaller parts                    |
| `dedupe`    | Remove duplicate rows by one or more columns                |
| `filter`    | Filter rows based on column values                          |
| `stats`     | Show summary statistics (rows, columns, empty/unique counts) |
| `preview`   | Display the first N rows of a CSV file                      |
| `merge`     | Combine multiple CSV files into one                         |

📚 See full usage details in [`docs/commands`](./docs/commands)

---

## 🧪 Examples

```bash
# Split a file into parts of 1000 rows each
csvops split --input big.csv --rows 1000 --output-dir ./out

# Dedupe by "email" column
csvops dedupe --input users.csv --output clean.csv --key email

# Filter rows where country == Egypt
csvops filter --input users.csv --column country --eq Egypt --output egyptians.csv

# View stats for a file
csvops stats --input users.csv

# Preview the first 10 rows
csvops preview --input data.csv --rows 10

# Merge multiple files
csvops merge --input file1.csv,file2.csv --output merged.csv
```

---

## 📦 Install From Source

```bash
git clone https://github.com/maherelgamil/csvops.git
cd csvops
go build -o csvops
./csvops --help
```

---

## 🧩 Contributing

We love contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## 📜 License

MIT © [Maher El Gamil](LICENSE)