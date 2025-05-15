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
			Input: in{
				p: "arn:aws:iam::213308312933:role/PandaRole",
				a: "s3:putobject",
				r: "arn:aws:s3:::crocodile-bucket-213308312933/yams.txt",
				c: map[string]string{
					"aws:CurrentTime": "2023-01-01T00:00:00Z",
				},
			},
			Want: false,
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
		// {
		// 	// Waiting on key to delete
		// 	Input: in{
		// 		p: "arn:aws:iam::777583092761:role/CoralRole",
		// 		a: "kms:decrypt",
		// 		r: "arn:aws:kms:us-east-1:777583092761:alias/chartreuse-key",
		// 		c: map[string]string{
		// 			"aws:RequestTag/Emergency": "true",
		// 		},
		// 	},
		// 	Want: true,
		// },
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
