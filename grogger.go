package main

import (
        "fmt"
        "github.com/blakesmith/go-grok"
        "github.com/ActiveState/tail"
        "encoding/json"
        "time"
        "sync"
        "gopkg.in/redis.v2"
        )

type logentry struct {
    logtext string
    logtime string
}

type JSONLogEntry struct {
    Host string
    Timestamp string
    Fields map[string][]string
}

type FullLogEntry struct {
    hostname string
    timestamp string
    fields map[string][]string
}

func GetChannel() chan logentry {
    newchan := make(chan logentry)
    return newchan
}

func GetJSONChannel() chan string {
    newchan := make(chan string)
    return newchan
}

func MonitorLog(logfile string, pattern string, jsonchan chan string){
    logchan := GetChannel()
    var wg sync.WaitGroup
    wg.Add(2)
    go taillog(logfile, logchan, &wg)
    go parseLogLine(logchan, jsonchan, pattern, &wg)
    wg.Wait()
}

func main() {
    jsonchan := GetJSONChannel()
    var wg sync.WaitGroup
    wg.Add(1)
    go MonitorLog("/tmp/test.txt","%{WORD}",jsonchan)
    go sendToRedis("lnx-logstash:6900", jsonchan, &wg)
    wg.Wait()
}

func taillog(file string, c chan logentry, wg *sync.WaitGroup){
    t, err := tail.TailFile(file, tail.Config{
        Follow: true,
        ReOpen: true})
        for line := range t.Lines {
            logline := logentry{}
            logline.logtext  = line.Text
            logline.logtime = time.Now().Format(time.RFC850)
            c <- logline
        }
    if err != nil {
        fmt.Println("error tailing file: ", err)
    }
    wg.Done()
}

func convertToJSON(jsondata FullLogEntry) string {
    l := JSONLogEntry {
        Host: jsondata.hostname,
        Timestamp: jsondata.timestamp,
        Fields: jsondata.fields,
        }
    j,err := json.Marshal(l)
    if err != nil {
        fmt.Println("test")
        }
    return string(j)
}

func parseLogLine(c chan logentry, jc chan string, pattern string, wg *sync.WaitGroup) {
    g := grok.New()
    g.AddPatternsFromFile("/tmp/base")
    err := g.Compile(pattern)
    if err != nil {
        fmt.Println("Error Compiling: ",err)
    }
    for {
        logline := <-c
        logdata := FullLogEntry{}
        logdata.hostname = "test"
        logdata.timestamp = logline.logtime
        logdata.fields= g.Match(logline.logtext).Captures()
        jsoncapture := convertToJSON(logdata)
        jc <- jsoncapture
        fmt.Println("parseLogLine_jsoncapture: ",jsoncapture)
    }
    wg.Done()
}

func sendToRedis(server string ,c chan string, wg *sync.WaitGroup){
    client := redis.NewTCPClient(&redis.Options {
        Addr:   "lnx-logstash:6900",
        Password: "",
        DB: 0,
    })
    for {
        d := <-c
        fmt.Println("Sending data to redis: ",d)
        if err := client.Append("grogger", d).Err(); err != nil{
            fmt.Println("RedisError: ",err)
        }
    }
    wg.Done()
}
