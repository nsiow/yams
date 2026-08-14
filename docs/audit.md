# Audit

The `audit` command evaluates configured actions across every matching resource and principal in
the supplied data sources. Audit runs locally against those sources and does not call the yams
server.

```shell
yams audit \
  --source config.json.gz \
  --source org.json.gz \
  --config misc/audit-config.json \
  --out audit.csv
```

The default `csv` format preserves action-level results:

```text
resource,action,principal
```

## Logical Action Groups

Actions can instead be mapped into logical groups. An action may belong to only one group within
a resource entry.

```json
[
  {
    "resource_type": "AWS::SQS::Queue",
    "action_groups": [
      {
        "name": "READ",
        "actions": ["sqs:ReceiveMessage"]
      },
      {
        "name": "WRITE",
        "actions": ["sqs:SendMessage", "sqs:DeleteMessage"]
      }
    ]
  }
]
```

Use `--format grouped-csv` for one row per resource, group, and allowed principal:

```text
resource,group,principal
arn:aws:sqs:us-east-1:111111111111:example,READ,arn:aws:iam::111111111111:role/reader
```

Use `--format grouped-jsonl` for one record per resource and group:

```json
{"resource":"arn:aws:sqs:us-east-1:111111111111:example","group":"READ","principals":[]}
```

A principal belongs to a group when at least one action in that group is allowed. Principals
allowed by multiple actions in a group appear once. Resources with no allowed principals are
retained: grouped JSONL uses an empty array and grouped CSV uses an empty principal field.

Grouped output is deterministic and uses bounded resource batches. Adjust the default batch size
of 256 with `--resource-batch-size` to trade aggregation memory for simulation overhead.
