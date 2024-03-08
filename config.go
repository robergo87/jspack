package main

import (
    "os"
    "bufio"
    "errors"
    "strings"
    "path/filepath"
)


type Config struct {
    source string
    basedir string
    filelist []string
}

func (self *Config) Load() error {
    file, err := os.Open(self.source)
    if err != nil {return err}
    defer file.Close()

    self.filelist = []string{}
    scanner := bufio.NewScanner(file)
     for scanner.Scan() {
        line := strings.Trim(scanner.Text(), " \t\n\r")
        if len(line) == 0 {continue}
        if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
            continue
        }
        self.filelist = append(self.filelist, line)
    }
    if len(self.filelist) == 0 {
        return errors.New("Empty file list")
    }
    return nil
}

func NewConfig(source string) Config {
    config := Config{}
    
    abssource, _ := filepath.Abs(source)
    config.source = abssource
    config.basedir = filepath.Dir(source)
    return config
}