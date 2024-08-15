package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Category an enumeration of the possible types of transactions handled by this program
type Category int64

// Currency an enumeration of possible currencies handled by this program
type Currency int64

const (
	Fee Category = iota
	Transfer
	Charge
	Order
	Others
)

const (
	Eur Currency = iota
	Usd
)

// PayeeVariations maps the canonical version of a payee to its known variations
type PayeeVariations map[string][]string

// Transaction holds data representing a single financial transaction, e.g. a bank transfer
type Transaction struct {
	BookingDate time.Time
	ValueDate   time.Time
	Category    Category
	Memo        string
	Amount      float64
	Currency    Currency
	Iban        string
	Payee       string
}

// ParseTransaction attempts to convert a single csv record into a Transaction
func ParseTransaction(record []string) (*Transaction, error) {
	// TODO make this bank-agnostic, currently the implementation is specific to Commerzbank exports
	wantLength := 8
	if len(record) != wantLength {
		return nil, fmt.Errorf("invalid record length (need %v): %v", wantLength, len(record))
	}
	parsedBookingDate, err1 := parseDate(record[0])
	parsedValueDate, err2 := parseDate(record[1])
	parsedCategory, err3 := parseTransactionCategory(record[2])
	payee, memo, err7 := extractPayeeAndMemo(record[3], parsedCategory)
	parsedAmount, err4 := parseAmount(record[4])
	parsedCurrency, err5 := parseCurrency(record[5])
	if err7 != nil {
		log.Println("error extracting payee:", err7)
	}
	transaction := &Transaction{
		BookingDate: parsedBookingDate,
		ValueDate:   parsedValueDate,
		Category:    parsedCategory,
		Memo:        memo,
		Amount:      parsedAmount,
		Currency:    parsedCurrency,
		Iban:        record[6],
		Payee:       payee,
	}
	if err := errors.Join(err1, err2, err3, err4, err5); err != nil {
		return nil, err
	}
	return transaction, nil
}

func parseAmount(amountStr string) (float64, error) {
	// TODO extract decimal separator to config
	return strconv.ParseFloat(strings.Replace(amountStr, ",", ".", -1), 64)
}

func parseDate(dateStr string) (time.Time, error) {
	layout := "02.01.2006"                              // TODO extract date layout to config
	location, err := time.LoadLocation("Europe/Berlin") // TODO extract location to config
	if err != nil {
		return time.Parse(layout, dateStr)
	}
	return time.ParseInLocation(layout, dateStr, location)
}

// parseTransactionCategory attempts to map the given text to a TransactionCategory
func parseTransactionCategory(categoryText string) (Category, error) {
	// TODO extract possible category names to config?
	switch categoryText {
	case "Zinsen/Entgelte":
		return Fee, nil
	case "Überweisung":
		return Transfer, nil
	case "Lastschrift":
		return Charge, nil
	case "Dauerauftrag":
		return Order, nil
	case "Sonstige":
		return Others, nil
	}
	return -1, errors.New("unknown transaction category")
}

// parseCurrency attempts to map the given text to a Currency
func parseCurrency(text string) (Currency, error) {
	switch text {
	case "EUR":
		return Eur, nil
	case "USD":
		return Usd, nil
	}
	return -1, errors.New("unknown currency")
}

// ExtractPayeeAndMemo attempts to separate payee and memo inside the given text
func extractPayeeAndMemo(text string, category Category) (payee string, memo string, err error) {
	if category == Fee {
		payee = "Commerzbank"
		memo = trimMemo(text)
		return payee, memo, nil
	}
	payee = "Unknown"
	payeeVariations, err := loadPayeeVariations("./testdata/known_payees.json") // TODO extract path to config
	if err != nil {
		log.Println("Error loading payee variations:", err)
	}
	memo = trimMemo(text)

	for canonicalPayee, variations := range payeeVariations {
		for _, variation := range variations {
			if strings.HasPrefix(strings.ToLower(memo), strings.ToLower(variation)) {
				payee = canonicalPayee
				memo = strings.TrimSpace(memo[len(variation):])
				return payee, memo, nil
			}
		}
	}

	// TODO extract corporate endings to config?
	// If no known payee is found, check for common corporate endings
	corporateEndings := []string{"e. V.", "KG", "GmbH", "AG", "OHG", "GbR", "PartG", "UG", "SE", "Inc.", "Ltd."}
	for _, ending := range corporateEndings {
		lowerEnding := strings.ToLower(ending)
		if strings.Contains(strings.ToLower(memo), lowerEnding) {
			payeeEndIndex := strings.LastIndex(strings.ToLower(memo), lowerEnding) + len(ending)
			return memo[:payeeEndIndex], strings.TrimSpace(memo[payeeEndIndex:]), nil
		}
	}
	return payee, memo, errors.New("payee unknown")
}

// trimMemo removes the e2e reference and excess whitespace
func trimMemo(memo string) string {
	memoRegex := regexp.MustCompile(` End-to-End-Ref\.: .*$`)
	memo = memoRegex.ReplaceAllString(memo, "")
	whiteSpaceRegex := regexp.MustCompile(`\s+`)
	memo = whiteSpaceRegex.ReplaceAllString(memo, " ")
	return strings.TrimSpace(memo)
}

// loadPayeeVariations loads the canonical payee map from the given json file
func loadPayeeVariations(filename string) (PayeeVariations, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening file: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close file: %v", err)
		}
	}(file)

	var payeeVariations PayeeVariations
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&payeeVariations)
	if err != nil {
		return nil, err
	}
	return payeeVariations, nil
}
