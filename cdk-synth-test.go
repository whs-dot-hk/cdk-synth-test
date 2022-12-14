package main

import (
	"os"
	"log"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsimagebuilder"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
)

type CdkSynthTestStackProps struct {
	awscdk.StackProps
}

func NewCdkSynthTestStack(scope constructs.Construct, id string, props *CdkSynthTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("CdkSynthTestQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })
	content, err := os.ReadFile("test.yaml")
	if err != nil {
		log.Fatal(err)
	}

	component := awsimagebuilder.NewCfnComponent(stack, jsii.String("Component"), &awsimagebuilder.CfnComponentProps{
		Name: jsii.String("install-cardano-node"),
		Platform: jsii.String("Linux"),
		Version: jsii.String("1.0.0"),
		Data: jsii.String(string(content)),
	})

	recipe := awsimagebuilder.NewCfnImageRecipe(stack, jsii.String("ImageRecipe"), &awsimagebuilder.CfnImageRecipeProps{
		Name: jsii.String("cardano-nodes"),
		ParentImage: jsii.String("arn:aws:imagebuilder:us-east-1:aws:image/amazon-linux-2-x86/x.x.x"),
		Version: jsii.String("1.0.0"),
		BlockDeviceMappings: []interface{}{
			&awsimagebuilder.CfnImageRecipe_InstanceBlockDeviceMappingProperty{
				DeviceName: jsii.String("/dev/xvda"),
				Ebs: &awsimagebuilder.CfnImageRecipe_EbsInstanceBlockDeviceSpecificationProperty{
					DeleteOnTermination: jsii.Bool(true),
					VolumeSize: jsii.Number(20),
				},
			},
		},
		Components: []interface{}{
			&awsimagebuilder.CfnImageRecipe_ComponentConfigurationProperty{
				ComponentArn: component.AttrArn(),
			},
		},
	})

	bucket := s3.NewBucket(stack, jsii.String("Bucket"), &s3.BucketProps{
		BlockPublicAccess: s3.BlockPublicAccess_BLOCK_ALL(),
	})

	role := awsiam.NewRole(stack, jsii.String("Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("EC2InstanceProfileForImageBuilder")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
		},
	})

	pattern := "*"

	role.AttachInlinePolicy(awsiam.NewPolicy(stack, jsii.String("Policy"), &awsiam.PolicyProps{
		Statements: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Actions: &[]*string{
					jsii.String("s3:PutObject"),
				},
				Resources: &[]*string{
					bucket.ArnForObjects(&pattern),
				},
			}),
		},
	}))

	instanceProfile := awsiam.NewCfnInstanceProfile(stack, jsii.String("InstanceProfile"), &awsiam.CfnInstanceProfileProps{
		Roles: &[]*string{
			role.RoleName(),
		},
	})

	infrastructureConfiguration := awsimagebuilder.NewCfnInfrastructureConfiguration(stack, jsii.String("InfrastructureConfiguration"), &awsimagebuilder.CfnInfrastructureConfigurationProps{
		InstanceProfileName: instanceProfile.Ref(),
		Name: jsii.String("cardano-node"),
		Logging: &awsimagebuilder.CfnInfrastructureConfiguration_LoggingProperty{
			S3Logs: &awsimagebuilder.CfnInfrastructureConfiguration_S3LogsProperty{
				S3BucketName: bucket.BucketName(),
			},
		},
	})

	awsimagebuilder.NewCfnImagePipeline(stack, jsii.String("ImagePipeline"), &awsimagebuilder.CfnImagePipelineProps{
		Name: jsii.String("cardano-node"),
		ImageRecipeArn: recipe.AttrArn(),
		InfrastructureConfigurationArn: infrastructureConfiguration.AttrArn(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkSynthTestStack(app, "CdkSynthTestStack", &CdkSynthTestStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
