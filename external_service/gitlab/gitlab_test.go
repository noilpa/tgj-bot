package gitlab_

import (
	"fmt"
	"testing"
)

func Test_string(t *testing.T) {
	description := "asdf\r\n\r\n// Reviewers: @o.plakhotnii @Viktor.Shkanov @Igor //\r\n\r\nasdfasdf"
	fmt.Println("1. ", removeReviewersFromDescription(description))
	fmt.Println("2. ", removeReviewersFromDescription(" // Reviewers:  "))
}
