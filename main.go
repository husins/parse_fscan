package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/tealeg/xlsx"
)

type ScanURL struct {
	url   string
	code  string
	len   string
	title string
	mvUrl string
}

type ScanIP struct {
	IP      string
	port    []string
	url     ScanURL
	netBios string
	info    string
}

type Info map[string]ScanIP

func removeSpace(s []string) (ret []string) {
	for _, value := range s {
		if len(value) > 2 {
			ret = append(ret, value)
		}
	}
	return ret
}

func regexpIP(line string) string {
	reg := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	if reg == nil {
		fmt.Println("reg is err")
		panic(1)
	}

	ip := reg.FindAllStringSubmatch(line, -1)

	return ip[0][0]
}

func regexpURL(line string) string {
	reg := regexp.MustCompile(`(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`)
	if reg == nil {
		fmt.Println("reg is err")
		panic(1)
	}
	url := reg.FindAllStringSubmatch(line, -1)

	return url[0][0]

}

func regexpTitle(line string) string {
	if strings.Contains(line, "跳转") {
		tmp := strings.Split(line, "跳转")[0]
		title := strings.Split(tmp, "title:")[1]
		return title
	} else {
		title := strings.Split(line, "title:")[1]
		return title
	}

}

func parsePort(line string, IPInfo Info) Info {
	ip := strings.Split(line, ":")
	port := strings.Split(ip[1], " ")
	var tempIP ScanIP
	if _, ok := IPInfo[ip[0]]; !ok {
		tempIP.IP = ip[0]
		tempIP.port = append(tempIP.port, port[0])
		IPInfo[ip[0]] = tempIP
	} else {
		tempIP = IPInfo[ip[0]]
		tempIP.port = append(IPInfo[ip[0]].port, port[0])
		IPInfo[ip[0]] = tempIP
	}

	return IPInfo

}

func parseWebtitle(line string, IPInfo Info) Info {
	var tempIP ScanIP
	ip := regexpIP(line)
	if _, ok := IPInfo[ip]; !ok {
		tempIP.IP = ip
		fmt.Println("IP: " + tempIP.IP)
		tempIP.url.url = regexpURL(line)
		fmt.Println("url: " + tempIP.url.url)
		url := strings.Split(line, " ")
		url = removeSpace(url)
		tempIP.url.code = strings.Split(url[3], ":")[1]
		fmt.Println("code: " + tempIP.url.code)
		tempIP.url.len = strings.Split(url[4], ":")[1]
		fmt.Println("len: " + tempIP.url.len)
		tempIP.url.title = regexpTitle(line)
		fmt.Println("title: " + tempIP.url.title)
		if strings.Contains(line, "跳转url:") {
			tempIP.url.mvUrl = strings.Split(line, "跳转url:")[1]
			fmt.Println("mvurl: " + tempIP.url.mvUrl)
		}
		// tempIP.url.mvUrl = strings.Split(line, "跳转url:")[1]
		// fmt.Println("mvurl: " + tempIP.url.mvUrl)
		IPInfo[ip] = tempIP
	} else {
		tempIP.url.url = regexpURL(line)
		url := strings.Split(line, " ")
		url = removeSpace(url)
		tempIP.url.code = strings.Split(url[3], ":")[1]
		tempIP.url.len = strings.Split(url[4], ":")[1]
		tempIP.url.title = regexpTitle(line)
		if strings.Contains(line, "跳转url:") {
			tempIP.url.mvUrl = strings.Split(line, "跳转url:")[1]
			IPInfo[tempIP.url.url] = tempIP
		} else {
			tempIP.url.mvUrl = "None"
		}

	}
	return IPInfo
}

func parseNetBios(line string, IPInfo Info) Info {
	var tempIP ScanIP
	ip := strings.Split(line, " ")
	ip = removeSpace(ip)
	// fmt.Println(url)
	if _, ok := IPInfo[ip[2]]; !ok {
		tempIP.IP = ip[2]
		tempIP.netBios = ip[3]
		if len(ip) > 4 {
			tempIP.info = ip[4]
		}
		IPInfo[ip[2]] = tempIP
	} else {
		tempIP = IPInfo[ip[2]]
		tempIP.netBios = ip[3]
		if len(ip) > 4 {
			tempIP.info = ip[4]
		}
		IPInfo[ip[2]] = tempIP
	}

	return IPInfo
}

func parseInfoScan(line string, IPInfo Info) Info {
	var tempIP ScanIP
	restIP := regexpIP(line)
	s := strings.Split(line, " ")

	if _, ok := IPInfo[restIP]; !ok {
		tempIP.IP = restIP
		tempIP.info = s[3]
		IPInfo[restIP] = tempIP
	} else {
		tempIP = IPInfo[restIP]
		tempIP.info = s[3]
		IPInfo[restIP] = tempIP
	}
	return IPInfo
}

func AddValue(row *xlsx.Row, value string) {
	cell := row.AddCell()
	cell.Value = value
}

func createXlsx(IPinfo Info) {
	file := xlsx.NewFile()
	defer file.Save("result.xlsx")
	sheet, err := file.AddSheet("Fscan")
	if err != nil {
		panic(err)
	}
	row := sheet.AddRow()
	title := []string{"IP", "Port", "Url", "Title", "Code", "Len", "Move Url", "NetBIOS", "INFO"}

	for _, value := range title {
		AddValue(row, value)
	}

	for _, value := range IPinfo {
		row := sheet.AddRow()
		AddValue(row, value.IP)
		ports := ""
		for _, port := range value.port {
			ports += port + ","
		}
		AddValue(row, ports)
		AddValue(row, value.url.url)
		AddValue(row, value.url.title)
		AddValue(row, value.url.code)
		AddValue(row, value.url.len)
		AddValue(row, value.url.mvUrl)
		AddValue(row, value.netBios)
		AddValue(row, value.info)
	}

}

func main() {
	IPInfo := make(Info)
	file, err := os.Open("result.txt")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer file.Close()

	br := bufio.NewReader(file)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		sLine := string(line)
		if strings.Contains(sLine, "open") {
			IPInfo = parsePort(sLine, IPInfo)
		} else if strings.Contains(sLine, "WebTitle:") {
			IPInfo = parseWebtitle(sLine, IPInfo)
		} else if strings.Contains(sLine, "NetBios:") {
			IPInfo = parseNetBios(sLine, IPInfo)
		} else if strings.Contains(sLine, "InfoScan:") {
			IPInfo = parseInfoScan(sLine, IPInfo)
		}
	}

	createXlsx(IPInfo)
}
