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
	Id       string   `json:"id"`
	Title    string   `json:"title"`
	Question string   `json:"question"`
	IsOpen   bool     `json:"isOpen"`
	MaxVotes int      `json:"maxVotes"`
	Options  []string `json:"options"`
	Votes    []Vote   `json:"votes"`
	Owner    string   `json:"owner"`
}

type Vote struct {
	Option string `json:"selectedOption"`
	User   string `json:"voteBy"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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

	// username, err := stub.ReadCertAttribute("username")
	// if err != nil {
	// 	return nil, errors.New("Failed to get username")
	// }
	// usernameStr := string(username)
	// fmt.Println("got username: " + usernameStr)
	// if usernameStr != "admin" {
	// 	return nil, errors.New("Only admin can create poll")
	// }
	if len(args) < 6 {
		return nil, errors.New("Minimum 6 arguments are need to create poll. Viz. id,title,question,maxVotes,option1,option2,option3(optional options followed)")
	}
	id := args[0]
	pollAsByte, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get poll with id as " + id)
	}

	res := Poll{}
	json.Unmarshal(pollAsByte, &res)
	if res.Id == id {
		return nil, errors.New("Id already exisit")
	}

	newPoll := Poll{}
	newPoll.Id = id
	newPoll.Title = args[1]
	newPoll.Question = args[2]
	newPoll.MaxVotes, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("4th Argument i.e max votes must be numeric string")
	}
	newPoll.IsOpen = true
	for i := 4; i < len(args); i++ {
		newPoll.Options = append(newPoll.Options, args[i])
	}
	newPoll.Owner = "admins"

	newPollAsByte, _ := json.Marshal(newPoll)

	err = stub.PutState(id, newPollAsByte)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) != 2 {
		return nil, errors.New("2 arguments are need to vote. Viz. id of poll,choice")
	}

	usernameStr := "admin"
	id := args[0]
	pollAsByte, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get poll with id as " + id)
	}
	res := Poll{}
	json.Unmarshal(pollAsByte, &res)
	if res.Id != id {
		return nil, errors.New("Poll id not found " + res.id)
	}
	if res.IsOpen == false {
		return nil, errors.New("Poll ended")
	}

	isValidOption := false
	for i := 0; i < len(res.Options); i++ {
		if res.Options[i] == args[1] {
			isValidOption = true
		}
	}
	if isValidOption == false {
		return nil, errors.New("Not a valid option")
	}

	newVote := Vote{}
	newVote.Option = args[1]
	newVote.User = usernameStr

	res.Votes = append(res.Votes, newVote)
	pollAsByte, _ = json.Marshal(res)
	err = stub.PutState(id, pollAsByte)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
