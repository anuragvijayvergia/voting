package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Poll struct {
	id       string   `json:"id"`
	title    string   `json:"title"`
	question string   `json:"question"`
	isOpen   bool     `json:"isOpen"`
	maxVotes int      `json:"maxVotes"`
	options  []string `json:"options"`
	votes    []Vote   `json:"votes"`
	owner    string   `json:"owner"`
}

type Vote struct {
	option string `json:"selectedOption"`
	user   string `json:"voteBy"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("in init fuction")
	var err error
		newPoll := Poll{}
		newPoll.id = "1"
		newPoll.title = "title"
		newPoll.question = "question"
		newPoll.maxVotes, err = strconv.Atoi("100")
		if err != nil {
		fmt.Println("error 4th args")
		return nil, errors.New("4th Argument i.e max votes must be numeric string")
		}
		newPoll.isOpen = true
		// for i := 4; i < len(args); i++ {
		// newPoll.options = append(options, args[i])
		// }
		newPoll.owner = "through init"
		fmt.Println("created poll object ")

		newPollAsByte, _ := json.Marshal(newPoll)
		fmt.Println("storing data")
		err = stub.PutState("1", newPollAsByte)
		if err != nil {
		fmt.Println("error while storing data")
		return nil, err
		}
	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "createPoll" {
		return t.createPoll(stub, args)
	} else if function == "vote" {
		return t.vote(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	// if function == "getVoteCount" {
	// 	return t.getVoteCount(stub, args)
	// } else if function == "getVotes" {
	// 	return t.getVotes(stub, args)
	// }
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) createPoll(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	username, err := stub.ReadCertAttribute("username")
	if err != nil {
		return nil, errors.New("Failed to get username")
	}
	usernameStr := string(username)
	fmt.Println("got username: " + usernameStr)
	if usernameStr != "admin" {
		return nil, errors.New("Only admin can create poll")
	}
	if len(args) < 6 {
		return nil, errors.New("Minimum 6 arguments are need to create poll. Viz. id,title,question,maxVotes,option1,option2,option3(optional options followed)")
	}
	id := args[0]
	//check if id already exists
	pollAsByte, err := stub.GetState(id)
	fmt.Println("got id as id " + id)
	if err != nil {
		fmt.Println("error getting poll id")
		return nil, errors.New("Failed to get poll with id as " + id)
	}

	res := Poll{}
	json.Unmarshal(pollAsByte, &res)
	if res.id == id {
		fmt.Println("error id exisit")
		return nil, errors.New("Id already exisit")
	}
	//str := `{"id":"` + id + `","title":"` + title + `","question":"` + question + `","isOpen":"` + isOpen + `","maxVotes":"` + maxVotes + `","options":"` + options + `","votes":"` + votes + `","owner":"` + username + `"}`
	fmt.Println("creating poll object")
	newPoll := Poll{}
	newPoll.id = id
	newPoll.title = args[1]
	newPoll.question = args[2]
	newPoll.maxVotes, err = strconv.Atoi(args[3])
	if err != nil {
		fmt.Println("error 4th args")
		return nil, errors.New("4th Argument i.e max votes must be numeric string")
	}
	newPoll.isOpen = true
	// for i := 4; i < len(args); i++ {
	// 	newPoll.options = append(options, args[i])
	// }
	newPoll.owner = usernameStr
	fmt.Println("created poll object ")

	newPollAsByte, _ := json.Marshal(newPoll)
	fmt.Println("storing data")
	err = stub.PutState(id, newPollAsByte)
	if err != nil {
		fmt.Println("error while storing data")
		return nil, err
	}
	fmt.Println("No Error")
	return nil, nil
}

func (t *SimpleChaincode) vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) != 2 {
		return nil, errors.New("2 arguments are need to vote. Viz. id of poll,choice")
	}
	username, err := stub.ReadCertAttribute("username")
	if err != nil {
		return nil, errors.New("Failed to get username")
	}
	usernameStr := string(username)
	id := args[0]
	pollAsByte, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get poll with id as " + id)
	}
	res := Poll{}
	json.Unmarshal(pollAsByte, &res)
	if res.id != id {
		return nil, errors.New("Poll id not found")
	}
	if res.isOpen == false {
		return nil, errors.New("Poll ended")
	}
	isValidOption := false
	for i := 0; i < len(res.options); i++ {
		if res.options[i] == args[1] {
			isValidOption = true
		}
	}
	if isValidOption == false {
		return nil, errors.New("Not a valid option")
	}

	newVote := Vote{}
	newVote.option = args[1]
	newVote.user = usernameStr

	res.votes = append(res.votes, newVote)
	pollAsByte, _ = json.Marshal(res)
	err = stub.PutState(id, pollAsByte)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
