# hermes

## Motivation

 In order to reduce AWS cost,
 It is necessary to effectively buy Reserved Instances.
 But AWS pricing is complicated and difficult.
 This library shows the RI that you should buy now,
 based on the future instance usage and the current RI purchase.

## Install

```sh
go install github.com/itsubaki/hermes@latest
```

## Required

```sh
# set aws credential "example" with iam policy "hermes"

$ cat ~/.aws/credentials
[example]
aws_access_key_id = ********************
aws_secret_access_key = ****************************************
```

```sh
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "hermes",
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeReserved*",
        "rds:DescribeReserved*",
        "elasticache:DescribeReserved*",
        "redshift:DescribeReserved*",
        "organizations:List*",
        "organizations:Describe*",
        "ce:Get*"
      ],
      "Resource": "*"
    }
  ]
}
```

## CommandLine Example

```sh
$ AWS_PROFILE=example hermes fetch
write: /var/tmp/hermes/pricing/ap-northeast-1.out
write: /var/tmp/hermes/pricing/us-west-2.out
write: /var/tmp/hermes/usage/2019-08.out
write: /var/tmp/hermes/usage/2019-07.out
write: /var/tmp/hermes/usage/2019-06.out
write: /var/tmp/hermes/usage/2019-04.out
write: /var/tmp/hermes/usage/2019-03.out
write: /var/tmp/hermes/usage/2019-02.out
write: /var/tmp/hermes/usage/2019-01.out
write: /var/tmp/hermes/usage/2018-12.out
write: /var/tmp/hermes/usage/2018-11.out
write: /var/tmp/hermes/usage/2018-10.out
write: /var/tmp/hermes/usage/2018-09.out
```

```sh
$ hermes pricing | jq .
[
  {
    "Version": "20190730012138",
    "SKU": "PDMPNVN5SPA5HWHH",
    "OfferTermCode": "6QCMYABX3D",
    "Region": "ap-northeast-1",
    "InstanceType": "ds1.8xlarge",
    "UsageType": "APN1-Node:dw.hs1.8xlarge",
    "LeaseContractLength": "1yr",
    "PurchaseOption": "All Upfront",
    "OnDemand": 9.52,
    "ReservedQuantity": 49020,
    "ReservedHrs": 0,
    "Tenancy": "",
    "PreInstalled": "",
    "OperatingSystem": "",
    "Operation": "RunComputeNode:0001",
    "CacheEngine": "",
    "DatabaseEngine": "",
    "OfferingClass": "standard",
    "NormalizationSizeFactor": ""
  }
  ...
]
```

```sh
$ hermes usage | jq .
[
  {
    "account_id": "123456789012",
    "description": "example",
    "region": "us-west-2",
    "usage_type": "USW2-NodeUsage:cache.t2.small",
    "cache_engine": "Redis",
    "date": "2019-08",
    "instance_hour": 101,
    "instance_num": 0.135752688172043
  }
  ...
]
```

```sh
$ hermes usage --format csv | column -t -s, | less -S
```

```sh
$ cat purchase.json | hermes recommend | jq .
{
  "region": "ap-northeast-1",
  "usage_type": "APN1-BoxUsage:c4.large",
  "platform": "Linux/UNIX",
  "instance_num": 1648
}
```
