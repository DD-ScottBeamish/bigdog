https://hub.docker.com/_/golang/

docker build -t bigdog .    
docker run -d --rm --name bigdog --link alphadog -e HOST_COUNT=<Host Count> -e API_KEY=<Your Api Key> -e APP_KEY=<Your App Key> bigdog