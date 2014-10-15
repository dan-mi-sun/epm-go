package epm

import (
    "fmt"
    "bufio"
    "os"
    "strings"
    "regexp"
    "strconv"
    "bytes"
    "encoding/binary"
    "encoding/hex"
    "github.com/project-douglas/lllc-server"
)

// an EPM Job
type Job struct{
   cmd string
   args []string 
}

// EPM object. maintains list of jobs and a symbols table
type EPM struct{
    eth ChainInterface

    lllcURL string

    jobs []Job
    vars map[string]string

    pkgdef string // latest pkgdef we are parsing
    state map[string]map[string]string// latest ethstate
    
    log string
}

// new empty epm
func NewEPM(eth ChainInterface, log string) *EPM{
    lllcserver.URL = "http://lllc.erisindustries.com/compile"
    return &EPM{
        eth:  eth,
        jobs: []Job{},
        vars: make(map[string]string),
        log: ".epm-log",
        state: nil,
    }
}

// allowed commands
var CMDS = []string{"deploy", "modify-deploy", "transact", "query", "log", "set", "endow"}

// make sure command is valid
func checkCommand(cmd string) bool{
    r := false
    for _, c := range CMDS{
        if c == cmd{
            r = true
        }
    }
    return r
}

//TODO: use Trim!
func shaveWhitespace(t string) string{
    // shave whitespace from front
    for ; t[0:1] == " " || t[0:1] == "\t"; t = t[1:]{
    }
    // shave whitespace from back...
    l := len(t)
    for ; t[l-1:] == " "; t = t[:l-1]{
    }
    return t
}

// peel off the next command and its args
func peelCmd(lines *[]string, startLine int) (*Job, error){
    job := Job{"", []string{}}
    for line, t := range *lines{
        // ignore comments and blank lines
        //fmt.Println("next line:", line, t)
        if len(t) == 0 || t[0:1] == "#" {
            continue
        }
        // if no cmd yet
        if job.cmd == ""{
            // cmd syntax check
            l := len(t)
            if t[l-1:] != ":"{
                return nil, fmt.Errorf("Syntax error: missing ':' on line %d", line+startLine)
            }
            cmd := t[:l-1]
            // ensure known cmd
            if !checkCommand(cmd){
                return nil, fmt.Errorf("Invalid command '%s' on line %d", cmd, line+startLine)
            }
            job.cmd = cmd
            continue
        }
        // if the line does not begin with white space, we're done
        if !(t[0:1] == " " || t[0:1] == "\t"){
            // peel off lines we've read
            *lines = (*lines)[line:]
            return &job, nil 
        } 
        
        // the line is args. parse them
        // first, eliminate prefix whitespace/tabs
        // TODO: use Trim
        t = shaveWhitespace(t)

        args := strings.Split(t, "=>")
        // should be 'arg1 => arg2'
        if len(args) != 2 && len(args) != 3{
            return nil, fmt.Errorf("Syntax error: improper argument formatting on line %d", line+startLine)
        }
        for _, a := range args{
            shaven := shaveWhitespace(a)
            job.args = append(job.args, shaven)
        }
    }
    // only gets here if we finish all the lines
    *lines = nil
    return &job, nil
}

//parse should open a file, read all lines, peel commands into jobs
func (e *EPM) Parse(filename string) error{
    // set current file to parse
    e.pkgdef = filename

    // temp dir
    // TODO: move it!
    CheckMakeTmp()

    lines := []string{}
    f, err := os.Open(filename)
    if err != nil{
        return err
    }
    scanner := bufio.NewScanner(f)
    // read in all lines
    for scanner.Scan(){
        t := scanner.Text()
        lines = append(lines, t)
    }

    l := 0
    startLength := len(lines)
    for lines != nil{
        job, err := peelCmd(&lines, l)
        if err != nil{
            return err
        }
        e.jobs = append(e.jobs, *job)
        l = startLength - len(lines)
    }
    return nil
}

// replaces any {{varname}} args with the variable value
func (e *EPM) VarSub(args []string) []string{
    r, _ := regexp.Compile(`\{\{(.+?)\}\}`)
    for i, a := range args{
        // if its a known var, replace it
        // else, leave alone
        args[i] = r.ReplaceAllStringFunc(a, func(s string) string{
            k := s[2:len(s)-2] // shave the brackets
            v, ok := e.vars[k]
            if ok{
                return v
            } else{
                return s
            }
        })
    }
    return args
}

func (e *EPM) Vars() map[string]string{
    return e.vars
}

func (e *EPM) Jobs() []Job{
    return e.jobs
}

func (e *EPM) StoreVar(key, val string){
    fmt.Println("storing var:", key, val)
    if key[:2] == "{{" && key[len(key)-2:] == "}}"{
        key = key[2:len(key)-2]
    }
    e.vars[key] = Coerce2Hex(val)
    fmt.Println("stored result:", e.vars[key])
}

// keeps N bytes of the conversion
func NumberToBytes(num interface{}, N int) []byte {
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, num)
    if err != nil {
        fmt.Println("NumberToBytes failed:", err)
    }
    fmt.Println("btyes!", buf.Bytes())
    if buf.Len() > N{
        return buf.Bytes()[buf.Len()-N:]
    }
    return buf.Bytes()
}

// s can be string, hex, or int.
// returns properly formatted 32byte hex value
func Coerce2Hex(s string) string{
    fmt.Println("coercing to hex:", s)
    // is int?
    i, err := strconv.Atoi(s)
    if err == nil{
        return "0x"+hex.EncodeToString(NumberToBytes(int32(i), i/256+1))
    }
    // is already prefixed hex?
    if len(s) > 1 && s[:2] == "0x"{
        if len(s) % 2 == 0{
            return s
        }
        return "0x0"+s[2:]
    }
    // is unprefixed hex?
    if len(s) > 32{
        return "0x"+s
    }
    pad := s + strings.Repeat("\x00", (32-len(s)))
    ret := "0x"+hex.EncodeToString([]byte(pad))
    fmt.Println("result:", ret)
    return ret
}

func addHex(s string) string{
    if len(s) < 2{
        return "0x"+s
    }

    if s[:2] != "0x"{
        return "0x"+s
    }
    
    return s
}

func stripHex(s string) string{
    if len(s) > 1{
        if s[:2] == "0x"{
            return s[2:]
        }
    }
    return s
}

// split line and trim space
func parseLine(line string) []string{
    line = strings.TrimSpace(line)
    line = strings.TrimRight(line, ";")

    args := strings.Split(line, ";")
    for i, a := range args{
        args[i] = strings.TrimSpace(a)
    }
    return args
}
