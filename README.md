# calculate-rds-reserved-instances

Calculate the number and price of rds reserved instances

# Usage
- go run main.go
  InstanceClass,ProductDescription,CurrencyCode,Price,Count,Amount,InstanceIdentifiers
  db.r5.xlarge,aurora-mysql,USD,3408,2,6816,[sample-xlarge-cluster-ins-1 sample-cluster-ins-2]
  db.r6g.large,aurora-mysql,USD,1525,1,1525,[sample-large-cluster-ins-1]

# Option

```shell
Usage of calculate-rds-reserved-instances:
  -duration string
    	reserved purchase duration| 1 | 3 | 31536000 | 94608000 (default "1")
  -multiaz
    	multiaz: true:false
  -offeringType string
    	offeringType: Partial Upfront|All Upfront|No Upfront (default "All Upfront")
  -profile string
    	profile: aws profile name
```