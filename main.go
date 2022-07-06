package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func init() {
	fmt.Println("#Welcom to Docter of Docker.")
}

type AddrObj struct {
	Index int
	Name  string
	Mac   string
	Ip    string
}

var addr = flag.String("addr", ":10086", "http service address") // name , value , usage

//和这个效果类似，用Must对其封装，返回一个模板
//如果发生错误，将会用panic(err)报错
var templ = template.Must(template.New("html_templ_qr").Parse(templateStr))

func main() {
	flag.Parse()
	http.Handle("/", http.HandlerFunc(QR))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

//QR  请求处理
func QR(w http.ResponseWriter, req *http.Request) {
	path := req.FormValue("fp")
	var fileContent string
	var err error

	if len(path) > 0 {
		fileContent, err = readContent(path)
		if err != nil {
			log.Fatal("File IO Error:", err)
		}
	}

	addrs, err := net.Interfaces()

	netaddrs := make([]AddrObj, 0)
	if err == nil {
		for idx, addr := range addrs {
			ip, _ := GetInterfaceIpv4Addr(addr.Name)
			netaddrs = append(netaddrs, AddrObj{Index: idx, Name: addr.Name, Mac: addr.HardwareAddr.String(), Ip: ip})
		}
	}

	templ.Execute(w, map[interface{}]interface{}{
		"netaddrs": netaddrs,
		"file":     fileContent,
	})
}

func readContent(textfile string) (string, error) {
	file, err := os.Open(textfile)
	if err != nil {
		log.Printf("Cannot open text file: %s, err: [%v]", textfile, err)
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var outstr string
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if len(line) > 0 {
			outstr += fmt.Sprint(line, "\n")
		}
	}
	outstr = strings.TrimSpace(outstr)

	return outstr, nil
}

func GetInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}
	if addrs, err = ief.Addrs(); err != nil { // get addresses
		return
	}
	for _, addr := range addrs { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		return "", errors.New(fmt.Sprintf("interface %s don't have an ipv4 address\n", interfaceName))
	}
	return ipv4Addr.String(), nil
}

const templateStr = `
<html>
<head>
<meta charset="UTF-8">
<title>Dockter</title>
<style type="text/css">
table, tr, td {
    border: 1px black solid;
}
</style>

</head>
<body>
{{if .netaddrs}}
<h1>服务器网卡信息</h1>
<table>
{{range $i, $v := .netaddrs}}
<tr>
  <td>{{$v.Index}}</td>
  <td>{{$v.Ip}}</td>
  <td>{{$v.Name}}</td>
  <td>{{$v.Mac}}</td>  
</tr>
{{end}}
</table>
{{end}}

<h1>读取文档：</h1>
<form action="/" name=f method="GET">
    <input maxLength=1024 size=70 name=fp value="" title="ReadFilePath">
    <input type=submit value="ReadFile" name=qr>
</form>
<h1>文档内容：</h1>
{{if .file}}
{{.file}}
{{end}}
</body>
</html>
`
