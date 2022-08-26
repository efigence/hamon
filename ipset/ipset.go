package ipset

import (
	"encoding/xml"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

var bin = ""

func init() {
	bin, _ = exec.LookPath("ipset")
}

func runError(cmd *exec.Cmd) error {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"error running %s: %s[%s])",
			strings.Join(cmd.Args, " "),
			string(out),
			err,
		)
	}
	return nil
}

func Create(name string, setType string) error {
	if len(name) > 31 {
		return fmt.Errorf("max 31 character name")
	}
	cmd := exec.Command(bin, "create", name, setType)
	return runError(cmd)
}
func Destroy(name string) error {
	cmd := exec.Command(bin, "destroy", name)
	return runError(cmd)
}

func Swap(name1, name2 string) error {
	return runError(exec.Command(bin, "swap", name1, name2))
}

func Add(set string, ip net.IP) error {
	return runError(exec.Command(bin, "add", set, ip.String()))
}

// https://www.onlinetool.io/xmltogo/
type Ipsets struct {
	XMLName xml.Name `xml:"ipsets"`
	Text    string   `xml:",chardata"`
	Ipset   struct {
		Text     string `xml:",chardata"`
		Name     string `xml:"name,attr"`
		Type     string `xml:"type"`
		Revision string `xml:"revision"`
		Header   struct {
			Text       string `xml:",chardata"`
			Family     string `xml:"family"`
			Hashsize   string `xml:"hashsize"`
			Maxelem    string `xml:"maxelem"`
			Memsize    string `xml:"memsize"`
			References string `xml:"references"`
			Numentries string `xml:"numentries"`
		} `xml:"header"`
		Members struct {
			Text   string `xml:",chardata"`
			Member []struct {
				Text string `xml:",chardata"`
				Elem string `xml:"elem"`
			} `xml:"member"`
		} `xml:"members"`
	} `xml:"ipset"`
}

func List(name string) ([]net.IP, error) {
	cmd := exec.Command(bin, "list", name, "-o", "xml")
	ipList := []net.IP{}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return []net.IP{}, err
	}
	//sc := bufio.NewScanner(strings.NewReader(string(out)))
	//for sc.Scan() {
	//	lines = append(lines, sc.Text())
	//}
	ipsXML := &Ipsets{}
	err = xml.Unmarshal(out, ipsXML)
	if err != nil {
		return []net.IP{}, err
	}
	for _, v := range ipsXML.Ipset.Members.Member {
		ip := net.ParseIP(v.Elem)
		if ip == nil {
			return []net.IP{}, fmt.Errorf("error parsing [%s] as IP", v.Elem)
		}
		ipList = append(ipList, ip)
	}
	return ipList, nil
}
