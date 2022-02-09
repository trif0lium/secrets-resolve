package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"gocloud.dev/gcp"
	"gocloud.dev/runtimevar"
	"gocloud.dev/runtimevar/gcpsecretmanager"
)

func main() {
	ctx := context.Background()
	var gcpInit sync.Once
	var gcpSecretManagerClient *secretmanager.Client
	var gcpSecretManagerClientCleanFunc func()

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envKey := pair[0]
		envValue := strings.TrimSpace(pair[1])
		if strings.HasPrefix(envValue, "gcpSecretManager://") {
			parts := strings.Split(strings.TrimPrefix(envValue, "gcpSecretManager://"), "/")
			if len(parts) != 2 {
				log.Printf("invalid reference format: %s", envValue)
				continue
			}
			gcpInit.Do(func() {
				creds, err := gcp.DefaultCredentials(ctx)
				if err != nil {
					panic(err)
				}
				client, cleanFunc, err := gcpsecretmanager.Dial(ctx, creds.TokenSource)
				if err != nil {
					panic(err)
				}
				gcpSecretManagerClient = client
				gcpSecretManagerClientCleanFunc = cleanFunc
			})
			gcpProjectID := parts[0]
			secretID := parts[1]
			variableKey := gcpsecretmanager.SecretKey(gcp.ProjectID(gcpProjectID), secretID)
			v, err := gcpsecretmanager.OpenVariable(gcpSecretManagerClient, variableKey, runtimevar.StringDecoder, nil)
			if err != nil {
				log.Printf("failed to construct a *runtimevar.Variable for '%s': %v", envValue, err)
				continue
			}
			secretValue, err := v.Latest(ctx)
			if err != nil {
				log.Printf("failed to resolve secret '%s': %v", envValue, err)
				continue
			}
			os.Setenv(envKey, secretValue.Value.(string))
			v.Close()
		}
	}

	if gcpSecretManagerClient != nil {
		gcpSecretManagerClientCleanFunc()
	}
}
