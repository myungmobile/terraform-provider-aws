package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAWSGlueCrawler_basic(t *testing.T) {
	const name = "aws_glue_catalog_crawler.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists(name, "test-basic"),
					resource.TestCheckResourceAttr(name, "name", "test-basic"),
					resource.TestCheckResourceAttr(name, "database_name", "test_db"),
					resource.TestCheckResourceAttr(name, "role", "AWSGlueServiceRole-tf"),
				),
			},
		},
	})
}

func TestAccAWSGlueCrawler_jdbcCrawler(t *testing.T) {
	const name = "aws_glue_catalog_crawler.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigJdbc,
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists(name, "test-jdbc"),
					resource.TestCheckResourceAttr(name, "name", "test-jdbc"),
					resource.TestCheckResourceAttr(name, "database_name", "test_db"),
					resource.TestCheckResourceAttr(name, "role", "tf-glue-service-role"),
					resource.TestCheckResourceAttr(name, "jdbc_target.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSGlueCrawler_customCrawlers(t *testing.T) {
	const name = "aws_glue_catalog_crawler.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigCustomClassifiers,
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists(name, "test_custom"),
					resource.TestCheckResourceAttr(name, "name", "test_custom"),
					resource.TestCheckResourceAttr(name, "database_name", "test_db"),
					resource.TestCheckResourceAttr(name, "role", "tf-glue-service-role"),
					resource.TestCheckResourceAttr(name, "table_prefix", "table_prefix"),
					resource.TestCheckResourceAttr(name, "schema_change_policy.0.delete_behavior", "DELETE_FROM_DATABASE"),
					resource.TestCheckResourceAttr(name, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(name, "s3_target.#", "2"),
				),
			},
		},
	})
}

func checkGlueCatalogCrawlerExists(name string, crawlerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		glueConn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueConn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(crawlerName),
		})

		if err != nil {
			return err
		}

		if out.Crawler == nil {
			return fmt.Errorf("no Glue Crawler found")
		}

		return nil
	}
}

const testAccGlueCrawlerConfigBasic = `
	resource "aws_glue_catalog_database" "test_db" {
  		name = "test_db"
	}

	resource "aws_glue_catalog_crawler" "test" {
	  name = "test-basic"
	  database_name = "${aws_glue_catalog_database.test_db.name}"
	  role = "${aws_iam_role.glue.name}"
	  description = "TF-test-crawler"
	  schedule="cron(0 1 * * ? *)"
	  s3_target {
		path = "s3://bucket"
	  }
	}
	
	resource "aws_iam_role_policy_attachment" "aws-glue-service-role-default-policy-attachment" {
  		policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  		role = "${aws_iam_role.glue.name}"
	}
	
	resource "aws_iam_role" "glue" {
  		name = "AWSGlueServiceRole-tf"
  		assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
	}
`

const testAccGlueCrawlerConfigJdbc = `
	resource "aws_glue_catalog_database" "test_db" {
  		name = "test_db"
	}

	resource "aws_glue_connection" "test" {
  		name = "tf-connection"
		connection_properties = {
    		JDBC_CONNECTION_URL = "jdbc:mysql://example.com/exampledatabase"
    		PASSWORD            = "examplepassword"
    		USERNAME            = "exampleusername"
  		}
	}
	
	resource "aws_iam_role_policy_attachment" "aws-glue-service-role-default-policy-attachment" {
  		policy_arn = "arn:aws:iam::aws:policy/AWSGlueConsoleFullAccess"
  		role = "${aws_iam_role.glue.name}"
	}

	data "aws_iam_policy_document" "all-glue-policy-document" {
  		statement {
    		actions = [
      			"glue:*",
                "s3:GetBucketLocation",
                "s3:ListBucket",
                "s3:ListAllMyBuckets",
                "s3:GetBucketAcl",
                "ec2:DescribeVpcEndpoints",
                "ec2:DescribeRouteTables",
                "ec2:CreateNetworkInterface",
                "ec2:DeleteNetworkInterface",				
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpcAttribute",
                "iam:ListRolePolicies",
                "iam:GetRole",
                "iam:GetRolePolicy"
			]
			principals = {
  				type = "service"
  				identifiers = ["glue.amazonaws.com"]
			}

    		resources = [
      			"*",
    		]
  		}
	}
	
	resource "aws_iam_role_policy_attachment" "aws-glue-all-glue-policy-attachment" {
  		policy_arn = "${data.aws_iam_policy_document.all-glue-policy-document.json}"
  		role = "${aws_iam_role.glue.name}"
	}

	resource "aws_iam_role" "glue" {
  		name = "tf-glue-service-role"
  		assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
	}

	resource "aws_glue_catalog_crawler" "test" {
	  name = "test-jdbc"
	  database_name = "${aws_glue_catalog_database.test_db.name}"
	  role = "${aws_iam_role.glue.name}"
	  description = "TF-test-crawler"
	  schedule="cron(0 1 * * ? *)"
	  jdbc_target {
		path = "s3://bucket"
		connection_name = "${aws_glue_connection.test.name}"
	  }
	}
`

//classifiers = [
//"${aws_glue_classifier.test.id}"
//]
//resource "aws_glue_classifier" "test" {
//name = "tf-example-123"
//
//grok_classifier {
//classification = "example"
//grok_pattern   = "example"
//}
//}
const testAccGlueCrawlerConfigCustomClassifiers = `
	resource "aws_glue_catalog_database" "test_db" {
  		name = "test_db"
	}

	resource "aws_glue_catalog_crawler" "test" {
	  name = "test_custom"
	  database_name = "${aws_glue_catalog_database.test_db.name}"
	  role = "${aws_iam_role.glue.name}"
	  s3_target {
		path = "s3://bucket1"
		exclusions = [
			"s3://bucket1/foo"
		]
	  }
	  s3_target {
		path = "s3://bucket2"
	  }
      table_prefix = "table_prefix"
	  schema_change_policy {
		delete_behavior = "DELETE_FROM_DATABASE"
		update_behavior = "UPDATE_IN_DATABASE"
      }
	}

	resource "aws_iam_role_policy_attachment" "aws-glue-service-role-default-policy-attachment" {
  		policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  		role = "${aws_iam_role.glue.name}"
	}
	
	resource "aws_iam_role" "glue" {
  		name = "tf-glue-service-role"
  		assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
	}
`
