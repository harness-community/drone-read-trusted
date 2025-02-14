# drone-read-trusted

## Synopsis

The **drone-read-trusted** plugin replicates the functionality of Jenkins’ **readTrusted** step in a Drone/Harness CI environment. It securely retrieves a file from a trusted branch in your Git repository and compares it against the file in the current branch. If the file in the current branch has been modified relative to its trusted version, the plugin fails—ensuring that only files from trusted sources are used in your pipeline.

## Functionality

- **Trusted File Retrieval:**  
  The plugin fetches the file content from a trusted branch using a lightweight method (`git show`) and falls back to a heavyweight checkout (with `git fetch` and `git checkout`) if necessary.

- **Current File Verification:**  
  It reads the file from the current branch directly from the local filesystem and compares it with the trusted branch’s version.

- **Output Variables:**  
  - `TRUSTED`: Set to `"true"` if the file content matches the trusted branch, or `"false"` otherwise.
  - `TRUSTED_FILE_CONTENT`: The Base64-encoded content of the trusted file. This encoded value can be decoded in subsequent steps.

- **Git PAT Support:**  
  For private repositories, the plugin accepts a Git Personal Access Token (PAT) to configure Git credentials correctly.

## Parameters

| Parameter          | Type     | Required/Default           | Description                                                                                     |
|--------------------|----------|----------------------------|-------------------------------------------------------------------------------------------------|
| `repo_path`        | string   | Default: `DRONE_WORKSPACE` | Filesystem path to the cloned repository.                                                     |
| `file_path`        | string   | **Required**               | Relative path to the file within the repository (e.g., `Jenkinsfile` or `config/settings.yml`).   |
| `trusted_branch`   | string   | **Required**               | Name of the trusted branch (e.g., `main` or `jenkins`) used for file verification.              |
| `current_branch`   | string   | Auto-detected              | Name of the current branch. If not provided, the plugin auto-detects it using Git.               |
| `git_pat`          | string   | Optional                   | Git Personal Access Token for accessing private repositories.                                 |

## Outputs

| Output Variable           | Description                                                                                                 |
|---------------------------|-------------------------------------------------------------------------------------------------------------|
| `TRUSTED`                 | `"true"` if the file content matches the trusted branch; `"false"` otherwise.                              |
| `TRUSTED_FILE_CONTENT`    | Base64-encoded content of the trusted file, available for use in subsequent pipeline steps.                   |

## Usage Example

Below is an example Drone pipeline step that uses the plugin:

```yaml
kind: pipeline
type: docker
name: read-trusted-example

steps:
  - name: read-trusted
    image: plugins/read-trusted:linux-amd64
    settings:
      file_path: "File1.txt"
      trusted_branch: "xyz"
      git_pat: "<+secrets.getValue('git_pat')>"
```

In this example, the plugin:
- Reads the File1.txt from the repository.
- Compares the file from the current branch (read directly from the filesystem) with that from the "xyz" branch.
- Exports TRUSTED=true and TRUSTED_FILE_CONTENT (Base64-encoded) if the file contents match.
- Fails the build if there is any discrepancy.

Notes
- Private Repositories:
When using private repositories, ensure you provide a valid Git PAT (with the proper scopes) via the git_pat input parameter. The plugin uses the format https://x-access-token:<git_pat>@github.com for authentication.

