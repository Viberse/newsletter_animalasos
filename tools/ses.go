package tools

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func createSesCfg(ctx context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
	)
}

func CreateSesSession(ctx context.Context) (*ses.Client, error) {
	if cfg, err := createSesCfg(ctx); err != nil {
		creds := aws.Credentials{
			CanExpire:       false,
			AccessKeyID:     "AKIAQIDNOFX37UJIOZXL",
			SecretAccessKey: "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",
		}
		cfg.Credentials = credentials.StaticCredentialsProvider{
			Value: creds,
		}

		return nil, err
	} else {
		return ses.NewFromConfig(cfg), nil
	}
}
