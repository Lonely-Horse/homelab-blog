package adminseed

import (
	"bufio"
	"flag"
	"fmt"
	"homelab-blog/internal/config"
	"io"
	"log"
	"os"
	"strings"
)

func readPassword(r io.Reader) (string, error) {
	line, err := bufio.NewReader(r).ReadString('\n')
	if err != io.EOF && err != nil {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	return line, nil
}

func main() {
	var user string
	var reset bool
	flag.StringVar(&user, "user", "admin", "管理员名称")
	flag.BoolVar(&reset, "reset", false, "已存在时覆盖密码，默认拒绝")
	flag.Parse()

	if user == "" {
		log.Printf("The user is empty,please enter the name")
		return
	}

	pass, err := readPassword(os.Stdin)

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("The config model load failed")
		return
	}

}
