package csv

import (
	"Schatzmeister/transaction"
	"encoding/csv"
	"fmt"
	"io"
	"log"
)

// ReadTransactionsFromCommerzbankCsv reads records from the given csv
func ReadTransactionsFromCommerzbankCsv(ioReader io.Reader) ([]*transaction.Transaction, error) {
	reader := csv.NewReader(ioReader)
	reader.Comma = ';' // TODO move to config
	var transactions []*transaction.Transaction
	// assume the first row of the .csv is a header and therefore skip it
	_, err := reader.Read()
	if err != nil {
		log.Println("Error reading header:", err)
		return nil, err
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		parsedTransaction, err := transaction.ParseTransaction(record)
		if err != nil {
			log.Println("Error parsing transaction:", err)
		}
		transactions = append(transactions, parsedTransaction)
	}
	return transactions, nil
}

// WriteYNABCsv writes transactions to .csv in the format required for import in YNAB
func WriteYNABCsv(ioWriter io.Writer, transactions []*transaction.Transaction) error {
	writer := csv.NewWriter(ioWriter)
	defer writer.Flush()
	layout := "2006-01-02" // TODO move to config
	// Date,Payee,Memo,Amount
	err := writer.Write([]string{"Date", "Payee", "Memo", "Amount"})
	if err != nil {
		return err
	}
	for _, t := range transactions {
		record := []string{
			t.ValueDate.Format(layout),
			t.Payee,
			t.Memo,
			fmt.Sprintf("%+.2f", t.Amount),
		}
		err := writer.Write(record)
		if err != nil {
			return err
		}
	}
	return nil
}
