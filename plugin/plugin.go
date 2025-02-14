package plugin

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Args represents the plugin input arguments.
type Args struct {
	RepoPath      string `envconfig:"PLUGIN_REPO_PATH"`
	FilePath      string `envconfig:"PLUGIN_FILE_PATH" required:"true"`
	TrustedBranch string `envconfig:"PLUGIN_TRUSTED_BRANCH" required:"true"`
	CurrentBranch string `envconfig:"PLUGIN_CURRENT_BRANCH"`
	GitPat        string `envconfig:"PLUGIN_GIT_PAT"`
}

// Exec runs the plugin logic.
func Exec(ctx context.Context, args Args) (err error) {
	// We'll write the final TRUSTED output only once at the end.
	resultTrusted := "false"
	defer func() {
		if werr := WriteEnvToFile("TRUSTED", resultTrusted); werr != nil {
			logrus.Warnf("Failed to write TRUSTED variable: %v", werr)
		}
	}()

	repoPath := args.RepoPath
	if repoPath == "" {
		repoPath = os.Getenv("DRONE_WORKSPACE")
		if repoPath == "" {
			return fmt.Errorf("repo_path is not set and DRONE_WORKSPACE is unavailable")
		}
	}

	if args.CurrentBranch == "" {
		var err error
		args.CurrentBranch, err = getCurrentBranch(repoPath)
		if err != nil {
			return fmt.Errorf("failed to determine current branch: %w", err)
		}
	}

	if args.GitPat != "" {
		if err := configureGitCredentials(args.GitPat); err != nil {
			return fmt.Errorf("failed to configure git credentials: %w", err)
		}
	}

	// Attempt lightweight access: get the file content from the trusted branch.
	trustedContent, err := getFileContentFromBranch(repoPath, args.TrustedBranch, args.FilePath)
	if err != nil {
		logrus.Warnf("Lightweight access failed: %v. Falling back to heavyweight checkout...", err)
		trustedContent, err = checkoutAndReadFile(repoPath, args.TrustedBranch, args.FilePath)
		if err != nil {
			return fmt.Errorf("heavyweight checkout failed: %w", err)
		}
	}

	// For the current branch, read the file directly from the filesystem.
	currentFilePath := filepath.Join(repoPath, args.FilePath)
	currentContentBytes, err := os.ReadFile(currentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file from current branch at %s: %w", currentFilePath, err)
	}
	currentContent := string(currentContentBytes)

	// Compare file contents.
	if trustedContent != currentContent {
		return fmt.Errorf("file content mismatch between branch '%s' and trusted branch '%s'", args.CurrentBranch, args.TrustedBranch)
	}

	// Verification succeeded.
	resultTrusted = "true"

	// Encode the file content in Base64.
	encodedContent := base64.StdEncoding.EncodeToString([]byte(trustedContent))

	// Export TRUSTED_FILE_CONTENT as an output variable.
	if err := WriteEnvToFile("TRUSTED_FILE_CONTENT", encodedContent); err != nil {
		return fmt.Errorf("failed to write TRUSTED_FILE_CONTENT: %w", err)
	}

	logrus.Info("File content matches the trusted branch. Validation succeeded.")
	return nil
}

func getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// configureGitCredentials sets up Git credentials in a cross-platform manner.
func configureGitCredentials(gitPat string) error {
	cmd := exec.Command("git", "config", "--global", "credential.helper", "store")
	if err := cmd.Run(); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	credFilePath := filepath.Join(home, ".git-credentials")
	// Use the recommended format for GitHub PAT authentication.
	credContent := fmt.Sprintf("https://x-access-token:%s@github.com", gitPat)
	return os.WriteFile(credFilePath, []byte(credContent), 0644)
}

func getFileContentFromBranch(repoPath, branch, filePath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "show", fmt.Sprintf("%s:%s", branch, filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func checkoutAndReadFile(repoPath, branch, filePath string) (string, error) {
	// Fetch the branch from remote.
	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "origin", branch)
	if err := fetchCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to fetch branch %s: %w", branch, err)
	}

	// Check out the branch, updating/creating the local branch from origin.
	checkoutCmd := exec.Command("git", "-C", repoPath, "checkout", "-B", branch, "origin/"+branch)
	if err := checkoutCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}

	fullPath := filepath.Join(repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	return string(content), nil
}

// package plugin

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"

// 	"github.com/sirupsen/logrus"
// )

// // Args represents the plugin input arguments.
// type Args struct {
// 	RepoPath      string `envconfig:"PLUGIN_REPO_PATH"`
// 	FilePath      string `envconfig:"PLUGIN_FILE_PATH" required:"true"`
// 	TrustedBranch string `envconfig:"PLUGIN_TRUSTED_BRANCH" required:"true"`
// 	CurrentBranch string `envconfig:"PLUGIN_CURRENT_BRANCH"`
// 	GitPat        string `envconfig:"PLUGIN_GIT_PAT"`
// }

// // Exec runs the plugin logic.
// func Exec(ctx context.Context, args Args) (err error) {
// 	// We'll write the final TRUSTED output only once at the end.
// 	resultTrusted := "false"
// 	defer func() {
// 		if werr := WriteEnvToFile("TRUSTED", resultTrusted); werr != nil {
// 			logrus.Warnf("Failed to write TRUSTED variable: %v", werr)
// 		}
// 	}()

// 	repoPath := args.RepoPath
// 	if repoPath == "" {
// 		repoPath = os.Getenv("DRONE_WORKSPACE")
// 		if repoPath == "" {
// 			return fmt.Errorf("repo_path is not set and DRONE_WORKSPACE is unavailable")
// 		}
// 	}

// 	if args.CurrentBranch == "" {
// 		var err error
// 		args.CurrentBranch, err = getCurrentBranch(repoPath)
// 		if err != nil {
// 			return fmt.Errorf("failed to determine current branch: %w", err)
// 		}
// 	}

// 	if args.GitPat != "" {
// 		if err := configureGitCredentials(args.GitPat); err != nil {
// 			return fmt.Errorf("failed to configure git credentials: %w", err)
// 		}
// 	}

// 	// Attempt lightweight access: get the file content from the trusted branch.
// 	trustedContent, err := getFileContentFromBranch(repoPath, args.TrustedBranch, args.FilePath)
// 	if err != nil {
// 		logrus.Warnf("Lightweight access failed: %v. Falling back to heavyweight checkout...", err)
// 		trustedContent, err = checkoutAndReadFile(repoPath, args.TrustedBranch, args.FilePath)
// 		if err != nil {
// 			return fmt.Errorf("heavyweight checkout failed: %w", err)
// 		}
// 	}

// 	// For the current branch, read the file directly from the filesystem.
// 	currentFilePath := filepath.Join(repoPath, args.FilePath)
// 	currentContentBytes, err := os.ReadFile(currentFilePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read file from current branch at %s: %w", currentFilePath, err)
// 	}
// 	currentContent := string(currentContentBytes)

// 	// Compare file contents.
// 	if trustedContent != currentContent {
// 		return fmt.Errorf("file content mismatch between branch '%s' and trusted branch '%s'", args.CurrentBranch, args.TrustedBranch)
// 	}

// 	// Verification succeeded.
// 	resultTrusted = "true"

// 	// Output the trusted file content.
// 	fmt.Printf("TRUSTED_FILE_CONTENT=%s\n", trustedContent)
// 	logrus.Info("File content matches the trusted branch. Validation succeeded.")
// 	return nil
// }

// func getCurrentBranch(repoPath string) (string, error) {
// 	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return "", err
// 	}
// 	return strings.TrimSpace(string(output)), nil
// }

// // configureGitCredentials sets up Git credentials in a cross-platform manner.
// // func configureGitCredentials(gitPat string) error {
// // 	cmd := exec.Command("git", "config", "--global", "credential.helper", "store")
// // 	if err := cmd.Run(); err != nil {
// // 		return err
// // 	}

// // 	home, err := os.UserHomeDir()
// // 	if err != nil {
// // 		return err
// // 	}
// // 	credFilePath := filepath.Join(home, ".git-credentials")
// // 	credContent := fmt.Sprintf("https://%s@github.com", gitPat)
// // 	return os.WriteFile(credFilePath, []byte(credContent), 0644)
// // }

// // configureGitCredentials sets up Git credentials in a cross-platform manner.
// func configureGitCredentials(gitPat string) error {
// 	cmd := exec.Command("git", "config", "--global", "credential.helper", "store")
// 	if err := cmd.Run(); err != nil {
// 		return err
// 	}

// 	home, err := os.UserHomeDir()
// 	if err != nil {
// 		return err
// 	}
// 	credFilePath := filepath.Join(home, ".git-credentials")
// 	// Use the recommended format for GitHub PAT authentication.
// 	credContent := fmt.Sprintf("https://x-access-token:%s@github.com", gitPat)
// 	return os.WriteFile(credFilePath, []byte(credContent), 0644)
// }

// func getFileContentFromBranch(repoPath, branch, filePath string) (string, error) {
// 	cmd := exec.Command("git", "-C", repoPath, "show", fmt.Sprintf("%s:%s", branch, filePath))
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(output), nil
// }

// func checkoutAndReadFile(repoPath, branch, filePath string) (string, error) {
// 	// Fetch the branch from remote.
// 	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "origin", branch)
// 	if err := fetchCmd.Run(); err != nil {
// 		return "", fmt.Errorf("failed to fetch branch %s: %w", branch, err)
// 	}

// 	// Check out the branch, updating/creating the local branch from origin.
// 	checkoutCmd := exec.Command("git", "-C", repoPath, "checkout", "-B", branch, "origin/"+branch)
// 	if err := checkoutCmd.Run(); err != nil {
// 		return "", fmt.Errorf("failed to checkout branch %s: %w", branch, err)
// 	}

// 	fullPath := filepath.Join(repoPath, filePath)
// 	content, err := os.ReadFile(fullPath)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
// 	}
// 	return string(content), nil
// }
