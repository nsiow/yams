package sim

import (
	"os"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

func buildTestUniverse() (*entities.Universe, error) {
	loader := awsconfig.NewLoader()

	// load resources
	file, err := os.Open("../../testdata/real-world/awsconfig.jsonl")
	if err != nil {
		return nil, err
	}
	err = loader.LoadJsonl(file)
	if err != nil {
		return nil, err
	}

	// load accounts, etc
	file, err = os.Open("../../testdata/real-world/org.jsonl")
	if err != nil {
		return nil, err
	}
	err = loader.LoadJsonl(file)
	if err != nil {
		return nil, err
	}

	return loader.Universe(), nil
}

func buildTestSimulator() (*Simulator, error) {
	uv, err := buildTestUniverse()
	if err != nil {
		return nil, err
	}

	sim, err := NewSimulator()
	if err != nil {
		return nil, err
	}
	sim.Universe = uv

	return sim, nil
}

func TestRealWorldData(t *testing.T) {
	type in struct {
		p string
		a string
		r string
		c map[string]string
	}

	tests := []testlib.TestCase[in, bool]{
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-magenta",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-magenta/object.txt",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-cyan",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-cyan",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-green",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-green/secrets.txt",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-green/secrets.txt",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-green/secrets.txt",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-bear",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-bear/secrets.txt",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-bear",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "sns:publish",
				r: "arn:aws:sns:us-east-1:213308312933:LemurTopic",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "sns:publish",
				r: "arn:aws:sns:us-east-1:777583092761:PurpleTopic",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "sns:publish",
				r: "arn:aws:sns:us-east-1:213308312933:LemurTopic",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "sns:gettopicattributes",
				r: "arn:aws:sns:us-east-1:213308312933:LemurTopic",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/DogUser",
				a: "sns:gettopicattributes",
				r: "arn:aws:sns:us-east-1:213308312933:LemurTopic",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/DogUser",
				a: "sqs:sendmessage",
				r: "arn:aws:sqs:us-east-1:213308312933:TurtleQueue",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/DogUser",
				a: "sqs:getqueueattributes",
				r: "arn:aws:sqs:us-east-1:213308312933:TurtleQueue",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/DogUser",
				a: "sqs:getqueueattributes",
				r: "arn:aws:sqs:us-east-1:777583092761:YellowQueue",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "sqs:sendmessage",
				r: "arn:aws:sqs:us-east-1:213308312933:TurtleQueue",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "sqs:sendmessage",
				r: "arn:aws:sqs:us-east-1:213308312933:TurtleQueue",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
			},
			Want: true,
		},
		// TODO(nsiow) add this test back when AWS Config supports DynamoDB table policies
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::213308312933:user/DogUser",
		// 		a: "dynamodb:getitem",
		// 		r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
		// 	},
		// 	Want: true,
		// },
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/CatUser",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/MustardRole",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/MustardRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/MustardRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-magenta",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/NoodleRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/NoodleRole",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/NoodleRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/BurgerRole",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SushiRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/BurgerRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/PizzaRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/SushiRole",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/PizzaRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/NoodleRole",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SushiRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/BurgerRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SushiRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/PizzaRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/DogUser",
				a: "sts:assumerole",
				r: "arn:aws:iam::213308312933:role/LionRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "iam:createrole",
				r: "arn:aws:iam::255082776537:role/EggRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "iam:deleterole",
				r: "arn:aws:iam::255082776537:role/SaladRole",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SaladRole",
				a: "iam:createrole",
				r: "arn:aws:iam::255082776537:role/EggRole",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:deleteobject",
				r: "arn:aws:s3:::crocodile-bucket-213308312933/yams.txt",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:deletebucket",
				r: "arn:aws:s3:::crocodile-bucket-213308312933",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::213308312933:role/PandaRole",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"sts:ExternalId":      "staging-access-789",
					"aws:SourceIp":        "10.0.0.0",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::213308312933:role/PandaRole",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"sts:ExternalId":      "bad-external-id",
					"aws:SourceIp":        "10.0.0.0",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::213308312933:role/PandaRole",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"sts:ExternalId":      "staging-access-789",
					"aws:SourceIp":        "8.8.8.8",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::213308312933:role/PandaRole",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-2",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::banana-bucket-255082776537/yams.txt",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::banana-bucket-255082776537/yams.txt",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::banana-bucket-255082776537/yams.txt",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-2",
				},
			},
			Want: false,
		},
		// LambdaPolicy not in config
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::255082776537:role/SandwichRole",
		// 		a: "lambda:invokefunction",
		// 		r: "arn:aws:lambda:us-east-1:255082776537:function:PieFunction",
		// 	},
		// 	Want: true,
		// },
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "dynamodb:putitem",
				r: "arn:aws:dynamodb:us-east-1:255082776537:table/TacoTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: true,
		},
		// No ec2 instances deployed
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::255082776537:role/SandwichRole",
		// 		a: "ec2:terminateinstances",
		// 		r: "arn:aws:ec2:us-east-1:255082776537:instances/i-1234567890abcdef0",
		// 	},
		// 	Want: false,
		// },
		// No ec2 instances deployed
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::255082776537:role/SaladRole",
		// 		a: "ec2:terminateinstances",
		// 		r: "arn:aws:ec2:us-east-1:255082776537:instances/i-1234567890abcdef0",
		// 	},
		// 	Want: true,
		// },
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::crocodile-bucket-213308312933/yams.txt",
				c: map[string]string{
					"aws:CurrentTime": "2025-01-01T00:00:00Z",
				},
			},
			Want: true,
		},
		{
			// Bucket policy directly grants PandaRole s3:PutObject, so boundary is bypassed
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:putobject",
				r: "arn:aws:s3:::crocodile-bucket-213308312933/yams.txt",
				c: map[string]string{
					"aws:CurrentTime": "2023-01-01T00:00:00Z",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:deleteobject",
				r: "arn:aws:s3:::crocodile-bucket-213308312933/yams.txt",
				c: map[string]string{
					"aws:CurrentTime": "2025-01-01T00:00:00Z",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::banana-bucket-255082776537/yams.txt",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/CoralRole",
				c: map[string]string{
					"aws:MultiFactorAuthPresent": "true",
					"aws:MultiFactorAuthAge":     "90",
					"sts:ExternalId":             "emergency-access-critical",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/CoralRole",
				c: map[string]string{
					"aws:MultiFactorAuthPresent": "true",
					"aws:MultiFactorAuthAge":     "9000",
					"sts:ExternalId":             "emergency-access-critical",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/CoralRole",
				c: map[string]string{
					"aws:MultiFactorAuthPresent": "true",
					"aws:MultiFactorAuthAge":     "90",
					"sts:ExternalId":             "emergency-access-critical",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/CoralRole",
				c: map[string]string{
					"aws:MultiFactorAuthPresent": "false",
					"aws:MultiFactorAuthAge":     "90",
					"sts:ExternalId":             "emergency-access-critical",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/CoralRole",
				c: map[string]string{
					"aws:MultiFactorAuthPresent": "true",
					"aws:MultiFactorAuthAge":     "90",
					"sts:ExternalId":             "emergency-access-critical",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/CoralRole",
				a: "sqs:createqueue",
				r: "arn:aws:sqs:us-east-1:777583092761/YamsQueue",
				c: map[string]string{
					"aws:RequestTag/Emergency": "true",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/CoralRole",
				a: "sqs:createqueue",
				r: "arn:aws:sqs:us-east-1:777583092761/YamsQueue",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/CoralRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::777583092761:role/TaupeRole",
				c: map[string]string{
					"aws:RequestTag/Emergency": "true",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/TaupeRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-magenta",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"aws:CurrentTime":     "2025-01-01T00:00:00Z",
					"aws:SourceIp":        "10.0.0.0",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/TaupeRole",
				a: "s3:putobject",
				r: "arn:aws:s3:::yams-magenta",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"aws:CurrentTime":     "2025-01-01T00:00:00Z",
					"aws:SourceIp":        "10.0.0.0",
				},
			},
			Want: false,
		},
		// KMS: CoralRole emergency access via root delegation + tag condition
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/CoralRole",
				a: "kms:decrypt",
				r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
				c: map[string]string{
					"aws:RequestTag/Emergency": "true",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/CoralRole",
				a: "kms:decrypt",
				r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:user/CatUser",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-bear",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/CoralRole",
				a: "sts:assumerole",
				r: "arn:aws:iam::255082776537:role/SaladRole",
				c: map[string]string{
					"sts:ExternalId":           "admin-access-456",
					"aws:RequestTag/Emergency": "true",
				},
			},
			Want: true,
		},

		// DDB: SandwichRole (acct2) with CupcakeBoundary
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:255082776537:table/TacoTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "dynamodb:deleteitem",
				r: "arn:aws:dynamodb:us-east-1:255082776537:table/TacoTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "dynamodb:scan",
				r: "arn:aws:dynamodb:us-east-1:255082776537:table/TacoTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::255082776537:role/SandwichRole",
				a: "dynamodb:query",
				r: "arn:aws:dynamodb:us-east-1:255082776537:table/TacoTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
				},
			},
			Want: true,
		},

		// DDB: MouseRole (acct1) with LlamaBoundary
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/MouseRole",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
				c: map[string]string{
					"aws:CurrentTime": "2025-01-01T00:00:00Z",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/MouseRole",
				a: "dynamodb:deleteitem",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/MouseRole",
				a: "dynamodb:deletetable",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
			},
			Want: false,
		},

		// DDB: PandaRole (acct1) with LlamaBoundary
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
				c: map[string]string{
					"aws:CurrentTime": "2025-01-01T00:00:00Z",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "dynamodb:putitem",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
				c: map[string]string{
					"aws:CurrentTime": "2025-01-01T00:00:00Z",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "dynamodb:putitem",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
				c: map[string]string{
					"aws:CurrentTime": "2023-01-01T00:00:00Z",
				},
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "dynamodb:deleteitem",
				r: "arn:aws:dynamodb:us-east-1:213308312933:table/ElephantTable",
			},
			Want: false,
		},

		// DDB: BlueRole (acct0) via GreyPolicy
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "dynamodb:query",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BlueRole",
				a: "dynamodb:putitem",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
			},
			Want: true,
		},

		// DDB: TaupeRole (acct0) with PinkBoundary
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/TaupeRole",
				a: "dynamodb:getitem",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"aws:CurrentTime":     "2025-01-01T00:00:00Z",
					"aws:SourceIp":        "10.0.0.0",
				},
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/TaupeRole",
				a: "dynamodb:putitem",
				r: "arn:aws:dynamodb:us-east-1:777583092761:table/OrangeTable",
				c: map[string]string{
					"aws:RequestedRegion": "us-east-1",
					"aws:CurrentTime":     "2025-01-01T00:00:00Z",
					"aws:SourceIp":        "10.0.0.0",
				},
			},
			Want: false,
		},

		// KMS: BeigeRole (acct0) - key policy directly names BeigeRole
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BeigeRole",
				a: "kms:decrypt",
				r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BeigeRole",
				a: "kms:describekey",
				r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/BeigeRole",
				a: "kms:encrypt",
				r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
			},
			Want: false,
		},

		// KMS: RedRole has no KMS identity policy, TurquoiseKey only has root policy
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "kms:encrypt",
				r: "arn:aws:kms:us-east-1:777583092761:key/803e5b22-eb7b-4ff2-863b-8a4ea1e4b5a3",
			},
			Want: false,
		},

		// KMS tests requiring redeploy: uncomment after `make cf && make real-world-data`
		// BlueRole → FoxKey (acct1) cross-account via key policy
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::777583092761:role/BlueRole",
		// 		a: "kms:decrypt",
		// 		r: "<FoxKey ARN from acct1>",
		// 	},
		// 	Want: true,
		// },
		// RedRole → FoxKey (acct1) - not in key policy
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::777583092761:role/RedRole",
		// 		a: "kms:decrypt",
		// 		r: "<FoxKey ARN from acct1>",
		// 	},
		// 	Want: false,
		// },
		// BlueRole → ChartreuseKey via GreyPolicy (root delegation)
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::777583092761:role/BlueRole",
		// 		a: "kms:decrypt",
		// 		r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
		// 	},
		// 	Want: true,
		// },
		// BlueRole → ChartreuseKey - GenerateDataKey not in GreyPolicy
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::777583092761:role/BlueRole",
		// 		a: "kms:generatedatakey",
		// 		r: "arn:aws:kms:us-east-1:777583092761:key/04379ae8-3ab9-4c17-bc9f-55c53dca02f0",
		// 	},
		// 	Want: false,
		// },
		// DogUser → WaffleKey (acct2) via acct1 root delegation + DogPolicy
		// {
		// 	Input: in{
		// 		p: "arn:aws:iam::213308312933:user/DogUser",
		// 		a: "kms:decrypt",
		// 		r: "<WaffleKey ARN from acct2>",
		// 	},
		// 	Want: true,
		// },
	}

	sim, err := buildTestSimulator()
	if err != nil {
		t.Fatalf("error creating simulator for testing: %v", err)
	}

	testlib.RunTestSuite(t, tests, func(i in) (bool, error) {
		opts := NewOptions(
			WithSkipServiceAuthorizationValidation(),
			WithTracing(),
			WithAdditionalProperties(i.c),
		)

		result, err := sim.SimulateByArnWithOptions(i.p, i.a, i.r, opts)
		if err != nil {
			return false, err
		}

		if os.Getenv("YAMS_TEST_DEBUG") == "1" {
			t.Log(result.Trace.Explain())
		}

		return result.IsAllowed, nil
	})
}
