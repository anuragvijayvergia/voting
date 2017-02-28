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
	Title    string      `json:"title"`
	Question string      `json:"question"`
	IsOpen   bool        `json:"isOpen"`
	MaxVotes int         `json:"maxVotes"`
	Options  []string    `json:"options"`
	Votes    []Vote      `json:"votes"`
	Owner    string      `json:"owner"`
	Count    []VoteCount `json:"voteCount"`
}

type Vote struct {
	Option string `json:"selectedOption"`
	User   string `json:"voteBy"`
}

type VoteCount struct {
	Option     string `json:"option"`
	CountTotal int    `json:"count"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return t.createPoll(stub, args)
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
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
	if function == "getVoteCount" {
		return t.getVoteCount(stub, args)
	} else if function == "getVotes" {
		return t.getVotes(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) createPoll(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) < 5 {
		return nil, errors.New("Minimum 5 arguments are need to create poll. Viz. title,question,maxVotes,option1,option2,option3(optional options followed)")
	}

	newPoll := Poll{}
	newPoll.Title = args[0]
	newPoll.Question = args[1]
	newPoll.MaxVotes, err = strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("4th Argument i.e max votes must be numeric string")
	}
	newPoll.IsOpen = true
	for i := 3; i < len(args); i++ {
		newPoll.Options = append(newPoll.Options, args[i])
		count := VoteCount{
			Option:     args[i],
			CountTotal: 0}
		newPoll.Count = append(newPoll.Count, count)
	}
	newPoll.Owner = "admins"

	newPollAsByte, _ := json.Marshal(newPoll)

	err = stub.PutState("poll", newPollAsByte)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) != 1 {
		return nil, errors.New("1 arguments are need to vote. Viz. choice")
	}

	usernameStr := "admin"
	pollAsByte, err := stub.GetState("poll")
	if err != nil {
		return nil, errors.New("Failed to get poll")
	}
	res := Poll{}
	json.Unmarshal(pollAsByte, &res)

	if res.IsOpen == false {
		return nil, errors.New("Poll ended")
	}

	isValidOption := false
	for i := 0; i < len(res.Options); i++ {
		if res.Options[i] == args[0] {
			isValidOption = true
		}
	}
	if isValidOption == false {
		return nil, errors.New("Not a valid option")
	}

	newVote := Vote{}
	newVote.Option = args[0]
	newVote.User = usernameStr

	res.Votes = append(res.Votes, newVote)

	for i := 0; i < len(res.Count); i++ {
		if res.Count[i].Option == args[0] {
			res.Count[i].CountTotal = res.Count[i].CountTotal + 1
		}
	}

	pollAsByte, _ = json.Marshal(res)
	err = stub.PutState("poll", pollAsByte)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *SimpleChaincode) getVoteCount(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	pollAsByte, err := stub.GetState("poll")
	if err != nil {
		return nil, errors.New("Failed to get poll with ")
	}
	res := Poll{}
	json.Unmarshal(pollAsByte, &res)
	votesB, _ := json.Marshal(res.Count)
	return votesB, nil
}

func (t *SimpleChaincode) getVotes(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	pollAsByte, err := stub.GetState("poll")
	if err != nil {
		return nil, errors.New("Failed to get poll ")
	}
	res := Poll{}
	json.Unmarshal(pollAsByte, &res)
	votesB, _ := json.Marshal(res.Votes)
	return votesB, nil
}
