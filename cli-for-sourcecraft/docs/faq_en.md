❓ Frequently Asked Questions (FAQ)

Q: I get an error like Error: token not found... when running commands. A: This means the CLI couldn't find your SourceCraft Personal Access Token (PAT). Make sure you have:

Run src auth login and pasted your valid PAT.

Alternatively, set the SOURCECRAFT_TOKEN environment variable.

Check that your configuration file (usually ~/.config/src/config.yaml or in the current directory) exists and contains the token: line if you used auth login with the non-secure version.

Q: Commands like repo list or repo create return HTML (<!DOCTYPE html>...) or a non-JSON response error. A: This usually indicates an issue with authentication or the API endpoint address:

Invalid Token: Your PAT might have expired or been revoked. Run src auth logout and then src auth login again with a valid token.

Incorrect API URL: Check your api_base_url setting. Run src config get api_base_url. It should be https://sourcecraft.dev/api/v1. If it's different, correct it with src config set api_base_url https://sourcecraft.dev/api/v1.

Q: I get a 404 Not Found error, often with Cannot POST /api/v1/... or workflow '' not found. A: This means the specific API endpoint wasn't found by the server.

For repo create: This might indicate a server-side issue with the POST /orgs/.../repos endpoint or a permissions problem. Check if you can create repos via the web UI.

For workflow run: Ensure the workflow name you provided (e.g., main) exactly matches the filename (without .yml) in your repository's .sourcecraft/ci/ directory and that the file has been pushed to the server. Allow a few minutes after pushing for the CI system to recognize it. Also, double-check the API path used in internal/api/client.go matches the Swagger definition (e.g., /org/repo/... vs /repos/org/repo/...).

General: Verify the api_base_url is correct (https://sourcecraft.dev/api/v1).

Q: Commands like repo sync or pr create say they could not detect repository. A: These commands try to guess the repository from your Git remote named origin. Make sure:

You are running the command from within your local Git repository clone.

You have a remote configured named origin that points to your SourceCraft repository (e.g., git@sourcecraft.dev:your-org/your-repo.git). Check with git remote -v.

Alternatively, use the --repo <org>/<repo> flag to specify the repository explicitly for the command.

Q: How do I get the User ID (UUID) for src access role add/remove? A: The CLI doesn't have a command to search for user UUIDs. You need to find the UUID in the SourceCraft web interface, typically on the user's profile page or in the organization/repository member list.

Q: Why aren't there commands for security scanning, package management, CI logs, etc.? A: The CLI can only implement features that are available through the SourceCraft REST API. Based on the provided OpenAPI (Swagger) specification, endpoints for those features do not currently exist.

Q: Where is my configuration file saved? A: By default, the CLI looks for config.yaml in the current directory (.) and then in ~/.config/src/ (where ~ is your home directory). You can see which file is being used by running a command with the -v (verbose) flag, e.g., src -v repo list. You can also specify a different path using the global --config flag.

Q: Is my token stored securely? A: No. The current implementation (after rolling back from keyring) stores the token in plain text in the config.yaml file. This is not secure. For better security, consider re-implementing the authentication using system keychains (like macOS Keychain, Windows Credential Manager, Linux Secret Service) via a library like go-keyring.