package api

import "encoding/json"
import "fmt"

func testResponse() {
	r := Response{
		Success: true,
		Data: map[string]interface{}{
			"token": "test-token",
		},
	}
	
	json, _ := json.Marshal(r)
	fmt.Println(string(json))
}
