package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
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
	region := os.Getenv("AWS_DEFAULT_REGION")

	flag.Parse()
	if region == "" {
		region = "ap-northeast-1"
	}
	awsConfig := &aws.Config{
		Region: aws.String(region),
	}

	if *profile != "" {
		awsConfig.Credentials = credentials.NewSharedCredentials("", *profile)
	}
	sess := session.Must(session.NewSession(awsConfig))

	err := extractRdsReservedInstances(sess, duration, multiAz, offeringType)
	if err != nil {
		log.Fatal(err)
	}

}

func extractRdsReservedInstances(sess *session.Session, duration *string, multiAz *bool, offeringType *string) (err error) {
	rdsSvc := rds.New(
		sess,
	)
	instanceList := map[string]*instanceInfo{}

	// TODO pagination
	instances, err := rdsSvc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		return err
	}

	for _, instance := range instances.DBInstances {
		if instanceList[*instance.DBInstanceClass] == nil {
			instanceList[*instance.DBInstanceClass] = &instanceInfo{
				Count:               1,
				InstanceIdentifiers: []string{*instance.DBInstanceIdentifier},
				ProductDescription:  *instance.Engine,
			}
		} else {
			instanceList[*instance.DBInstanceClass].Count = instanceList[*instance.DBInstanceClass].Count + 1
			instanceList[*instance.DBInstanceClass].InstanceIdentifiers = append(instanceList[*instance.DBInstanceClass].InstanceIdentifiers, *instance.DBInstanceIdentifier)
		}

	}

	fmt.Println("InstanceClass,ProductDescription,CurrencyCode,Price,OfferringId,Count,Amount,InstanceIdentifiers")
	for instanceClass, instance := range instanceList {
		// TODO pagination
		offerings, err := rdsSvc.DescribeReservedDBInstancesOfferings(&rds.DescribeReservedDBInstancesOfferingsInput{
			DBInstanceClass:    &instanceClass,
			Duration:           duration,
			ProductDescription: &instance.ProductDescription,
			MultiAZ:            multiAz,
			OfferingType:       offeringType,
		},
		)
		if err != nil {
			return err
		}
		for _, offering := range offerings.ReservedDBInstancesOfferings {
			fmt.Printf("%v,%v,%v,%v,%v,%v,%v,%v\n", *offering.DBInstanceClass, *offering.ProductDescription, *offering.CurrencyCode, *offering.ReservedDBInstancesOfferingId, *offering.FixedPrice, instance.Count, *offering.FixedPrice*float64(instance.Count), instance.InstanceIdentifiers)
		}
	}
	return err
}
