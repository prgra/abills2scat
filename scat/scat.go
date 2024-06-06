package scat

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/ssh"
)

// App is the main application struct
type App struct {
	Abills *Abills
	Nases  []*Nas
}

type Abills struct {
	DB *sqlx.DB
}

type Nas struct {
	Host string
	SSH  *ssh.Client
}

func NewApp(conf Config) (*App, error) {
	abills, err := sqlx.Connect("mysql", conf.AbillsDB)
	if err != nil {
		fmt.Println("db error", err)
		return nil, err
	}
	if conf.AbillsBDNames != "" {
		_, nerr := abills.Exec(fmt.Sprintf("SET NAMES %s", conf.AbillsBDNames))
		if nerr != nil {
			fmt.Println("set names error", nerr)
		}
	}

	nases := make([]*Nas, len(conf.Nases))
	for i, n := range conf.Nases {
		keyfile := conf.NasKeyFile
		if n.Key != "" {
			keyfile = n.Key
		}
		key, err := os.ReadFile(filepath.Clean(keyfile))
		if err != nil {
			return nil, err
		}
		ns, err := NewNas(n.Host, n.User, key)
		if err != nil {
			return nil, err
		}
		nases[i] = ns
	}
	return &App{
		Abills: &Abills{
			DB: abills,
		},
		Nases: nases,
	}, nil
}

func NewNas(host, user string, key []byte) (*Nas, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("parse key error", err)
		return nil, err
	}
	conf := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	conn, cerr := ssh.Dial("tcp", host, conf)
	if cerr != nil {
		return nil, cerr
	}
	return &Nas{
		Host: host,
		SSH:  conn,
	}, nil
}

type ShortScatUser struct {
	ID string
	IP string
}

var uregexp = regexp.MustCompile(`(\w+):(((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4})`) // user:ip

// GetUserList returns a list of users from the NAS
func (n *Nas) GetUserList() (users []ShortScatUser, err error) {
	s, err := n.SSH.NewSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	sshOut, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}

	sshErr, err := s.StderrPipe()
	if err != nil {
		return nil, err
	}
	err = s.Run("fdpi_ctrl list all --bind_multi")
	if err != nil {
		return nil, err
	}
	be, _ := io.ReadAll(sshErr)
	if len(be) > 0 {
		fmt.Println("ParseErr GetUserList", string(be))
	}
	scanner := bufio.NewScanner(sshOut)
	for scanner.Scan() {
		if uregexp.MatchString(scanner.Text()) {
			m := uregexp.FindStringSubmatch(scanner.Text())
			users = append(users, ShortScatUser{ID: m[1], IP: m[2]})
		}
	}
	return users, nil
}

type ScatUser struct {
	ID     string
	IP     string
	TPID   string
	TPName string
	KV     map[string]string
}

// GetUserProfilesList fdpi_ctrl list all --policing показывает список пользователей и их скорости
func (n *Nas) GetUserProfilesList() (users []ScatUser, err error) {
	s, serr := n.SSH.NewSession()
	if serr != nil {
		return nil, serr
	}
	defer s.Close()

	sshOut, _ := s.StdoutPipe()
	sshErr, _ := s.StderrPipe()
	err = s.Run("fdpi_ctrl list all --policing")
	if err != nil {
		return nil, err
	}
	be, _ := io.ReadAll(sshErr)
	if len(be) > 0 {
		fmt.Println("getuserprofilelist", "ParseErr", string(be))
	}
	scanner := bufio.NewScanner(sshOut)
	for scanner.Scan() {
		var u ScatUser
		a := scanner.Text()
		strs := strings.Split(a, "\t")
		if len(strs) < 2 {
			continue
		}
		u.ID = strs[0]
		u.TPID = strs[1]
		u.TPName = strs[len(strs)-2]
		kv := make(map[string]string)
		for i := range strs {
			if strings.Contains(strs[i], "=") {
				kvs := strings.Split(strs[i], "=")
				if len(kvs) != 2 {
					continue
				}
				kv[kvs[0]] = kvs[1]
			}
			u.KV = kv
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserCGNat fdpi_ctrl list all --policing показывает список пользователей и их скорости
func (n *Nas) GetUserCGNat() (users map[string]bool, err error) {
	s, serr := n.SSH.NewSession()
	if serr != nil {
		return nil, serr
	}
	defer s.Close()
	users = make(map[string]bool)
	sshOut, _ := s.StdoutPipe()
	sshErr, _ := s.StderrPipe()
	err = s.Run("fdpi_ctrl list all --service 11")
	if err != nil {
		return nil, err
	}
	be, _ := io.ReadAll(sshErr)
	if len(be) > 0 {
		fmt.Println("GetUserCGNat", "ParseErr", string(be))
	}
	scanner := bufio.NewScanner(sshOut)
	for scanner.Scan() {
		a := scanner.Text()
		strs := strings.Split(a, "\t")
		if len(strs) < 2 {
			continue
		}
		users[strs[0]] = true

	}
	return users, nil
}

func (a *App) AddTariff(in, out int) {

}

// user3	HTB	dnlnk_rate=0.00mbit	dnlnk_ceil=0.00mbit rrate=2500000(20.00mbit)	rburst=1250000(10.00mbit)	rceil=2500000(20.00mbit)	rcburst=1250000(10.00mbit)	rate0=0.00mbit	ceil0=20.00mbit	rate1=0.00mbit	ceil1=20.00mbit	rate2=0.00mbit	ceil2=20.00mbit	rate3=0.00mbit	ceil3=20.00mbit	rate4=0.00mbit	ceil4=20.00mbit	rate5=0.00mbit	ceil5=20.00mbit	rate6=0.00mbit	ceil6=20.00mbit	rate7=0.00mbit	ceil7=20.00mbit	HTB_INBOUND	rrate=2500000(20.00mbit)	rburst=1250000(10.00mbit)	rceil=2500000(20.00mbit)	rcburst=1250000(10.00mbit)	rate0=0.00mbit	ceil0=20.00mbit	rate1=0.00mbit	ceil1=20.00mbit	rate2=0.00mbit	ceil2=20.00mbit	rate3=0.00mbit	ceil3=20.00mbit	rate4=0.00mbit	ceil4=20.00mbit	rate5=0.00mbit	ceil5=20.00mbit	rate6=0.00mbit	ceil6=20.00mbit	rate7=0.00mbit	ceil7=20.00mbit	название_тарифного_плана

func (n *Nas) Run(cmd string) (string, error) {
	log.Println("Running", cmd)
	s, err := n.SSH.NewSession()
	if err != nil {
		return "", err
	}
	defer s.Close()
	// modes := ssh.TerminalModes{
	// 	ssh.ECHO:          0,     // disable echoing
	// 	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	// 	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	// }
	// err = s.RequestPty("linux", 80, 40, modes)
	// if err != nil {
	// 	return "", err
	// }

	sshOut, err := s.StdoutPipe()
	if err != nil {
		return "", err
	}
	sshErr, err := s.StderrPipe()
	if err != nil {
		return "", err
	}
	err = s.Run(cmd)
	if err != nil {
		return "", err
	}
	b, _ := io.ReadAll(sshErr)
	if len(b) > 0 {
		fmt.Printf("RunErr cmd: %s\n%s", cmd, string(b))
	}
	b, _ = io.ReadAll(sshOut)
	log.Println("status", string(b))

	return string(b), nil
}
