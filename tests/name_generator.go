package tests

import (
	"k8s.io/apimachinery/pkg/util/rand"
)

func randName(name string) string {
	return name + "-" + rand.String(5)
}
