package provider

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	sns "github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func checkDeploymentAccountExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resource_, ok := state.RootModule().Resources[name]

		if !ok || resource_.Type != "vy_deployment_account" {
			return fmt.Errorf("Deployment Account '%s' not found", name)
		}

		return nil
	}
}

func testAccDeploymentAccount(sns_topic_arn string) string {
	var resource = fmt.Sprintf(`
	resource "vy_deployment_account" "test" {
		topics = {
			trigger_events  = "%s"
			pipeline_events = "%s"
		}
	}
	`, sns_topic_arn, sns_topic_arn)
	return testAcc_ProviderConfig + resource
}

func TestAccDeploymentAccount(t *testing.T) {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("eu-west-1")},
		SharedConfigState: session.SharedConfigEnable,
	}))
	sns_client := sns.New(session)

	result, err := sns_client.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String("test-topic-for-terraform-provider-deployment-account.fifo"),
		Attributes: map[string]*string{
			"FifoTopic": aws.String("true"),
		},
	})
	if err != nil {
		panic(err.Error())
	}

	// The deployment-test account needs permissions to the SNS topic.
	_, err = sns_client.AddPermission(&sns.AddPermissionInput{
		TopicArn: result.TopicArn,
		Label:    aws.String("allow-deployment-account"),
		ActionName: []*string{
			aws.String("Subscribe"),
		},
		AWSAccountId: []*string{aws.String("846274634169")},
	})
	if err != nil {
		panic(err.Error())
	}

	t.Cleanup(func() {
		_, err := sns_client.DeleteTopic(
			&sns.DeleteTopicInput{TopicArn: result.TopicArn},
		)
		if err != nil {
			panic(err.Error())
			return
		}
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentAccount(*result.TopicArn),
				Check: resource.ComposeTestCheckFunc(
					checkDeploymentAccountExists("vy_deployment_account.test"),
				),
			},
		},
	})
}
