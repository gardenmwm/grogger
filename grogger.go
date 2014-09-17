package main

import (
        "fmt"
        "github.com/blakesmith/go-grok"
        "github.com/ActiveState/tail"
)



//Channels

func main() {
    g := grok.New()
//    g.AddPatternsFromFile("/tmp/base")
    g.AddPattern("WORD", ".*")
    pattern := "%{WORD}"
    err := g.Compile(pattern)
    if err != nil {
        fmt.Println("Error:",err)
    }

    t, err := tail.TailFile("/tmp/test.txt", tail.Config{
        Follow: true,
        ReOpen: true})
    for line := range t.Lines {
        fmt.Println(line.Text)
    }
    if err != nil {
    fmt.Println("all goo")}
}


