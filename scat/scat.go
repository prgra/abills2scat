package scat

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/ssh"
)

// App is the main application struct
type App struct {
	Abills *Abills
	Nases  []Nas
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
	nases := make([]Nas, len(conf.Nases))

	for i, n := range conf.Nases {
		keyfile := conf.NasKeyFile
		if n.Key != "" {
			keyfile = n.Key
		}
		key, err := os.ReadFile(keyfile)
		if err != nil {
			return nil, err
		}
		ns, err := NewNas(n.Host, n.User, key)
		if err != nil {
			return nil, err
		}
		nases[i] = *ns
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

type ScatUser struct {
	ID string
	IP string
}

var uregexp = regexp.MustCompile(`(\w+):(((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4})`) // user:ip

// GetUserList returns a list of users from the NAS
func (n *Nas) GetUserList() (users []ScatUser, err error) {
	session, err := n.SSH.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := session.RequestPty("linux", 80, 40, modes); err != nil {
		log.Fatal("request for pseudo terminal failed: ", err)
	}
	sshOut, _ := session.StdoutPipe()
	sshErr, _ := session.StderrPipe()

	err = session.Run("fdpi_ctrl list all --bind_multi")
	if err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	be, _ := io.ReadAll(sshErr)

	fmt.Println("ParseErr", string(be))
	scanner := bufio.NewScanner(sshOut)
	for scanner.Scan() {
		if uregexp.MatchString(scanner.Text()) {
			m := uregexp.FindStringSubmatch(scanner.Text())
			users = append(users, ScatUser{ID: m[1], IP: m[2]})
		}
	}
	return users, nil
}

func (a *App) AddTariff(in, out int) {

}
