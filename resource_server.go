package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getIDH(address, name, tarballChecksum string, numThreads int) string {
	return fmt.Sprintf("%v %v %v %v", address, name, tarballChecksum, numThreads)
}

func getID(address, name, tarballPath string, numThreads int) string {
	tarballChecksum := hashFile(tarballPath)
	return getIDH(address, name, tarballChecksum, numThreads)
}

type uploadPackage struct {
	Name        string                 `json:"name"`
	Threads     int                    `json:"threads"`
	Tarball     string                 `json:"tarball"`
	Environment map[string]interface{} `json:"environment"`
}

func uploadFunction(address, name, tarballPath string, numThreads int, env map[string]interface{}) {
	dat, _ := ioutil.ReadFile(tarballPath)
	b64 := base64.StdEncoding.EncodeToString(dat)
	up := uploadPackage{Name: name, Threads: numThreads, Tarball: b64}
	up.Environment = env
	ups, _ := json.Marshal(up)
	http.Post("http://"+address+":8080/upload", "application/json", bytes.NewBuffer(ups))
}

func deleteFunction(address, name string) {
	http.Post("http://"+address+":8080/delete", "application/json", bytes.NewBuffer([]byte(name)))
}

type funcList []struct {
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	Threads  int    `json:"threads"`
	Resource string `json:"resource"`
}

func getFuncID(address, name string) string {
	var funcs funcList
	resp, _ := http.Get("http://" + address + ":8080/list")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &funcs)

	for _, fun := range funcs {
		if name == fun.Name {
			return getIDH(address, fun.Name, fun.Hash, fun.Threads)
		}
	}
	return ""
}

func findFuncID(address, id string) string {
	var funcs funcList
	resp, _ := http.Get("http://" + address + ":8080/list")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &funcs)

	for _, fun := range funcs {
		rID := getIDH(address, fun.Name, fun.Hash, fun.Threads)
		if id == rID {
			return fun.Name
		}
	}
	return ""
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	address := os.Getenv("TINYFAAS_ADDRESS")
	name := d.Get("name").(string)
	tarballPath := d.Get("tarball_path").(string)
	numThreads := d.Get("num_threads").(int)
	environment := d.Get("environment").(map[string]interface{})
	uploadFunction(address, name, tarballPath, numThreads, environment)
	d.SetId(getID(address, name, tarballPath, numThreads))

	return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	address := os.Getenv("TINYFAAS_ADDRESS")
	name := d.Get("name").(string)
	tarballPath := d.Get("tarball_path").(string)
	numThreads := d.Get("num_threads").(int)

	if getFuncID(address, name) != getID(address, name, tarballPath, numThreads) {
		d.SetId("")
	} else {
		d.SetId(getID(address, name, tarballPath, numThreads))
	}

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	address := os.Getenv("TINYFAAS_ADDRESS")
	name := d.Get("name").(string)
	tarballPath := d.Get("tarball_path").(string)
	numThreads := d.Get("num_threads").(int)
	environment := d.Get("environment").(map[string]interface{})
	oldName := findFuncID(address, d.Id())
	if oldName != "" {
		deleteFunction(address, oldName)
	}
	uploadFunction(address, name, tarballPath, numThreads, environment)
	d.SetId(getID(address, name, tarballPath, numThreads))

	return resourceServerRead(d, m)
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	address := os.Getenv("TINYFAAS_ADDRESS")
	name := d.Get("name").(string)
	deleteFunction(address, name)
	d.SetId("")
	return nil
}

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"tarball_path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"num_threads": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"environment": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
			},
		},
	}
}
