package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

const logLevel slog.Level = slog.LevelInfo

func initLogger() {
	o := slog.HandlerOptions{Level: logLevel}
	h := slog.NewTextHandler(os.Stderr, &o)
	lg := slog.New(h)
	slog.SetDefault(lg)
}

func parseCmdline() (paths []string) {
	var usage string
	var ret []string
	usage = fmt.Sprintf("usage: %s path [path [path [...]]]\n", filepath.Base(os.Args[0]))
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}
	ret = append(ret, os.Args[1:]...)
	return ret
}

func main() {
	var pp []string
	initLogger()
	pp = parseCmdline()
	slog.Info("Starting", "paths", pp)
	for _, p := range pp {
		SaveFile(p)
	}
}

type tContext struct {
	sNow string
	sMe  string
}

func getContextSingleton() *tContext {
	var now time.Time
	var me *user.User
	var sNow, sMe string
	var err error
	if contextSingleton == nil {
		now = time.Now()
		sNow = fmt.Sprintf("%04d%02d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		if me, err = user.Current(); err != nil {
			slog.Error("Cannot get current user information", "error", err)
			panic(err)
		}
		sMe = me.Name
		if len(sMe) == 0 {
			slog.Debug("Current user real name unknown, falling back to login name")
			sMe = me.Username
		}
		contextSingleton = &tContext{sNow, sMe}
		slog.Debug("Created context singleton", "sNow", contextSingleton.sNow, "sMe", contextSingleton.sMe)
	}
	return contextSingleton
}

var contextSingleton *tContext

func SaveFile(sourcePath string) {
	// Takes a backup of the file privided in argument
	var err error
	var targetPath string
	if sourcePath, err = filepath.Abs(sourcePath); err != nil {
		slog.Error("Cannot convert to absolute path", "path", sourcePath, "error", err)
		panic(err)
	}
	targetPath = computeTopLevelTargetPath(sourcePath)
	processEntry(sourcePath, targetPath)
}

func processEntry(sourcePath string, targetPath string) {
	var sourceFile *os.File
	var sourceInfo os.FileInfo
	var err error
	slog.Debug(
		"Processing",
		"sourcePath", sourcePath, "targetPath", targetPath,
	)
	if sourceFile, err = os.Open(sourcePath); err != nil {
		slog.Error("Cannot open for reading", "path", sourcePath, "error", err)
		panic(err)
	}
	defer sourceFile.Close()
	if sourceInfo, err = sourceFile.Stat(); err != nil {
		slog.Error("Cannot get file info", "path", sourcePath, "error", err)
		panic(err)
	}
	if sourceInfo.IsDir() {
		processDirectory(sourcePath, sourceFile, sourceInfo, targetPath)
	} else {
		processFile(sourcePath, sourceFile, sourceInfo, targetPath)
	}
}

func mirrorMtimeAndMode(path string, info os.FileInfo) {
	var err error
	var mtime, atime time.Time
	slog.Info("Setting properties", "path", path)
	mtime = info.ModTime()
	atime = time.Time{} // time 0 -> not changing
	if os.Chtimes(path, atime, mtime) != nil {
		slog.Error("Cannot set time", "mtime", mtime, "atime", atime, "path", path, "error", err)
	}
	os.Chmod(path, info.Mode())
}

func processDirectory(
	sourcePath string, sourceFile *os.File, sourceInfo os.FileInfo,
	targetPath string,
) {
	slog.Info("Creating", "directory", targetPath, "source", sourcePath)
	defer mirrorMtimeAndMode(targetPath, sourceInfo)
	os.Mkdir(targetPath, 0777) // umask will remove what's needed
	diveDirectory(sourceFile, targetPath)
}

func diveDirectory(dir *os.File, targetDirName string) {
	var childs []os.DirEntry
	var err error
	var sourcePath, targetPath string
	if childs, err = dir.ReadDir(-1); err != nil {
		slog.Error(
			"Cannot get directory content", "path", dir.Name(),
		)
		panic(err)
	}
	for _, child := range childs {
		sourcePath = fmt.Sprintf("%s%c%s", dir.Name(), os.PathSeparator, child.Name())
		targetPath = fmt.Sprintf("%s%c%s", targetDirName, os.PathSeparator, child.Name())
		processEntry(sourcePath, targetPath)
	}
}

func processFile(
	sourcePath string, sourceFile *os.File, sourceInfo os.FileInfo,
	targetPath string,
) {
	var targetFile *os.File
	var err error
	slog.Info("Creating", "file", targetPath, "source", sourcePath)

	if targetFile, err = os.Create(targetPath); err != nil {
		slog.Error("Cannot open file for writing", "path", targetPath, "error", err)
		panic(err)
	}

	defer func() {
		targetFile.Close()
		mirrorMtimeAndMode(targetPath, sourceInfo)
	}()

	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		slog.Error("Cannot open file for writing", "path", targetPath, "error", err)
		panic(err)
	}
}

func computeTopLevelTargetPath(sourcePath string) string {
	const extSep rune = '.'
	var posn = -1
	var base, dir, stem, ext string
	base = filepath.Base(sourcePath)
	dir = filepath.Dir(sourcePath)
	for i, c := range base {
		// Searching for the extension. Since we want to have the last
		// extension separator, we must go to the end of the loop
		if c == extSep {
			posn = i
		}
	}
	if posn > 0 { // -1 if not found, 0 if starting with dot (=> N/A)
		stem = base[:posn]
		ext = base[posn:]
	} else {
		ext = ""
		stem = base
	}
	tmp := getContextSingleton()
	return fmt.Sprintf("%s%c%s.bak_%s_%s%s", dir, os.PathSeparator, stem, tmp.sNow, tmp.sMe, ext)
}
