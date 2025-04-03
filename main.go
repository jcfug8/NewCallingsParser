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
	columnCount    = 6
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
	// get the input file name from the user with a default value
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("error getting executable path: %v", err)
	}
	log.Println("Executable path:", execPath)

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
	}

	log.Printf("input file: %s", inputFileName)

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
	from := time.Date(time.Now().Year(), time.Now().Month()-2, 1, 0, 0, 0, 0, time.Local)
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
