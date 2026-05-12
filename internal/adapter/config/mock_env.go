package config

import (
	"fmt"
	"os"
)

func SetMockEnv(envs map[string]string) {
	for k, v := range envs {
		if err := os.Setenv(k, v); err != nil {
			fmt.Printf("warning: cannot set env %s: %v\n", k, err)
		}
	}
}

func UnsetMockEnv(keys []string) {
	for _, k := range keys {
		if err := os.Unsetenv(k); err != nil {
			fmt.Printf("warning: cannot unset env %s: %v\n", k, err)
		}
	}
}
