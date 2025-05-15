# Simulation

One of the core features of **yams** is the ability to accurately simulate IAM access decisions.
These simulation APIs revolve around four key inputs we will designate as the "`PARC`" tuple:

- `Principal`: who is making the request
- `Action`: what AWS API call is being made
- `Resource`: the target of the AWS API call
- `Context`: any additional request context (IP, User-Agent, etc)

### Basic Simulation

The most straightforward type of simulation is for a single request, providing all aspects of
`PARC`

Once your data is loaded, basic simulation can be facilitated via the **yams** CLI as follows:

```shell
yams sim
  -principal <principal_arn>
  -action <api_action>
  -resource <resource_arn>
```

**Example**
```shell
yams sim \
  -principal arn:aws:iam::777583092761:role/RedRole \
  -action s3.getobject \
  -resource arn:aws:s3:::yams-magenta/secret.txt
```
```json
{
  "result": "ALLOW",
  "principal": "arn:aws:iam::777583092761:role/RedRole",
  "action": "s3:GetObject",
  "resource": "arn:aws:s3:::yams-magenta/secret.txt"
}
```

### Extended Simulation

A more flexible form of simulation can be achieved by omitting one of the `PAR` identifiers. Doing
so will affect the results of simulation in the following ways:

| Data Provided              | Results |
| -------------------------- | ------- |
| `Principal` and `Action`   | "against which resources can this `Principal` perform this `Action`"
| `Action` and `Resource`    | "which `Principals` can perform this `Action` against this `Resource`"
| `Principal` and `Resource` | "which actions can this `Principal` perform against this `Resource`"

**Example: "which resources?"**
```shell
yams sim \
  -p arn:aws:iam::777583092761:role/RedRole \
  -a s3.getobject
```
```json
[
  "arn:aws:s3:::yams-magenta/*"
]
```

**Example: "which principals?"**
```shell
yams sim \
  -a s3.getobject \
  -r arn:aws:s3:::yams-cyan/secret.txt
```
```json
[
  "arn:aws:iam::777583092761:role/BlueRole"
]
```

**Example: "which actions?"**
```shell
yams sim \
  -p arn:aws:iam::777583092761:role/RedRole \
  -r arn:aws:s3:::yams-magenta/secret.txt
```
```json
[
  "s3:DeleteObject",
  "s3:GetObject",
  "s3:ListBucket",
  "s3:PutObject"
]
```

### Explain & Trace

Beyond knowing the result of an access decision, it is often useful to know _why_ that decision
came to be. **yams** comes with two additional simulation flags which help users to understand
which policies and statements are contributing towards a particular decision.

- `-e/--explain`: provides a concise, human-readable explanation of how the access decision was
  reached
- `-t/--trace`: provides a detailed walkthrough for each step of the policy evaluation

!!! warning

    These flags are only available for basic simulation, as recording and storing the data for
    extended simulation becomes prohibitive at sufficiently large scale

**Example: `explain`**

```shell
yams sim \
  -p arn:aws:iam::777583092761:role/BlueRole \
  -a sns.publish \
  -r arn:aws:sns:us-east-1:213308312933:LemurTopic \
  -explain
```
```json
{
  "result": "ALLOW",
  "principal": "arn:aws:iam::777583092761:role/BlueRole",
  "action": "sns:Publish",
  "resource": "arn:aws:sns:us-east-1:213308312933:LemurTopic",
  "explain": [
    "allow in inline principal policy: 1",
    "[allow] access granted via x-account identity + resource policies"
  ]
}
```

**Example: `trace`**

```shell
yams sim \
  -p arn:aws:iam::777583092761:role/BlueRole \
  -a sns.publish \
  -r arn:aws:sns:us-east-1:213308312933:LemurTopic \
  -trace
```
```json
{
  "result": "ALLOW",
  "principal": "arn:aws:iam::777583092761:role/BlueRole",
  "action": "sns:Publish",
  "resource": "arn:aws:sns:us-east-1:213308312933:LemurTopic",
  "explain": [
    "allow in inline principal policy: 1",
    "[allow] access granted via x-account identity + resource policies"
  ],
  "trace": [
    "(allow) begin: root",
    "  begin: evaluating resource policies",
    "    begin: evaluating policy: 0",
    "      begin: evaluating statement: AllowBlueRole",
    "        begin: evaluating Action",
    "          using Action block",
    "          match: sns:Publish and sns:Publish",
    "        end: evaluating Action",
    "        begin: evaluating Resource",
    "          using Resource block",
  ...
}
```

### Overlays

For more information about overlays, please refer to [Concepts > Overlays](./concepts.md#overlay)

Overlays are supported as input files to the `yams sim` subcommand, allowing you to override
**entity** definitions as needed.

**Example: Using an overlay**
```shell
yams sim \
  -p arn:aws:iam::777583092761:role/RedRole \
  -r arn:aws:s3:::yams-magenta/secret.txt
```
```json
[
  "s3:DeleteObject",
  "s3:GetObject",
  "s3:ListBucket",
  "s3:PutObject"
]
```

```shell
yams principals \
  -k arn:aws:iam::777583092761:role/RedRole \
  > RedRole.json
```
```shell
vim redrole.json  # make some edits, add s3:PutObjectAcl permission
```
```shell
yams sim \
  -p arn:aws:iam::777583092761:role/RedRole \
  -r arn:aws:s3:::yams-magenta/secret.txt \
  -overlay RedRole.json
```
```json
[
  "s3:DeleteObject",
  "s3:GetObject",
  "s3:ListBucket",
  "s3:PutObject",
  "s3:PutObjectAcl"
]
```

!!! note

    When defining an **entity** for an overlay, make sure to use the non-frozen version. Overriding
    managed policy definitions should by overwriting the policy itself

### Entity Autocomplete

To avoid having excessive copy-pasting of ARNs, **yams** will attempt to autocomplete any provided
ARN fragments into full **entity** ARNs.

- If a single match is found, that ARN will be used
- If multiple matches are found, **yams** will return an error

The return value of the simulation will echo the full **Principal** and **Resource** ARNs used

**Example: Successful Autocomplete**

```shell
yams sim \
  -p arn:aws:iam::777583092761:role/BlueRole \
  -a sns.publish \
  -r arn:aws:sns:us-east-1:213308312933:LemurTopic \
```

can be shortened to:

```shell
yams sim \
  -p bluerole \
  -a sns.publish \
  -r lemurtopic
```

**Example: Failed Autocomplete**

```shell
yams sim -p role -a s3.listbucket -r bucket
```
```json
{
  "error": "simulation error: error resolving principal for simulation: too many matches for 'role': [arn:aws:iam::777583092761:role/CoralRole arn:aws:iam::213308312933:role/LionRole arn:aws:iam::777583092761:role/TaupeRole arn:aws:iam::213308312933:role/MouseRole arn:aws:iam::213308312933:role/PandaRole arn:aws:iam::255082776537:role/BurgerRole arn:aws:iam::255082776537:role/SushiRole arn:aws:iam::777583092761:role/BlueRole arn:aws:iam::777583092761:role/RedRole arn:aws:iam::777583092761:role/GreenRole]"
}
```

### FAQ

**Q: How do I simulate API actions without resources?**

A: For actions that do not accept a Resource, you can simply omit it from the command. **yams**
will skip resource-level evaluation and return the simulation result directly. For example,
```shell
yams sim -p arn:aws:iam::777583092761:role/RedRole -a s3:listallmybuckets`
```

**Q: I am getting weird results when simulating a "Create" API operation**

A: "Create" operations are tricky. While we usually do not think of them as being performed
_against_ a resource (after all, they are creating one), IAM authorization does act as if they are.

A common use case for this is limiting a Principal to create resources with a specified prefix,
which means the authorization decision depends heavily

For example, if you want to model whether or not an IAM role can create a bucket, you would
