grogger
=======

Logwatcher that parses logs with grok before sending to Redis

I was running into an issue with syslog and active4j not playing nice, and I wanted something more elegant than tail -F, so I wrote grogger to follow any file and then parse it with grok, then converts to the logstash format and dumps it in redis.


## Install ##
Dependencies:
:   Grok
:   Go-Grok
:   ActiveState/Tail
:   PCRE <= 7.8

go get github.com/gardenmwm/grogger 


Notes for OSX Dev:
    If using brew modify both the grok recipe and the prce recipe, the grok recipe needs to be only a make, so the library is created, and pcre needs to be for version 7.8.

## Config ##
The ini file is a simple format, of which you can add as many files as you want. Note that if the grok pattern uses \ then they must be escaped like so \\
```
[file "<filename>"]
path = <path to file>
pattern = <grok pattern>
```

## Usage ##
grogger -server <Redis Server> -config <config file>

## Todo ##
* Add error handling
* Add daemon mode
* Clean up code


