package main

import (
    "errors"
    "slices"
    "fmt"
    "path/filepath"
    "github.com/fsnotify/fsnotify"
)

type FileWatch struct {
    watcher *fsnotify.Watcher
    files map[string]string
    dirs map[string]int
}

func (self *FileWatch) Open() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {return err}
    self.watcher = watcher
    return err
}
func NewFileWatch() (FileWatch, error) {
    watcher := FileWatch{}
    err := watcher.Open()
    watcher.files = make(map[string]string)
    watcher.dirs = make(map[string]int)
    return watcher, err
}

func (self *FileWatch) Close() {
    self.watcher.Close()
}

func (self *FileWatch) Add(filename string) error {
    file, err := filepath.Abs(filename)
    if err != nil {return err}
    filedir := filepath.Dir(file)
    
    _, found := self.files[file]
    if found {return errors.New(fmt.Sprintf("File %s already tracked", filename))}
    
    counter, found := self.dirs[filedir]
    if !found {
        counter = 0
        err = self.watcher.Add(filedir)
    }
    if err != nil {return err}
    self.dirs[filedir] = counter+1
    self.files[file] = filename
    return nil
}

func (self *FileWatch) Delete(filename string) error {
    file, err := filepath.Abs(filename)
    if err != nil {return err}
    filedir := filepath.Dir(file)

    _, found := self.files[file]
    if found {return errors.New(fmt.Sprintf("File %s is not tracked", filename))}

    counter, found := self.dirs[filedir]
    if !found {return errors.New(fmt.Sprintf("Directory %s is not tracked", filedir))}
    
    if counter == 1 {
        err = self.watcher.Remove(filedir)
        if err != nil {return err}
        delete(self.dirs, filedir)
    } else {
        self.dirs[filedir] = counter-1
    }
    delete(self.files, file)
    return nil
}

func (self *FileWatch) ReplaceAll(filenames []string) error {
    newfiles := make(map[string]string)
    for _, newfilename := range(filenames) {
        newfile, err := filepath.Abs(newfilename)
        if err != nil {return err}
        _, found := self.files[newfile]
        if !found {
            err = self.Add(newfilename)
            if err != nil {return err}
        }
        newfiles[newfile] = newfilename
    }
    for oldfile, oldfilename := range(self.files) {
        _, found := newfiles[oldfile]
        if found {continue}
        err := self.Delete(oldfilename)
        if err != nil {return err}
    }
    return nil
}

func (self *FileWatch) Run(callback func(string, string) bool) error {
    tracked := []string{"CREATE", "WRITE","RENAME", "REMOVE"}
    for {
        select {
            case err, ok := <-self.watcher.Errors:
                if !ok {return err}
            case event, ok := <-self.watcher.Events:
                if !ok {return errors.New("Watcher error 1")}
                
                relpath, found := self.files[event.Name]
                if !found {continue}

                op := event.Op.String()
                if slices.Index(tracked, op) < 0 {continue}
                if !callback(relpath, event.Name) {break}
        }
    }  
    return nil
}
