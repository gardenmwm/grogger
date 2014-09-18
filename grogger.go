package main

import (
        "fmt"
        "github.com/blakesmith/go-grok"
        "github.com/ActiveState/tail"
        "encoding/json"
        "sync"
)



//Channels


func main() {
    logtext := make(chan string)
    var wg sync.WaitGroup
    wg.Add(2)
    fmt.Println("Calling taillog")
    go taillog("/tmp/test.txt",logtext, &wg)
    fmt.Println("Calling parseLogLine")
    go parseLogLine(logtext, &wg)
    wg.Wait()
}

func taillog(file string, c chan string, wg *sync.WaitGroup){
    fmt.Println("Enterting Taillog")
    t, err := tail.TailFile(file, tail.Config{
        Follow: true,
        ReOpen: true})
        for line := range t.Lines {
            c <- line.Text
            fmt.Println(line.Text)
        }
    if err != nil {
        fmt.Println("error tailing file")
    }
    wg.Done()
}

func convertToJSON(matches *grok.Match) string {
    str, err := json.Marshal(matches.Captures())
    if err != nil {
        fmt.Println("Error encoding json: ", err)
    }
    return string(str)
}

func parseLogLine(c chan string, wg *sync.WaitGroup) {
    fmt.Println("entering parseLogLine")
    g := grok.New()
    g.AddPatternsFromFile("/tmp/base")
    pattern := "%{WORD}"
    err := g.Compile(pattern)
    if err != nil {
        fmt.Println("Error Compiling: ",err)
    }
    for {
        line := <-c
        jsoncapture := convertToJSON(g.Match(line))
        fmt.Println(jsoncapture)
    }
    wg.Done()
}
