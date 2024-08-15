package main

import (
	"Schatzmeister/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	inputPath := flag.String("input", "", "Path to the input Commerzbank CSV file")
	outputPath := flag.String("output",
		fmt.Sprintf("%v-ynab.csv", time.Now().Format("2006-01-02")),
		"Path to the output YNAB CSV file")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		log.Println("Usage: your_program_name -input <input_csv_path> -output <output_csv_path>")
		flag.PrintDefaults()
		os.Exit(1) // Indicate an error
	}

	inputFile, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalln("Error opening input file:", err)
	}
	defer func(inputFile *os.File) {
		err := inputFile.Close()
		if err != nil {
			log.Fatalln("error closing input file:", err)
		}
	}(inputFile)

	transactions, err := csv.ReadTransactionsFromCommerzbankCsv(inputFile)
	if err != nil {
		log.Fatalln("error reading transactions:", err)
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalln("error creating output file:", err)
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			log.Fatalln("error closing output file:", err)
		}
	}(outputFile)

	err = csv.WriteYNABCsv(outputFile, transactions)
	if err != nil {
		log.Fatalln("Error writing YNAB CSV:", err)
	}

	log.Println("Conversion successful!")
}
