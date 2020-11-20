package common

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Helper Test Suite")
}

var _ = Describe("Common Helper Tests", func() {
	Context("Checking CompareAvailableVsNeeded functions correctly in all scenarios.", func() {
		Specify("The function should return nothing when all needles are found", func() {
			needleArray := []string{"apples", "bananas", "oranges"}
			haystackArray := []string{"apples", "bananas", "dates", "grapes", "oranges", "pears"}

			notFound, compareErr := CompareAvailableVsNeeded(haystackArray, needleArray)

			Expect(notFound).Should(BeEmpty())
			Expect(compareErr).Should(BeNil())
		})

		Specify("The function should error and return an array of missing values when one needle is not found", func() {
			needleArray := []string{"apples", "oranges", "raisins"}
			haystackArray := []string{"apples", "bananas", "dates", "grapes", "oranges", "pears"}

			notFound, compareErr := CompareAvailableVsNeeded(haystackArray, needleArray)

			Expect(notFound).Should(ContainElement(ContainSubstring("raisins")))
			Expect(compareErr).Should(MatchError("One or more Needed Items Not Found"))
		})

		Specify("The function should error and return an array of missing values when multiple needles are not found", func() {
			needleArray := []string{"apples", "oranges", "raisins"}
			haystackArray := []string{"apples", "bananas", "dates", "grapes", "pears"}

			notFound, compareErr := CompareAvailableVsNeeded(haystackArray, needleArray)

			Expect(notFound).Should(ContainElements(ContainSubstring("raisins"), ContainSubstring("oranges")))
			Expect(compareErr).Should(MatchError("One or more Needed Items Not Found"))
		})
	})

	Context("Checking ConvertKeyValuesToMap converts an array of key values to a map", func() {
		Specify("The function should return an expected map of transformed key-value pairs", func() {
			sourceArray := []string{"key1=value1", "key2=value2", "key3=value3"}

			expectedMap := make(map[string]string)
			expectedMap["key1"] = "value1"
			expectedMap["key2"] = "value2"
			expectedMap["key3"] = "value3"

			Expect(ConvertKeyValuesToMap(sourceArray)).Should(BeEquivalentTo(expectedMap))
		})
	})

	Context("Checking IsCommandAvailable reports true for a valid command", func() {
		Specify("The Command should return true", func() {
			Expect(IsCommandAvailable("ls")).Should(BeTrue())
		})
	})

	Context("Checking IsCommandAvailable reports false for an invalid command", func() {
		Specify("The Command should return false", func() {
			Expect(IsCommandAvailable("ls123")).Should(BeFalse())
		})
	})
})
