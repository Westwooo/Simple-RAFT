package main

type mockNode struct {
	conns         []string
	id            string
	heart         chan int
	state         string
	currentTerm   int
	svrs          int
	votesRecieved int
	votedFor      string
}

func (n *mockNode) appendEntries(aReq AppendRequest, c string) AppendResponse {
	return AppendResponse{}
}

func (n *mockNode) requestVotes(vReq VoteRequest, c string) VoteResponse {
	// aRes := VoteResponse{}
	// switch c {
	// case "A":
	// }
	// return
	return VoteResponse{}
}

func (n *mockNode) runServer() error {
	//set the node id as appropriate here
	return nil
}

func (n *mockNode) Heart() chan int {
	return n.heart
}

func (n *mockNode) SetState(s string) {
	n.state = s
}

func (n *mockNode) State() string {
	return n.state
}

func (n *mockNode) Term() int {
	return n.currentTerm
}

func (n *mockNode) SetTerm(t int) {
	n.currentTerm = t
}

func (n *mockNode) Svrs() int {
	return n.svrs
}

func (n *mockNode) SetSvrs(s int) {
	n.svrs = s
}

func (n *mockNode) Votes() int {
	return n.votesRecieved
}

func (n *mockNode) SetVotes(v int) {
	n.votesRecieved = v
}

func (n *mockNode) Id() string {
	return n.id
}

func (n *mockNode) SetVotedFor(voted string) {
	n.votedFor = voted
}

func (n *mockNode) Conns() []string {
	return n.conns
}

func (n *mockNode) print() {
	// if n.state == "Leader" {
	// 	fmt.Print(clearLine, upLine, clearLine, upLine, clearLine, upLine, clearLine)
	// 	fmt.Println("Status: ", n.state)
	// 	fmt.Println("Term: ", n.currentTerm)
	// 	fmt.Println("Servers in cluster :", n.svrs)
	// } else if n.state == "Follower" {
	// 	fmt.Print(clearLine, upLine, clearLine, upLine, clearLine)
	// 	fmt.Println("Status: ", n.state)
	// 	fmt.Println("Term: ", n.currentTerm)
	// } else if n.state == "Candidate" {
	// 	fmt.Print(clearLine, upLine, clearLine, upLine, clearLine, upLine, clearLine)
	// 	fmt.Println("Status: ", n.state)
	// 	fmt.Println("Term: ", n.currentTerm)
	// 	fmt.Println("Votes Recieved: ", n.votesRecieved)
	// }
}
