package ssh

import (
	"bytes"
	"fmt"
	"go-aws/m/v2/aws"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	defaultUser = "ubuntu"           // default username for ubuntu AMI TODO: make dynamic
	keyPath     = "/keys/go-aws.pem" // location of ssh secret TODO: cmd arg
)

/*InitializeWorker connects to the given instance over ssh and installs the application and its dependencies*/
func InitializeWorker(svc *ec2.EC2, instanceID string) (err error) {
	fmt.Println("Initializing worker", instanceID)
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: defaultUser,
		Auth: []ssh.AuthMethod{
			authenticate(keyPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Wait until the instance has a public dns assigned
	for aws.CheckInstanceState(svc, instanceID) != true {
	}
	publicDNS := aws.GetPublicDNS(svc, instanceID)

	// TODO: fix this
	// Wait an additional 60 seconds to be sure the instance is open for connections
	time.Sleep(60 * time.Second)

	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDNS+":22", config)
	if err != nil {
		return
	}
	defer conn.Close()

	// Install docker and pull the application dockerfile
	err = runCommand("/usr/bin/sudo apt-get update", conn)
	if err != nil {
		return
	}
	err = runCommand("sudo apt-get --assume-yes install sysstat", conn) // TODO: sometimes it breaks on this
	if err != nil {
		return
	}
	err = runCommand("cd ~", conn)
	if err != nil {
		return
	}
	err = runCommand("curl -fsSL https://get.docker.com -o get-docker.sh", conn)
	if err != nil {
		return
	}
	err = runCommand("sudo sh get-docker.sh", conn)
	if err != nil {
		return
	}
	err = runCommand("sudo docker pull bobray/style-transfer:firsttry", conn)
	if err != nil {
		return
	}
	copyFileToHost("./heartbeat.sh", "./heartbeat.sh", conn)
	if err != nil {
		return
	}
	err = runCommand("sudo nohup sudo sh ./heartbeat.sh 80.114.173.4 8000 > /dev/null 2>&1 &", conn) // TODO: fix hardcoded ip
	if err != nil {
		return
	}
	fmt.Println("Initialized  worker", instanceID)
	return
}

func authenticate(path string) ssh.AuthMethod {
	pwd, _ := os.Getwd()
	key, err := ioutil.ReadFile(pwd + path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err) //something went wrong and we can't recover
	}
	return ssh.PublicKeys(signer)
}

func runCommand(cmd string, conn *ssh.Client) (err error) {
	sess, err := conn.NewSession()
	if err != nil {
		return
	}
	defer sess.Close()
	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return
	}
	// go io.Copy(os.Stdout, sessStdOut)
	_ = sessStdOut
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		return
	}
	// go io.Copy(os.Stderr, sessStderr)
	_ = sessStderr

	err = sess.Run(cmd) // eg., /usr/bin/whoami
	if err != nil {
		return
	}
	return
}

func copyFileToHost(srcFilePath string, dstFilePath string, conn *ssh.Client) (err error) {
	client, err := sftp.NewClient(conn)
	if err != nil {
		return
	}
	defer client.Close()

	dstFile, err := client.Create(dstFilePath)
	if err != nil {
		return
	}
	defer dstFile.Close()

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return
	}
	// fmt.Println(bytes, "bytes copied")
	_ = bytes
	return
}

func copyFileFromHost(srcFilePath string, dstFilePath string, conn *ssh.Client) (err error) {
	client, err := sftp.NewClient(conn)
	if err != nil {
		return
	}
	defer client.Close()

	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		return
	}
	defer dstFile.Close()

	srcFile, err := client.Open(srcFilePath)
	if err != nil {
		return
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return
	}
	// fmt.Println(bytes, "bytes copied")
	_ = bytes

	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		return
	}
	return
}

/*RunApplication copies input files from folder to instance, processes input, and copies resulting files back.
The instance needs to be already provisioned with the right dependencies and docker image*/
func RunApplication(svc *ec2.EC2, instanceID string, folder string) (err error) {
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: defaultUser,
		Auth: []ssh.AuthMethod{
			authenticate(keyPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	publicDNS := aws.GetPublicDNS(svc, instanceID)
	// Set up the ssh connection
	conn, err := ssh.Dial("tcp", publicDNS+":22", config)
	if err != nil {
		return
	}
	defer conn.Close()

	time.Sleep(10 * time.Second)

	// Copy the input images to the worker
	runCommand("mkdir -p "+folder+"/{input,results}", conn)
	if err != nil {
		return
	}
	copyFileToHost("./data/"+folder+"/style.jpg", "./"+folder+"/input/style.jpg", conn)
	if err != nil {
		return
	}
	copyFileToHost("./data/"+folder+"/content.jpg", "./"+folder+"/input/content.jpg", conn)
	if err != nil {
		return
	}

	// TODO: parametrize iterations and size
	// Run the application on the docker image
	runCommand("sudo docker run -v /home/ubuntu/"+folder+"/input:/input -v /home/ubuntu/"+folder+"/results:/results bobray/style-transfer:firsttry -i 1 -s 50", conn)
	if err != nil {
		return
	}

	// Copy the results back
	copyFileFromHost("./"+folder+"/results/combined.png", "./data/"+folder+"/combined.png", conn)
	if err != nil {
		return
	}
	copyFileFromHost("./"+folder+"/results/output.png", "./data/"+folder+"/output.png", conn)
	if err != nil {
		return
	}
	return
}

/*InstanceRunning runs docker ps on the instance and returns true if there are containers running.
Returns true if any connection failed so we can check again later.*/
func InstanceRunning(svc *ec2.EC2, instanceID string) bool {
	// Make the ssh connection
	config := &ssh.ClientConfig{
		User: defaultUser,
		Auth: []ssh.AuthMethod{
			authenticate(keyPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	publicDNS := aws.GetPublicDNS(svc, instanceID)
	conn, err := ssh.Dial("tcp", publicDNS+":22", config)
	if err != nil {
		return true
	}
	defer conn.Close()
	sess, err := conn.NewSession()
	if err != nil {
		return true
	}
	defer sess.Close()

	//get docker ps results from stdout
	var b bytes.Buffer
	sess.Stdout = &b
	err = sess.Run("sudo docker ps | grep style-transfer | wc -l")
	//parse result
	applications := strings.Split(b.String(), "\n")[0]
	fmt.Println(applications, "applications running on worker", instanceID)

	if applications == "0" {
		return false
	}
	return true
}

/*GetCPUUtilization connects to the instance and gets system CPU using mpstat.
WARNING: only works with unix based instances*/
func GetCPUUtilization(svc *ec2.EC2, instanceID string) (float64, error) {
	// Make the config for the ssh connection
	config := &ssh.ClientConfig{
		User: defaultUser,
		Auth: []ssh.AuthMethod{
			authenticate(keyPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	publicDNS := aws.GetPublicDNS(svc, instanceID)
	conn, err := ssh.Dial("tcp", publicDNS+":22", config)
	if err != nil {
		return -1, err
	}
	defer conn.Close()
	sess, err := conn.NewSession()
	if err != nil {
		return -1, err
	}
	defer sess.Close()
	var b bytes.Buffer
	sess.Stdout = &b
	//https://web.archive.org/web/20160403064806/http://linuxcommand.org/man_pages/mpstat1.html
	err = sess.Run("mpstat -P ALL | grep all | awk '{print $3}'")
	if err != nil {
		return -1, err
	}
	percentage := strings.Split(b.String(), "\n")
	return strconv.ParseFloat(percentage[0], 64)
}
