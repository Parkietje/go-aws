package ssh

import (
	"bytes"
	"fmt"
	"go-aws/m/v2/aws"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	defaultUser = "ubuntu"                         // default username for ubuntu AMI TODO: make dynamic
	keyPath     = "/keys/awskey"                   // location of ssh secret TODO: cmd arg
	data        = "data"                           // data upload folder
	combined    = "combined.png"                   // result file
	home        = "/home/ubuntu"                   // worker home folder
	docker      = "bobray/style-transfer:firsttry" // docker image to pull
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
		fmt.Println("runcmd: connection fails")
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
		fmt.Println("runcmd: cmd fails")
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
		fmt.Println("copyto: dst not found: " + dstFilePath)
		return
	}
	defer dstFile.Close()

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println("copyto: src not found: " + srcFilePath)
		return
	}
	defer srcFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println("copyto: copy NOK")
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
		fmt.Println("copyfrom: dst not found: " + dstFilePath)
		return
	}
	defer dstFile.Close()

	srcFile, err := client.Open(srcFilePath)
	if err != nil {
		fmt.Println("copyfrom: src not found: " + srcFilePath)
		return
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println("copyfrom: copy NOK")
		return
	}
	// fmt.Println(bytes, "bytes copied")
	_ = bytes

	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		fmt.Println("copyfrom: sync NOK")
		return
	}
	return
}

/*RunApplication copies input files from folder to instance, processes input, and copies resulting files back.
The instance needs to be already provisioned with the right dependencies and docker image.*/
func RunApplication(svc *ec2.EC2, instanceID string, folder string, styleFile string, contentFile string) (err error) {
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

	//make folders on worker
	workerInputFolder := filepath.Join(home, folder, "input")
	workerResultFolder := filepath.Join(home, folder, "results")
	//fmt.Println("worker inputFolder: " + workerInputFolder)
	//fmt.Println("worker resultsFolder: " + workerResultFolder)
	err = runCommand("mkdir -p "+workerInputFolder, conn)
	if err != nil {
		fmt.Println("ssh run app: mkdir1 NOK")
		return
	}
	err = runCommand("mkdir -p "+workerResultFolder, conn)
	if err != nil {
		fmt.Println("ssh run app: mkdir2 NOK")
		return
	}

	//dynamic paths
	styleSrc := filepath.Join(data, folder, styleFile)
	//fmt.Println("host styleSrc: " + "./" + styleSrc)
	contentSrc := filepath.Join(data, folder, contentFile)
	//fmt.Println("host contentSrc: " + "./" + contentSrc)
	styleDest := filepath.Join(workerInputFolder, styleFile)
	//fmt.Println("worker styleDest: " + styleDest)
	contentDest := filepath.Join(workerInputFolder, contentFile)
	//fmt.Println("worker contentDest: " + contentDest)

	//copy to worker
	err = copyFileToHost("./"+styleSrc, styleDest, conn)
	if err != nil {
		return
	}
	err = copyFileToHost("./"+contentSrc, contentDest, conn)
	if err != nil {
		return
	}

	// TODO: parametrize iterations and size
	// Run the application on the docker image
	cmd := "sudo docker run -v " + workerInputFolder + ":/input -v " + workerResultFolder + ":/results " + docker + " -i 1 -s 50"
	err = runCommand(cmd, conn)
	if err != nil {
		fmt.Println("ssh run app: following docker run cmd failed:")
		fmt.Println(cmd)
		return
	}

	// Copy the results back
	resultFileSrc := filepath.Join(workerResultFolder, combined)
	//fmt.Println("worker resultFileSrc: " + resultFileSrc)
	resultFileDest := filepath.Join(data, folder, combined)
	//fmt.Println("host resultFileDest: " + resultFileDest)
	err = copyFileFromHost(resultFileSrc, resultFileDest, conn)
	if err != nil {
		fmt.Println("ssh run app: copy result back failed")
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
