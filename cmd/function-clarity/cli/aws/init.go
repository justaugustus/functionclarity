// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"bufio"
	"context"
	"fmt"
	"github.com/openclarity/function-clarity/pkg/clients"
	i "github.com/openclarity/function-clarity/pkg/init"
	"github.com/sigstore/cosign/cmd/cosign/cli/generate"
	"os"
	"strings"
)

func ReceiveParameters(i *i.AWSInput) error {
	awsClient, err := receiveAndValidateCredentials(i)
	if err != nil {
		return err
	}

	if err := receiveAndValidateBucketName(i, awsClient); err != nil {
		return err
	}

	if err := inputStringArrayParameter("enter tag keys of functions to include in the verification (leave empty to include all): ", &i.IncludedFuncTagKeys, true); err != nil {
		return err
	}
	if err := inputStringArrayParameter("enter the function regions to include in the verification, i.e: us-east-1,us-west-1 (leave empty to include all): ", &i.IncludedFuncRegions, true); err != nil {
		return err
	}

	if err := inputMultipleChoiceParameter("post verification action", &i.Action, map[string]string{"1": "detect", "2": "block"}, true); err != nil {
		return err
	}

	if err := receiveAndValidateSNSTopicArn(i, awsClient); err != nil {
		return err
	}

	if err := receiveAndValidateCloudTrail(i, awsClient); err != nil {
		return err
	}

	if err := inputYesNoParameter("do you want to work in keyless mode (y/n): ", &i.IsKeyless, false); err != nil {
		return err
	}

	if !i.IsKeyless {
		if err := inputKeyPair(i); err != nil {
			return err
		}
	}

	if err := digestParameters(i); err != nil {
		return err
	}
	return nil
}

func digestParameters(i *i.AWSInput) error {
	if i.PublicKey == "" && !i.IsKeyless {
		if err := generate.GenerateKeyPairCmd(context.Background(), "", []string{}); err != nil {
			return err
		}
		i.PublicKey = "cosign.pub"
		i.PrivateKey = "cosign.key"
	}
	return nil
}

func receiveAndValidateCloudTrail(i *i.AWSInput, awsClient *clients.AwsClient) error {
	if err := inputStringParameter("is there existing trail in CloudTrail (in the region selected above) which you would like to use? (if no, please press enter): ", &i.CloudTrail.Name, true); err != nil {
		return err
	}
	trailName := i.CloudTrail.Name
	if trailName != "" && !awsClient.IsCloudTrailExist(trailName) {
		return fmt.Errorf("validation error: SNS topic doesn't exist or you don't have permissions")
	}
	return nil
}

func receiveAndValidateSNSTopicArn(i *i.AWSInput, awsClient *clients.AwsClient) error {
	if err := inputStringParameter("enter SNS arn if you would like to be notified when signature verification fails, otherwise press enter: ", &i.SnsTopicArn, true); err != nil {
		return err
	}
	if i.SnsTopicArn != "" && !awsClient.IsSnsTopicExist(i.SnsTopicArn) {
		return fmt.Errorf("validation error: SNS topic doesn't exist or you don't have permissions")
	}
	return nil
}

func receiveAndValidateBucketName(i *i.AWSInput, awsClient *clients.AwsClient) error {
	if err := inputStringParameter("enter default bucket (you can leave empty and a bucket with name functionclarity will be created): ", &i.Bucket, true); err != nil {
		return err
	}
	if i.Bucket != "" && !awsClient.IsBucketExist(i.Bucket) {
		return fmt.Errorf("validation error: bucket doesn't exist or you don't have permissions")
	}
	return nil
}

func receiveAndValidateCredentials(i *i.AWSInput) (*clients.AwsClient, error) {
	if err := inputStringParameter("enter Access Key: ", &i.AccessKey, false); err != nil {
		return nil, err
	}
	if err := inputStringParameter("enter Secret Key: ", &i.SecretKey, false); err != nil {
		return nil, err
	}
	if err := inputStringParameter("enter region: ", &i.Region, false); err != nil {
		return nil, err
	}
	awsClient := clients.NewAwsClientInit(i.AccessKey, i.SecretKey, i.Region)
	if credentials := awsClient.ValidateCredentials(); !credentials {
		return nil, fmt.Errorf("validation error: credentials aren't valid")
	}
	return awsClient, nil
}

func inputKeyPair(i *i.AWSInput) error {
	if err := inputStringParameter("enter path to custom public key for code signing? (if you want us to generate key pair, please press enter): ", &i.PublicKey, true); err != nil {
		return err
	}
	if i.PublicKey != "" {
		if err := inputStringParameter("enter path to custom private key for code signing: ", &i.PrivateKey, false); err != nil {
			return err
		}
	}
	return nil
}

func inputStringParameter(q string, p *string, em bool) error {
	fmt.Print(q)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	if !em && input == "" {
		return fmt.Errorf("this is a compulsory parameter")
	}
	*p = strings.TrimSuffix(input, "\n")
	return err
}

func inputStringArrayParameter(q string, p *[]string, em bool) error {
	fmt.Print(q)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSpace(input)
	if !em && input == "" {
		return fmt.Errorf("this is a compulsory parameter")
	}
	*p = strings.Split(input, ",")
	for index := range *p {
		(*p)[index] = strings.TrimSpace((*p)[index])
	}
	return err
}

func inputYesNoParameter(q string, p *bool, em bool) error {
	fmt.Print(q)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	if !em && input == "" {
		return fmt.Errorf("this is a compulsory parameter")
	}
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "y" {
		*p = true
	} else if input == "n" {
		*p = false
	}
	return err
}

func inputMultipleChoiceParameter(action string, p *string, m map[string]string, em bool) error {
	message := "select " + action + " : "
	for key, element := range m {
		message = message + "(" + key + ")" + " for " + element + "; "
	}
	if em {
		message = message + "leave empty for no " + action + " to perform: "
	}
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.TrimSuffix(input, "\n")
	if !em && input == "" {
		return fmt.Errorf("this is a compulsory parameter")
	}
	for key, element := range m {
		if input == key {
			*p = element
		}
	}
	if input == "" {
		if !em {
			return fmt.Errorf("this is a compulsory parameter")
		} else {
			*p = ""
		}
	}
	return nil
}
