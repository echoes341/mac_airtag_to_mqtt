package main

import (
	"fmt"
	"path"

	"github.com/fsnotify/fsnotify"
)

func watcher(cfg Config) (chan struct{}, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	changed := make(chan struct{})
	go func() {
		defer close(changed)
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if debug {
					fmt.Println("watchfile event:", event)
				}
				if event.Has(fsnotify.Write) && event.Name == cfg.AirtagsDataFile {
					select {
					// non-blocking send
					case changed <- struct{}{}:
					default:
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("watchfile error:", err)
			}
		}
	}()

	if err := watcher.Add(path.Dir(cfg.AirtagsDataFile)); err != nil {
		return nil, err
	}
	return changed, nil
}
