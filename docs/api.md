# API Reference

The following documentation provides examples of integrating with the HTTP API for **yams**.

### Healthcheck API

`GET /api/v1/healthcheck`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/healthcheck
```
```shell
OK
```

### Status API

`GET /api/v1/status`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/status
```
```json
{
  "accounts": 4,
  "entities": 1448,
  "groups": 0,
  "policies": 1371,
  "principals": 20,
  "resources": 53,
  "sources": [
    {
      "source": "testdata/real-world/awsconfig.jsonl",
      "updated": "2025-03-15T15:04:35.173468943-07:00"
    },
    {
      "source": "testdata/real-world/org.jsonl",
      "updated": "2025-03-15T15:04:35.173687682-07:00"
    }
  ]
}
```

### Actions API

**List**
`GET /api/v1/actions`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/actions
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

**Lookup**
`GET /api/v1/actions/{key...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/actions/dynamodb.putitem
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

**Search**
`GET /api/v1/actions/search/{search...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/actions/search/networkinterface
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

### Principals API

**List**
`GET /api/v1/principals`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/principals
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

**Lookup**
`GET /api/v1/principals/{key...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/principals/arn:aws:iam::213308312933:user/CatUser
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

**Search**
`GET /api/v1/principals/search/{search...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/principals/search/cat
```
```json
[
  "arn:aws:iam::213308312933:user/CatUser"
]
```

### Resources API

**List**
`GET /api/v1/resources`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/resources
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

**Lookup**
`GET /api/v1/resources/{key...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/resources/arn:aws:s3:::yams-cyan
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

**Search**
`GET /api/v1/resources/search/{search...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/resources/search/s3
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

### Policies API

**List**
`GET /api/v1/policies`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/policies
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

**Lookup**
`GET /api/v1/policies/{key...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/policies/arn:aws:iam::213308312933:policy/yams-test-infra-DogPolicy-pX0mgCedLaeo
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

**Search**
`GET /api/v1/policies/search/{search...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/policies/search/test
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

### Accounts API

**List**
`GET /api/v1/accounts`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/accounts
```
```json
[
  "213308312933",
  "255082776537",
  "777583092761",
  "810970970902"
]
```

**Lookup**
`GET /api/v1/accounts/{key...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/accounts/213308312933
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

**Search**
`GET /api/v1/accounts/search/{search...}`
```shell
curl ${YAMS_SERVER_ADDRESS}/api/v1/accounts/search/213
```
```json
[
  "213308312933"
]
```

### Basic Simulation

`POST /api/v1/sim`
```shell
curl -X POST ${YAMS_SERVER_ADDRESS}/api/v1/sim -d '{
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "sns:publish",
  "resource": "arn:aws:sns:us-east-1:777583092761:PurpleTopic"
}'
```
```json
{
  "result": "DENY",
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "sns:Publish",
  "resource": "arn:aws:sns:us-east-1:777583092761:PurpleTopic"
}
```

### Extended Simulation

**Which Principals?**
`POST /api/v1/sim/whichPrincipals`
```shell
curl -X POST ${YAMS_SERVER_ADDRESS}/api/v1/sim/whichPrincipals -d '{
  "action": "sns:publish",
  "resource": "arn:aws:sns:us-east-1:777583092761:PurpleTopic"
}'
```
```json
[
  "arn:aws:iam::777583092761:role/BlueRole"
]
```

**Which Resources?**
`POST /api/v1/sim/whichResources`
```shell
curl -X POST ${YAMS_SERVER_ADDRESS}/api/v1/sim/whichResources -d '{
  "principal": "arn:aws:iam::777583092761:role/BlueRole",
  "action": "sns:publish"
}'
```
```json
[
  "arn:aws:sns:us-east-1:213308312933:LemurTopic",
  "arn:aws:sns:us-east-1:777583092761:PurpleTopic"
]
```

**Which Actions?**
`POST /api/v1/sim/whichActions`
```shell
curl -X POST ${YAMS_SERVER_ADDRESS}/api/v1/sim/whichActions -d '{
  "principal": "arn:aws:iam::777583092761:role/BlueRole",
  "resource": "arn:aws:sns:us-east-1:777583092761:PurpleTopic"
}'
```
```json
[
  "sns:AddPermission",
  "sns:ConfirmSubscription",
  "sns:CreateTopic",
  "sns:DeleteTopic",
  "sns:GetDataProtectionPolicy",
  "sns:GetTopicAttributes",
  "sns:ListSubscriptionsByTopic",
  "sns:ListTagsForResource",
  "sns:Publish",
  "sns:PutDataProtectionPolicy",
  "sns:RemovePermission",
  "sns:SetTopicAttributes",
  "sns:Subscribe",
  "sns:TagResource",
  "sns:UntagResource"
]
```

### Explain & Trace

`POST /api/v1/sim`
```shell
curl -X POST ${YAMS_SERVER_ADDRESS}/api/v1/sim -d '{
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "s3.GetObject",
  "resource": "arn:aws:s3:::yams-cyan/foo.txt",
  "explain": true
}'
```
```json
{
  "result": "DENY",
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "s3:GetObject",
  "resource": "arn:aws:s3:::yams-cyan/foo.txt",
  "explain": [
    "[explicit deny] in resource policy"
  ]
}
```

### Overlays

`POST /api/v1/sim`
```shell
curl -X POST ${YAMS_SERVER_ADDRESS}/api/v1/sim -d '{
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "s3.GetObject",
  "resource": "arn:aws:s3:::yams-cyan/foo.txt",
  "overlay": {
    "principals": [
      {
        "Type": "AWS::IAM::Role",
        "AccountId": "777583092761",
        "Name": "RedRole",
        "Arn": "arn:aws:iam::777583092761:role/RedRole",
        "Tags": [
          {
            "Key": "is-yams-test-resource",
            "Value": "true"
          },
          {
            "Key": "Color",
            "Value": "Blue"
          }
        ],
        "InlinePolicies": [
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Effect": "Allow",
                "Action": "sts:assumerole",
                "Resource": "*"
              }
            ]
          },
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Effect": "Deny",
                "Action": "dynamodb:*",
                "Resource": "*"
              }
            ]
          },
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Effect": "Allow",
                "Action": "s3:ListAllMyBuckets",
                "Resource": "*"
              },
              {
                "Effect": "Allow",
                "Action": [
                  "s3:GetObject",
                  "s3:PutObject",
                  "s3:DeleteObject",
                  "s3:ListBucket"
                ],
                "Resource": [
                  "arn:aws:s3:::yams-*",
                  "arn:aws:s3:::yams-*/*"
                ]
              }
            ]
          }
        ],
        "AttachedPolicies": null,
        "Groups": null
      }
    ]
  }
}'
```
```json
{
  "result": "ALLOW",
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "s3:GetObject",
  "resource": "arn:aws:s3:::yams-cyan/foo.txt"
}
```
