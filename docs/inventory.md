# Inventory

**yams** provides the ability to query and investigate various sources of IAM policy data in your
environment. This data is populated via [Sources](./data_sources.md) and includes the following
**Entity** types:

- AWS API Actions
- Principals
- Resources
- Policies
- and Accounts

The inventorying commands for each **Entity** type follow the same common pattern:

- `yams <type>` will list all known instances of that **Entity** type
- `yams <type> -q/-query <term>` will do a case-insensitive search using the provided term across
  that Entity type
- `yams <type> -k/-key <name/ARN>` will look up the **Entity** by its primary ID (typically ARN,
  name, or ID); exact match

### Command Aliases

For convenience, inventory commands support short aliases:

| Command | Alias |
|---------|-------|
| `principals` | `p` |
| `resources` | `r` |
| `actions` | `a` |
| `accounts` | `acc` |
| `policies` | `pol` |

### Output Formats

By default, inventory commands output JSON. For human-readable output, use `--format table`:

```shell
yams principals --format table
```
```
TYPE           NAME       ACCOUNT       ARN
------------   --------   -----------   --------------------------------------------
AWS::IAM::Role LionRole   213308312933  arn:aws:iam::213308312933:role/LionRole
AWS::IAM::Role MouseRole  213308312933  arn:aws:iam::213308312933:role/MouseRole
AWS::IAM::User CatUser    213308312933  arn:aws:iam::213308312933:user/CatUser
...
```

You can set the default format in your config file (`~/.config/yams/config.json`):
```json
{
  "format": "table"
}
```

### AWS API Actions

##### List

```shell
yams actions
```
```json
[
  "a2c:GetContainerizationJobDetails",
  "a2c:GetDeploymentJobDetails",
  "a2c:StartContainerizationJob",
  "a2c:StartDeploymentJob",
  "a4b:ApproveSkill",
  "a4b:AssociateContactWithAddressBook",
  "a4b:AssociateDeviceWithNetworkProfile",
  "a4b:AssociateDeviceWithRoom",
  "a4b:AssociateSkillGroupWithRoom",
  ...
]
```

##### Search

```shell
yams actions -q networkinterface
```
```json
[
  "ec2:AttachNetworkInterface",
  "ec2:CreateNetworkInterface",
  "ec2:CreateNetworkInterfacePermission",
  "ec2:DeleteNetworkInterface",
  "ec2:DeleteNetworkInterfacePermission",
  "ec2:DescribeNetworkInterfaceAttribute",
  "ec2:DescribeNetworkInterfacePermissions",
  "ec2:DescribeNetworkInterfaces",
  "ec2:DetachNetworkInterface",
  ...
]
```

##### Lookup

```shell
yams actions -k dynamodb:PutItem

// also valid
yams actions -k dynamodb.putitem
```
```json
{
  "Name": "PutItem",
  "Service": "dynamodb",
  "ActionConditionKeys": [
    "dynamodb:attributes",
    "dynamodb:enclosingoperation",
    "dynamodb:leadingkeys",
    "dynamodb:returnconsumedcapacity",
    "dynamodb:returnvalues"
  ],
  "ResolvedResources": [
    {
      "Name": "table",
      "ARNFormats": [
        "arn:*:dynamodb:*:*:table/*"
      ],
      "ConditionKeys": [
        "aws:resourcetag"
      ]
    }
  ]
}
```

### Principals

##### List

```shell
yams principals
```
```json
[
  "arn:aws:iam::213308312933:role/LionRole",
  "arn:aws:iam::213308312933:role/MouseRole",
  "arn:aws:iam::213308312933:role/PandaRole",
  "arn:aws:iam::213308312933:user/CatUser",
  "arn:aws:iam::213308312933:user/DogUser",
  "arn:aws:iam::213308312933:user/FishUser",
  "arn:aws:iam::255082776537:role/BurgerRole",
  "arn:aws:iam::255082776537:role/NoodleRole",
  "arn:aws:iam::255082776537:role/PizzaRole",
  ...
]
```

##### Search

```shell
yams principals -q cat
```
```json
[
  "arn:aws:iam::213308312933:user/CatUser"
]
```

##### Lookup

```shell
yams principals -k arn:aws:iam::213308312933:user/CatUser
```
```json
{
  "Type": "AWS::IAM::User",
  "AccountId": "213308312933",
  "Name": "CatUser",
  "Arn": "arn:aws:iam::213308312933:user/CatUser",
  ...
}
```

### Resources

##### List

```shell
yams resources
```
```json
[
  "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
  "arn:aws:dynamodb:us-east-1:255082776537:table/TacoTable",
  "arn:aws:dynamodb:us-east-1:777583092761:table/NavyTable",
  "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
  "arn:aws:iam::213308312933:policy/yams-test-infra-DogPolicy-pX0mgCedLaeo",
  "arn:aws:iam::213308312933:policy/yams-test-infra-LlamaBoundary-mvVoctsE53pG",
  "arn:aws:iam::213308312933:role/LionRole",
  "arn:aws:iam::213308312933:role/MouseRole",
  "arn:aws:iam::213308312933:role/PandaRole",
  ...
]
```

##### Search

```shell
yams resources -q s3
```
```json
[
  "arn:aws:s3:::banana-bucket-255082776537",
  "arn:aws:s3:::crocodile-bucket-213308312933",
  "arn:aws:s3:::peach-bucket-777583092761",
  "arn:aws:s3:::yams-bear",
  "arn:aws:s3:::yams-cyan",
  "arn:aws:s3:::yams-green",
  "arn:aws:s3:::yams-magenta"
]
```

##### Lookup

```shell
yams resources -k arn:aws:s3:::yams-cyan
```
```json
{
  "Type": "AWS::S3::Bucket",
  "AccountId": "777583092761",
  "Region": "us-east-1",
  "Name": "yams-cyan",
  "Arn": "arn:aws:s3:::yams-cyan",
  "Policy": {
    "Version": "2012-10-17",
    "Id": "",
    "Statement": [
      {
        "Sid": "",
        "Effect": "Deny",
        "Principal": "*",
        "Action": [
          "s3:listbucket",
          "s3:getobject"
        ],
        "Resource": [
          "arn:aws:s3:::yams-cyan",
          "arn:aws:s3:::yams-cyan/*"
        ],
        "Condition": {
          "StringNotEquals": {
            "aws:PrincipalTag/Color": "Blue"
          }
        }
      }
    ]
  },
  ...
}
```

### Policies

##### List

```shell
yams policies
```
```json
[
  "arn:aws:iam::213308312933:policy/yams-test-infra-DogPolicy-pX0mgCedLaeo",
  "arn:aws:iam::213308312933:policy/yams-test-infra-LlamaBoundary-mvVoctsE53pG",
  "arn:aws:iam::255082776537:policy/yams-test-infra-CupcakeBoundary-udxeJjTH6ebJ",
  "arn:aws:iam::255082776537:policy/yams-test-infra-SaladPolicy-opE0edVZrSWR",
  "arn:aws:iam::255082776537:policy/yams-test-infra-SoupPolicy-65QKm40EPh1y",
  "arn:aws:iam::777583092761:policy/yams-test-infra-GreyPolicy-gLf7j3ZwJYBm",
  "arn:aws:iam::777583092761:policy/yams-test-infra-MustardBoundary-47JW6znulEXt",
  "arn:aws:iam::777583092761:policy/yams-test-infra-PinkBoundary-xuINwerkCuZ3",
  "arn:aws:iam::aws:policy/AIOpsAssistantPolicy",
  ...
]
```

##### Search

```shell
yams policies -q test
```
```json
[
  "arn:aws:iam::213308312933:policy/yams-test-infra-DogPolicy-pX0mgCedLaeo",
  "arn:aws:iam::213308312933:policy/yams-test-infra-LlamaBoundary-mvVoctsE53pG",
  "arn:aws:iam::255082776537:policy/yams-test-infra-CupcakeBoundary-udxeJjTH6ebJ",
  "arn:aws:iam::255082776537:policy/yams-test-infra-SaladPolicy-opE0edVZrSWR",
  "arn:aws:iam::255082776537:policy/yams-test-infra-SoupPolicy-65QKm40EPh1y",
  "arn:aws:iam::777583092761:policy/yams-test-infra-GreyPolicy-gLf7j3ZwJYBm",
  "arn:aws:iam::777583092761:policy/yams-test-infra-MustardBoundary-47JW6znulEXt",
  "arn:aws:iam::777583092761:policy/yams-test-infra-PinkBoundary-xuINwerkCuZ3",
  "arn:aws:iam::aws:policy/AWSIoTDeviceTesterForFreeRTOSFullAccess",
  ...
]
```

##### Lookup

```shell
yams policies -k arn:aws:iam::213308312933:policy/yams-test-infra-DogPolicy-pX0mgCedLaeo
```
```json
{
  "Type": "AWS::IAM::Policy",
  "AccountId": "213308312933",
  "Arn": "arn:aws:iam::213308312933:policy/yams-test-infra-DogPolicy-pX0mgCedLaeo",
  "Name": "yams-test-infra-DogPolicy-pX0mgCedLaeo",
  "Policy": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
      }
    ]
  }
}
```

### Accounts

##### List

```shell
yams accounts
```
```json
[
  "213308312933",
  "255082776537",
  "777583092761",
  "810970970902"
]
```

##### Search

```shell
yams accounts -q 213
```
```json
[
  "213308312933"
]
```

##### Lookup

```shell
yams accounts -k 213308312933
```
```json
{
  "Id": "213308312933",
  "Name": "yams1",
  "OrgId": "o-9hmw0uhxs4",
  "OrgPaths": [
    "o-9hmw0uhxs4/r-m4x4/",
    "o-9hmw0uhxs4/r-m4x4/ou-m4x4-onrzr6t1/"
  ],
  "OrgNodes": ...
}
```
