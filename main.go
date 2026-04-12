package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type instanceInfo struct {
	Count               int
	ProductDescription  string
	InstanceIdentifiers []string
}

func main() {
	duration := flag.String("duration", "1", "reserved purchase duration| 1 | 3 | 31536000 | 94608000")
	multiAz := flag.Bool("multiaz", false, "multiaz: true:false")
	offeringType := flag.String("offeringType", "All Upfront", "offeringType: Partial Upfront|All Upfront|No Upfront")
	profile := flag.String("profile", "", "profile: aws profile name")
	region := flag.String("region", "", "region: aws region name")

	flag.Parse()

	resolvedRegion := *region
	if resolvedRegion == "" {
		resolvedRegion = os.Getenv("AWS_DEFAULT_REGION")
	}
	if resolvedRegion == "" {
		resolvedRegion = "ap-northeast-1"
	}

	awsConfig := &aws.Config{
		Region: aws.String(resolvedRegion),
	}

	if *profile != "" {
		awsConfig.Credentials = credentials.NewSharedCredentials("", *profile)
	}
	sess := session.Must(session.NewSession(awsConfig))

	err := extractRdsReservedInstances(rds.New(sess), os.Stdout, duration, multiAz, offeringType)
	if err != nil {
		log.Fatal(err)
	}
}

func extractRdsReservedInstances(rdsSvc rdsiface.RDSAPI, out io.Writer, duration *string, multiAz *bool, offeringType *string) error {
	instanceList := map[string]*instanceInfo{}

	err := rdsSvc.DescribeDBInstancesPages(&rds.DescribeDBInstancesInput{},
		func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
			for _, instance := range page.DBInstances {
				if instanceList[*instance.DBInstanceClass] == nil {
					instanceList[*instance.DBInstanceClass] = &instanceInfo{
						Count:               1,
						InstanceIdentifiers: []string{*instance.DBInstanceIdentifier},
						ProductDescription:  *instance.Engine,
					}
				} else {
					instanceList[*instance.DBInstanceClass].Count++
					instanceList[*instance.DBInstanceClass].InstanceIdentifiers = append(instanceList[*instance.DBInstanceClass].InstanceIdentifiers, *instance.DBInstanceIdentifier)
				}
			}
			return true
		},
	)
	if err != nil {
		return err
	}

	w := csv.NewWriter(out)
	w.Write([]string{"InstanceClass", "ProductDescription", "CurrencyCode", "OfferingId", "Price", "Count", "Amount", "InstanceIdentifiers"})

	for instanceClass, instance := range instanceList {
		err = rdsSvc.DescribeReservedDBInstancesOfferingsPages(
			&rds.DescribeReservedDBInstancesOfferingsInput{
				DBInstanceClass:    &instanceClass,
				Duration:           duration,
				ProductDescription: &instance.ProductDescription,
				MultiAZ:            multiAz,
				OfferingType:       offeringType,
			},
			func(page *rds.DescribeReservedDBInstancesOfferingsOutput, lastPage bool) bool {
				for _, offering := range page.ReservedDBInstancesOfferings {
					amount := *offering.FixedPrice * float64(instance.Count)
					w.Write([]string{
						*offering.DBInstanceClass,
						*offering.ProductDescription,
						*offering.CurrencyCode,
						*offering.ReservedDBInstancesOfferingId,
						fmt.Sprintf("%v", *offering.FixedPrice),
						fmt.Sprintf("%d", instance.Count),
						fmt.Sprintf("%v", amount),
						strings.Join(instance.InstanceIdentifiers, " "),
					})
				}
				return true
			},
		)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
