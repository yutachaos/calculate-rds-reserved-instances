package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func main() {
	sess := session.Must(session.NewSession())
	durarion := flag.String("durarion", "1", "reserved purchase duration")
	productdescription := flag.String("productdescription", "aurora-mysql", "DB type: e.g:aurora-mysql")
	multiAz := flag.Bool("multiaz", false, "multiaz: true:false")
	offeringType := flag.String("offeringType", "All Upfront", "offeringType: Partial Upfront|All Upfront|No Upfront")

	region := os.Getenv("AWS_DEFAULT_REGION")

	if region == "" {
		region = "ap-northeast-1"
	}

	svc := rds.New(
		sess,
		aws.NewConfig().WithRegion(region),
	)
	// TODO pagination
	instances, err := svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Instance identifer,Instance Class,CurrencyCode,Price")
	for _, instance := range instances.DBInstances {
		// TODO pagination
		offerings, err := svc.DescribeReservedDBInstancesOfferings(&rds.DescribeReservedDBInstancesOfferingsInput{
			DBInstanceClass:    instance.DBInstanceClass,
			Duration:           durarion,
			ProductDescription: productdescription,
			MultiAZ:            multiAz,
			OfferingType:       offeringType,
		},
		)
		if err != nil {
			log.Fatal(err)
		}
		for _, offering := range offerings.ReservedDBInstancesOfferings {
			// fmt.Printf("offering: %v \n", offering)
			fmt.Printf("%v,%v,%v,%v\n", *instance.DBInstanceIdentifier, *offering.DBInstanceClass, *offering.CurrencyCode, *offering.FixedPrice)
		}
	}
}
