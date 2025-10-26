# 🤖 SourceCraft CLI: Command Line Interface

A command-line tool for **SourceCraft** that allows you to manage your repositories, pull requests, issues, and CI/CD directly from your terminal.

---

## 🚀 Usage Examples

A quick start with the most popular commands.

| Action | Command |
| :--- | :--- |
| **List** open pull requests in the current repository | `src pr list --state open` |
| **Create** a new repository named "My Project" | `src repo create --name "My Project" --description "My new project" --private` |
| **View** details of PR #10 in `your-org/your-repo` | `src pr view 10 --repo your-org/your-repo` |
| **Merge** PR #12 (squash and delete branch) | `src pr merge 12 --squash --delete-branch` |
| **Create** a new issue interactively | `src issue create` |
| **Close** issue #3 | `src issue close 3` |
| **Run** the CI workflow 'main' on the 'develop' branch | `src workflow run main --ref develop` |
| **Add** a user (by ID) with the 'developer' role | `src access role add <User ID> --role developer` |
| **Install** the pre-commit hook | `src pre-commit install` |

---

## ❓ Frequently Asked Questions (FAQ)

### Authentication and Token (PAT) Issues

**Q:** I get the error `Error: token not found...` when running commands.
**A:** This means the CLI couldn't find your **SourceCraft Personal Access Token (PAT)**. Ensure you have:
1.  Executed `src auth login` and inserted a valid PAT.
2.  **OR** set the environment variable `SOURCECRAFT_TOKEN`.
3.  Checked that your configuration file (usually `~/.config/src/config.yaml` or in the current folder) exists and contains the line `token:` if you used `auth login`.

### API Errors and Non-JSON Response

**Q:** Commands like `repo list` or `repo create` return HTML (`<!DOCTYPE html>...`) or a `non-JSON response` error.
**A:** This usually indicates an authentication or API address problem:
1.  **Invalid Token:** Your PAT might have expired or been revoked. Run `src auth logout`, then `src auth login` with a valid token.
2.  **Incorrect API URL:** Check the `api_base_url` setting. Run `src config get api_base_url`. The value should be **`https://sourcecraft.dev/api/v1`**. If it's different, correct it using `src config set api_base_url https://sourcecraft.dev/api/v1`.

**Q:** I get a **404 Not Found** error, often with the message `Cannot POST /api/v1/...` or `workflow '' not found`.
**A:** This means the server could not find the specific API endpoint.
* **For `repo create`:** May indicate a server-side issue with the `POST /orgs/.../repos` endpoint or a permissions problem. Check if you can create repositories via the web interface.
* **For `workflow run`:** Ensure the workflow name (e.g., `main`) **exactly matches** the file name (without `.yml`) in your repository's `.sourcecraft/ci/` folder, and that the file has been pushed to the server (`git push`). Use the correct name (`main`, without `.yml`).
* **General:** Make sure the `api_base_url` is correct (`https://sourcecraft.dev/api/v1`).

### Repository and Git Errors

**Q:** Commands like `repo sync` or `pr create` report that **the repository could not be determined**.
**A:** These commands attempt to determine the repository from your Git remote named `origin`. Ensure that:
1.  You are running the command **inside your local Git repository clone**.
2.  You have an `origin` remote configured that points to your SourceCraft repository (e.g., `git@sourcecraft.dev:your-org/your-repo.git`). Check with `git remote -v`.
3.  **Alternatively**, use the flag `--repo <org>/<repo>` to explicitly specify the repository for the command.

### Other Questions

**Q:** Where is my configuration file saved?
**A:** By default, the CLI looks for `config.yaml` in the **current directory** (`.`), and then in **`~/.config/src/`**.

---

## 🏗️ Architecture

The CLI is built using **Go** and the **Cobra** library.

| Component | Location | Description |
| :--- | :--- | :--- |
| **Command Definitions** | `/cmd` | Contains the logic for all commands. |
| **Base Command** | `root.go` | Defines the base `src` command, handles global flags, and initializes the API client. |
| **Action Commands** | `*.go` in `/cmd` | Implement the execution logic within the `RunE` function. |
| **API Client** | `internal/api/client.go` | Logic for API requests (`makeRequest`), data structures, and API call methods. |
| **Git Utility Functions** | `internal/git/git.go` | Functions for executing local `git` commands. |

---

## 🛠️ Extending Functionality

To add a new command (e.g., `src newthing doaction`):

1.  **Define the Action Command**: Create `cmd/newthing_doaction.go`, define `newthingDoactionCmd`, and implement the core logic in the **`RunE`** field.
2.  **API Interaction**: Add the necessary method (e.g., `DoAction(...)`) to the `api.Client` struct in `internal/api/client.go` and call it from `RunE`.
3.  **Git Interaction**: Use functions from `internal/git/git.go` if the command needs to interact with the local repository.

---

## 📦 Building and Deployment

### Building
1.  **Install Go**: Ensure you have Go version 1.18+ installed.
2.  **Clone the repository**:
    ```bash
    git clone <repository address>
    cd <folder name>
    ```
3.  **Build the executable**:
    ```bash
    go build -o src .
    ```
4.  **(Optional) Embed Version**:
    ```bash
    go build -o src -ldflags "-X main.version=v1.2.3" .
    ```

### Cross-Compilation
Use the environment variables **`GOOS`** and **`GOARCH`**:
```bash
# Build for Windows x64
env GOOS=windows GOARCH=amd64 go build -o src.exe .

# Build for Linux x64
env GOOS=linux GOARCH=amd64 go build -o src .