# 🙌 Contributing to csvops

First off, thanks for taking the time to contribute! We welcome PRs, bug reports, and feature requests.

---

## 🧠 What is csvops?

`csvops` is a modular CLI toolkit for working with CSV files — blazing fast, cleanly organized, and built with Go. It’s designed to help anyone quickly split, filter, dedupe, and inspect CSV data.

---

## 🛠 How to Contribute

### 1. Fork the Repo

```bash
git clone https://github.com/maherelgail/csvops.git
cd csvops
```

### 2. Install & Run

```bash
go build -o csvops
./csvops --help
```

### 3. Make Your Changes

Add a new command under `cmd/`, or enhance an existing one.

### 4. Lint & Format

```bash
go fmt ./...
go vet ./...
```

### 5. Test Locally

Try your new or updated command using sample CSV files.

### 6. Open a Pull Request

- Give your PR a clear title and description
- Link to any related issues
- Keep it focused — one feature/fix per PR

---

## 🧪 Code Style

- Keep commands modular under `cmd/`
- Use Cobra CLI for all commands
- Keep CLI help text and flags well-documented
- Use `tablewriter` and `progressbar` when it improves UX

---

## 💬 Need Help?

Feel free to open an issue or start a discussion in GitHub. We’re happy to support contributors!

---

Thanks for helping make `csvops` awesome 🚀

