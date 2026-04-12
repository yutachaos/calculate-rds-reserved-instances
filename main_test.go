package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type mockRDSClient struct {
	rdsiface.RDSAPI
	dbInstances  []*rds.DBInstance
	offerings    []*rds.ReservedDBInstancesOffering
	describeErr  error
	offeringsErr error
}

func (m *mockRDSClient) DescribeDBInstancesPages(_ *rds.DescribeDBInstancesInput, fn func(*rds.DescribeDBInstancesOutput, bool) bool) error {
	if m.describeErr != nil {
		return m.describeErr
	}
	fn(&rds.DescribeDBInstancesOutput{DBInstances: m.dbInstances}, true)
	return nil
}

func (m *mockRDSClient) DescribeReservedDBInstancesOfferingsPages(_ *rds.DescribeReservedDBInstancesOfferingsInput, fn func(*rds.DescribeReservedDBInstancesOfferingsOutput, bool) bool) error {
	if m.offeringsErr != nil {
		return m.offeringsErr
	}
	fn(&rds.DescribeReservedDBInstancesOfferingsOutput{ReservedDBInstancesOfferings: m.offerings}, true)
	return nil
}

func parseCSV(t *testing.T, output string) [][]string {
	t.Helper()
	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV output: %v", err)
	}
	return records
}

func strPtr(s string) *string { return aws.String(s) }
func f64Ptr(f float64) *float64 { return aws.Float64(f) }
func boolPtr(b bool) *bool { return aws.Bool(b) }

func defaultArgs() (duration, offeringType *string, multiAz *bool) {
	return strPtr("1"), strPtr("All Upfront"), boolPtr(false)
}

func TestSingleInstance(t *testing.T) {
	mock := &mockRDSClient{
		dbInstances: []*rds.DBInstance{
			{DBInstanceClass: strPtr("db.r5.xlarge"), DBInstanceIdentifier: strPtr("my-instance"), Engine: strPtr("aurora-mysql")},
		},
		offerings: []*rds.ReservedDBInstancesOffering{
			{
				DBInstanceClass:                strPtr("db.r5.xlarge"),
				ProductDescription:             strPtr("aurora-mysql"),
				CurrencyCode:                   strPtr("USD"),
				ReservedDBInstancesOfferingId:  strPtr("offer-abc"),
				FixedPrice:                     f64Ptr(3408),
			},
		},
	}

	var buf bytes.Buffer
	duration, offeringType, multiAz := defaultArgs()
	if err := extractRdsReservedInstances(mock, &buf, duration, multiAz, offeringType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, buf.String())
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1 data), got %d", len(records))
	}

	header := records[0]
	expected := []string{"InstanceClass", "ProductDescription", "CurrencyCode", "OfferingId", "Price", "Count", "Amount", "InstanceIdentifiers"}
	for i, col := range expected {
		if header[i] != col {
			t.Errorf("header[%d]: got %q, want %q", i, header[i], col)
		}
	}

	row := records[1]
	if row[0] != "db.r5.xlarge" {
		t.Errorf("InstanceClass: got %q, want %q", row[0], "db.r5.xlarge")
	}
	if row[3] != "offer-abc" {
		t.Errorf("OfferingId: got %q, want %q", row[3], "offer-abc")
	}
	if row[4] != "3408" {
		t.Errorf("Price: got %q, want %q", row[4], "3408")
	}
	if row[5] != "1" {
		t.Errorf("Count: got %q, want %q", row[5], "1")
	}
	if row[6] != "3408" {
		t.Errorf("Amount: got %q, want %q", row[6], "3408")
	}
	if row[7] != "my-instance" {
		t.Errorf("InstanceIdentifiers: got %q, want %q", row[7], "my-instance")
	}
}

func TestMultipleInstancesSameClass(t *testing.T) {
	mock := &mockRDSClient{
		dbInstances: []*rds.DBInstance{
			{DBInstanceClass: strPtr("db.r5.xlarge"), DBInstanceIdentifier: strPtr("instance-1"), Engine: strPtr("aurora-mysql")},
			{DBInstanceClass: strPtr("db.r5.xlarge"), DBInstanceIdentifier: strPtr("instance-2"), Engine: strPtr("aurora-mysql")},
		},
		offerings: []*rds.ReservedDBInstancesOffering{
			{
				DBInstanceClass:               strPtr("db.r5.xlarge"),
				ProductDescription:            strPtr("aurora-mysql"),
				CurrencyCode:                  strPtr("USD"),
				ReservedDBInstancesOfferingId: strPtr("offer-abc"),
				FixedPrice:                    f64Ptr(3408),
			},
		},
	}

	var buf bytes.Buffer
	duration, offeringType, multiAz := defaultArgs()
	if err := extractRdsReservedInstances(mock, &buf, duration, multiAz, offeringType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, buf.String())
	if len(records) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(records))
	}

	row := records[1]
	if row[5] != "2" {
		t.Errorf("Count: got %q, want %q", row[5], "2")
	}
	if row[6] != "6816" {
		t.Errorf("Amount: got %q, want %q", row[6], "6816")
	}
	identifiers := row[7]
	if !strings.Contains(identifiers, "instance-1") || !strings.Contains(identifiers, "instance-2") {
		t.Errorf("InstanceIdentifiers %q should contain both instance-1 and instance-2", identifiers)
	}
}

func TestNoInstances(t *testing.T) {
	mock := &mockRDSClient{
		dbInstances: []*rds.DBInstance{},
	}

	var buf bytes.Buffer
	duration, offeringType, multiAz := defaultArgs()
	if err := extractRdsReservedInstances(mock, &buf, duration, multiAz, offeringType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, buf.String())
	if len(records) != 1 {
		t.Fatalf("expected only header row, got %d rows", len(records))
	}
}

func TestDescribeDBInstancesError(t *testing.T) {
	mock := &mockRDSClient{
		describeErr: errors.New("AWS API error"),
	}

	var buf bytes.Buffer
	duration, offeringType, multiAz := defaultArgs()
	err := extractRdsReservedInstances(mock, &buf, duration, multiAz, offeringType)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDescribeOfferingsError(t *testing.T) {
	mock := &mockRDSClient{
		dbInstances: []*rds.DBInstance{
			{DBInstanceClass: strPtr("db.r5.xlarge"), DBInstanceIdentifier: strPtr("my-instance"), Engine: strPtr("aurora-mysql")},
		},
		offeringsErr: errors.New("offerings API error"),
	}

	var buf bytes.Buffer
	duration, offeringType, multiAz := defaultArgs()
	err := extractRdsReservedInstances(mock, &buf, duration, multiAz, offeringType)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
