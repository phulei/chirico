/*

Create a program to track cpu per pid over time, but you also
want it to do the following:

  - output in .csv format
  - load into sqlite3 database
  - provide custom scripts to execute

By automatically putting the data in a database you find
processes that use a lot of cpu during 2.5min, 5min or
10min intervals across the whole day.


*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	//	"strings"
	//	"time"
)

/*
   -- STANDARD HEADING --
*/
var VERSION string = "0.0.1(alpha)"
var AUTHORS string = "mchirico@gmail.com"
var SRC string = "mmc/src/go/proc.go"

var debug bool
var version bool
var DEBUG bool = false

var interval int = 0
var counts int64 = 0
var buildScripts bool

var help_msg = `
The following files are created with the program:
_loadproc.sql  proc  _proc.awk  proc.csv  proc.db  proc.go  proc.log  _procReport.sql  proc.sh


Here are some common commands:


./proc -i 3  -c 10

nohup ./proc -i 60 -c 1000  >/dev/null 2>&1 &

nohup ./proc -i 60 -c 10000000000000000  >/dev/null 2>&1 &  # You can use very large counts. Check disk space.

./proc -b
./proc.sh

sqlite3 proc.db < _procReports.sql



` // End of help

var Usage = func() {
	fmt.Fprintf(os.Stderr, "%s,%s,%s\n", VERSION, SRC, AUTHORS)
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	//Remove this
	fmt.Fprintf(os.Stderr, help_msg)
}

func versionTxt() {
	s := `

`
	if version {
		fmt.Fprintf(os.Stderr, "VERSION: %s, AUTHORS: %s, SRC: %s\n%s\n", VERSION, AUTHORS, SRC, s)
		os.Exit(0)
	}

}
func init() {
	const (
		defaultDebug   = false
		debugText      = " turns on/off debugging. Normally just use -d with not arguments to turn on."
		defaultVersion = false
		versionText    = " version information.  -v to show the version and exit. "
	)

	flag.IntVar(&interval, "i", 1, "intervals (secs) Default is 1 second.")
	flag.Int64Var(&counts, "c", 5, "counts num of times to sample. Default is 5")
	flag.BoolVar(&buildScripts, "b", false, " Build script: proc.sh ")

	flag.BoolVar(&debug, "d", false, " This is for debugging. Writes out raw logs")
	flag.BoolVar(&version, "v", false, versionText+" ")

	flag.Usage = Usage

}

/*
   -- END STANDARD HEADING --
*/

func setupSignal(d chan int) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("\nCaptured signal %v\n", sig)
			fmt.Printf("Output in %v\n", "proc.log")
			os.Exit(1) // Will exit immediately.
			d <- 0
			os.Exit(1)
		}
	}()

}

var MyDebug = false

var re = regexp.MustCompile("(?P<server>[a-zA-Z]+72[3-7])\\.(?P<key>key)\\.(?P<date>[0-9]+)\\.(?P<time>[0-9]+)")
var rep = regexp.MustCompile("(?P<server>[a-zA-Z]+72[3-7])\\.(?P<pid>pid)\\.(?P<pid>[0-9]+)\\.(?P<date>[0-9]+)\\.(?P<time>[0-9]+)")

func Open(name string) (file *os.File, err error) {
	file, err = os.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	return file, err
}

func Read(fileFrom *os.File) string {
	var byteLen = 1024
	var s string
	buf := make([]byte, byteLen)
	for {
		n, err := fileFrom.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		if n > byteLen {
			log.Printf("Problem in ReadPid. n > bytelen n=%d bytelen=%d\n", n, byteLen)
			panic(err)

		}
		s = s + string(buf[:(n-1)])

	}
	//fileFrom.Close()
	return s
}

var fcsv *os.File

func CsvIze(data string) {
	const layout = "2006-01-02 03:04:06"
	t := time.Now().Local()
	s := fmt.Sprintf("%s,%s\n", t.Format(layout), strings.Replace(data, " ", ",", -1))

	n, err := fcsv.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to proc.csv. n=%v", n)
		return
	}
	fcsv.Sync()

}

/* ProcRead == Process Read
 */
var log2 *log.Logger

func ProcRead(file string) {

	f, err := Open(file)
	if err != nil {
		panic(err)
	}
	data := Read(f)
	CsvIze(data)
	if debug {
		log.Printf("%v\n", data)
	}
	log2.Printf("%v\n", data)
	f.Close()
}

/*
   sleep: time in seconds
   iterations: loops
*/
func GetStats(sleep int, iterations int64) {
	matches, err := filepath.Glob("/proc/[0-9]*/stat")
	if err != nil {
		panic(err)
	}
	for j := int64(0); j < iterations; j++ {
		for i := range matches {
			if 1 == 0 {
				fmt.Printf("matches[%v]=\t%v\n", i, matches[i])
			}
			ProcRead(matches[i])
		}
		time.Sleep(time.Duration(sleep) * time.Second)
		//time.Sleep(1000 * time.Millisecond)
	}
}

func CreateSupportScripts() {
	f, err := os.OpenFile("proc.sh", os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		fmt.Println("Error opening proc.sh")
	}
	defer f.Close()
	s := `#!/bin/bash
#
echo -e "
BEGIN{FS=\",\"}
{c=NF;}
END {
     printf(\"create table ps (d datetime,p int,v varchar(10),s varchar(1)\");
     for (i=4; i <c; i++)
     {
        printf(\",n%i int\",i)
     }
     print \");\"
}
"> _proc.awk

awk -f _proc.awk proc.csv >  _loadproc.sql

echo -e "
.separator \",\"
.import proc.csv ps
">>_loadproc.sql
rm -f ./proc.db
sqlite3 proc.db < _loadproc.sql

echo -e "

select 'Version 0.0.0(alpha)';
select 'mchirico@gmail.com';
select 'Generated from ./proc.go -b  && ./prog.sh ';
select '';


select '';
select '----------------------------------------------------------------------------------------';
select '----------------------------------------------------------------------------------------';
select 'Summary each 250(s) interval';
select '';

select (strftime(\"%s\",d)/250)*250,max(d),min(d),p,v,max(n14)-min(n14),max(n15)-min(n15),max(n16)-min(n16),max(n17)-min(n17)
from ps 
group by (strftime(\"%s\",d)/250)*250,p,v
having   (max(n15)-min(n15))+ (max(n15)-min(n15))      > 2
order by (strftime(\"%s\",d)/250)*250,max(n14)-min(n14) desc,max(n15)-min(n15);


select '';
select '----------------------------------------------------------------------------------------';
select '----------------------------------------------------------------------------------------';
select 'Summary each 300(s) interval';
select '';



select (strftime(\"%s\",d)/300)*300,max(d),min(d),p,v,max(n14)-min(n14),max(n15)-min(n15),max(n16)-min(n16),max(n17)-min(n17)
from ps 
group by (strftime(\"%s\",d)/300)*300,p,v
having   (max(n15)-min(n15))+ (max(n15)-min(n15))      > 20
order by (strftime(\"%s\",d)/300)*300,max(n14)-min(n14) desc,max(n15)-min(n15);




select '';
select '----------------------------------------------------------------------------------------';
select '----------------------------------------------------------------------------------------';
select 'Summary each 10 minute interval';
select '';

select (strftime(\"%s\",d)/600)*600,max(d),min(d),p,v,max(n14)-min(n14),max(n15)-min(n15),max(n16)-min(n16),max(n17)-min(n17)
from ps 
group by (strftime(\"%s\",d)/600)*600,p,v
having   (max(n15)-min(n15))+ (max(n15)-min(n15))      > 5
order by (strftime(\"%s\",d)/600)*600,max(n14)-min(n14) desc,max(n15)-min(n15);



-- select p,v, max(n14) utime,max(n15) stime,max(n16) cutime,max(n17) cstime from ps group by p order by max(n14) desc limit 30;

" > _procReport.sql
sqlite3 proc.db < _procReport.sql

`

	n, err := f.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to proc.sh n=%v", n)
		return
	}
	fcsv.Sync()

	fmt.Printf("\n\nNow just run the following:\n\n./proc.sh")
	fmt.Printf("\nsqlite3 proc.db \"select count(*) from ps;\"\n\n")

}

func main() {

	flag.Parse()
	versionTxt()
	sigChan := make(chan int, 0)
	setupSignal(sigChan)

	if buildScripts {
		CreateSupportScripts()
		os.Exit(1)
	}

	logf, err := os.OpenFile("proc.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer logf.Close()
	log.SetOutput(logf)

	log2 = log.New(os.Stdout, "", log.LstdFlags)

	fcsv, err = os.OpenFile("proc.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening proc.csv")
	}
	defer fcsv.Close()

	GetStats(interval, counts)

}
