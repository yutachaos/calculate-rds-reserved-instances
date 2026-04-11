# calculate-rds-reserved-instances

A CLI tool that lists your running RDS DB instances and looks up the matching Reserved Instance offering prices, helping you estimate the cost and count needed before purchasing reserved instances.

## Features

- Queries all running RDS DB instances in your AWS account (with pagination support)
- Looks up Reserved Instance offering prices for each instance class/engine combination
- Outputs CSV-formatted results showing offering ID, count, price, and total amount

## Prerequisites

- Go 1.19+
- AWS credentials configured (via environment variables, `~/.aws/credentials`, or IAM role)

## Installation

### Homebrew (macOS/Linux)

```shell
brew install yutachaos/calculate-rds-reserved-instances/calculate-rds-reserved-instances
```

### Go install

```shell
go install github.com/yutachaos/calculate-rds-reserved-instances@latest
```

### Run directly

```shell
go run main.go
```

## Usage

```shell
calculate-rds-reserved-instances [options]
```

Output is CSV format printed to stdout:

```
InstanceClass,ProductDescription,CurrencyCode,Price,OfferringId,Count,Amount,InstanceIdentifiers
db.r5.xlarge,aurora-mysql,USD,3408,xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,2,6816,[sample-xlarge-cluster-ins-1 sample-cluster-ins-2]
db.r6g.large,aurora-mysql,USD,1525,yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy,1,1525,[sample-large-cluster-ins-1]
```

| Column | Description |
|---|---|
| InstanceClass | DB instance class (e.g. `db.r5.xlarge`) |
| ProductDescription | DB engine (e.g. `aurora-mysql`) |
| CurrencyCode | Currency of the price (e.g. `USD`) |
| Price | Fixed upfront price per instance |
| OfferringId | Reserved DB instance offering ID |
| Count | Number of running instances of this class |
| Amount | Total price (Price × Count) |
| InstanceIdentifiers | List of DB instance identifiers |

## Options

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

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `AWS_DEFAULT_REGION` | AWS region to query | `ap-northeast-1` |