package common

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helper Test Suite")

	if IsCommandAvailable("ls") == false {
		t.Error("ls command does not exist!")
	}
	if IsCommandAvailable("ls111") == true {
		t.Error("ls111 command should not exist!")
	}
}

var _ = Describe("Helper Tests", func() {
	Context("Checking isCommandAvailable reports true for a valid command", func() {
		Specify("The Command should return true", func() {
			Expect(IsCommandAvailable("ls")).Should(BeTrue())
		})
	})

	Context("Checking isCommandAvailable reports false for an invalid command", func() {
		Specify("The Command should return false", func() {
			Expect(IsCommandAvailable("ls123")).Should(BeFalse())
		})
	})
})
