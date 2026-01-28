package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const modulePath = "github.com/solean/letterboxd-tui"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func String() string {
	version := resolvedVersion()
	commit := resolvedCommit()
	date := resolvedDate()
	details := []string{}
	if commit != "" {
		details = append(details, "commit "+commit)
	}
	if date != "" {
		details = append(details, "built "+date)
	}
	if len(details) == 0 {
		return version
	}
	return fmt.Sprintf("%s (%s)", version, strings.Join(details, ", "))
}

func UserAgent() string {
	return fmt.Sprintf("%s/%s", modulePath, resolvedVersion())
}

func resolvedVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	if Version == "" {
		return "dev"
	}
	return Version
}

func resolvedCommit() string {
	if Commit != "" && Commit != "none" {
		return shortCommit(Commit)
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				return shortCommit(setting.Value)
			}
		}
	}
	return ""
}

func resolvedDate() string {
	if Date != "" && Date != "unknown" {
		return Date
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" && setting.Value != "" {
				return setting.Value
			}
		}
	}
	return ""
}

func shortCommit(commit string) string {
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}
