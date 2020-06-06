package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	ec2_helper "go-aws/m/v2/ec2"

	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/crypto/ssh"
)

func InitializeWorker(svc *ec2.EC2, instanceId string) {
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/LuppesKey.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for ec2_helper.CheckInstanceState(svc, instanceId) != true {
		// fmt.Println("waiting")
	}
	publicDns := ec2_helper.RetrivePublicDns(svc, instanceId)
	fmt.Println("Public dns is ", publicDns)

	conn, err := ssh.Dial("tcp", publicDns+":22", config)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

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

	err = sess.Run("/usr/bin/whoami") // eg., /usr/bin/whoami
	if err != nil {
		panic(err)
	}
}

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
