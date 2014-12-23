package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
    "os/user"
    "log"
    "io"
    "io/ioutil"
	"path"
	"strconv"
	"strings"
    "encoding/json"
    "github.com/eris-ltd/thelonious/monklog"
)

var (
	GoPath  = os.Getenv("GOPATH")
	ErisLtd = path.Join(GoPath, "src", "github.com", "eris-ltd")

	usr, _      = user.Current() // error?!
	Decerver    = path.Join(usr.HomeDir, ".decerver")
	Apps        = path.Join(Decerver, "dapps")
	Blockchains = path.Join(Decerver, "blockchains")
	Filesystems = path.Join(Decerver, "filesystems")
	Logs        = path.Join(Decerver, "logs")
	Modules     = path.Join(Decerver, "modules")
	Scratch     = path.Join(Decerver, "scratch")
	HEAD        = path.Join(Blockchains, "HEAD")
	Refs        = path.Join(Blockchains, "refs")
	Epm         = path.Join(Scratch, "epm")
	Lllc        = path.Join(Scratch, "lllc")
)

func exit(err error) {
	status := 0
	if err != nil {
		fmt.Println(err)
		status = 1
	}
	os.Exit(status)
}

func AbsolutePath(Datadir string, filename string) string {
	if path.IsAbs(filename) {
		return filename
	}
	return path.Join(Datadir, filename)
}


// common golang, really?
func Copy(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil
}

func InitDataDir(Datadir string) error {
	_, err := os.Stat(Datadir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(Datadir, 0777)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func WriteJson(config interface{}, config_file string) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config_file, out.Bytes(), 0600)
	return err
}

// keeps N bytes of the conversion
func NumberToBytes(num interface{}, N int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
        // TODO: get this guy a return error?
		// logger.Errorln("NumberToBytes failed:", err)
	}
	//fmt.Println("btyes!", buf.Bytes())
	if buf.Len() > N {
		return buf.Bytes()[buf.Len()-N:]
	}
	return buf.Bytes()
}

// s can be string, hex, or int.
// returns properly formatted 32byte hex value
func Coerce2Hex(s string) string {
	//fmt.Println("coercing to hex:", s)
	// is int?
	i, err := strconv.Atoi(s)
	if err == nil {
		return "0x" + hex.EncodeToString(NumberToBytes(int32(i), i/256+1))
	}
	// is already prefixed hex?
	if len(s) > 1 && s[:2] == "0x" {
		if len(s)%2 == 0 {
			return s
		}
		return "0x0" + s[2:]
	}
	// is unprefixed hex?
	if len(s) > 32 {
		return "0x" + s
	}
	pad := strings.Repeat("\x00", (32-len(s))) + s
	ret := "0x" + hex.EncodeToString([]byte(pad))
	//fmt.Println("result:", ret)
	return ret
}

func AddHex(s string) string {
	if len(s) < 2 {
		return "0x" + s
	}

	if s[:2] != "0x" {
		return "0x" + s
	}

	return s
}

func StripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			return s[2:]
		}
	}
	return s
}

func Usr() string {
	u, _ := user.Current()
	return u.HomeDir
}

var MajorDirs = []string{
	Decerver, Apps, Blockchains, Filesystems, Logs, Modules, Scratch, Refs, Epm, Lllc,
}

func InitLogging(Datadir string, LogFile string, LogLevel int, DebugFile string) {
	if !monklog.IsNil() {
		return
	}
	var writer io.Writer
	if LogFile == "" {
		writer = os.Stdout
	} else {
		writer = openLogFile(Datadir, LogFile)
	}
	monklog.AddLogSystem(monklog.NewStdLogSystem(writer, log.LstdFlags, monklog.LogLevel(LogLevel)))
	if DebugFile != "" {
		writer = openLogFile(Datadir, DebugFile)
		monklog.AddLogSystem(monklog.NewStdLogSystem(writer, log.LstdFlags, monklog.DebugLevel))
	}
}

func openLogFile(Datadir string, filename string) *os.File {
	path := AbsolutePath(Datadir, filename)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening log file '%s': %v", filename, err))
	}
	return file
}

func InitDecerverDir() error {
	for _, d := range MajorDirs {
		err := InitDataDir(d)
		if err != nil {
			return err
		}
	}
	err := InitDataDir(Refs)
	if err != nil {
		return err
	}
	if _, err = os.Stat(HEAD); err != nil {
		_, err = os.Create(HEAD)
	}
	return err
}

