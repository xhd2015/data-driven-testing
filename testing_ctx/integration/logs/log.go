package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

// Watch monitors a file for changes and calls the provided callback function
// for each new line detected.
//
// Parameters:
//   - ctx: Context for cancellation
//   - filepath: Path to the file to monitor
//   - callback: Function called for each new line with the line content as parameter
//
// The callback function should return an error if processing should stop.
// The function continues until the context is cancelled, an error occurs,
// or the callback returns an error.
//
// If the file doesn't exist initially, it will wait for it to be created.
func Watch(ctx context.Context, filepath string, callback func(content []byte) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Add the file's directory to watch for file creation
	dir := path.Dir(filepath)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

	// Keep track of last read position
	lastPos := int64(0)

	// If file exists, add it to the watcher and initialize lastPos to current size
	if fileExists(filepath) {
		if err := watcher.Add(filepath); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", filepath, err)
		}

		// Get the current file size to start watching only new content
		file, err := os.Open(filepath)
		if err == nil {
			fileInfo, err := file.Stat()
			if err == nil {
				lastPos = fileInfo.Size()
			}
			file.Close()
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// If the file was created, start watching it
			if event.Op&fsnotify.Create == fsnotify.Create && event.Name == filepath {
				if err := watcher.Add(filepath); err != nil {
					return fmt.Errorf("failed to watch newly created file %s: %w", filepath, err)
				}
			}

			// If the file was modified or created, read new content
			if event.Name == filepath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				newPos, err := readNewContent(filepath, lastPos,callback)
				if err != nil {
					// Just log the error and continue watching
					fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filepath, err)
					continue
				}

				lastPos = newPos
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %w", err)

		case <-ctx.Done():
			return nil
		}
	}
}

// Pipe monitors a file for changes and outputs new content to the provided
// writer, prefixing each line with the specified prefix.
//
// Parameters:
//   - ctx: Context for cancellation
//   - filepath: Path to the file to monitor
//   - prefix: Prefix to prepend to each output line
//   - output: Writer where output will be written (typically os.Stdout)
//
// The function continues until the context is cancelled or an error occurs.
// If the file doesn't exist initially, it will wait for it to be created.
func Pipe(ctx context.Context, filepath string, prefix string, output io.Writer) error {
	return Watch(ctx, filepath, func(line []byte) error {
		if prefix != "" {
			_, err := fmt.Fprint(output, prefix)
			if err != nil {
				return err
			}
		}
		_, err := output.Write(line)
		if err != nil {
			return err
		}
		if  prefix!="" && !bytes.HasSuffix(line, []byte("\n")) {
			_, err = fmt.Fprintln(output)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// readNewContent reads content from a file starting at lastPos position.
// It returns the new lines read, the new file position, and any error encountered.
func readNewContent(filepath string, lastPos int64, callback func(content []byte) error) (int64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return  lastPos, err
	}
	defer file.Close()

	// Get current file size
	fileInfo, err := file.Stat()
	if err != nil {
		return lastPos, err
	}

	// Handle file truncation - reset position to beginning
	if fileInfo.Size() < lastPos {
		lastPos = 0
	}

	// If no new content, return early
	if fileInfo.Size() == lastPos {
		return  lastPos, nil
	}

	// Seek to last read position
	_, err = file.Seek(lastPos, io.SeekStart)
	if err != nil {
		return lastPos, err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return lastPos, err
	}

	if err := callback(content); err != nil {
		return lastPos, err
	}

	return lastPos + int64(len(content)), nil
}

// fileExists checks if a file exists at the given path
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}
