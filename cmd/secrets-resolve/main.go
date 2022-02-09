package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"gocloud.dev/gcp"
	"gocloud.dev/runtimevar/gcpsecretmanager"
	"google.golang.org/genproto/googleapis/api"
)

func main() {
	ctx := context.Background()
	var gcpInit sync.Once
	var gcpSecretManagerClient *api.Client
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
				gcpSecretManagerClient, gcpSecretManagerClientCleanFunc, err := gcpsecretmanager.Dial(ctx, gcp.TokenSource(creds))
				if err != nil {
					panic(err)
				}
			})
		}
	}

	if gcpSecretManagerClient != nil {
		gcpSecretManagerClientCleanFunc()
	}
}
