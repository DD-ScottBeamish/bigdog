## Alphadog
JSON/REST API server used issue a container count.  Bigdog calls Alphadog to to scale the number of hosts each instance needs to provision.

### golang container used to host alphadog
https://hub.docker.com/_/golang/

### Command used to build the alphadog container
docker build -t bigdog .  

### Command to run Alphadog.
docker run -d --rm --name bigdog --link alphadog -e HOST_COUNT=HostCount -e API_KEY=YourApiKey -e APP_KEY=YourAppKey bigdog

**Note 
`--link` allows bigdog to call alphadog directly.
`HOST_COUNT` is used to scale the number of hosts a contaner can create.
