package ssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	aws_helper "go-aws/m/v2/aws"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Connect to the given instance over ssh and install the application and its dependencies
func InitializeWorker(svc *ec2.EC2, instanceId string) {

	fmt.Println("Initializing worker", instanceId)

	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/go-aws.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Wait until the instance has a public dns assigned
	for aws_helper.CheckInstanceState(svc, instanceId) != true {
		// fmt.Println("waiting")
	}
	publicDns := aws_helper.RetrievePublicDns(svc, instanceId)
	// fmt.Println("Public dns is ", publicDns)

	// TODO: fix this
	// Wait an additional 60 seconds to be sure the instance is open for connections
	time.Sleep(60 * time.Second)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDns+":22", config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Install docker and pull the application dockerfile
	runCommand("/usr/bin/sudo apt-get update", conn)
	runCommand("sudo apt-get --assume-yes install sysstat", conn) // TODO: sometimes it breaks on this
	runCommand("cd ~", conn)
	runCommand("curl -fsSL https://get.docker.com -o get-docker.sh", conn)
	runCommand("sudo sh get-docker.sh", conn)
	runCommand("sudo docker pull bobray/style-transfer:firsttry", conn)
	copyFileToHost("./heartbeat.sh", "./heartbeat.sh", conn)
	runCommand("sudo nohup sudo sh ./heartbeat.sh 80.114.173.4 8000 > /dev/null 2>&1 &", conn) // TODO: fix hardcoded ip

	fmt.Println("Initialized  worker", instanceId)
}

// Read public key for ssh configuration
func publicKey(path string) ssh.AuthMethod {
	pwd, _ := os.Getwd()
	key, err := ioutil.ReadFile(pwd + path)
	if err != nil {
		fmt.Println(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println(err)
	}
	return ssh.PublicKeys(signer)
}

// Run a command using the ssh client
func runCommand(cmd string, conn *ssh.Client) {
	sess, err := conn.NewSession()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sess.Close()
	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}
	// go io.Copy(os.Stdout, sessStdOut)
	_ = sessStdOut
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return
	}
	// go io.Copy(os.Stderr, sessStderr)
	_ = sessStderr

	err = sess.Run(cmd) // eg., /usr/bin/whoami
	if err != nil {
		fmt.Println(err)
		return
	}
}

func copyFileToHost(srcFilePath string, dstFilePath string, conn *ssh.Client) {
	client, err := sftp.NewClient(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	dstFile, err := client.Create(dstFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dstFile.Close()

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(bytes, "bytes copied")
	_ = bytes
}

func copyFileFromHost(srcFilePath string, dstFilePath string, conn *ssh.Client) {
	client, err := sftp.NewClient(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dstFile.Close()

	srcFile, err := client.Open(srcFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(bytes, "bytes copied")
	_ = bytes

	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func RunApplication(svc *ec2.EC2, instanceId string, folder string) {
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/go-aws.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	publicDns := aws_helper.RetrievePublicDns(svc, instanceId)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDns+":22", config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	time.Sleep(10 * time.Second)

	// Copy the input images to the worker
	runCommand("mkdir -p "+folder+"/{input,results}", conn)
	copyFileToHost("./data/"+folder+"/style.jpg", "./"+folder+"/input/style.jpg", conn)
	copyFileToHost("./data/"+folder+"/content.jpg", "./"+folder+"/input/content.jpg", conn)

	// TODO: parametrize iterations and size
	// Run the application on the docker image
	runCommand("sudo docker run -v /home/ubuntu/"+folder+"/input:/input -v /home/ubuntu/"+folder+"/results:/results bobray/style-transfer:firsttry -i 1 -s 50", conn)

	// Copy the results back
	copyFileFromHost("./"+folder+"/results/combined.png", "./data/"+folder+"/combined.png", conn)
	copyFileFromHost("./"+folder+"/results/output.png", "./data/"+folder+"/output.png", conn)
}

func CheckIfApplicationsAreRunning(svc *ec2.EC2, instanceId string) bool {
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/go-aws.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	publicDns := aws_helper.RetrievePublicDns(svc, instanceId)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDns+":22", config)
	if err != nil {
		fmt.Println(err)
		return true
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		fmt.Println(err)
		return true
	}
	defer sess.Close()

	var b bytes.Buffer
	sess.Stdout = &b

	err = sess.Run("sudo docker ps | grep style-transfer | wc -l")

	applications := strings.Split(b.String(), "\n")[0]
	fmt.Println(applications, "applications running on worker", instanceId)

	var running bool = true
	if applications == "0" {
		running = false
	}

	return running
}

func GetCpuUtilization(svc *ec2.EC2, instanceId string) float64 {
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/go-aws.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	publicDns := aws_helper.RetrievePublicDns(svc, instanceId)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDns+":22", config)
	if err != nil {
		// fmt.Println(err)
		return 0.0
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		// fmt.Println(err)
		return 0.0
	}
	defer sess.Close()

	var b bytes.Buffer
	sess.Stdout = &b

	err = sess.Run("mpstat -P ALL | grep all | awk '{print $3}'")

	percentage := strings.Split(b.String(), "\n")

	utilization, err := strconv.ParseFloat(percentage[0], 64)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	return utilization
}

func RestartWorker(svc *ec2.EC2, instanceId string) {

	fmt.Println("Restarting worker", instanceId)

	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			publicKey("/keys/go-aws.pem"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Wait until the instance has a public dns assigned
	for aws_helper.CheckInstanceState(svc, instanceId) != true {
		// fmt.Println("waiting")
	}
	publicDns := aws_helper.RetrievePublicDns(svc, instanceId)
	// fmt.Println("Public dns is ", publicDns)

	// TODO: fix this
	// Wait an additional 60 seconds to be sure the instance is open for connections
	time.Sleep(60 * time.Second)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDns+":22", config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	runCommand("cd ~", conn)
	runCommand("sudo nohup sudo sh ./heartbeat.sh 80.114.173.4 8000 > /dev/null 2>&1 &", conn) // TODO: fix hardcoded ip

	fmt.Println("Restarted  worker", instanceId)
}
