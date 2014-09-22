package main

import (
        "fmt"
        "flag"
        "github.com/blakesmith/go-grok"
        "github.com/ActiveState/tail"
        "encoding/json"
        "time"
        "sync"
        "gopkg.in/redis.v2"
        "code.google.com/p/gcfg"
        "os"
        "strings"
        //"reflect"
        )

var server = flag.String("server", "lnx-logstash:6900", "Server:Port for Redis Server")
var conffile = flag.String("config", "./grogger.ini", "Path to Config file")
var patternfile = flag.String("patternfile", "/tmp/base", "Path to Paterns file")

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
    flag.Parse()
    cfg := getfiles()
    jsonchan := GetJSONChannel()
    var wg sync.WaitGroup
    for k,v := range cfg.File {
        fmt.Println("Added: ",k,v.Path,v.Pattern)
        wg.Add(1)
        go MonitorLog(v.Path,v.Pattern,jsonchan)
    }
    go sendToRedis("lnx-logstash:6900", jsonchan, &wg)
    wg.Wait()
}

func taillog(file string, c chan logentry, wg *sync.WaitGroup){
    endfile := tail.SeekInfo{
        Offset: 0,
        Whence: 2,
        }
    t, err := tail.TailFile(file, tail.Config{
        Follow: true,
        ReOpen: true,
        Location: &endfile,
        })
        for line := range t.Lines {
            logline := logentry{}
            logline.logtext  = line.Text
            logline.logtime = time.Now().Format(time.RFC822Z)
            c <- logline
        }
    if err != nil {
        fmt.Println("error tailing file: ", err)
    }
    wg.Done()
}

func convertToJSON(jsondata FullLogEntry) string {
    //Get rid of everything before : in field list since that has the grok pattern name
    newjson := make(map[string][]string)
    for k,v := range(jsondata.fields) {
        keysplit := strings.Split(k,":")
        newkey := keysplit[len(keysplit)-1]
        newjson[newkey]= v
    }
    //Convert to a single map, this ugly because of a late design change
    newjson["logsource"] = []string{jsondata.hostname}
    newjson["groggertime"] = []string{jsondata.timestamp}
    //Convert to JSON
    j,err := json.Marshal(newjson)
    if err != nil {
        fmt.Println("test")
        }
    return string(j)
}

func parseLogLine(c chan logentry, jc chan string, pattern string, wg *sync.WaitGroup) {
    hostname, herr := os.Hostname()
    if herr != nil {
        fmt.Println("Getting hostname failed, wtf")
        }
    g := grok.New()
    g.AddPatternsFromFile(*patternfile)
    err := g.Compile(pattern)
    if err != nil {
        fmt.Println("Error Compiling: ",err)
    }
    for {
        logline := <-c
        logdata := FullLogEntry{}
        logdata.hostname = hostname
        logdata.timestamp = logline.logtime
        logdata.fields = grokline(*g,logline.logtext)
        if len(logdata.fields) > 0 {
            jsoncapture := convertToJSON(logdata)
            jc <- jsoncapture
            fmt.Println("parseLogLine_jsoncapture: ",jsoncapture)
        }
    }
    wg.Done()
}

func grokline(g grok.Grok, line string) map[string][]string {
    defer func() {
        if e:= recover(); e!= nil {
            fmt.Println("Error in Groking, pattern probably has no matches",e)
        }
    }()
    match := g.Match(line)
    fields := match.Captures()
    return fields
}

func sendToRedis(server string ,c chan string, wg *sync.WaitGroup){
    client := redis.NewTCPClient(&redis.Options {
        Addr:   server,
        Password: "",
        DB: 0,
    })
    for {
        d := <-c
        if err := client.RPush("grogger", d).Err(); err != nil{
            fmt.Println("RedisError: ",err)
        }
    }
    wg.Done()
}

type Config struct {
        File map[string]*struct {
            Path string
            Pattern string
        }
    }

func getfiles() Config {
    cfg := Config{}
    err := gcfg.ReadFileInto(&cfg, *conffile)
    if err != nil {
        fmt.Println("Config Error: ",err)
    }

    return cfg
}

