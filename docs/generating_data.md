# Generating Data

The [data sources](./data_sources.md) used by **yams** can be created by any tooling adhering to the
schema and naming conventions. For convenience, **yams** itself can also serve as a generator of
this data.

At this time, **yams** supports the generation of two types of data:

- **Entities**: principals, resources, policies, etc
- **Org Data**: accounts, OUs, SCPs, etc

### Entity Data

##### Using yams

!!! note

    Using **yams** to generate **Entity** data requires providing the program with credentials to
    perform `config:SelectAggregateResourceConfig` in the same AWS account as the AWS Config
    Aggregator

**Entity** data can be generated via the `yams dump -t config` subcommand, which requires specifying
both:

- `-a/-aggregator` the name of the AWS Config aggregator to use
- `-r/-rtype` the AWS Config resource type(s) that you would like to dump (allows multiple!)


###### Examples

> Dump all known SQS Queues to stdout:
```shell
yams dump -t config \
  -a my-aggregator \
  -r AWS::SQS::Queue
```

> Dump all IAM **Entities** to a file:
```shell
yams dump -t config \
  -a my-aggregator \
  -r AWS::IAM::Role \
  -r AWS::IAM::User \
  -r AWS::IAM::Group \
  -r AWS::IAM::Policy \
  -o out.json
```

> Dump all IAM **Entities** and key resources to an S3 bucket; compressed:
```shell
yams dump -t config \
  -a my-aggregator \
  -r AWS::IAM::Role \
  -r AWS::IAM::User \
  -r AWS::IAM::Group \
  -r AWS::IAM::Policy \
  -r AWS::S3::Bucket \
  -r AWS::SQS::Queue \
  -r AWS::SNS::Topic \
  -r AWS::DynamoDB::Table \
  -r AWS::KMS::Key \
  -o s3://my-bucket/resources.json.gz
```

##### Alternatives

You can also use basic command-line tools such as `awscli` and `jq` to construct valid **Sources**
with highly customized subsets of data:

```shell
aws configservice select-aggregate-resource-config \
--configuration-aggregator-name my-aggregator \
--expression "SELECT *, configuration, supplementaryConfiguration, tags WHERE ..." \
| jq -c '.Results[] | fromjson' \
> resources.jsonl
```

### Org Data

##### Using yams

!!! note

    Using **yams** to generate **Org** data requires providing the program with credentials to
    access read-only `organizations` APIs in the org master account

**Org** data can be generated via the `yams dump -t org` subcommand.

###### Examples

> Dump org data to stdout:
```shell
yams dump -t org
```

> Dump org data to an S3 bucket; compressed:
```shell
yams dump -t org \
  -o s3://my-bucket/org.json.gz
```

##### Alternatives

At this time, there are no valid alternatives to generating Org data, due to a lack of standard
schema and definition for these entity types.
