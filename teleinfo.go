package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("EDF Teleinfo to AirVantage\n")
	if len(os.Args) != 4 {
		fmt.Println(`
        Usage: ./teleinfo [tty] [platform] [deviceid] [password]

        Example: ./teleinfo /dev/ttyUSB0 na MYHOME mypassword
        `)
		os.Exit(1)
	}

	fmt.Println("TeleInfo to AirVantage\n")

	fi, err := os.Open(os.Args[0])

	if err != nil {
		panic(err)
	}

	defer func() {
		fi.Close()
	}()

	// scan the file for lines
	scanner := bufio.NewScanner(fi)

	// container for 10 data read
	container := make([]map[string]interface{}, 10)
	index := -1

	myWay := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Index(line, "***") == 0 {

			if strings.Index(line, "*** Voie 1 ***") == 0 {
				fmt.Println("my way")

				myWay = true
				if index == 9 {
					index = 0

					// push the batch of data to AV
					b, err := json.Marshal(container)
					if err != nil {
						panic(err)
					}
					fmt.Printf("JSon: %s\n", b)
					reader := bytes.NewBuffer(b)
					r, _ := http.NewRequest("POST", "http://"+os.Args[1]+".m2mop.net/device/messages", reader)
					r.SetBasicAuth(os.Args[2], os.Args[3])
					client := &http.Client{}
					resp, err := client.Do(r)

					if err == nil {
						fmt.Printf("HTTP status %d\n", resp.StatusCode)
						if resp.StatusCode != 200 {
							fmt.Printf("Server error: %s\n", resp.Status)
						}
					} else {
						fmt.Printf("HTTP error: %s\n", err.Error())
					}
				} else {
					index++
				}
				container[index] = make(map[string]interface{})
			} else {
				fmt.Println("another way")
				myWay = false
			}
		} else {

			if myWay && strings.Index(line, "PAPP") == 0 {
				parts := strings.Split(line, " ")
				addData("PAPP", parts[1], index, &container)
			}

			if myWay && strings.Index(line, "HCHC") == 0 {
				parts := strings.Split(line, " ")
				addData("HCHC", parts[1], index, &container)
			}

			if myWay && strings.Index(line, "HCHP") == 0 {
				parts := strings.Split(line, " ")
				addData("HCHP", parts[1], index, &container)
			}
		}
	}
}

func addData(key string, value string, idx int, container *[]map[string]interface{}) {
	var err error
	(*container)[idx][key], err = createEntry(value)
	if err != nil {
		fmt.Println("Parsing error: %s\n", err.Error())
	}
}

func createEntry(value string) (dataKey []map[string]interface{}, err error) {
	dataKey = make([]map[string]interface{}, 1)
	dataKey[0]["timestamp"] = time.Now().Unix() * 1000
	dataKey[0]["value"], err = strconv.Atoi(value)
	return dataKey, err
}
