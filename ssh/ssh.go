package ssh

import (
	"fmt"
	"go-aws/m/v2/aws"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/crypto/ssh"
)

/*InitializeWorker connects to the given instance over ssh and install the application and its dependencies
TODO: pull the docker image from the docker hub and run it*/
func InitializeWorker(svc *ec2.EC2, id string) {
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/LuppesKey.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Wait until the instance has a public dns assigned
	for aws.CheckInstanceState(svc, id) != true {
		// fmt.Println("waiting")
	}
	dns := aws.GetPublicDNS(svc, id)
	fmt.Println("Public dns is ", dns)

	// TODO: fix this
	// Wait an additional 60 seconds to be sure the instance is open for connections
	time.Sleep(60 * time.Second)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", dns+":22", config)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Install docker
	runCommand("/usr/bin/sudo apt-get update", conn)
	runCommand("cd ~", conn)
	runCommand("curl -fsSL https://get.docker.com -o get-docker.sh", conn)
	runCommand("sudo sh get-docker.sh", conn)
}

// Read public key for ssh configuration
func publicKey(path string) ssh.AuthMethod {
	pwd, _ := os.Getwd()
	key, err := ioutil.ReadFile(pwd + path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}

// Run a command using the ssh client
func runCommand(cmd string, conn *ssh.Client) {
	sess, err := conn.NewSession()
	if err != nil {
		panic(err)
	}
	defer sess.Close()
	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go io.Copy(os.Stdout, sessStdOut)
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		panic(err)
	}
	go io.Copy(os.Stderr, sessStderr)

	err = sess.Run(cmd) // eg., /usr/bin/whoami
	if err != nil {
		panic(err)
	}
}
