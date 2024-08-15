package transaction

import (
	"errors"
	"fmt"
	"testing"
)

func ExampleExtractPayeeAndMemo() {
	buchungstext := "ACME Corporation Zahlung 123456789 End-to-End-Ref.: NOTPROVIDED"
	payee, memo, err := extractPayeeAndMemo(buchungstext, Transfer)
	fmt.Println(payee)
	fmt.Println(memo)
	fmt.Println(err)
	// Output: ACME Corporation
	// Zahlung 123456789
	// <nil>
}

func ExampleExtractPayeeAndMemo_fee() {
	buchungstext := "Kontoführung Konto  1235757"
	payee, memo, err := extractPayeeAndMemo(buchungstext, Fee)
	fmt.Println(payee)
	fmt.Println(memo)
	fmt.Println(err)
	// Output: Commerzbank
	// Kontoführung Konto 1235757
	// <nil>
}

// TODO create unit tests for loadPayeeVariations
// TODO create unit tests for trimMemo
// TODO create unit tests for parseCurrency
// TODO create unit tests for parseTransactionCategory

func TestExtractPayeeAndMemoWithKnownPayeeWithVariations(t *testing.T) {
	given := "ACME Zahlung 123456789 End-to-End-Ref.: NOTPROVIDED"
	wantPayee, wantMemo := "ACME Corporation", "Zahlung 123456789"
	testExtractPayeeAndMemo(t, given, Charge, wantPayee, wantMemo, false)
}

func TestExtractPayeeAndMemoWithKnownPayeeExactMatch(t *testing.T) {
	given := "ACME Corporation Rechnung 4711"
	wantPayee, wantMemo := "ACME Corporation", "Rechnung 4711"
	testExtractPayeeAndMemo(t, given, Transfer, wantPayee, wantMemo, false)
}

func TestExtractPayeeAndMemoWithCorporateEnding(t *testing.T) {
	given := "ACME Inc. Mitgliedsbeitrag 2023"
	wantPayee, wantMemo := "ACME Corporation", "Mitgliedsbeitrag 2023"
	testExtractPayeeAndMemo(t, given, Transfer, wantPayee, wantMemo, false)
}

func TestExtractPayeeAndMemoWithUnknownPayee(t *testing.T) {
	given := "Some Random Person Spende für das Sommerfest"
	wantPayee, wantMemo := "Unknown", "Some Random Person Spende für das Sommerfest"
	testExtractPayeeAndMemo(t, given, Transfer, wantPayee, wantMemo, true)
}

func TestExtractPayeeAndMemoWithFeeTransaction(t *testing.T) {
	given := "Kontoführungsgebühr August 2023"
	wantPayee, wantMemo := "Commerzbank", "Kontoführungsgebühr August 2023"
	testExtractPayeeAndMemo(t, given, Fee, wantPayee, wantMemo, false)
}

func TestExtractPayeeAndMemoWithEmptyString(t *testing.T) {
	given := ""
	wantPayee, wantMemo := "Unknown", ""
	testExtractPayeeAndMemo(t, given, Transfer, wantPayee, wantMemo, true)
}

func TestExtractPayeeAndMemoWithKnownPayeeAndEmptyMemo(t *testing.T) {
	given := "ACME Corporation"
	wantPayee, wantMemo := "ACME Corporation", ""
	testExtractPayeeAndMemo(t, given, Transfer, wantPayee, wantMemo, false)
}

func TestExtractPayeeAndMemoWithExcessWhitespace(t *testing.T) {
	given := "  ACME   \t\n Re. Nr. 123\t\n"
	wantPayee, wantMemo := "ACME Corporation", "Re. Nr. 123"
	testExtractPayeeAndMemo(t, given, Transfer, wantPayee, wantMemo, false)
}

func testExtractPayeeAndMemo(t *testing.T, givenText string, givenCategory Category, wantPayee string, wantMemo string, wantError bool) {
	actualPayee, actualMemo, err := extractPayeeAndMemo(givenText, givenCategory)
	if err != nil {
		if !wantError {
			t.Fatalf("failed to extract payee and Memo: (%q, %q)->(%q, %q, %v)", givenText, givenCategory, actualPayee, actualMemo, err)
		}
	} else {
		if wantError {
			t.Fatalf("expected error but got none: (%q, %v)->(%q, %q, %v)", givenText, givenCategory, actualPayee, actualMemo, err)
		}
	}
	var errs []error
	if actualPayee != wantPayee {
		errs = append(errs, errors.New(fmt.Sprintf("payee mismatch: got %q, want %q", actualPayee, wantPayee)))
	}
	if actualMemo != wantMemo {
		errs = append(errs, errors.New(fmt.Sprintf("memo mismatch: got %q, want %q", actualMemo, wantMemo)))
	}
	if len(errs) != 0 {
		t.Fatalf("ExtractPayeeAndMemo returned unexpected errors: %q", errs)
	}
}
