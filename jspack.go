package main


import (
    "os"
    "fmt"
    "github.com/alexflint/go-arg"
)

var config Config
var bundle Bundle
var watcher FileWatch


func watch_callback(filename string, abspath string) bool{
    if abspath == bundle.destination {
        return true
    }
    if abspath == config.source {
        config.Load()
        bundle.ReplaceAll(config.filelist)
        watcher.ReplaceAll(config.filelist)
        watcher.Add(config.source)
        bundle.Render()
        return true
    }
    bundle.Reload(filename)
    bundle.Render()
    return true
}

func main() {
    var args struct {
    	Config string `default:"jspack.conf"`
    	Destination string `default:"dist.js"`
    	Watch bool
    }
    arg.MustParse(&args)
    fmt.Println("args", args.Config, args.Destination)
    
    config = NewConfig(args.Config)
    config.Load()
    
    err := os.Chdir(config.basedir)
    if err != nil {
        log_error(err)
        return
    }
    log_notice(fmt.Sprintf("Working directory: %s", config.basedir))
    
    bundle = NewBundle(args.Destination)
    bundle.ReplaceAll(config.filelist)
    bundle.Render()
    
    watcher, err = NewFileWatch()
    if err != nil {
        log_error(err)
        return
    }
    watcher.ReplaceAll(config.filelist)
    watcher.Add(config.source)
    watcher.Run(watch_callback)
}