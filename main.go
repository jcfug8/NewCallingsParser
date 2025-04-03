package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	inputFileName  = "callings.in.tsv"
	outputFileName = "callings.out.csv"
	columnCount    = 4
	numberOfMonths = 1
)

const (
	CallingsWithDateSustainedColumn = "Callings with Date Sustained"
	IndividualPhoneColumn           = "Individual Phone"
)

type Calling struct {
	Name          string
	DateSustained time.Time
	// DateSetApart  time.Time
}

var columnNames = []string{}
var callingsIndex = -1
var phoneIndex = -1

type Record struct {
	Data     []string
	Callings []Calling
}

func main() {
	printLogo()
	// get the input file name from the user with a default value
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("error getting executable path: %v", err)
	}
	// log.Println("Executable path:", execPath)

	// Check if the executable is in a temporary directory
	if strings.Contains(execPath, os.TempDir()) {
		log.Println("Running with `go run`")
	} else if strings.Contains(execPath, "__debug_bin") {
		log.Println("Running with `dlv debug`")
	} else {
		inputFileName = path.Join(filepath.Dir(execPath), inputFileName)
		outputFileName = path.Join(filepath.Dir(execPath), outputFileName)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter input file name (callings.in.tsv): ")
		userInputFileName, _ := reader.ReadString('\n')
		if userInputFileName != "\n" {
			inputFileName = strings.TrimSpace(userInputFileName)
		}

		fmt.Print("Enter output file name (callings.out.csv): ")
		userOutputFileName, _ := reader.ReadString('\n')
		if userOutputFileName != "\n" {
			outputFileName = strings.TrimSpace(userOutputFileName)
		}

		fmt.Print("How many columns are in the input file?: ")
		columnCountString, _ := reader.ReadString('\n')
		columnCount, err = strconv.Atoi(strings.TrimSpace(columnCountString))
		if err != nil {
			log.Fatalf("please enter a number for the number of columns: %v", err)
		}

		fmt.Print("How many months back do you want to filter?: ")
		numberOfMonthsString, _ := reader.ReadString('\n')
		numberOfMonths, err = strconv.Atoi(strings.TrimSpace(numberOfMonthsString))
		if err != nil {
			log.Fatalf("please enter a number for the number of months: %v", err)
		}
		numberOfMonths = numberOfMonths - 1
	}

	// log.Printf("input file: %s", inputFileName)

	log.Print("reading file...")
	file, err := os.Open(inputFileName)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}
	defer file.Close()

	log.Print("parsing records...")
	records, err := parseRecords(file)
	if err != nil {
		log.Fatalf("error parsing records: %v", err)
	}

	log.Print("filtering records...")
	// filter records by calling date
	from := time.Date(time.Now().Year(), time.Now().Month()-time.Month(numberOfMonths), 1, 0, 0, 0, 0, time.Local)
	to := time.Date(time.Now().Year(), time.Now().Month(), 31, 23, 59, 59, 0, time.Local)
	records = filterRecordsByCallingDate(records, from, to)

	log.Print("writing records...")
	// write records to csv file
	err = writeRecordsToCSVFile(records)
	if err != nil {
		log.Fatalf("error writing records to csv file: %v", err)
	}

	log.Print("done")
}

func printLogo() {
	// blue backgorund and white text
	fmt.Println("\033[44;37m")
	fmt.Println("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY55PP")
	fmt.Println("PPP55555555YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY555PPGGGG")
	fmt.Println("PPPPPPPPPPPPPPP5555555YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY55555PPPGGGGGGG")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPPPPPPP55555555YYYYYYYYYYYYYYYYYYYYY55555PPPPPPGGGGGGGGG")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPPPPPPPPPP5YJ????7777777?JYY555555555555PPPPGGGGGGGGGGGG")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPPPPPPP5J77?JY55PPPPPP55YJ??J5GGPPPPPPPPGGPPPPPPPPPGGGGG")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPPPPPY77J5PPPPPPPPPPGGGGBBBG5??YGBGBBBBBBBBGGGGGGGGGPPPP")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPPP5!75PPPPPPPPGGGGGBGGGGGGGGBPJ7P#BBBBBBBBBBBBBBBBBBBGG")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPPJ~YPPPPPGGGGGGP77!YGGGGGGGGBB#G!Y#BBBBBBBBBBBBBBBBBBBB")
	fmt.Println("PPPPPPPPPPPPPPPPPPPPJ~PGGGGGGGGGGGG7~:~JPGGGGBBBBBBBB!Y#BBBBBBBBBBBBBBBBBBB")
	fmt.Println("PPPPPPPPPPPPPPPPPPGP^5BGGGGGGGGGGB57~^~Y?GBBBBBBBBBBBB^PBBBBBBBBBBBBBBBBBBB")
	fmt.Println("PPPPPPPPPPPPGGGGGGBJ!BGGGGGGGGGGG5~!^^75??5BBBBBBBBBB#??#BBBBBBBBBBBBBBBBBB")
	fmt.Println("PPPPPPGGGGGGGGGGGGB7?BGGGGGGGGB5!^.   ~^^???BBBBBBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("PGGGGGGGGGGGGGGGGGB7?BGGGGGGGBY^^7. :~:^7.?!JGBBBBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGGGGGGGB7?BGGGGBBP! :??~^::!7: !.:JG#BBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGGGGGGGB7?BGG55Y!:^J5~^.:!?7^~ !:.?.?GBGBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGGGGGGGB7?BY~!7?5P~~7!~!77~.:! ~~:PY77J5P5B5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGGGGGGGB7?BGGBB##B:.!^~!!!. !: :!!!7:GBBPP#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGGGGGGGB7J#BBBBBB#Y ^!~^!^  7. .Y!:! Y#BBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGGGGBBB#7J#BBBBBBB#5!^!7:  :7  .G!:! 7#BBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGGGGBBBBBB#7J#BBBBBBBBB##J    7^  .P::~ ?#BBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGGGGBBBBBBBBB#7J#BBBBBBBBBB#7   .J   :! .: 5#BBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGGGGBBBBBBBBBBBB#7J#BBBBBBBBBB#!   ?7   ^7   :BBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("GGGBBBBBBBBBBBBBBB#7J#BBBBBBBBBBB:  :5.   ~? .^.BBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBB#7J#BBBBBBBBBBG.  Y~    ?P~~?.GBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBB#7J#BBBBBBBBBBP..7~   :!5#BBB?GBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBB#7J#BBBBBBBBB#Y ^. :~YG?J#BBBBBBBBB#5!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBB#7J#####BBB###J~7?5!!#P^G#BBB#######5~#BBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBB#?75YYYYYYYYY7:!55Y:.7Y??YYYYYYYYYY5?!#BBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBBBGPPPPPPPPPPPPGGPPPGGPPGPPPPPPPPPPPPPGBBBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	// change text back to default
	// fmt.Println("\033[0m")
	fmt.Println("Welcome to the Callings Parser!")
	fmt.Println("This program will parse a tab delimited file of callings and output a csv file.")
}

func parseRecords(file *os.File) ([]Record, error) {
	// unmarshall fileContents into a slice of Records
	var records []Record
	rawRecords, err := readRecords(file)
	if err != nil {
		return nil, fmt.Errorf("error reading records: %v", err)
	}
	// get the column names
	columnNames = rawRecords[0]
	if len(columnNames) != columnCount {
		return nil, fmt.Errorf("expected %d header columns, found %d", columnCount, len(columnNames))
	}
	// get the index of the callings column
	for i, columnName := range columnNames {
		if columnName == CallingsWithDateSustainedColumn {
			callingsIndex = i
		}
		if columnName == IndividualPhoneColumn {
			phoneIndex = i
		}
	}
	if callingsIndex == -1 {
		return nil, fmt.Errorf("callings column not found in csv file")
	}

	// parse the records
	for i, rawRecord := range rawRecords[1:] {
		if len(rawRecord) != columnCount {
			return nil, fmt.Errorf("record %d has %d columns, expected %d", i+1, len(rawRecord), columnCount)
		}
		// parse the callings
		callings, err := parseCallings(rawRecord[callingsIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing callings for record %d: %v", i+1, err)
		}

		// format the phone number
		if phoneIndex != -1 {
			rawRecord[phoneIndex] = formatPhoneNumber(rawRecord[phoneIndex])
		}

		record := Record{
			Data:     rawRecord,
			Callings: callings,
		}

		// add the record to the list of records
		records = append(records, record)
	}

	return records, nil
}

func readRecords(file *os.File) ([][]string, error) {
	// manually parse the file based on the tab delimiter
	// This is because the std lib csv reader does not support unquoted fields with newlines
	var records [][]string

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	tabCount := 0
	record := make([]string, columnCount)
	for _, char := range contents {
		if char == '\t' {
			tabCount++
		} else if tabCount == columnCount-1 && char == '\n' {
			records = append(records, record)
			tabCount = 0
			record = make([]string, columnCount)
		} else {
			record[tabCount] += string(char)
		}
	}

	return records, nil
}

// formatPhoneNumber formats the phone number to a standard format
func formatPhoneNumber(phoneNumber string) string {
	// consistantly format phone numbers
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ".", "")
	if len(phoneNumber) == 10 {
		phoneNumber = "1" + phoneNumber
	}
	if len(phoneNumber) == 11 {
		phoneNumber = "+" + phoneNumber
	}
	return phoneNumber
}

func parseCallings(rawCallings string) ([]Calling, error) {
	lineCallings := strings.Split(rawCallings, "\n")
	var callings []Calling
	for j, calling := range lineCallings {
		if len(calling) == 0 {
			continue
		}
		calling, err := parseCalling(calling)
		if err != nil {
			return nil, fmt.Errorf("error parsing calling %d: %v", j, err)
		}
		callings = append(callings, calling)
	}
	return callings, nil
}

func parseCalling(callingBytes string) (Calling, error) {
	callingColumns := strings.Split(callingBytes, "(")
	if len(callingColumns) != 2 {
		return Calling{}, fmt.Errorf("unable to parse calling: not 2 sections: %s", callingBytes)
	}
	callingName := callingColumns[0]
	callingDateString := strings.Replace(callingColumns[1], ")", "", -1)
	callingTime, err := time.Parse("2 Jan 2006", callingDateString)
	if err != nil {
		return Calling{}, err
	}
	return Calling{Name: callingName, DateSustained: callingTime.In(time.Local)}, nil
}

func filterRecordsByCallingDate(records []Record, from, to time.Time) []Record {
	var filteredRecords []Record
	for _, record := range records {
		for _, calling := range record.Callings {
			if (calling.DateSustained.After(from) || calling.DateSustained.Equal(from)) && (calling.DateSustained.Before(to) || calling.DateSustained.Equal(to)) {
				filteredRecords = append(filteredRecords, record)
				break
			}
		}
	}
	return filteredRecords
}

func writeRecordsToCSVFile(records []Record) error {
	log.Printf("writing %d records to %s", len(records), outputFileName)
	csvFile, err := os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("error creating csv file: %v", err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// write headers
	err = csvWriter.Write(columnNames)
	if err != nil {
		return fmt.Errorf("error writing headers: %v", err)
	}

	// write records
	for _, record := range records {
		var callings []string
		for _, calling := range record.Callings {
			callings = append(callings, fmt.Sprintf("%s (%s)", calling.Name, calling.DateSustained.Format("2 Jan 2006")))
		}
		err = csvWriter.Write(record.Data)
		if err != nil {
			return fmt.Errorf("error writing record: %v", err)
		}
	}
	return nil
}
