package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// debounceWindow is the quiet period that must elapse after the last write
// before we surface a single FileChangedMsg. bd rewrites several Dolt files per
// mutation, so without coalescing a single `bd close` would fan out into a
// burst of refreshes.
const debounceWindow = 300 * time.Millisecond

// FileChangedMsg signals that the watched beads directory changed and the model
// should refetch. It carries no payload: the watcher only knows "something
// moved," the model owns deciding what to re-query.
type FileChangedMsg struct{}

// WatchBeadsCmd watches dir for write/create/remove events and emits exactly
// one FileChangedMsg per quiet window (~300ms debounce). It is a one-shot
// tea.Cmd: after emitting, the caller re-issues it. If dir is "" or cannot be
// watched, it returns a command that blocks without emitting (no crash) so the
// program never spins or panics on an unwatchable workspace.
func WatchBeadsCmd(dir string) tea.Cmd {
	if dir == "" {
		return blockForever
	}
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			// Can't watch: park this command so Init/re-issue never busy-loops.
			return blockForever()
		}
		defer watcher.Close()

		if err := watcher.Add(dir); err != nil {
			return blockForever()
		}

		// timer is nil until the first relevant event arrives; we lazily start
		// it so an idle watcher never fires.
		var timer *time.Timer
		var fired <-chan time.Time
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return blockForever()
				}
				// Only writes, creates, and removes signal a content change;
				// chmod-only events (e.g. lock churn) are ignored.
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) == 0 {
					continue
				}
				if timer == nil {
					timer = time.NewTimer(debounceWindow)
					fired = timer.C
				} else {
					// Drain a possibly-already-expired timer before reset.
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(debounceWindow)
				}
			case err, ok := <-watcher.Errors:
				if !ok || err != nil {
					return blockForever()
				}
			case <-fired:
				return FileChangedMsg{}
			}
		}
	}
}

// blockForever is a tea.Msg-producing function that never returns. Used as the
// degraded watcher: the program keeps running, just without live refresh.
func blockForever() tea.Msg {
	select {}
}
