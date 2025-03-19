/*
Package logs provides utilities for real-time log file monitoring and output formatting.

The package offers functionality to efficiently watch log files for changes and
output new content with customizable prefixes. It uses fsnotify for efficient
file system event monitoring instead of polling.

Key features:
- Real-time monitoring of log files using fsnotify
- Outputs only new content when files are modified
- Handles file creation, modification, and truncation
- Support for callback functions to process new lines
- Prefixes each output line with customizable text
- Context-based cancellation for graceful shutdown

Primary functions:
- Watch: Monitor a file and process new lines with a callback function
- Pipe: Monitor any file with a custom prefix (uses Watch internally)
- WatchLogTextFile: Convenient wrapper for watching log.txt with a standard prefix

Example usage with callback:

	// Watch a log file with custom processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logs.Watch(ctx, "/path/to/app.log", func(line string) error {
		// Process the line however you want
		if strings.Contains(line, "ERROR") {
			fmt.Printf("!!! ERROR: %s\n", line)
		}
		// Return an error to stop watching
		if strings.Contains(line, "FATAL") {
			return errors.New("fatal error detected")
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Watcher stopped: %v", err)
	}

Example usage with prefix:

	// Watch a log file with custom prefix
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logs.Pipe(ctx, "/path/to/app.log", "app: ", os.Stdout)
	if err != nil {
		log.Fatalf("Error watching log: %v", err)
	}

For more examples, see the example directory.
*/
package logs