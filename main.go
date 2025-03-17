package main

import (
	"fmt"
	"io"
	"os"
	"net/http"
	"encoding/xml"
	"reflect"
	"slices"
	"strings"
	// "log"
)

// Custom Structs - For XML Parsing

type Member struct {
	XMLName 	xml.Name	`xml:"member"`
	FullName 	string		`xml:"member_full"`
	LastName	string		`xml:"last_name"`
	FirstName	string		`xml:"first_name"`
	Party		string		`xml:"party"` // You could define a "Party" struct if you wanted...
	State		string		`xml:"state"`
	Vote		string		`xml:"vote_cast"`	
	MemberID	string		`xml:"lis_member"`
}


type Members struct {
	XMLName xml.Name `xml:"members"`
	Electors []Member `xml:"member"`
}


type rollCallVote struct {
	XMLName 	xml.Name 	`xml:"roll_call_vote"`
	Members 	Members 	`xml:"members"`
	Congress	string		`xml:"congress"`
	Session		int			`xml:"session"`
	CongressYear int		`xml:"congress_year"`
	VoteNumber	int			`xml:"vote_number"`
	VoteDate	string		`xml:"vote_date"`
	VoteQuestion string		`xml:"vote_question_text"`
	VoteDocText	string		`xml:"vote_document_text"`

}

// ------------------------------------------ //

func main() {
	apiKey := getAPIkey()
	congressAPIURL := "https://www.senate.gov/legislative/LIS/roll_call_votes/vote1191/vote_119_1_00124.xml" // "https://api.congress.gov/v3/bill"
	body, err := clientConnect(congressAPIURL, apiKey)
	if err != nil {
		fmt.Printf("GET request failure: %s", err)
	}

	parseResponse(body)



}

func getAPIkey() string {
	api_key_env := "DATA_DOT_GOV_API_KEY"
	apiKey := os.Getenv(api_key_env)
    if apiKey == "" {
        fmt.Printf("%s not set", api_key_env)
    }
	return apiKey
}

func clientConnect(congressAPIURL string, apiKey string) ([]byte, error) {

	// Create Request Format
	req, err := http.NewRequest("GET",congressAPIURL, nil)
	if err != nil{
		fmt.Printf("GET request failure: %s", err)
	}
	// Add API Key to the Header
	req.Header.Set("x-api-key", apiKey)

	// String up Client and make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("GET request failure: %s\n", err)
	}
	fmt.Printf("Response: %v\n", resp)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error reading response: %s\n", err)
        return body, err
    }

    // fmt.Println(string(body))
	return body, nil
}

// Parse the Roll Call
// URL Format:	https://www.senate.gov/legislative/LIS/roll_call_votes/vote[CONGRESS_NUM][SESSION_NUM]/vote_[CONGRESS_NUM]_[SESSION_NUM]_[ROLL_NUM_5_PADDED].xml
// Senate Ex:	https://www.senate.gov/legislative/LIS/roll_call_votes/vote1172/vote_117_2_00071.xml
// House Ex:	https://clerk.house.gov/evs/2025/roll070.xml
//
func parseResponse(resp []byte) {

	// Unmarshall
	var rollCallVote rollCallVote
	err := xml.Unmarshal([]byte(resp), &rollCallVote) 
	if err != nil {
		fmt.Println("Error unmarshalling response: ", err)
	}

	// MetaData
	printFields(rollCallVote)
	// Get Idv. Votes
	for _, member := range rollCallVote.Members.Electors {
		fmt.Printf("%s (%s) - %s\n", member.LastName, member.State, member.Vote)
	}
}

func printFields(v interface{}) {
	t := reflect.TypeOf(v)

	// Check if Pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check if Struct
	if t.Kind() != reflect.Struct {
		fmt.Println("Value Type is not a struct!")
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		xmlTag := field.Tag.Get("xml")
		// Don't print these two fields
		if slices.Contains([]string{"members","roll_call_vote"}, xmlTag) {continue}
		rfl := reflect.Indirect(reflect.ValueOf(v)).FieldByName(field.Name)
		fmt.Printf("%s: %v\n", field.Name, rfl)
		}
	fmt.Println(strings.Repeat("-", 35))
}