package main

import (
    "time"
    "fmt"
    "strconv"
    "net/http"
    "math"
    "math/rand"
    "io/ioutil"
    "bytes"
    "os"
    "encoding/json"
)

const (
    DatadogMetricAPI = "https://app.datadoghq.com/api/v1/series"
    DatadogCheckAPI = "https://app.datadoghq.com/api/v1/check_run"
    DatadogTagAPI = "https://app.datadoghq.com/api/v1/tags/hosts/"
    BigdogCheckTag = "bigdog.is_ok"
    BigDogMaxCPU = 90
    BigDogMinCPU = 5
    BigDogMaxDisk = 30000
    BigDogMinDisk = 1000
    BigDogMaxMem = 99
    BigDogMinMem = 20
    CheckInInterval = 60
)

type Tag struct {
    name string
    value string 
}

type Host struct {
    name string
    tags []Tag
    tagged bool
}

// Call alphadog to a container count.  Use this to set value to create the host name number
// Ex. If we pass in 5 as the host count and get container value of 2.  We name the hosts 10-19
type Container struct {
    Count int `json:count"`
}

var (
    datadogAPIKey = os.Getenv("API_KEY")
    datadogAppKey = os.Getenv("APP_KEY")
    totalHosts = os.Getenv("HOST_COUNT")
    hosts []Host
    services = [...]string {"redis", "cassandra", "mongodb", "nginx", "rabbitmq", "consul", "kafka", "elasticsearch"}
    cloudProviders = [...]string {"aws", "gcp", "azure"}
    metricApiUrl = DatadogMetricAPI+"?api_key="+datadogAPIKey+"&application_key="+datadogAppKey
    tagApiUrl = DatadogTagAPI
    myClient = &http.Client{Timeout: 10 * time.Second}
)

func getJson(url string, target interface{}) error {
    r, err := myClient.Get(url)
    if err != nil {
        return err
    }
    defer r.Body.Close()

    return json.NewDecoder(r.Body).Decode(target)
}

func initializeHosts() {

    //fmt.Fprintf(w, "Creating %s!", r.URL.Path[1:])
    fmt.Println("Building hosts...")

    container := &Container{}
    getJson(alphadogApiUrl, container)
    fmt.Println("Container Number: " + strconv.Itoa(container.Count))

    // convert environment variable host count to int
    hostCount,err  := strconv.Atoi(totalHosts)
    if err != nil {
        fmt.Println("Error")
    }
    
    go func() {
        for i := 0; i < hostCount; i++ {
            name := "bigdog_"+(strconv.Itoa(hostCount*container.Count-i))
            role := services[rand.Intn(len(services))]
            cp := cloudProviders[rand.Intn(len(cloudProviders))]
            tags := []Tag {Tag{name: "role", value:role}, Tag{name:"cloud_provider", value:cp}}
            newHost := Host{name: name, tags: tags, tagged: false}
            
            hosts = append(hosts,newHost)
            go func(host *Host){
                // Add host tags
                var jsonStr = []byte(fmt.Sprintf(`{"tags" : ["cloud_provider:%s", "role:%s"]}`,host.tags[1].value,host.tags[0].value))
                apiUrl := tagApiUrl+host.name+"?api_key="+datadogAPIKey+"&application_key="+datadogAppKey

                fmt.Println("Deleting Host Tags for "+host.name)
                req, err := http.NewRequest("DELETE", apiUrl, nil)
                req.Header.Set("Content-Type", "application/json")

                client := &http.Client{}
                resp, err := client.Do(req)

                fmt.Println("Submitting Host Tags for "+host.name)
                req, err = http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonStr))
                req.Header.Set("Content-Type", "application/json")

                client = &http.Client{}
                resp, err = client.Do(req)
                if err != nil {
                    return
                }
                defer resp.Body.Close()   

                fmt.Println("response Status:", resp.Status)
                fmt.Println("response Headers:", resp.Header)

                body, _ := ioutil.ReadAll(resp.Body)
                fmt.Println("response Body:", string(body))

            }(&newHost)
        }
    }()
}

func main() {

    rand.Seed(time.Now().Unix()) // Set up a random picker for tags
     // Start a host check in process. We'll check in every CheckInInterval seconds 
    initializeHosts()
    //time.Sleep(30 * time.Second) 
    hostCheckIn()     
}

// Random Range
func random(min, max int) int {
    return rand.Intn(max - min) + min
}

// Return Host Metrics JSON
func hostMetrics(host *Host, time int32) string {
    
    cpu := random(BigDogMinCPU,BigDogMaxCPU)
    disk := random(BigDogMinDisk,BigDogMaxDisk)
    mem := random(BigDogMinMem,BigDogMaxMem)
    json := fmt.Sprintf(`{"series" : 
    [
        {
          "metric":"system.cpu.stolen",
          "points":[[%d,0]],
          "type":"gauge",
          "host":"%s",
          "tags":["%s:%s","%s:%s"]
        },
        {
          "metric":"system.cpu.user",
          "points":[[%d,%d]],
          "type":"gauge",
          "host":"%s",
          "tags":["%s:%s","%s:%s"]
        },
        {
          "metric":"system.disk.used",
          "points":[[%d,%d]],
          "type":"gauge",
          "host":"%s",
          "tags":["%s:%s","%s:%s"]
        },
        {
          "metric":"system.mem.used",
          "points":[[%d,%d]],
          "type":"gauge",
          "host":"%s",
          "tags":["%s:%s","%s:%s"]
        }
    ]
    }`,time,host.name,host.tags[0].name,host.tags[0].value,host.tags[1].name,host.tags[1].value,time,cpu,host.name,host.tags[0].name,host.tags[0].value,host.tags[1].name,host.tags[1].value,time,disk,host.name,host.tags[0].name,host.tags[0].value,host.tags[1].name,host.tags[1].value,time,mem,host.name,host.tags[0].name,host.tags[0].value,host.tags[1].name,host.tags[1].value)
    fmt.Println(json)

    return json
}

// Lets just run this for. ev. er.
func hostCheckIn() {
    
    fmt.Println("Host Check in!")
    currentTime := int32(time.Now().Unix())
    for index, _ := range hosts { 
        go func(i int) {
            fmt.Println("Submitting Metrics for "+hosts[i].name)
            metrics := hostMetrics(&hosts[i],currentTime)
            var jsonStr = []byte(metrics)
            req, err := http.NewRequest("POST", metricApiUrl, bytes.NewBuffer(jsonStr))
            req.Header.Set("Content-Type", "application/json")

            client := &http.Client{}
            resp, err := client.Do(req)
            if err != nil {
                return
            }
            defer resp.Body.Close()

            fmt.Println("response Status:", resp.Status)
            fmt.Println("response Headers:", resp.Header)
            body, _ := ioutil.ReadAll(resp.Body)
            fmt.Println("response Body:", string(body))
            
        }(index)
        if(math.Mod(float64(index), 100) == 0) {
            time.Sleep(100*time.Millisecond)
        }
    }

    // Report in every 30 seconds
    time.Sleep(CheckInInterval * time.Second)  
    hostCheckIn()  
}
