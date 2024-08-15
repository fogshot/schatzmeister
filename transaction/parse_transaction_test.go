package transaction

import (
	"encoding/csv"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseTransaction(t *testing.T) {
	// given
	sampleCsv := `Buchungstag;Wertstellung;Umsatzart;Buchungstext;Betrag;Währung;IBAN Kontoinhaber;Kategorie
05.08.2024;05.08.2024;Lastschrift;ACME Inc. RE. 0031055169          37,90 1 Stk 240 ltr. Behaelter End-to-End-Ref.: Zahl.Beleg 0000039102 Mandatsref: 000560-1004082357 Gläubiger-ID: AB11ZZZ00000444444 SEPA-BASISLASTSCHRIFT wiederholend;-37,9;EUR;DE11100400410211111111;
02.08.2024;02.08.2024;Überweisung;Max Mustermann Miete Zimmer 1, Max Mustermann End-to-End-Ref.: NOTPROVIDED Kundenreferenz: NSCT1111111129340000000000000000003;200;EUR;DE11111100410245111111;`
	ioReader := strings.NewReader(sampleCsv)
	csvReader := csv.NewReader(ioReader)
	csvReader.Comma = ';'
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return
	}
	// when
	_, err = csvReader.Read()
	if err != nil {
		t.Fatalf("Failed to read header row: %v", err)
	}
	record, err := csvReader.Read()
	if err != nil {
		t.Fatalf("Failed to read row: %v", err)
	}
	transaction, err := ParseTransaction(record)
	// then
	want := &Transaction{
		BookingDate: time.Date(2024, 8, 5, 0, 0, 0, 0, location),
		ValueDate:   time.Date(2024, 8, 5, 0, 0, 0, 0, location),
		Category:    2,
		Memo:        "RE. 0031055169 37,90 1 Stk 240 ltr. Behaelter",
		Amount:      -37.9,
		Currency:    0,
		Iban:        "DE11100400410211111111",
		Payee:       "ACME Corporation",
	}
	if err != nil {
		if errors.Is(err, csv.ErrFieldCount) {
			t.Fatalf("%v: expect %v", err, csvReader.FieldsPerRecord)
		}
		t.Fatalf("error during parseTransaction: %v", err)
	}
	if !reflect.DeepEqual(transaction, want) {
		t.Fatalf("parseTransaction returned unexpected result.\ngot: %+v\nwant: %+v", transaction, want)
	}
}
