package init

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sigstore/cosign/cmd/cosign/cli/generate"
	"os"
	"strings"
)

type AWSInput struct {
	AccessKey  string
	SecretKey  string
	Region     string
	Bucket     string
	Action     string
	PublicKey  string
	PrivateKey string
	CloudTrail CloudTrail
	IsKeyless  bool
}

type CloudTrail struct {
	Name string
}

func (i *AWSInput) digestParameters() error {
	if i.PublicKey == "" && i.IsKeyless == false {
		if err := generate.GenerateKeyPairCmd(context.Background(), "", []string{}); err != nil {
			return err
		}
		i.PublicKey = "cosign.pub"
		i.PrivateKey = "cosign.key"
	}
	return nil
}

func (i *AWSInput) ReceiveParameters() error {
	if err := inputStringParameter("enter Access Key: ", &i.AccessKey, false); err != nil {
		return err
	}
	if err := inputStringParameter("enter Secret Key: ", &i.SecretKey, false); err != nil {
		return err
	}
	if err := inputStringParameter("enter region: ", &i.Region, false); err != nil {
		return err
	}
	if err := inputStringParameter("enter default bucket: ", &i.Bucket, true); err != nil {
		return err
	}
	if err := inputStringParameter("select post verification action: (1) for detect; (2) for block and detect; leave empty for no action to perform", &i.Action, true); err != nil {
		return err
	}
	if err := inputStringParameter("is there existing trail in CloudTrail which you would like to use? (if no, please press enter): ", &i.CloudTrail.Name, true); err != nil {
		return err
	}
	if err := inputYesNoParameter("do you want to work in keyless mode: ", &i.IsKeyless, false); err != nil {
		return err
	}

	if !i.IsKeyless {
		if err := inputKeyPair(i); err != nil {
			return err
		}
	}

	if err := i.digestParameters(); err != nil {
		return err
	}
	return nil
}

func inputKeyPair(i *AWSInput) error {
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
